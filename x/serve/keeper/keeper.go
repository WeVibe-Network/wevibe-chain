package keeper

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/gogoproto/proto"
	"github.com/wevibe-network/wevibe-chain/x/serve/types"
)

type Keeper struct {
	storeService     store.KVStoreService
	logger           log.Logger
	authority        string
	orgKeeper        types.OrgKeeper
	memoryKeeper     types.MemoryKeeper
	bandwidthKeeper  types.BandwidthKeeper
	reputationKeeper types.ReputationKeeper
}

func NewKeeper(
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,
	orgKeeper types.OrgKeeper,
	memoryKeeper types.MemoryKeeper,
	bandwidthKeeper types.BandwidthKeeper,
	reputationKeeper types.ReputationKeeper,
) *Keeper {
	return &Keeper{
		storeService:     storeService,
		logger:           logger,
		authority:        authority,
		orgKeeper:        orgKeeper,
		memoryKeeper:     memoryKeeper,
		bandwidthKeeper:  bandwidthKeeper,
		reputationKeeper: reputationKeeper,
	}
}

func (k *Keeper) getStore(ctx context.Context) store.KVStore {
	return k.storeService.OpenKVStore(ctx)
}

func nullifierKey(nullifier []byte) []byte {
	return []byte(fmt.Sprintf("nullifier/%s", types.ContentHashToHex(nullifier)))
}

func denialNullifierKey(nullifier []byte) []byte {
	return []byte(fmt.Sprintf("denynull/%s", types.ContentHashToHex(nullifier)))
}

func attestationKey(orgID string, epoch uint64, nullifier []byte) []byte {
	return []byte(fmt.Sprintf("attestation/%s/%d/%s", orgID, epoch, types.ContentHashToHex(nullifier)))
}

func attestationPrefix(orgID string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("attestation/%s/%d/", orgID, epoch))
}

func denialAttestationKey(orgID string, epoch uint64, nullifier []byte) []byte {
	return []byte(fmt.Sprintf("denial/%s/%d/%s", orgID, epoch, types.ContentHashToHex(nullifier)))
}

func denialAttestationPrefix(orgID string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("denial/%s/%d/", orgID, epoch))
}

func statsKey(orgID string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("stats/%s/%d", orgID, epoch))
}

func contributorKey(contributorID string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("contributor/%s/%d", contributorID, epoch))
}

func memCountKey(orgID string, contentHash []byte, epoch uint64) []byte {
	return []byte(fmt.Sprintf("memcount/%s/%s/%d", orgID, types.ContentHashToHex(contentHash), epoch))
}

func denialCountKey(orgID string, contentHash []byte, epoch uint64) []byte {
	return []byte(fmt.Sprintf("denycount/%s/%s/%d", orgID, types.ContentHashToHex(contentHash), epoch))
}

func memFirstKey(orgID string, contentHash []byte, epoch uint64) []byte {
	return []byte(fmt.Sprintf("memfirst/%s/%s/%d", orgID, types.ContentHashToHex(contentHash), epoch))
}

func keyFirstKey(orgID string, serveKey string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("keyfirst/%s/%s/%d", orgID, serveKey, epoch))
}

const ParamsKey = "params"

func (k *Keeper) SetParams(ctx context.Context, params types.Params) error {
	store := k.getStore(ctx)
	bz, err := proto.Marshal(&params)
	if err != nil {
		return fmt.Errorf("marshal params: %w", err)
	}
	return store.Set([]byte(ParamsKey), bz)
}

func (k *Keeper) GetParams(ctx context.Context) (types.Params, error) {
	store := k.getStore(ctx)
	bz, err := store.Get([]byte(ParamsKey))
	if err != nil {
		return types.Params{}, fmt.Errorf("get params: %w", err)
	}
	if bz == nil {
		return types.DefaultParams(), nil
	}
	var params types.Params
	if err := proto.Unmarshal(bz, &params); err != nil {
		return types.Params{}, fmt.Errorf("unmarshal params: %w", err)
	}
	return params, nil
}

func (k *Keeper) HasNullifier(ctx context.Context, nullifier []byte) bool {
	store := k.getStore(ctx)
	has, err := store.Has(nullifierKey(nullifier))
	if err != nil {
		return false
	}
	return has
}

func (k *Keeper) HasDenialNullifier(ctx context.Context, nullifier []byte) bool {
	store := k.getStore(ctx)
	has, err := store.Has(denialNullifierKey(nullifier))
	if err != nil {
		return false
	}
	return has
}

