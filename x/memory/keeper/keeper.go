package keeper

import (
	"context"
	"encoding/binary"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/gogoproto/proto"
	"github.com/wevibe-network/wevibe-chain/x/memory/types"
)

type Keeper struct {
	storeService     store.KVStoreService
	logger           log.Logger
	authority        string
	orgKeeper        types.OrgKeeper
	serveKeeper      types.ServeKeeper
	reputationKeeper types.ReputationKeeper
}

func NewKeeper(storeService store.KVStoreService, logger log.Logger, authority string, orgKeeper types.OrgKeeper, reputationKeeper types.ReputationKeeper) *Keeper {
	return &Keeper{
		storeService:     storeService,
		logger:           logger,
		authority:        authority,
		orgKeeper:        orgKeeper,
		serveKeeper:      nil,
		reputationKeeper: reputationKeeper,
	}
}

func (k *Keeper) SetServeKeeper(serveKeeper types.ServeKeeper) {
	if serveKeeper == nil {
		panic("serve keeper cannot be nil")
	}
	k.serveKeeper = serveKeeper
}

func (k *Keeper) getStore(ctx context.Context) store.KVStore {
	return k.storeService.OpenKVStore(ctx)
}

func pendingKey(orgID string, contentHash []byte) []byte {
	return []byte(fmt.Sprintf("pending/%s/%s", orgID, types.ContentHashToHex(contentHash)))
}

func pendingPrefix(orgID string) []byte {
	return []byte(fmt.Sprintf("pending/%s/", orgID))
}

func approvedKey(orgID string, contentHash []byte) []byte {
	return []byte(fmt.Sprintf("approved/%s/%s", orgID, types.ContentHashToHex(contentHash)))
}

func approvedPrefix(orgID string) []byte {
	return []byte(fmt.Sprintf("approved/%s/", orgID))
}

func merkleKey(orgID string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("merkle/%s/%d", orgID, epoch))
}

var reportPrefixKey = []byte("report/")

func reportKey(orgID string, contentHash []byte, reporterID string) []byte {
	return []byte(fmt.Sprintf("report/%s/%s/%s", orgID, types.ContentHashToHex(contentHash), reporterID))
}

func reportPrefix(orgID string, contentHash []byte) []byte {
	return []byte(fmt.Sprintf("report/%s/%s/", orgID, types.ContentHashToHex(contentHash)))
}

func (k *Keeper) IterateUpheldReports(ctx context.Context, cb func(*types.StoredMemoryReport) bool) error {
	store := k.getStore(ctx)
	iter, err := store.Iterator(reportPrefixKey, storetypes.PrefixEndBytes(reportPrefixKey))
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var report types.StoredMemoryReport
		if err := proto.Unmarshal(iter.Value(), &report); err != nil {
			return err
		}
		if !cb(&report) {
			return nil
		}
	}
	return nil
}

func (k *Keeper) GetUpheldReport(ctx context.Context, orgID string, memoryHash []byte) (*types.StoredMemoryReport, error) {
	store := k.getStore(ctx)
	memHashHex := types.ContentHashToHex(memoryHash)
	iter, err := store.Iterator(reportPrefixKey, storetypes.PrefixEndBytes(reportPrefixKey))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var report types.StoredMemoryReport
		if err := proto.Unmarshal(iter.Value(), &report); err != nil {
			continue
		}
		if report.OrgId == orgID && types.ContentHashToHex(report.ContentHash) == memHashHex {
			return &report, nil
		}
	}
	return nil, nil
}

