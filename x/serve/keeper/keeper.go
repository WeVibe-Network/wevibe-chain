package keeper

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

func serveFingerprintKey(fingerprint []byte) []byte {
	return []byte(fmt.Sprintf("fingerprint/%s", types.ContentHashToHex(fingerprint)))
}

func denialFingerprintKey(fingerprint []byte) []byte {
	return []byte(fmt.Sprintf("denyfingerprint/%s", types.ContentHashToHex(fingerprint)))
}

func receiptKey(orgID string, epoch uint64, fingerprint []byte) []byte {
	return []byte(fmt.Sprintf("receipt/%s/%d/%s", orgID, epoch, types.ContentHashToHex(fingerprint)))
}

func receiptPrefix(orgID string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("receipt/%s/%d/", orgID, epoch))
}

func denialReceiptKey(orgID string, epoch uint64, fingerprint []byte) []byte {
	return []byte(fmt.Sprintf("denial/%s/%d/%s", orgID, epoch, types.ContentHashToHex(fingerprint)))
}

func denialReceiptPrefix(orgID string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("denial/%s/%d/", orgID, epoch))
}

// eventKey stores the immutable recall-pivot event log entry.
// Layout: event/{org_id}/{epoch}/{fingerprint_hex} -> StoredEvent.
func eventKey(orgID string, epoch uint64, fingerprint []byte) []byte {
	return []byte(fmt.Sprintf("%s%s/%d/%s", types.EventPrefix, orgID, epoch, types.ContentHashToHex(fingerprint)))
}

func eventPrefix(orgID string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("%s%s/%d/", types.EventPrefix, orgID, epoch))
}

// eventFingerprintKey stores global event dedup presence.
// Layout: eventfp/{fingerprint_hex} -> 0x01.
func eventFingerprintKey(fingerprint []byte) []byte {
	return []byte(fmt.Sprintf("%s%s", types.EventFingerprintPrefix, types.ContentHashToHex(fingerprint)))
}

