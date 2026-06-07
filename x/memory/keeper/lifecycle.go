package keeper

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strings"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/wevibe-network/wevibe-chain/x/memory/types"
)

const currentEpochKey = "current_epoch"

type dec struct {
	value *big.Rat
}

type orgIdleDecayConfig struct {
	idleScale    float64
	suppressIdle bool
	serves       uint64
	denials      uint64
}

func newDec(s string) dec {
	r := new(big.Rat)
	r.SetString(s)
	return dec{value: r}
}

func (d dec) Add(other dec) dec {
	r := new(big.Rat).Add(d.value, other.value)
	return dec{value: r}
}

func (d dec) Sub(other dec) dec {
	r := new(big.Rat).Sub(d.value, other.value)
	return dec{value: r}
}

func (d dec) GT(other dec) bool {
	return d.value.Cmp(other.value) > 0
}

func (d dec) IsNegative() bool {
	return d.value.Sign() < 0
}

func (d dec) IsPositive() bool {
	return d.value.Sign() > 0
}

func (d dec) String() string {
	return d.value.FloatString(4)
}

var zeroDec = dec{value: big.NewRat(0, 1)}
var oneDec = dec{value: big.NewRat(1, 1)}

func parseWeight(s string) dec {
	if s == "" {
		return newDec("1.0")
	}
	s = strings.TrimSpace(s)
	if !strings.Contains(s, ".") && !strings.Contains(s, "/") {
		if i, ok := new(big.Int).SetString(s, 10); ok {
			return dec{value: new(big.Rat).SetFrac(i, big.NewInt(1))}
		}
	}
	return newDec(s)
}

func clampFloat64(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (k *Keeper) saveMemoryCommitment(ctx context.Context, memory *types.MemoryCommitment) error {
	store := k.getStore(ctx)
	bz, err := proto.Marshal(memoryToStored(memory))
	if err != nil {
		return fmt.Errorf("marshal memory: %w", err)
	}
	if err := store.Set(approvedKey(memory.OrgID, memory.ContentHash), bz); err != nil {
		return fmt.Errorf("persist memory: %w", err)
	}
	return nil
}

func (k *Keeper) loadMemoryByCID(ctx context.Context, orgID, cid string) (*types.MemoryCommitment, error) {
	contentHash, err := decodeCID(cid)
	if err != nil {
		return nil, err
	}

	store := k.getStore(ctx)
	bz, err := store.Get(approvedKey(orgID, contentHash))
	if err != nil {
		return nil, fmt.Errorf("get memory: %w", err)
	}
	if bz == nil {
		return nil, types.ErrMemoryNotFound
	}

	var stored types.StoredMemoryCommitment
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal memory: %w", err)
	}

	memory := storedToMemory(stored)
	return &memory, nil
}

// applyDecay is the canonical Earned Trust decay handler.
// Source: wevibe-sim/ranking-fix.js applyDecay.
// D-4.2 Implementation Clarifications (DMO-006 + DMO-007) is the binding spec.
func (k *Keeper) applyDecay(
	memory *types.MemoryCommitment,
	currentEpoch uint64,
	servesThisEpoch uint64,
	denialsThisEpoch uint64,
	kwIDsMatched map[string]bool,
	params types.Params,
	idleScale float64,
	suppressIdle bool,
) error {
	if isInGracePeriod(currentEpoch, memory.Epoch, params.GraceEpochs) {
		return nil
	}

	totalEvents := memory.ServeCountTotal + memory.DenialCountTotal
	var denialRate float64
	if totalEvents > 0 {
		denialRate = float64(memory.DenialCountTotal) / float64(totalEvents)
	}

	trust := math.Max(0, 1.0-denialRate)
	trustSq := trust * trust
	trustEarned := memory.ServeCountTotal >= params.TrustMinServes &&
		denialRate < (float64(params.TrustMaxRateBps)/10000.0)

	serveD := float64(params.ServeDBps) / 10000.0
	denialD := float64(params.DenialDBps) / 10000.0
	idleD := float64(params.IdleDBps) / 10000.0
	serveFloor := float64(params.ServeFloorBps) / 10000.0
	denialFloor := float64(params.DenialFloorBps) / 10000.0
	idleProtect := float64(params.IdleProtectBps) / 10000.0
	idleUntrusted := float64(params.IdleUntrustedBps) / 10000.0

	for i := range memory.Keywords {
		kw := memory.Keywords[i]
		matched := kwIDsMatched[kw.Keyword]
		weight := parseWeight(kw.Weight)

		if servesThisEpoch > 0 && matched {
			delta := serveD * float64(servesThisEpoch) *
				(serveFloor + (1.0-serveFloor)*trustSq)
			weight = weight.Add(dec{value: new(big.Rat).SetFloat64(delta)})
		}

		if denialsThisEpoch > 0 && matched {
			delta := denialD * float64(denialsThisEpoch) *
				(denialFloor + (1.0-denialFloor)*denialRate)
			weight = weight.Sub(dec{value: new(big.Rat).SetFloat64(delta)})
		}

		if !matched || (servesThisEpoch == 0 && denialsThisEpoch == 0) {
			if !suppressIdle {
				idleMult := idleUntrusted * idleScale
				if trustEarned {
					idleMult = idleProtect
				}
				weight = weight.Sub(dec{value: new(big.Rat).SetFloat64(idleD * idleMult)})
			}
		}

		if weight.IsNegative() {
			weight = zeroDec
		} else if weight.GT(oneDec) {
			weight = oneDec
		}

		memory.Keywords[i].Weight = weight.String()
	}

	memory.LastActiveEpoch = currentEpoch

	retrievalThreshold := dec{value: new(big.Rat).SetFloat64(float64(params.RetrievalThresholdBps) / 10000.0)}
	allBelow := true
	for _, kw := range memory.Keywords {
		weight := parseWeight(kw.Weight)
		if weight.GT(retrievalThreshold) {
			allBelow = false
			break
		}
	}

	if allBelow && memory.State != types.MemoryState_MEMORY_STATE_ARCHIVED {
		memory.State = types.MemoryState_MEMORY_STATE_ARCHIVED
		memory.ArchivedEpoch = currentEpoch
	}

	return nil
}