func (k *Keeper) ReportMemory(ctx context.Context, orgID string, contentHash []byte, contributorPubkey string, approvingModerators []string, upholdingModerators []string, reporterPubkey string, committingLeader string, reason string, plaintext []byte, ciphertext []byte, capsule []byte, plaintextHash []byte, plaintextOversized bool, epoch uint64) error {
	if len(plaintextHash) == 0 {
		return types.ErrInvalidContentHash
	}

	if plaintextOversized {
		if len(plaintext) > 0 || len(ciphertext) > 0 || len(capsule) > 0 {
			return types.ErrBlobTooLarge
		}
	} else {
		if len(plaintext) > 4096 || len(ciphertext) > 8192 || len(capsule) == 0 {
			return types.ErrBlobTooLarge
		}
	}

	store := k.getStore(ctx)
	memKey := approvedKey(orgID, contentHash)
	has, err := store.Has(memKey)
	if err != nil {
		return err
	}
	if !has {
		return types.ErrMemoryNotFound
	}

	rKey := reportKey(orgID, contentHash, reporterPubkey)
	has, err = store.Has(rKey)
	if err != nil {
		return err
	}
	if has {
		return types.ErrReportExists
	}

	report := &types.StoredMemoryReport{
		OrgId:                  orgID,
		ContentHash:            contentHash,
		ContributorPubkey:      contributorPubkey,
		ApprovingModerators:    approvingModerators,
		UpholdingModerators:    upholdingModerators,
		ReporterPubkey:         reporterPubkey,
		CommittingLeaderPubkey: committingLeader,
		Reason:                 reason,
		Plaintext:              plaintext,
		Ciphertext:             ciphertext,
		Capsule:                capsule,
		PlaintextHash:          plaintextHash,
		PlaintextOversized:     plaintextOversized,
		UpheldAtEpoch:          epoch,
		UpheldAtTimestamp:      0,
	}

	bz, err := proto.Marshal(report)
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	if err := store.Set(rKey, bz); err != nil {
		return err
	}

	// Update org aggregate
	if err := k.orgKeeper.IncrementOrgUpheldReports(ctx, orgID); err != nil {
		return fmt.Errorf("increment org upheld reports: %w", err)
	}

	// Update contributor's upheld report count
	if err := k.reputationKeeper.IncrementContribution(ctx, contributorPubkey, orgID, types.ContentHashToHex(contentHash)); err != nil {
		return fmt.Errorf("increment contribution: %w", err)
	}

	// Update each approving moderator's upheld count (they approved something that got deleted)
	for _, modPubkey := range approvingModerators {
		if err := k.reputationKeeper.IncrementModeratorUpheld(ctx, modPubkey, orgID); err != nil {
			return fmt.Errorf("increment moderator upheld: %w", err)
		}
	}

	// Update committing leader's upheld report count
	if err := k.reputationKeeper.IncrementLeaderUpheldReport(ctx, committingLeader, orgID); err != nil {
		return fmt.Errorf("increment leader upheld report: %w", err)
	}

	k.logger.Info("memory reported",
		"org_id", orgID,
		"content_hash", types.ContentHashToHex(contentHash),
		"reporter", reporterPubkey,
		"reason", reason,
	)
	return nil
}

func (k *Keeper) GetMemoryReports(ctx context.Context, orgID string, contentHash []byte) ([]*types.StoredMemoryReport, error) {
	store := k.getStore(ctx)
	prefix := reportPrefix(orgID, contentHash)
	iter, err := store.Iterator(prefix, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var reports []*types.StoredMemoryReport
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredMemoryReport
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		reports = append(reports, &stored)
	}
	return reports, nil
}

