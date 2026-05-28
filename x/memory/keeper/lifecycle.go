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
			idleMult := idleUntrusted
			if trustEarned {
				idleMult = idleProtect
			}
			weight = weight.Sub(dec{value: new(big.Rat).SetFloat64(idleD * idleMult)})
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

func (k *Keeper) ApplyIdleDecay(ctx context.Context, _ string, epoch uint64) error {
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
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var stored types.StoredMemoryCommitment
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			return fmt.Errorf("unmarshal memory commitment: %w", err)
		}

		memory := storedToMemory(stored)
		if !isDecayEligible(memory.State) {
			continue
		}

		cidHex := types.ContentHashToHex(memory.ContentHash)
		servesThisEpoch, err := k.getMemoryServeCount(ctx, memory.OrgID, cidHex, epoch)
		if err != nil {
			return err
		}
		denialsThisEpoch, err := k.getMemoryDenialCount(ctx, memory.OrgID, cidHex, epoch)
		if err != nil {
			return err
		}

		if servesThisEpoch > 0 || denialsThisEpoch > 0 {
			continue
		}

		if err := k.applyDecay(&memory, epoch, 0, 0, nil, params); err != nil {
			return err
		}
		if err := k.saveMemoryCommitment(ctx, &memory); err != nil {
			return err
		}
	}

	return nil
}

func (k *Keeper) ApplyServeBoost(ctx context.Context, orgID string, contentHash []byte) error {
	memory, err := k.GetApprovedMemory(ctx, orgID, contentHash)
	if err != nil {
		return err
	}
	if memory.State == types.MemoryState_MEMORY_STATE_ARCHIVED {
		return nil
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	currentEpoch := k.getCurrentEpoch(ctx)
	cidHex := types.ContentHashToHex(contentHash)
	servesThisEpoch, err := k.getMemoryServeCount(ctx, orgID, cidHex, currentEpoch)
	if err != nil {
		return err
	}
	denialsThisEpoch, err := k.getMemoryDenialCount(ctx, orgID, cidHex, currentEpoch)
	if err != nil {
		return err
	}
	kwIDsMatched, err := k.getMatchedKeywords(ctx, orgID, cidHex, currentEpoch)
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

	if err := k.applyDecay(memory, currentEpoch, servesThisEpoch, denialsThisEpoch, kwIDsMatched, params); err != nil {
		return err
	}

	return k.saveMemoryCommitment(ctx, memory)
}

func (k *Keeper) ApplyDenialDecay(ctx context.Context, orgID string, contentHash []byte) error {
	memory, err := k.GetApprovedMemory(ctx, orgID, contentHash)
	if err != nil {
		return err
	}
	if memory.State == types.MemoryState_MEMORY_STATE_ARCHIVED {
		return nil
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	currentEpoch := k.getCurrentEpoch(ctx)
	cidHex := types.ContentHashToHex(contentHash)
	servesThisEpoch, err := k.getMemoryServeCount(ctx, orgID, cidHex, currentEpoch)
	if err != nil {
		return err
	}
	denialsThisEpoch, err := k.getMemoryDenialCount(ctx, orgID, cidHex, currentEpoch)
	if err != nil {
		return err
	}
	kwIDsMatched, err := k.getMatchedKeywords(ctx, orgID, cidHex, currentEpoch)
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

	if err := k.applyDecay(memory, currentEpoch, servesThisEpoch, denialsThisEpoch, kwIDsMatched, params); err != nil {
		return err
	}

	return k.saveMemoryCommitment(ctx, memory)
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