// ApplyEpochDecay runs once at epoch end. It iterates all non-archived
// committed memories, reads their aggregated serve/denial counts and
// matched keywords for this epoch, and calls applyDecay exactly once
// per memory. This is the ONLY call site for applyDecay - event-time
// handlers (ApplyServeBoost, ApplyDenialDecay) update counters only.
func (k *Keeper) ApplyEpochDecay(ctx context.Context, epoch uint64) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	store := k.getStore(ctx)
	prefix := []byte("approved/")
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return fmt.Errorf("iterate approved memories: %w", err)
	}

	// R-CACHEKV-ITER: never mutate the store while iterating it under a
	// cache-wrapped context. Collect the decay-eligible memories first, close
	// the iterator, then load+decay+persist each one. saveMemoryCommitment
	// writes back to the same "approved/" prefix we are iterating here.
	var eligible []types.MemoryCommitment
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredMemoryCommitment
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			iter.Close()
			return fmt.Errorf("unmarshal memory commitment: %w", err)
		}

		memory := storedToMemory(stored)
		if !isDecayEligible(memory.State) {
			continue
		}
		eligible = append(eligible, memory)
	}
	iter.Close()

	activeCountByOrg := make(map[string]uint64)
	for i := range eligible {
		activeCountByOrg[eligible[i].OrgID]++
	}

	orgIdleConfig := make(map[string]orgIdleDecayConfig)

	for i := range eligible {
		memory := eligible[i]
		cidHex := types.ContentHashToHex(memory.ContentHash)

		config, ok := orgIdleConfig[memory.OrgID]
		if !ok {
			config, err = k.resolveOrgIdleDecayConfig(ctx, memory.OrgID, epoch, params)
			if err != nil {
				return err
			}
			orgIdleConfig[memory.OrgID] = config
		}

		graceRemaining := graceEpochsRemaining(epoch, memory.Epoch, params.GraceEpochs)
		denialRate, trustEarned := calculateDenialRateAndTrust(memory, params)

		servesThisEpoch, err := k.getMemoryServeCount(ctx, memory.OrgID, cidHex, epoch)
		if err != nil {
			return err
		}
		denialsThisEpoch, err := k.getMemoryDenialCount(ctx, memory.OrgID, cidHex, epoch)
		if err != nil {
			return err
		}

		kwIDsMatched, err := k.getMatchedKeywords(ctx, memory.OrgID, cidHex, epoch)
		if err != nil {
			return err
		}

		if err := k.applyDecay(&memory, epoch, servesThisEpoch, denialsThisEpoch, kwIDsMatched, params, config.idleScale, config.suppressIdle); err != nil {
			return err
		}

		if err := k.saveMemoryCommitment(ctx, &memory); err != nil {
			return err
		}

		_ = graceRemaining
		_ = denialRate
		_ = trustEarned
	}

	return nil
}