func countKey(orgID string) []byte {
	return []byte(fmt.Sprintf("count/%s", orgID))
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

func (k *Keeper) SubmitCommitment(ctx context.Context, commitment *types.PendingCommitment) error {
	if err := commitment.Validate(); err != nil {
		return err
	}

	hasOrg, err := k.orgKeeper.HasOrg(ctx, commitment.OrgID)
	if err != nil {
		return err
	}
	if !hasOrg {
		return types.ErrOrgNotFound
	}

	store := k.getStore(ctx)
	key := pendingKey(commitment.OrgID, commitment.ContentHash)
	has, err := store.Has(key)
	if err != nil {
		return err
	}
	if has {
		return types.ErrCommitmentExists
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	pendingCount := 0
	iter, err := store.Iterator(pendingPrefix(commitment.OrgID), nil)
	if err != nil {
		return err
	}
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		pendingCount++
	}
	if uint64(pendingCount) >= params.MaxPendingPerOrg {
		return types.ErrMaxPendingExceeded
	}

	bz, err := proto.Marshal(pendingToStored(commitment))
	if err != nil {
		return fmt.Errorf("marshal commitment: %w", err)
	}
	if err := store.Set(key, bz); err != nil {
		return err
	}

	k.logger.Info("commitment submitted",
		"org_id", commitment.OrgID,
		"content_hash", types.ContentHashToHex(commitment.ContentHash),
		"contributor", commitment.Contributor,
	)
	return nil
}

func (k *Keeper) ApproveMemory(ctx context.Context, orgID string, contentHash, encryptedBlob []byte, committingLeader string, wrappedDekEnc []byte, memoryType types.MemoryType) error {
	if !types.ValidMemoryType(memoryType) {
		return types.ErrInvalidMemoryType
	}

	store := k.getStore(ctx)
	pendingStoreKey := pendingKey(orgID, contentHash)
	has, err := store.Has(pendingStoreKey)
	if err != nil {
		return err
	}
	if !has {
		return types.ErrCommitmentNotFound
	}

	isLeader, err := k.orgKeeper.IsLeader(ctx, orgID, committingLeader)
	if err != nil {
		return err
	}
	if !isLeader {
		return types.ErrUnauthorized
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if len(encryptedBlob) > int(params.MaxBlobSizeBytes) {
		return types.ErrBlobTooLarge
	}

	var pendingStored types.StoredPendingCommitment
	bz, err := store.Get(pendingStoreKey)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(bz, &pendingStored); err != nil {
		return fmt.Errorf("unmarshal pending: %w", err)
	}
	pending := storedToPending(pendingStored)

	approvedKey := approvedKey(orgID, contentHash)
	has, err = store.Has(approvedKey)
	if err != nil {
		return err
	}
	if has {
		return types.ErrMemoryExists
	}

	currentEpoch := pending.Epoch
	approved := &types.MemoryCommitment{
		OrgID:              orgID,
		ContentHash:        contentHash,
		EncryptedBlob:      encryptedBlob,
		Keywords:           pending.Keywords,
		Contributor:        pending.Contributor,
		ContributorAddress: pending.ContributorAddress,
		Epoch:              currentEpoch,
		CommittedAtHeight:  pending.SubmittedAt,
		CommittingLeader:   committingLeader,
		State:              types.MemoryState_MEMORY_STATE_COMMITTED,
		LastActiveEpoch:    currentEpoch,
		ApprovedAtEpoch:    k.getCurrentEpoch(ctx),
		WrappedDekEnc:      wrappedDekEnc,
		MemoryType:         memoryType,
	}

	bz, err = proto.Marshal(memoryToStored(approved))
	if err != nil {
		return fmt.Errorf("marshal approved: %w", err)
	}
	if err := store.Set(approvedKey, bz); err != nil {
		return err
	}

	if err := store.Delete(pendingStoreKey); err != nil {
		return err
	}

	countKeyBuf := countKey(orgID)
	countBuf, err := store.Get(countKeyBuf)
	if err != nil {
		return err
	}
	var count uint64
	if countBuf != nil {
		count = binary.BigEndian.Uint64(countBuf)
	}
	count++
	countBuf = make([]byte, 8)
	binary.BigEndian.PutUint64(countBuf, count)
	if err := store.Set(countKeyBuf, countBuf); err != nil {
		return err
	}

	// Update org aggregates
	if err := k.orgKeeper.IncrementOrgCommittedMemories(ctx, orgID); err != nil {
		return fmt.Errorf("increment org committed memories: %w", err)
	}
	if err := k.orgKeeper.SetOrgLastActivityEpoch(ctx, orgID, currentEpoch); err != nil {
		return fmt.Errorf("set org last activity: %w", err)
	}

	// Update contributor reputation
	if err := k.reputationKeeper.IncrementContribution(ctx, pending.Contributor, orgID, types.ContentHashToHex(contentHash)); err != nil {
		return fmt.Errorf("increment contribution: %w", err)
	}

	// Update leader's profile
	if err := k.reputationKeeper.IncrementLeaderChainCommit(ctx, committingLeader, orgID); err != nil {
		return fmt.Errorf("increment leader chain commit: %w", err)
	}

	k.logger.Info("memory approved",
		"org_id", orgID,
		"content_hash", types.ContentHashToHex(contentHash),
		"committing_leader", committingLeader,
	)
	return nil
}

func (k *Keeper) GetApprovedMemory(ctx context.Context, orgID string, contentHash []byte) (*types.MemoryCommitment, error) {
	store := k.getStore(ctx)
	key := approvedKey(orgID, contentHash)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrMemoryNotFound
	}

	var stored types.StoredMemoryCommitment
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal approved memory: %w", err)
	}
	memory := storedToMemory(stored)
	return &memory, nil
}

