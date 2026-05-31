package keeper

import (
	"context"
	"testing"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/store/cachekv"
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/wevibe-network/wevibe-chain/x/memory/types"
)

// cacheKVStoreService routes every keeper store operation through a
// cachekv.Store layered over an IAVL store. This reproduces the exact store
// type used inside BeginBlock / epoch hooks on the live chain, where
// cachekv/internal/mergeiterator.cacheMergeIterator.Error() returns a NON-NIL
// error at normal end-of-iteration. The default unit-test harness iterates a
// direct IAVL store, which returns nil at end -- which is precisely why the
// pre-CO-041 iter.Error()-as-failure pattern was invisible to the existing
// test suite. (R-CACHEKV-ITER)
type cacheKVStoreService struct {
	cache storetypes.KVStore
}

func (s *cacheKVStoreService) OpenKVStore(_ context.Context) corestore.KVStore {
	return cacheKVWrap{kv: s.cache}
}

type cacheKVWrap struct {
	kv storetypes.KVStore
}

func (w cacheKVWrap) Get(key []byte) ([]byte, error) { return w.kv.Get(key), nil }
func (w cacheKVWrap) Has(key []byte) (bool, error)   { return w.kv.Has(key), nil }
func (w cacheKVWrap) Set(key, value []byte) error    { w.kv.Set(key, value); return nil }
func (w cacheKVWrap) Delete(key []byte) error        { w.kv.Delete(key); return nil }

func (w cacheKVWrap) Iterator(start, end []byte) (corestore.Iterator, error) {
	return w.kv.Iterator(start, end), nil
}

func (w cacheKVWrap) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	return w.kv.ReverseIterator(start, end), nil
}

func makeCacheKVTestKeeper(t *testing.T) *Keeper {
	t.Helper()
	storeKey := storetypes.NewKVStoreKey("memory")
	db := dbm.NewMemDB()
	cms := rootmulti.NewStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, nil)
	if err := cms.LoadLatestVersion(); err != nil {
		t.Fatalf("failed to load latest version: %v", err)
	}
	iavlKV := cms.GetKVStore(storeKey)
	cacheStore := cachekv.NewStore(iavlKV)

	sdk.DefaultBondDenom = "uvibe"
	mockOrg := &mockOrgKeeper{
		orgs:    map[string]bool{"test-org": true},
		leaders: map[string]string{"test-org": "leader-pubkey"},
	}
	mockRep := &mockReputationKeeper{}
	return NewKeeper(&cacheKVStoreService{cache: cacheStore}, log.NewNopLogger(), "gov", mockOrg, mockRep)
}

// TestCacheKVEpochDecayPersists is the CO-041 regression test for R-CACHEKV-ITER.
// It MUST fail against the pre-fix code (where ApplyEpochDecay returns the
// cacheMergeIterator false-failure error) and pass after the collect-then-mutate
// fix. It asserts (a) ApplyEpochDecay returns no error under a cache-wrapped
// store, and (b) the idle decay actually persisted to the store.
func TestCacheKVEpochDecayPersists(t *testing.T) {
	k := makeCacheKVTestKeeper(t)
	ctx := context.Background()

	contentHash := []byte("12345678901234567890123456789012")
	commitment := newPendingCommitment(
		"test-org",
		contentHash,
		[]string{"keyword1", "keyword2"},
		"contributor-pubkey",
		1,
		100,
	)
	if err := k.SubmitCommitment(ctx, commitment); err != nil {
		t.Fatalf("SubmitCommitment failed: %v", err)
	}
	if err := approveMemory(k, ctx, "test-org", contentHash, []byte("blob"), "leader-pubkey", nil); err != nil {
		t.Fatalf("ApproveMemory failed: %v", err)
	}

	before, err := k.GetApprovedMemory(ctx, "test-org", contentHash)
	if err != nil {
		t.Fatalf("GetApprovedMemory(before) failed: %v", err)
	}
	beforeWeight := before.Keywords[0].Weight

	// Run well past the default grace period (20 epochs) so idle decay applies.
	const epoch = uint64(40)
	if err := k.ApplyEpochDecay(ctx, epoch); err != nil {
		t.Fatalf("ApplyEpochDecay returned error under cache-wrapped store "+
			"(cacheMergeIterator false-failure not fixed): %v", err)
	}

	after, err := k.GetApprovedMemory(ctx, "test-org", contentHash)
	if err != nil {
		t.Fatalf("GetApprovedMemory(after) failed: %v", err)
	}
	if after.Keywords[0].Weight == beforeWeight {
		t.Fatalf("epoch decay did not persist under cache-wrapped store: weight stayed %s", beforeWeight)
	}
	if after.LastActiveEpoch != epoch {
		t.Fatalf("LastActiveEpoch not persisted: got %d want %d", after.LastActiveEpoch, epoch)
	}
}