func (k *Keeper) ApplyServeBoost(ctx context.Context, orgID string, contentHash []byte, epoch uint64) error {
	memory, err := k.GetApprovedMemory(ctx, orgID, contentHash)
	if err != nil {
		return err
	}
	if memory.State == types.MemoryState_MEMORY_STATE_ARCHIVED {
		return nil
	}

	cidHex := types.ContentHashToHex(contentHash)
	kwIDsMatched, err := k.getMatchedKeywords(ctx, orgID, cidHex, epoch)
	if err != nil {
		return err
	}
	if len(kwIDsMatched) == 0 {
		return fmt.Errorf("matched keywords not found for serve event")
	}

	for i := range memory.Keywords {
		if kwIDsMatched[memory.Keywords[i].Keyword] {
			memory.Keywords[i].ServeCount++
		}
	}
	memory.ServeCountTotal++

	return k.saveMemoryCommitment(ctx, memory)
}

func (k *Keeper) ApplyDenialDecay(ctx context.Context, orgID string, contentHash []byte, epoch uint64) error {
	memory, err := k.GetApprovedMemory(ctx, orgID, contentHash)
	if err != nil {
		return err
	}
	if memory.State == types.MemoryState_MEMORY_STATE_ARCHIVED {
		return nil
	}

	cidHex := types.ContentHashToHex(contentHash)
	kwIDsMatched, err := k.getMatchedKeywords(ctx, orgID, cidHex, epoch)
	if err != nil {
		return err
	}
	if len(kwIDsMatched) == 0 {
		return fmt.Errorf("matched keywords not found for denial event")
	}

	for i := range memory.Keywords {
		if kwIDsMatched[memory.Keywords[i].Keyword] {
			memory.Keywords[i].DenialCount++
		}
	}
	memory.DenialCountTotal++

	return k.saveMemoryCommitment(ctx, memory)
}

// GetContributorsWithApprovalsInEpoch returns, network-wide, the number of
// committed (non-archived, non-denied) approved memories per contributor
// pubkey that were approved in the given epoch. It is consumed by the
// emissions keeper to determine the qualifying contributor set for per-epoch
// contributor emissions.
//
// R-CACHEKV-ITER: this is a NEW network-wide iteration. It is read-only and
// MUST NOT use a post-loop iter.Error() block (which returns a false failure
// at normal end-of-iteration under cache-wrapped stores). The Valid() loop
// terminates correctly on its own.
func (k *Keeper) GetContributorsWithApprovalsInEpoch(ctx context.Context, epoch uint64) (map[string]uint64, error) {
	store := k.getStore(ctx)
	prefix := []byte("approved/")
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, fmt.Errorf("iterate approved memories: %w", err)
	}
	defer iter.Close()

	counts := make(map[string]uint64)
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredMemoryCommitment
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		if stored.ApprovedAtEpoch != epoch {
			continue
		}
		if stored.State != types.MemoryState_MEMORY_STATE_COMMITTED {
			continue
		}
		if stored.ContributorPubkey == "" {
			continue
		}
		counts[stored.ContributorPubkey]++
	}

	return counts, nil
}

// GetActiveMemoryCountByOrg returns the number of ACTIVE memories for one org.
// ACTIVE means committed (non-archived, non-denied).
//
// R-CACHEKV-ITER: this is read-only iteration over the approved prefix and
// intentionally uses only the Valid() loop with no post-loop iter.Error().
func (k *Keeper) GetActiveMemoryCountByOrg(ctx context.Context, orgID string) (uint64, error) {
	store := k.getStore(ctx)
	prefix := approvedPrefix(orgID)
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return 0, fmt.Errorf("iterate approved memories for org %s: %w", orgID, err)
	}
	defer iter.Close()

	var count uint64
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredMemoryCommitment
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		if stored.State == types.MemoryState_MEMORY_STATE_COMMITTED {
			count++
		}
	}

	return count, nil
}

func (k *Keeper) resolveOrgIdleDecayConfig(ctx context.Context, orgID string, epoch uint64, params types.Params) (orgIdleDecayConfig, error) {
	config := orgIdleDecayConfig{
		idleScale:    1.0,
		suppressIdle: false,
		serves:       0,
		denials:      0,
	}

	if k.serveKeeper == nil {
		return config, nil
	}

	serves, denials, err := k.serveKeeper.GetEpochTrafficStats(ctx, orgID, epoch)
	if err != nil {
		k.logger.Warn("epoch decay: org traffic stats unavailable; using safe idle defaults",
			"org_id", orgID,
			"epoch", epoch,
			"error", err,
		)
		return config, nil
	}

	config.serves = serves
	config.denials = denials

	orgEvents := serves + denials
	if orgEvents == 0 {
		config.suppressIdle = true
		return config, nil
	}

	activeCount, err := k.GetActiveMemoryCountByOrg(ctx, orgID)
	if err != nil {
		return orgIdleDecayConfig{}, err
	}
	if activeCount == 0 {
		return config, nil
	}

	trafficRef := float64(params.IdleTrafficRefBpsPerMem) / 10000.0
	if trafficRef <= 0 {
		return config, nil
	}

	trafficFloor := float64(params.IdleTrafficFloorBps) / 10000.0
	tOrg := float64(orgEvents) / float64(activeCount)
	config.idleScale = clampFloat64(tOrg/trafficRef, trafficFloor, 1.0)

	return config, nil
}