func (k *Keeper) GetPendingCommitment(ctx context.Context, orgID string, contentHash []byte) (*types.PendingCommitment, error) {
	store := k.getStore(ctx)
	key := pendingKey(orgID, contentHash)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrCommitmentNotFound
	}

	var stored types.StoredPendingCommitment
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal pending: %w", err)
	}
	pending := storedToPending(stored)
	return &pending, nil
}

func (k *Keeper) GetAllPendingForOrg(ctx context.Context, orgID string) ([]*types.PendingCommitment, error) {
	store := k.getStore(ctx)
	prefix := pendingPrefix(orgID)
	iter, err := store.Iterator(prefix, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var pending []*types.PendingCommitment
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredPendingCommitment
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		p := storedToPending(stored)
		pending = append(pending, &p)
	}
	return pending, nil
}

func (k *Keeper) GetApprovedCount(ctx context.Context, orgID string) (uint64, error) {
	store := k.getStore(ctx)
	key := countKey(orgID)
	bz, err := store.Get(key)
	if err != nil {
		return 0, err
	}
	if bz == nil {
		return 0, nil
	}
	return binary.BigEndian.Uint64(bz), nil
}

func (k *Keeper) GetApprovedCountByContributor(ctx context.Context, orgID string, epoch uint64) (map[string]uint64, error) {
	store := k.getStore(ctx)
	iter, err := store.Iterator(approvedPrefix(orgID), nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	counts := make(map[string]uint64)
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredMemoryCommitment
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		if stored.Epoch != epoch {
			continue
		}
		counts[stored.ContributorPubkey]++
	}

	return counts, nil
}

func (k *Keeper) ComputeAndStoreEpochMerkleRoot(ctx context.Context, orgID string, epoch uint64) error {
	store := k.getStore(ctx)
	prefix := approvedPrefix(orgID)
	iter, err := store.Iterator(prefix, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	var contentHashes [][]byte
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredMemoryCommitment
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		if stored.Epoch == epoch {
			contentHashes = append(contentHashes, stored.ContentHash)
		}
	}

	merkleRoot := types.ComputeMerkleRoot(contentHashes)
	count := uint64(len(contentHashes))

	bz, err := proto.Marshal(&types.StoredEpochMerkleRoot{
		OrgId:       orgID,
		Epoch:       epoch,
		MerkleRoot:  merkleRoot,
		MemoryCount: count,
	})
	if err != nil {
		return fmt.Errorf("marshal merkle root: %w", err)
	}
	if err := store.Set(merkleKey(orgID, epoch), bz); err != nil {
		return err
	}

	k.logger.Info("epoch merkle root computed",
		"org_id", orgID,
		"epoch", epoch,
		"memory_count", count,
	)
	return nil
}

func (k *Keeper) GetEpochMerkleRoot(ctx context.Context, orgID string, epoch uint64) (*types.EpochMerkleRoot, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(merkleKey(orgID, epoch))
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrMerkleRootNotFound
	}

	var stored types.StoredEpochMerkleRoot
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal merkle root: %w", err)
	}
	return &types.EpochMerkleRoot{
		OrgID:       stored.OrgId,
		Epoch:       stored.Epoch,
		MerkleRoot:  stored.MerkleRoot,
		MemoryCount: stored.MemoryCount,
	}, nil
}