func (k *Keeper) StoreDenialAttestation(ctx context.Context, orgID string, epoch uint64, entry *types.DenialEntry) error {
	store := k.getStore(ctx)
	att := &types.StoredDenialAttestation{
		OrgId:      orgID,
		MemoryHash: entry.MemoryHash,
		DenyKey:    entry.DenyKey,
		Reason:     entry.Reason,
		Epoch:      epoch,
		Nullifier:  entry.Nullifier,
	}
	bz, err := proto.Marshal(att)
	if err != nil {
		return fmt.Errorf("marshal denial attestation: %w", err)
	}
	store.Set(denialAttestationKey(orgID, epoch, entry.Nullifier), bz)
	return nil
}

func (k *Keeper) ProcessServeBatch(ctx context.Context, orgID string, epoch uint64, serves []*types.ServeEntry) (accepted, rejectedDuplicate, rejectedInvalid uint64, err error) {
	params, _ := k.GetParams(ctx)

	if len(serves) > int(params.MaxServesPerBatch) {
		return 0, 0, 0, types.ErrBatchTooLarge
	}

	if err := k.bandwidthKeeper.ConsumeServeBandwidth(ctx, orgID, epoch, uint64(len(serves))); err != nil {
		return 0, 0, 0, err
	}

	for _, serve := range serves {
		if k.HasNullifier(ctx, serve.Nullifier) {
			rejectedDuplicate++
			continue
		}

		_, err := k.memoryKeeper.GetApprovedMemory(ctx, orgID, serve.MemoryContentHash)
		if err != nil {
			rejectedInvalid++
			continue
		}

		canonicalCID := types.ContentHashToHex(serve.MemoryContentHash)
		valid, err := k.memoryKeeper.IsValidInEpoch(ctx, orgID, canonicalCID, epoch)
		if err != nil {
			return accepted, rejectedDuplicate, rejectedInvalid, err
		}
		if !valid {
			rejectedInvalid++
			continue
		}

		isSelfServe := serve.ServeKey == serve.ContributorId

		currentCount := k.GetMemoryServeCount(ctx, orgID, serve.MemoryContentHash, epoch)
		if currentCount >= uint64(params.MaxServesPerMemoryPerEpoch) {
			rejectedInvalid++
			continue
		}

		store := k.getStore(ctx)

		store.Set(nullifierKey(serve.Nullifier), []byte{1})

		attestation := types.NewServeAttestation(orgID, serve.MemoryContentHash, serve.ServeKey, serve.ContributorId, epoch, serve.Nullifier, isSelfServe, serve.ModelId, serve.TurnCount)
		attBz, err := proto.Marshal(types.ServeAttestationToStored(attestation))
		if err != nil {
			return 0, 0, 0, fmt.Errorf("marshal attestation: %w", err)
		}
		store.Set(attestationKey(orgID, epoch, serve.Nullifier), attBz)

		k.incrementMemoryServeCount(ctx, orgID, serve.MemoryContentHash, epoch)

		k.updateEpochStats(ctx, orgID, epoch, serve.MemoryContentHash, serve.ServeKey, isSelfServe, serve.ModelId)

		k.updateContributorServes(ctx, serve.ContributorId, epoch, orgID, isSelfServe, serve.TurnCount)

		accepted++

		if err := k.reputationKeeper.RecordServe(ctx, []byte(serve.ContributorWallet), orgID, epoch, isSelfServe); err != nil {
			k.logger.Info("failed to record serve reputation",
				"contributor", serve.ContributorWallet,
				"org", orgID,
				"error", err,
			)
		}
	}

	return accepted, rejectedDuplicate, rejectedInvalid, nil
}

func (k *Keeper) incrementMemoryServeCount(ctx context.Context, orgID string, contentHash []byte, epoch uint64) {
	store := k.getStore(ctx)
	key := memCountKey(orgID, contentHash, epoch)
	bz, _ := store.Get(key)
	var count uint64
	if bz != nil {
		count = binary.BigEndian.Uint64(bz)
	}
	count++
	bz = make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(key, bz)
}

func (k *Keeper) updateEpochStats(ctx context.Context, orgID string, epoch uint64, contentHash []byte, serveKey string, isSelfServe bool, modelID string) {
	store := k.getStore(ctx)

	memFirst := memFirstKey(orgID, contentHash, epoch)
	if has, _ := store.Has(memFirst); !has {
		store.Set(memFirst, []byte{1})
		stats, _ := k.GetEpochServeStats(ctx, orgID, epoch)
		if stats == nil {
			stats = types.NewEpochServeStats(orgID, epoch)
		}
		stats.UniqueMemoriesServed++
		esBz, _ := proto.Marshal(types.EpochServeStatsToStored(stats))
		store.Set(statsKey(orgID, epoch), esBz)
	}

	keyFirst := keyFirstKey(orgID, serveKey, epoch)
	if has, _ := store.Has(keyFirst); !has {
		store.Set(keyFirst, []byte{1})
		stats, _ := k.GetEpochServeStats(ctx, orgID, epoch)
		if stats == nil {
			stats = types.NewEpochServeStats(orgID, epoch)
		}
		stats.UniqueServeKeys++
		esBz, _ := proto.Marshal(types.EpochServeStatsToStored(stats))
		store.Set(statsKey(orgID, epoch), esBz)
	}

	stats, _ := k.GetEpochServeStats(ctx, orgID, epoch)
	if stats == nil {
		stats = types.NewEpochServeStats(orgID, epoch)
	}
	stats.TotalServes++
	if isSelfServe {
		stats.SelfServes++
	}
	if modelID != "" {
		if stats.ModelBreakdown == nil {
			stats.ModelBreakdown = make(map[string]uint64)
		}
		stats.ModelBreakdown[modelID]++
	}
	esBz, _ := proto.Marshal(types.EpochServeStatsToStored(stats))
	store.Set(statsKey(orgID, epoch), esBz)
}