// TestCacheKVAfterEpochEndRuns drives the full epoch hook (setCurrentEpoch,
// CheckEpochExpiry, ApplyEpochDecay, getAllOrgsWithMemories) over a
// cache-wrapped store. Per R-EPOCH-HOOK-RESILIENCE the hook returns nil
// unconditionally, but the decay it performs must still persist.
func TestCacheKVAfterEpochEndRuns(t *testing.T) {
	k := makeCacheKVTestKeeper(t)
	ctx := context.Background()

	contentHash := []byte("abcdefabcdefabcdefabcdefabcdef12")
	commitment := newPendingCommitment(
		"test-org",
		contentHash,
		[]string{"keyword1"},
		"contributor-pubkey",
		1,
		100,
	)
	if err := k.SubmitCommitment(ctx, commitment); err != nil {
		t.Fatalf("SubmitCommitment failed: %v", err)
	}
	if err := approveMemory(k, ctx, "test-org", contentHash, []byte("blob"), "leader-pubkey", nil); err != nil {
		t.Fatalf("ApproveMemory failed: %v", err)
	}

	before, err := k.GetApprovedMemory(ctx, "test-org", contentHash)
	if err != nil {
		t.Fatalf("GetApprovedMemory(before) failed: %v", err)
	}
	beforeWeight := before.Keywords[0].Weight

	if err := k.AfterEpochEnd(ctx, WeVibeEpochIdentifier, int64(40)); err != nil {
		t.Fatalf("AfterEpochEnd returned error under cache-wrapped store: %v", err)
	}

	after, err := k.GetApprovedMemory(ctx, "test-org", contentHash)
	if err != nil {
		t.Fatalf("GetApprovedMemory(after) failed: %v", err)
	}
	if after.Keywords[0].Weight == beforeWeight {
		t.Fatalf("AfterEpochEnd did not persist decay under cache-wrapped store: weight stayed %s", beforeWeight)
	}
}

// TestCacheKVEpochDecayZeroSignalGuardSkipsIdle verifies CO-042 quiet-epoch
// guard behavior under a cache-wrapped store context (R-CACHEKV-ITER).
func TestCacheKVEpochDecayZeroSignalGuardSkipsIdle(t *testing.T) {
	k := makeCacheKVTestKeeper(t)
	ctx := context.Background()

	hash := []byte("quietcachekvguardmemoryhash000000")
	storeMemoryWithKeywords(
		t,
		k,
		ctx,
		"test-org",
		hash,
		types.MemoryState_MEMORY_STATE_COMMITTED,
		0,
		withKeywords(&types.KeywordWeight{Keyword: "keyword1", Weight: "0.1600"}),
	)

	k.SetServeKeeper(newMockServeKeeper())

	before, err := k.GetApprovedMemory(ctx, "test-org", hash)
	if err != nil {
		t.Fatalf("GetApprovedMemory(before) failed: %v", err)
	}
	beforeWeight := before.Keywords[0].Weight

	if err := k.ApplyEpochDecay(ctx, 40); err != nil {
		t.Fatalf("ApplyEpochDecay failed: %v", err)
	}

	after, err := k.GetApprovedMemory(ctx, "test-org", hash)
	if err != nil {
		t.Fatalf("GetApprovedMemory(after) failed: %v", err)
	}
	if after.Keywords[0].Weight != beforeWeight {
		t.Fatalf("quiet-epoch guard did not suppress idle decay: before=%s after=%s", beforeWeight, after.Keywords[0].Weight)
	}
	if after.State != types.MemoryState_MEMORY_STATE_COMMITTED {
		t.Fatalf("memory should remain committed under quiet-epoch guard, got %v", after.State)
	}
}

// seedApprovedForContributor submits + approves one memory for the given
// contributor wallet at the current epoch (ApprovedAtEpoch is taken from the
// keeper's current_epoch). Each call needs a unique 32-byte content hash.
func seedApprovedForContributor(t *testing.T, k *Keeper, ctx context.Context, contentHash []byte, wallet string) {
	t.Helper()
	commitment := newPendingCommitment(
		"test-org",
		contentHash,
		[]string{"keyword1"},
		"contributor-pubkey",
		1,
		100,
	)
	commitment.ContributorAddress = wallet
	if err := k.SubmitCommitment(ctx, commitment); err != nil {
		t.Fatalf("SubmitCommitment failed: %v", err)
	}
	if err := approveMemory(k, ctx, "test-org", contentHash, []byte("blob"), "leader-pubkey", nil); err != nil {
		t.Fatalf("ApproveMemory failed: %v", err)
	}
}