func (k *Keeper) InitGenesis(ctx context.Context, state *types.GenesisState) error {
	store := k.getStore(ctx)

	for _, pc := range state.PendingCommitments {
		bz, err := proto.Marshal(pendingToStored(pc))
		if err != nil {
			return err
		}
		if err := store.Set(pendingKey(pc.OrgID, pc.ContentHash), bz); err != nil {
			return err
		}
	}

	for _, am := range state.MemoryCommitments {
		bz, err := proto.Marshal(memoryToStored(am))
		if err != nil {
			return err
		}
		if err := store.Set(approvedKey(am.OrgID, am.ContentHash), bz); err != nil {
			return err
		}
	}

	for _, rel := range state.Relationships {
		if rel == nil {
			continue
		}
		if err := k.saveRelationship(ctx, rel.OrgID, rel); err != nil {
			return err
		}
	}

	for _, vm := range state.ValidityMetadata {
		if vm == nil {
			continue
		}
		bz, err := proto.Marshal(vm)
		if err != nil {
			return fmt.Errorf("marshal validity metadata: %w", err)
		}
		if err := store.Set(validityKey(vm.OrgId, vm.MemoryCid), bz); err != nil {
			return err
		}
	}

	for _, mr := range state.MerkleRoots {
		bz, err := proto.Marshal(&types.StoredEpochMerkleRoot{
			OrgId:       mr.OrgID,
			Epoch:       mr.Epoch,
			MerkleRoot:  mr.MerkleRoot,
			MemoryCount: mr.MemoryCount,
		})
		if err != nil {
			return err
		}
		if err := store.Set(merkleKey(mr.OrgID, mr.Epoch), bz); err != nil {
			return err
		}
	}

	params := state.Params
	if params == (types.Params{}) {
		params = types.DefaultParams()
	}
	if err := k.SetParams(ctx, params); err != nil {
		return err
	}

	return nil
}