func (k *Keeper) updateContributorServes(ctx context.Context, contributorID string, epoch uint64, orgID string, isSelfServe bool, turnCount uint32) {
	store := k.getStore(ctx)
	key := contributorKey(contributorID, epoch)

	var cs *types.ContributorEpochServes
	bz, _ := store.Get(key)
	if bz != nil {
		var stored types.StoredContributorEpochServes
		proto.Unmarshal(bz, &stored)
		storedCs := types.StoredToContributorEpochServes(stored)
		cs = &storedCs
	} else {
		cs = types.NewContributorEpochServes(contributorID, epoch)
	}

	cs.ServeCount++
	if isSelfServe {
		cs.SelfServeCount++
	}
	cs.AddOrgID(orgID)
	cs.TotalTurns += uint64(turnCount)

	csBz, _ := proto.Marshal(types.ContributorEpochServesToStored(cs))
	store.Set(key, csBz)
}

func (k *Keeper) GetEpochServeStats(ctx context.Context, orgID string, epoch uint64) (*types.EpochServeStats, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(statsKey(orgID, epoch))
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrStatsNotFound
	}

	var stored types.StoredEpochServeStats
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal stats: %w", err)
	}
	es := types.StoredToEpochServeStats(stored)
	return &es, nil
}

func (k *Keeper) GetContributorEpochServes(ctx context.Context, contributorID string, epoch uint64) (*types.ContributorEpochServes, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(contributorKey(contributorID, epoch))
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrContributorNotFound
	}

	var stored types.StoredContributorEpochServes
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal contributor serves: %w", err)
	}
	cs := types.StoredToContributorEpochServes(stored)
	return &cs, nil
}

func (k *Keeper) GetEpochServeStatsRaw(ctx context.Context, orgID string, epoch uint64) (uint64, uint64, uint64, map[string]uint64, error) {
	stats, err := k.GetEpochServeStats(ctx, orgID, epoch)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	return stats.TotalServes, stats.UniqueMemoriesServed, stats.SelfServes, stats.ModelBreakdown, nil
}

func (k *Keeper) GetContributorEpochServesRaw(ctx context.Context, contributorID string, epoch uint64) (uint64, uint64, uint64, []string, error) {
	cs, err := k.GetContributorEpochServes(ctx, contributorID, epoch)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	return cs.ServeCount, cs.SelfServeCount, cs.TotalTurns, cs.OrgIDs, nil
}