// TestCacheKVGetContributorsWithApprovalsInEpoch exercises the CO-041 Task E
// network-wide contributor-by-epoch query over a cache-wrapped store
// (R-CACHEKV-ITER). It asserts per-address counts for the target epoch,
// exclusion of other epochs, and an empty (error-free) result for an epoch with
// no approvals.
func TestCacheKVGetContributorsWithApprovalsInEpoch(t *testing.T) {
	k := makeCacheKVTestKeeper(t)
	ctx := context.Background()

	// Approvals committed during epoch 5.
	if err := k.setCurrentEpoch(ctx, 5); err != nil {
		t.Fatalf("setCurrentEpoch failed: %v", err)
	}
	seedApprovedForContributor(t, k, ctx, []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa1"), "wallet-A")
	seedApprovedForContributor(t, k, ctx, []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa2"), "wallet-A")
	seedApprovedForContributor(t, k, ctx, []byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb1"), "wallet-B")

	// One approval committed during a different epoch (6) — must be excluded.
	if err := k.setCurrentEpoch(ctx, 6); err != nil {
		t.Fatalf("setCurrentEpoch failed: %v", err)
	}
	seedApprovedForContributor(t, k, ctx, []byte("ccccccccccccccccccccccccccccccc1"), "wallet-A")

	counts, err := k.GetContributorsWithApprovalsInEpoch(ctx, 5)
	if err != nil {
		t.Fatalf("GetContributorsWithApprovalsInEpoch(5) error: %v", err)
	}
	if counts["wallet-A"] != 2 {
		t.Fatalf("wallet-A count: got %d want 2", counts["wallet-A"])
	}
	if counts["wallet-B"] != 1 {
		t.Fatalf("wallet-B count: got %d want 1", counts["wallet-B"])
	}
	if len(counts) != 2 {
		t.Fatalf("distinct contributors in epoch 5: got %d want 2", len(counts))
	}

	// Epoch with no approvals returns empty, no error.
	empty, err := k.GetContributorsWithApprovalsInEpoch(ctx, 99)
	if err != nil {
		t.Fatalf("GetContributorsWithApprovalsInEpoch(99) error: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected no contributors in empty epoch, got %d", len(empty))
	}
}

func TestCacheKVGetActiveMemoryCountByOrg(t *testing.T) {
	k := makeCacheKVTestKeeper(t)
	ctx := context.Background()

	if err := k.setCurrentEpoch(ctx, 7); err != nil {
		t.Fatalf("setCurrentEpoch failed: %v", err)
	}

	seedApprovedForContributor(t, k, ctx, []byte("ddddddddddddddddddddddddddddddd1"), "wallet-A")
	seedApprovedForContributor(t, k, ctx, []byte("ddddddddddddddddddddddddddddddd2"), "wallet-A")
	seedApprovedForContributor(t, k, ctx, []byte("ddddddddddddddddddddddddddddddd3"), "wallet-A")

	archivedHash := []byte("ddddddddddddddddddddddddddddddd2")
	deniedHash := []byte("ddddddddddddddddddddddddddddddd3")

	archived, err := k.GetApprovedMemory(ctx, "test-org", archivedHash)
	if err != nil {
		t.Fatalf("GetApprovedMemory archived failed: %v", err)
	}
	archived.State = types.MemoryState_MEMORY_STATE_ARCHIVED
	if err := k.saveMemoryCommitment(ctx, archived); err != nil {
		t.Fatalf("saveMemoryCommitment archived failed: %v", err)
	}

	denied, err := k.GetApprovedMemory(ctx, "test-org", deniedHash)
	if err != nil {
		t.Fatalf("GetApprovedMemory denied failed: %v", err)
	}
	denied.State = types.MemoryState_MEMORY_STATE_DENIED
	if err := k.saveMemoryCommitment(ctx, denied); err != nil {
		t.Fatalf("saveMemoryCommitment denied failed: %v", err)
	}

	count, err := k.GetActiveMemoryCountByOrg(ctx, "test-org")
	if err != nil {
		t.Fatalf("GetActiveMemoryCountByOrg failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("active memory count mismatch: got %d want 1", count)
	}
}