func (k *Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	store := k.getStore(ctx)

	var pending []*types.PendingCommitment
	pendingPrefix := []byte("pending/")
	pendingIter, err := store.Iterator(pendingPrefix, storetypes.PrefixEndBytes(pendingPrefix))
	if err != nil {
		return nil, err
	}
	defer pendingIter.Close()
	for ; pendingIter.Valid(); pendingIter.Next() {
		var stored types.StoredPendingCommitment
		if err := proto.Unmarshal(pendingIter.Value(), &stored); err != nil {
			continue
		}
		p := storedToPending(stored)
		pending = append(pending, &p)
	}

	var approved []*types.MemoryCommitment
	approvedPrefix := []byte("approved/")
	approvedIter, err := store.Iterator(approvedPrefix, storetypes.PrefixEndBytes(approvedPrefix))
	if err != nil {
		return nil, err
	}
	defer approvedIter.Close()
	for ; approvedIter.Valid(); approvedIter.Next() {
		var stored types.StoredMemoryCommitment
		if err := proto.Unmarshal(approvedIter.Value(), &stored); err != nil {
			continue
		}
		a := storedToMemory(stored)
		approved = append(approved, &a)
	}

	var merkle []*types.EpochMerkleRoot
	merklePrefix := []byte("merkle/")
	merkleIter, err := store.Iterator(merklePrefix, storetypes.PrefixEndBytes(merklePrefix))
	if err != nil {
		return nil, err
	}
	defer merkleIter.Close()
	for ; merkleIter.Valid(); merkleIter.Next() {
		var stored types.StoredEpochMerkleRoot
		if err := proto.Unmarshal(merkleIter.Value(), &stored); err != nil {
			continue
		}
		merkle = append(merkle, &types.EpochMerkleRoot{
			OrgID:       stored.OrgId,
			Epoch:       stored.Epoch,
			MerkleRoot:  stored.MerkleRoot,
			MemoryCount: stored.MemoryCount,
		})
	}

	var relationships []*types.MemoryRelationship
	relPrefix := types.RelationshipKeyPrefix
	relIter, err := store.Iterator(relPrefix, storetypes.PrefixEndBytes(relPrefix))
	if err != nil {
		return nil, err
	}
	defer relIter.Close()
	for ; relIter.Valid(); relIter.Next() {
		var stored types.StoredMemoryRelationship
		if err := proto.Unmarshal(relIter.Value(), &stored); err != nil {
			continue
		}
		rel := storedToRelationship(stored)
		relationships = append(relationships, &rel)
	}

	var validity []*types.StoredValidityMetadata
	valPrefix := types.ValidityKeyPrefix
	valIter, err := store.Iterator(valPrefix, storetypes.PrefixEndBytes(valPrefix))
	if err != nil {
		return nil, err
	}
	defer valIter.Close()
	for ; valIter.Valid(); valIter.Next() {
		var stored types.StoredValidityMetadata
		if err := proto.Unmarshal(valIter.Value(), &stored); err != nil {
			continue
		}
		copy := stored
		validity = append(validity, &copy)
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	return &types.GenesisState{
		PendingCommitments: pending,
		MemoryCommitments:  approved,
		Relationships:      relationships,
		ValidityMetadata:   validity,
		MerkleRoots:        merkle,
		Params:             params,
	}, nil
}

func (k *Keeper) Logger(ctx context.Context) log.Logger {
	return k.logger.With("module", "x/memory")
}

func pendingToStored(pc *types.PendingCommitment) *types.StoredPendingCommitment {
	return &types.StoredPendingCommitment{
		OrgId:              pc.OrgID,
		ContentHash:        pc.ContentHash,
		Keywords:           pc.Keywords,
		ContributorId:      pc.Contributor,
		Epoch:              pc.Epoch,
		SubmittedAtHeight:  pc.SubmittedAt,
		MemoryType:         pc.MemoryType,
		ContributorAddress: pc.ContributorAddress,
	}
}

func storedToPending(stored types.StoredPendingCommitment) types.PendingCommitment {
	return types.PendingCommitment{
		OrgID:              stored.OrgId,
		ContentHash:        stored.ContentHash,
		Keywords:           stored.Keywords,
		Contributor:        stored.ContributorId,
		ContributorAddress: stored.ContributorAddress,
		Epoch:              stored.Epoch,
		SubmittedAt:        stored.SubmittedAtHeight,
		MemoryType:         stored.MemoryType,
	}
}

func memoryToStored(con *types.MemoryCommitment) *types.StoredMemoryCommitment {
	return &types.StoredMemoryCommitment{
		OrgId:                  con.OrgID,
		ContentHash:            con.ContentHash,
		EncryptedBlob:          con.EncryptedBlob,
		Keywords:               con.Keywords,
		ContributorPubkey:      con.Contributor,
		ContributorAddress:     con.ContributorAddress,
		Epoch:                  con.Epoch,
		CommittedAtHeight:      con.CommittedAtHeight,
		CommittingLeaderPubkey: con.CommittingLeader,
		State:                  con.State,
		LastActiveEpoch:        con.LastActiveEpoch,
		WrappedDekEnc:          con.WrappedDekEnc,
		PlaintextHash:          con.PlaintextHash,
		Salt:                   con.Salt,
		CiphertextHash:         con.CiphertextHash,
		WrappedDekHash:         con.WrappedDekHash,
		ContributorSig:         con.ContributorSig,
		MemoryType:             con.MemoryType,
		ApprovedAtEpoch:        con.ApprovedAtEpoch,
		ServeCountTotal:        con.ServeCountTotal,
		DenialCountTotal:       con.DenialCountTotal,
		ArchivedEpoch:          con.ArchivedEpoch,
	}
}

func storedToMemory(stored types.StoredMemoryCommitment) types.MemoryCommitment {
	return types.MemoryCommitment{
		OrgID:              stored.OrgId,
		ContentHash:        stored.ContentHash,
		EncryptedBlob:      stored.EncryptedBlob,
		Keywords:           stored.Keywords,
		Contributor:        stored.ContributorPubkey,
		ContributorAddress: stored.ContributorAddress,
		Epoch:              stored.Epoch,
		CommittedAtHeight:  stored.CommittedAtHeight,
		CommittingLeader:   stored.CommittingLeaderPubkey,
		State:              stored.State,
		LastActiveEpoch:    stored.LastActiveEpoch,
		WrappedDekEnc:      stored.WrappedDekEnc,
		PlaintextHash:      stored.PlaintextHash,
		Salt:               stored.Salt,
		CiphertextHash:     stored.CiphertextHash,
		WrappedDekHash:     stored.WrappedDekHash,
		ContributorSig:     stored.ContributorSig,
		MemoryType:         stored.MemoryType,
		ApprovedAtEpoch:    stored.ApprovedAtEpoch,
		ServeCountTotal:    stored.ServeCountTotal,
		DenialCountTotal:   stored.DenialCountTotal,
		ArchivedEpoch:      stored.ArchivedEpoch,
	}
}