func (k *Keeper) GetMemoryServeCount(ctx context.Context, orgID string, contentHash []byte, epoch uint64) uint64 {
	store := k.getStore(ctx)
	bz, err := store.Get(memCountKey(orgID, contentHash, epoch))
	if err != nil || bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (k *Keeper) GetMemoryDenialCount(ctx context.Context, orgID string, contentHash []byte, epoch uint64) uint64 {
	store := k.getStore(ctx)
	bz, err := store.Get(denialCountKey(orgID, contentHash, epoch))
	if err != nil || bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (k *Keeper) IncrementDenialCount(ctx context.Context, orgID string, contentHash []byte, epoch uint64) {
	store := k.getStore(ctx)
	key := denialCountKey(orgID, contentHash, epoch)
	bz, _ := store.Get(key)
	var count uint64
	if bz != nil {
		count = binary.BigEndian.Uint64(bz)
	}
	count++
	bz = make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(key, bz)
}

func (k *Keeper) GetMemoryServeCountForEpoch(ctx context.Context, orgID string, memoryCID string, epoch uint64) (uint64, error) {
	contentHash, err := hex.DecodeString(memoryCID)
	if err != nil {
		return 0, fmt.Errorf("decode memory cid: %w", err)
	}
	if len(contentHash) != 32 {
		return 0, fmt.Errorf("invalid content hash length: %d", len(contentHash))
	}
	return k.GetMemoryServeCount(ctx, orgID, contentHash, epoch), nil
}

func (k *Keeper) GetMemoryDenialCountForEpoch(ctx context.Context, orgID string, memoryCID string, epoch uint64) (uint64, error) {
	contentHash, err := hex.DecodeString(memoryCID)
	if err != nil {
		return 0, fmt.Errorf("decode memory cid: %w", err)
	}
	if len(contentHash) != 32 {
		return 0, fmt.Errorf("invalid content hash length: %d", len(contentHash))
	}
	return k.GetMemoryDenialCount(ctx, orgID, contentHash, epoch), nil
}

func (k *Keeper) GetServeAttestations(ctx context.Context, orgID string, epoch uint64) ([]*types.ServeAttestation, error) {
	store := k.getStore(ctx)
	prefix := attestationPrefix(orgID, epoch)
	iter, err := store.Iterator(prefix, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var attestations []*types.ServeAttestation
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredServeAttestation
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		sa := types.StoredToServeAttestation(stored)
		attestations = append(attestations, &sa)
	}
	return attestations, nil
}

func (k *Keeper) InitGenesis(ctx context.Context, state *types.GenesisState) error {
	store := k.getStore(ctx)

	for _, att := range state.Attestations {
		store.Set(nullifierKey(att.Nullifier), []byte{1})
		attBz, _ := proto.Marshal(types.ServeAttestationToStored(att))
		store.Set(attestationKey(att.OrgID, att.Epoch, att.Nullifier), attBz)
	}

	for _, denial := range state.DenialAttestations {
		if denial == nil {
			continue
		}
		store.Set(denialNullifierKey(denial.Nullifier), []byte{1})
		denialBz, err := proto.Marshal(denial)
		if err != nil {
			return fmt.Errorf("marshal denial attestation: %w", err)
		}
		store.Set(denialAttestationKey(denial.OrgId, denial.Epoch, denial.Nullifier), denialBz)
		k.IncrementDenialCount(ctx, denial.OrgId, denial.MemoryHash, denial.Epoch)
	}

	for _, es := range state.EpochStats {
		esBz, _ := proto.Marshal(types.EpochServeStatsToStored(es))
		store.Set(statsKey(es.OrgID, es.Epoch), esBz)
	}

	for _, cs := range state.ContributorServes {
		csBz, _ := proto.Marshal(types.ContributorEpochServesToStored(cs))
		store.Set(contributorKey(cs.ContributorID, cs.Epoch), csBz)
	}

	return nil
}

func (k *Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	store := k.getStore(ctx)

	var attestations []*types.ServeAttestation
	attPrefix := []byte("attestation/")
	attIter, err := store.Iterator(attPrefix, nil)
	if err != nil {
		return nil, err
	}
	defer attIter.Close()
	for ; attIter.Valid(); attIter.Next() {
		var stored types.StoredServeAttestation
		if err := proto.Unmarshal(attIter.Value(), &stored); err != nil {
			continue
		}
		sa := types.StoredToServeAttestation(stored)
		attestations = append(attestations, &sa)
	}

	var denialAttestations []*types.StoredDenialAttestation
	denialPrefix := []byte("denial/")
	denialIter, err := store.Iterator(denialPrefix, nil)
	if err != nil {
		return nil, err
	}
	defer denialIter.Close()
	for ; denialIter.Valid(); denialIter.Next() {
		var stored types.StoredDenialAttestation
		if err := proto.Unmarshal(denialIter.Value(), &stored); err != nil {
			continue
		}
		copied := stored
		denialAttestations = append(denialAttestations, &copied)
	}

	var epochStats []*types.EpochServeStats
	statsPrefix := []byte("stats/")
	statsIter, err := store.Iterator(statsPrefix, nil)
	if err != nil {
		return nil, err
	}
	defer statsIter.Close()
	for ; statsIter.Valid(); statsIter.Next() {
		var stored types.StoredEpochServeStats
		if err := proto.Unmarshal(statsIter.Value(), &stored); err != nil {
			continue
		}
		es := types.StoredToEpochServeStats(stored)
		epochStats = append(epochStats, &es)
	}

	var contributorServes []*types.ContributorEpochServes
	contribPrefix := []byte("contributor/")
	contribIter, err := store.Iterator(contribPrefix, nil)
	if err != nil {
		return nil, err
	}
	defer contribIter.Close()
	for ; contribIter.Valid(); contribIter.Next() {
		var stored types.StoredContributorEpochServes
		if err := proto.Unmarshal(contribIter.Value(), &stored); err != nil {
			continue
		}
		cs := types.StoredToContributorEpochServes(stored)
		contributorServes = append(contributorServes, &cs)
	}

	return &types.GenesisState{
		Attestations:       attestations,
		DenialAttestations: denialAttestations,
		EpochStats:         epochStats,
		ContributorServes:  contributorServes,
	}, nil
}

func (k *Keeper) Logger(ctx context.Context) log.Logger {
	return k.logger.With("module", "x/serve")
}