func (k *Keeper) getCurrentEpoch(ctx context.Context) uint64 {
	store := k.getStore(ctx)
	bz, err := store.Get([]byte(currentEpochKey))
	if err != nil || bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (k *Keeper) setCurrentEpoch(ctx context.Context, epoch uint64) error {
	store := k.getStore(ctx)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, epoch)
	if err := store.Set([]byte(currentEpochKey), buf); err != nil {
		return fmt.Errorf("set current epoch: %w", err)
	}
	return nil
}

func (k *Keeper) getMemoryServeCount(ctx context.Context, orgID, cid string, epoch uint64) (uint64, error) {
	if k.serveKeeper == nil {
		return 0, nil
	}
	count, err := k.serveKeeper.GetMemoryServeCountForEpoch(ctx, orgID, cid, epoch)
	if err != nil {
		return 0, fmt.Errorf("serve keeper: %w", err)
	}
	return count, nil
}

func (k *Keeper) getMemoryDenialCount(ctx context.Context, orgID, cid string, epoch uint64) (uint64, error) {
	if k.serveKeeper == nil {
		return 0, nil
	}
	count, err := k.serveKeeper.GetMemoryDenialCountForEpoch(ctx, orgID, cid, epoch)
	if err != nil {
		return 0, fmt.Errorf("serve keeper denial count: %w", err)
	}
	return count, nil
}

func (k *Keeper) getMatchedKeywords(ctx context.Context, orgID, cid string, epoch uint64) (map[string]bool, error) {
	if k.serveKeeper == nil {
		return map[string]bool{}, nil
	}
	matches, err := k.serveKeeper.GetMatchedKeywordsForEpoch(ctx, orgID, cid, epoch)
	if err != nil {
		return nil, fmt.Errorf("serve keeper matched keywords: %w", err)
	}
	if matches == nil {
		return map[string]bool{}, nil
	}
	return matches, nil
}

func isDecayEligible(state types.MemoryState) bool {
	return state == types.MemoryState_MEMORY_STATE_COMMITTED
}

func isInGracePeriod(epoch, memoryEpoch, graceEpochs uint64) bool {
	if graceEpochs == 0 {
		return false
	}
	if epoch < memoryEpoch {
		return true
	}
	return epoch-memoryEpoch < graceEpochs
}

func decodeCID(cid string) ([]byte, error) {
	contentHash, err := hex.DecodeString(cid)
	if err != nil {
		return nil, fmt.Errorf("decode cid: %w", err)
	}
	if len(contentHash) != types.ContentHashLen {
		return nil, types.ErrInvalidContentHash
	}
	return contentHash, nil
}

func graceEpochsRemaining(epoch, memoryEpoch, graceEpochs uint64) uint64 {
	if graceEpochs == 0 {
		return 0
	}
	if !isInGracePeriod(epoch, memoryEpoch, graceEpochs) {
		return 0
	}
	if epoch < memoryEpoch {
		return (memoryEpoch - epoch) + graceEpochs
	}
	return graceEpochs - (epoch - memoryEpoch)
}

func calculateDenialRateAndTrust(memory types.MemoryCommitment, params types.Params) (float64, bool) {
	totalEvents := memory.ServeCountTotal + memory.DenialCountTotal
	denialRate := 0.0
	if totalEvents > 0 {
		denialRate = float64(memory.DenialCountTotal) / float64(totalEvents)
	}
	trustEarned := memory.ServeCountTotal >= params.TrustMinServes &&
		denialRate < (float64(params.TrustMaxRateBps)/10000.0)
	return denialRate, trustEarned
}

func minKeywordWeight(keywords []*types.KeywordWeight) float64 {
	if len(keywords) == 0 {
		return 0
	}

	minWeight := 0.0
	for i, kw := range keywords {
		weight, _ := parseWeight(kw.Weight).value.Float64()
		if i == 0 || weight < minWeight {
			minWeight = weight
		}
	}

	return minWeight
}