// policyAnchorKey stores policy anchors by published edge-policy version.
// Layout: policy_anchor/{policy_version} -> StoredPolicyAnchor.
func policyAnchorKey(policyVersion string) []byte {
	return []byte(fmt.Sprintf("%s%s", types.PolicyAnchorPrefix, policyVersion))
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

func (k *Keeper) HasServeFingerprint(ctx context.Context, fingerprint []byte) bool {
	store := k.getStore(ctx)
	has, err := store.Has(serveFingerprintKey(fingerprint))
	if err != nil {
		return false
	}
	return has
}

func (k *Keeper) HasDenialFingerprint(ctx context.Context, fingerprint []byte) bool {
	store := k.getStore(ctx)
	has, err := store.Has(denialFingerprintKey(fingerprint))
	if err != nil {
		return false
	}
	return has
}

func (k *Keeper) GetServeReceiptByFingerprint(ctx context.Context, fingerprint []byte) (*types.StoredServeReceipt, bool, error) {
	store := k.getStore(ctx)
	receiptStoreKey, err := store.Get(serveFingerprintKey(fingerprint))
	if err != nil {
		return nil, false, fmt.Errorf("get serve receipt pointer: %w", err)
	}
	if len(receiptStoreKey) == 0 {
		return nil, false, nil
	}

	receiptBz, err := store.Get(receiptStoreKey)
	if err != nil {
		return nil, false, fmt.Errorf("get serve receipt: %w", err)
	}
	if receiptBz == nil {
		return nil, false, nil
	}

	var stored types.StoredServeReceipt
	if err := proto.Unmarshal(receiptBz, &stored); err != nil {
		return nil, false, fmt.Errorf("unmarshal serve receipt: %w", err)
	}

	return &stored, true, nil
}

func (k *Keeper) StoreDenialReceipt(ctx context.Context, orgID string, epoch uint64, denialFingerprint []byte, entry *types.DenialEntry) error {
	store := k.getStore(ctx)
	denialReceipt := &types.StoredDenialReceipt{
		OrgId:            orgID,
		MemoryHash:       entry.MemoryHash,
		DenyKey:          hex.EncodeToString(entry.ServeKeyPubkey),
		Reason:           entry.Reason,
		Epoch:            epoch,
		ServeFingerprint: entry.ServeFingerprint,
		ServeKeyPubkey:   entry.ServeKeyPubkey,
	}
	bz, err := proto.Marshal(denialReceipt)
	if err != nil {
		return fmt.Errorf("marshal denial receipt: %w", err)
	}
	store.Set(denialReceiptKey(orgID, epoch, denialFingerprint), bz)
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
		if len(serve.ServeKeyPubkey) != ed25519.PublicKeySize || len(serve.ServeSig) != ed25519.SignatureSize {
			rejectedInvalid++
			continue
		}

		canonicalBody := types.CanonicalServeBody(orgID, serve.MemoryContentHash, epoch, serve.ServeKeyPubkey, serve.Nonce)
		if !ed25519.Verify(ed25519.PublicKey(serve.ServeKeyPubkey), canonicalBody, serve.ServeSig) {
			rejectedInvalid++
			continue
		}

		serveFingerprint := types.ComputeServeFingerprint(serve.MemoryContentHash, serve.ServeKeyPubkey, epoch)
		if k.HasServeFingerprint(ctx, serveFingerprint) {
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

		serveKeyID := hex.EncodeToString(serve.ServeKeyPubkey)
		isSelfServe := serveKeyID == serve.ContributorId

		currentCount := k.GetMemoryServeCount(ctx, orgID, serve.MemoryContentHash, epoch)
		if currentCount >= uint64(params.MaxServesPerMemoryPerEpoch) {
			rejectedInvalid++
			continue
		}

		store := k.getStore(ctx)

		receipt := types.NewServeReceipt(orgID, serve.MemoryContentHash, serveKeyID, serve.ServeKeyPubkey, serve.ContributorId, epoch, serveFingerprint, isSelfServe, serve.ModelId, serve.TurnCount)
		receiptBz, err := proto.Marshal(types.ServeReceiptToStored(receipt))
		if err != nil {
			return 0, 0, 0, fmt.Errorf("marshal receipt: %w", err)
		}
		receiptStoreKey := receiptKey(orgID, epoch, serveFingerprint)
		store.Set(receiptStoreKey, receiptBz)
		store.Set(serveFingerprintKey(serveFingerprint), receiptStoreKey)

		k.incrementMemoryServeCount(ctx, orgID, serve.MemoryContentHash, epoch)

		k.updateEpochStats(ctx, orgID, epoch, serve.MemoryContentHash, serveKeyID, isSelfServe, serve.ModelId)

		k.updateContributorServes(ctx, serve.ContributorId, epoch, orgID, isSelfServe, serve.TurnCount)

		accepted++

		if mem, memErr := k.memoryKeeper.GetApprovedMemory(ctx, orgID, serve.MemoryContentHash); memErr != nil || mem == nil || mem.Contributor == "" {
			// Attribution is derived solely from the authoritative committed
			// memory record (CO-041 Task F). If the stored contributor pubkey
			// is unavailable, skip the reputation record — never fall back to
			// the untrusted serve payload wallet (R-ONE-PATH).
			k.logger.Info("serve attribution skipped: stored contributor pubkey unavailable",
				"org", orgID,
				"hash", types.ContentHashToHex(serve.MemoryContentHash),
				"error", memErr,
			)
		} else if err := k.reputationKeeper.RecordServe(ctx, []byte(mem.Contributor), orgID, epoch, isSelfServe); err != nil {
			k.logger.Info("failed to record serve reputation",
				"contributor", mem.Contributor,
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

func (k *Keeper) updateEpochDenialStats(ctx context.Context, orgID string, epoch uint64) {
	store := k.getStore(ctx)
	stats, _ := k.GetEpochServeStats(ctx, orgID, epoch)
	if stats == nil {
		stats = types.NewEpochServeStats(orgID, epoch)
	}
	stats.TotalDenials++
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

func (k *Keeper) ProcessEventBatch(ctx context.Context, orgID string, epoch uint64, events []*types.EventEntry) (accepted, rejectedDuplicate, rejectedInvalid uint64, err error) {
	store := k.getStore(ctx)

	for _, entry := range events {
		body, err := types.CanonicalEventBody(entry.EventType, orgID, entry.MemoryContentHash, epoch, entry.SignerPubkey, entry.Nonce, entry)
		if err != nil {
			rejectedInvalid++
			continue
		}
		if !ed25519.Verify(ed25519.PublicKey(entry.SignerPubkey), body, entry.Signature) {
			rejectedInvalid++
			continue
		}

		fingerprint := types.ComputeEventFingerprint(body)
		has, err := store.Has(eventFingerprintKey(fingerprint))
		if err != nil {
			return accepted, rejectedDuplicate, rejectedInvalid, fmt.Errorf("check event fingerprint: %w", err)
		}
		if has {
			rejectedDuplicate++
			continue
		}

		if _, err := k.memoryKeeper.GetApprovedMemory(ctx, orgID, entry.MemoryContentHash); err != nil {
			rejectedInvalid++
			continue
		}

		stored := storedEventFromEntry(orgID, epoch, entry, fingerprint)
		bz, err := proto.Marshal(stored)
		if err != nil {
			return accepted, rejectedDuplicate, rejectedInvalid, fmt.Errorf("marshal event: %w", err)
		}
		store.Set(eventKey(orgID, epoch, fingerprint), bz)
		store.Set(eventFingerprintKey(fingerprint), []byte{0x01})
		accepted++
	}

	return accepted, rejectedDuplicate, rejectedInvalid, nil
}

func storedEventFromEntry(orgID string, epoch uint64, entry *types.EventEntry, fingerprint []byte) *types.StoredEvent {
	stored := &types.StoredEvent{
		OrgId:             orgID,
		Epoch:             epoch,
		EventType:         entry.EventType,
		MemoryContentHash: entry.MemoryContentHash,
		SignerPubkey:      entry.SignerPubkey,
		Nonce:             entry.Nonce,
		Signature:         entry.Signature,
		Fingerprint:       fingerprint,
	}
	switch body := entry.GetBody().(type) {
	case *types.EventEntry_Outcome:
		stored.Body = &types.StoredEvent_Outcome{Outcome: body.Outcome}
	case *types.EventEntry_ValidityPredicate:
		stored.Body = &types.StoredEvent_ValidityPredicate{ValidityPredicate: body.ValidityPredicate}
	case *types.EventEntry_CostToDiscover:
		stored.Body = &types.StoredEvent_CostToDiscover{CostToDiscover: body.CostToDiscover}
	case *types.EventEntry_Convergence:
		stored.Body = &types.StoredEvent_Convergence{Convergence: body.Convergence}
	}
	return stored
}

func (k *Keeper) GetEventsForEpoch(ctx context.Context, orgID string, epoch uint64) ([]*types.StoredEvent, error) {
	store := k.getStore(ctx)
	prefix := eventPrefix(orgID, epoch)
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var events []*types.StoredEvent
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredEvent
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		copied := stored
		events = append(events, &copied)
	}
	return events, nil
}

func (k *Keeper) SetPolicyAnchor(ctx context.Context, policyVersion string, policyHash []byte) error {
	store := k.getStore(ctx)
	key := policyAnchorKey(policyVersion)
	existingBz, err := store.Get(key)
	if err != nil {
		return fmt.Errorf("get policy anchor: %w", err)
	}
	if existingBz != nil {
		var existing types.StoredPolicyAnchor
		if err := proto.Unmarshal(existingBz, &existing); err != nil {
			return fmt.Errorf("unmarshal policy anchor: %w", err)
		}
		if !bytes.Equal(existing.PolicyHash, policyHash) {
			return fmt.Errorf("policy anchor already exists with different hash")
		}
		return nil
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	anchor := &types.StoredPolicyAnchor{
		PolicyVersion:    policyVersion,
		PolicyHash:       append([]byte(nil), policyHash...),
		AnchoredAtEpoch:  0, // x/serve has no epoch source; height is the authoritative on-chain ordinal.
		AnchoredAtHeight: sdkCtx.BlockHeight(),
	}
	bz, err := proto.Marshal(anchor)
	if err != nil {
		return fmt.Errorf("marshal policy anchor: %w", err)
	}
	store.Set(key, bz)
	store.Set([]byte(types.LatestPolicyAnchorKey), []byte(policyVersion))
	return nil
}

func (k *Keeper) GetPolicyAnchor(ctx context.Context, policyVersion string) (*types.StoredPolicyAnchor, bool, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(policyAnchorKey(policyVersion))
	if err != nil {
		return nil, false, fmt.Errorf("get policy anchor: %w", err)
	}
	if bz == nil {
		return nil, false, nil
	}
	var anchor types.StoredPolicyAnchor
	if err := proto.Unmarshal(bz, &anchor); err != nil {
		return nil, false, fmt.Errorf("unmarshal policy anchor: %w", err)
	}
	return &anchor, true, nil
}

func (k *Keeper) GetLatestPolicyAnchor(ctx context.Context) (*types.StoredPolicyAnchor, bool, error) {
	store := k.getStore(ctx)
	versionBz, err := store.Get([]byte(types.LatestPolicyAnchorKey))
	if err != nil {
		return nil, false, fmt.Errorf("get latest policy anchor pointer: %w", err)
	}
	if len(versionBz) == 0 {
		return nil, false, nil
	}
	return k.GetPolicyAnchor(ctx, string(versionBz))
}

func (k *Keeper) GetServeReceipts(ctx context.Context, orgID string, epoch uint64) ([]*types.ServeReceipt, error) {
	store := k.getStore(ctx)
	prefix := receiptPrefix(orgID, epoch)
	iter, err := store.Iterator(prefix, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var receipts []*types.ServeReceipt
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredServeReceipt
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		sr := types.StoredToServeReceipt(stored)
		receipts = append(receipts, &sr)
	}
	return receipts, nil
}

func (k *Keeper) InitGenesis(ctx context.Context, state *types.GenesisState) error {
	store := k.getStore(ctx)

	for _, receipt := range state.ServeReceipts {
		receiptBz, _ := proto.Marshal(types.ServeReceiptToStored(receipt))
		receiptStoreKey := receiptKey(receipt.OrgID, receipt.Epoch, receipt.Fingerprint)
		store.Set(receiptStoreKey, receiptBz)
		store.Set(serveFingerprintKey(receipt.Fingerprint), receiptStoreKey)
	}

	for _, denial := range state.DenialReceipts {
		if denial == nil {
			continue
		}
		denialFingerprint := types.ComputeDenialFingerprint(denial.OrgId, denial.MemoryHash, denial.Epoch, denial.ServeKeyPubkey, denial.ServeFingerprint)
		store.Set(denialFingerprintKey(denialFingerprint), []byte{1})
		denialBz, err := proto.Marshal(denial)
		if err != nil {
			return fmt.Errorf("marshal denial receipt: %w", err)
		}
		store.Set(denialReceiptKey(denial.OrgId, denial.Epoch, denialFingerprint), denialBz)
		k.IncrementDenialCount(ctx, denial.OrgId, denial.MemoryHash, denial.Epoch)
		k.updateEpochDenialStats(ctx, denial.OrgId, denial.Epoch)
	}

	for _, event := range state.Events {
		if event == nil {
			continue
		}
		eventBz, err := proto.Marshal(event)
		if err != nil {
			return fmt.Errorf("marshal event: %w", err)
		}
		store.Set(eventKey(event.OrgId, event.Epoch, event.Fingerprint), eventBz)
		store.Set(eventFingerprintKey(event.Fingerprint), []byte{0x01})
	}

	for _, anchor := range state.PolicyAnchors {
		if anchor == nil {
			continue
		}
		anchorBz, err := proto.Marshal(anchor)
		if err != nil {
			return fmt.Errorf("marshal policy anchor: %w", err)
		}
		store.Set(policyAnchorKey(anchor.PolicyVersion), anchorBz)
		store.Set([]byte(types.LatestPolicyAnchorKey), []byte(anchor.PolicyVersion))
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

	var receipts []*types.ServeReceipt
	receiptPrefix := []byte("receipt/")
	receiptIter, err := store.Iterator(receiptPrefix, nil)
	if err != nil {
		return nil, err
	}
	defer receiptIter.Close()
	for ; receiptIter.Valid(); receiptIter.Next() {
		var stored types.StoredServeReceipt
		if err := proto.Unmarshal(receiptIter.Value(), &stored); err != nil {
			continue
		}
		sr := types.StoredToServeReceipt(stored)
		receipts = append(receipts, &sr)
	}

	var denialReceipts []*types.StoredDenialReceipt
	denialPrefix := []byte("denial/")
	denialIter, err := store.Iterator(denialPrefix, nil)
	if err != nil {
		return nil, err
	}
	defer denialIter.Close()
	for ; denialIter.Valid(); denialIter.Next() {
		var stored types.StoredDenialReceipt
		if err := proto.Unmarshal(denialIter.Value(), &stored); err != nil {
			continue
		}
		copied := stored
		denialReceipts = append(denialReceipts, &copied)
	}

	var events []*types.StoredEvent
	eventIter, err := store.Iterator([]byte(types.EventPrefix), nil)
	if err != nil {
		return nil, err
	}
	defer eventIter.Close()
	for ; eventIter.Valid(); eventIter.Next() {
		var stored types.StoredEvent
		if err := proto.Unmarshal(eventIter.Value(), &stored); err != nil {
			continue
		}
		copied := stored
		events = append(events, &copied)
	}

	var policyAnchors []*types.StoredPolicyAnchor
	anchorIter, err := store.Iterator([]byte(types.PolicyAnchorPrefix), nil)
	if err != nil {
		return nil, err
	}
	defer anchorIter.Close()
	for ; anchorIter.Valid(); anchorIter.Next() {
		var stored types.StoredPolicyAnchor
		if err := proto.Unmarshal(anchorIter.Value(), &stored); err != nil {
			continue
		}
		copied := stored
		policyAnchors = append(policyAnchors, &copied)
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
		ServeReceipts:     receipts,
		DenialReceipts:    denialReceipts,
		Events:            events,
		PolicyAnchors:     policyAnchors,
		EpochStats:        epochStats,
		ContributorServes: contributorServes,
	}, nil
}

func (k *Keeper) Logger(ctx context.Context) log.Logger {
	return k.logger.With("module", "x/serve")
}
