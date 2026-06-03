package keeper

import (
	"context"
	"testing"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/wevibe-network/wevibe-chain/testutil/keeper"
	"github.com/wevibe-network/wevibe-chain/x/memory/types"
	orgtypes "github.com/wevibe-network/wevibe-chain/x/org/types"
)

type mockOrgKeeper struct {
	orgs    map[string]bool
	leaders map[string]string
}

func (m *mockOrgKeeper) HasOrg(ctx context.Context, orgID string) (bool, error) {
	return m.orgs[orgID], nil
}

func (m *mockOrgKeeper) IsLeader(ctx context.Context, orgID string, memberPubkey string) (bool, error) {
	return m.leaders[orgID] == memberPubkey, nil
}

func (m *mockOrgKeeper) IsModerator(ctx context.Context, orgID string, memberPubkey string) (bool, error) {
	return false, nil
}

func (m *mockOrgKeeper) GetMember(ctx context.Context, orgID, memberPubkey string) (*orgtypes.MemberRecord, error) {
	if m.leaders[orgID] == memberPubkey {
		return &orgtypes.MemberRecord{OrgID: orgID, Pubkey: memberPubkey, Role: "leader"}, nil
	}
	return nil, orgtypes.ErrMemberNotFound
}

func (m *mockOrgKeeper) GetOrgConfig(ctx context.Context, orgID string) (*orgtypes.OrgConfig, error) {
	return &orgtypes.OrgConfig{OrgID: orgID}, nil
}

func (m *mockOrgKeeper) GetLeaderWallet(ctx context.Context, orgID string) (string, error) {
	return m.leaders[orgID], nil
}

func makeTestKeeper(t *testing.T) (*Keeper, *mockOrgKeeper, *mockReputationKeeper) {
	storeKey := storetypes.NewKVStoreKey("memory")
	storeService, _ := keeper.NewTestStoreService(t, storeKey)
	logger := keeper.NewTestLogger()
	sdk.DefaultBondDenom = "uvibe"
	mockOrg := &mockOrgKeeper{
		orgs:    map[string]bool{"test-org": true},
		leaders: map[string]string{"test-org": "leader-pubkey"},
	}
	mockRep := &mockReputationKeeper{}
	return NewKeeper(storeService, logger, "gov", mockOrg, mockRep), mockOrg, mockRep
}

func newPendingCommitment(orgID string, contentHash []byte, keywords []string, contributor string, epoch, submittedAt uint64) *types.PendingCommitment {
	keywordWeights := make([]*types.KeywordWeight, len(keywords))
	for i, kw := range keywords {
		keywordWeights[i] = &types.KeywordWeight{Keyword: kw, Weight: "1.0", ServeCount: 0, DenialCount: 0}
	}
	return types.NewPendingCommitment(orgID, contentHash, keywordWeights, contributor, epoch, submittedAt, types.MemoryType_MEMORY_TYPE_MEMORY)
}

func approveMemory(k *Keeper, ctx context.Context, orgID string, contentHash, encryptedBlob []byte, leader string, wrappedDekEnc []byte) error {
	return k.ApproveMemory(ctx, orgID, contentHash, encryptedBlob, leader, wrappedDekEnc, types.MemoryType_MEMORY_TYPE_MEMORY)
}

func TestSubmitCommitment(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	commitment := newPendingCommitment(
		"test-org",
		[]byte("12345678901234567890123456789012"),
		[]string{"keyword1", "keyword2"},
		"contributor-pubkey",
		1,
		100,
	)

	err := k.SubmitCommitment(ctx, commitment)
	if err != nil {
		t.Fatalf("SubmitCommitment failed: %v", err)
	}

	retrieved, err := k.GetPendingCommitment(ctx, "test-org", commitment.ContentHash)
	if err != nil {
		t.Fatalf("GetPendingCommitment failed: %v", err)
	}
	if retrieved.OrgID != commitment.OrgID {
		t.Errorf("OrgID mismatch: got %s, want %s", retrieved.OrgID, commitment.OrgID)
	}
}

func TestSubmitCommitment_InvalidHash(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	commitment := newPendingCommitment(
		"test-org",
		[]byte("invalid"),
		[]string{"keyword1"},
		"contributor-pubkey",
		1,
		100,
	)

	err := k.SubmitCommitment(ctx, commitment)
	if err != types.ErrInvalidContentHash {
		t.Fatalf("expected ErrInvalidContentHash, got: %v", err)
	}
}

func TestSubmitCommitment_DuplicateCommitment(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	contentHash := []byte("12345678901234567890123456789012")
	commitment := newPendingCommitment(
		"test-org",
		contentHash,
		[]string{"keyword1"},
		"contributor-pubkey",
		1,
		100,
	)

	err := k.SubmitCommitment(ctx, commitment)
	if err != nil {
		t.Fatalf("first SubmitCommitment failed: %v", err)
	}

	err = k.SubmitCommitment(ctx, commitment)
	if err != types.ErrCommitmentExists {
		t.Fatalf("expected ErrCommitmentExists, got: %v", err)
	}
}

func TestSubmitCommitment_OrgNotFound(t *testing.T) {
	k, mockOrg, _ := makeTestKeeper(t)
	ctx := context.Background()
	mockOrg.orgs["nonexistent-org"] = false

	commitment := newPendingCommitment(
		"nonexistent-org",
		[]byte("12345678901234567890123456789012"),
		[]string{"keyword1"},
		"contributor-pubkey",
		1,
		100,
	)

	err := k.SubmitCommitment(ctx, commitment)
	if err != types.ErrOrgNotFound {
		t.Fatalf("expected ErrOrgNotFound, got: %v", err)
	}
}

func TestSubmitCommitment_PersistsContributorAddress(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
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
	commitment.ContributorAddress = "wallet-addr-xyz"

	if err := k.SubmitCommitment(ctx, commitment); err != nil {
		t.Fatalf("SubmitCommitment failed: %v", err)
	}

	retrieved, err := k.GetPendingCommitment(ctx, "test-org", contentHash)
	if err != nil {
		t.Fatalf("GetPendingCommitment failed: %v", err)
	}
	if retrieved.ContributorAddress != "wallet-addr-xyz" {
		t.Errorf("ContributorAddress mismatch on pending: got %q, want %q", retrieved.ContributorAddress, "wallet-addr-xyz")
	}
}

func TestApproveMemory_CarriesContributorAddress(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
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
	commitment.ContributorAddress = "wallet-addr-xyz"
	if err := k.SubmitCommitment(ctx, commitment); err != nil {
		t.Fatalf("SubmitCommitment failed: %v", err)
	}

	encryptedBlob := []byte("encrypted blob data")
	if err := approveMemory(k, ctx, "test-org", contentHash, encryptedBlob, "leader-pubkey", nil); err != nil {
		t.Fatalf("ApproveMemory failed: %v", err)
	}

	approved, err := k.GetApprovedMemory(ctx, "test-org", contentHash)
	if err != nil {
		t.Fatalf("GetApprovedMemory failed: %v", err)
	}
	if approved.ContributorAddress != "wallet-addr-xyz" {
		t.Errorf("ContributorAddress mismatch on approved: got %q, want %q", approved.ContributorAddress, "wallet-addr-xyz")
	}
}

func TestApproveMemory(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
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
	_ = k.SubmitCommitment(ctx, commitment)

	encryptedBlob := []byte("encrypted blob data")
	err := approveMemory(k, ctx, "test-org", contentHash, encryptedBlob, "leader-pubkey", nil)
	if err != nil {
		t.Fatalf("ApproveMemory failed: %v", err)
	}

	_, err = k.GetPendingCommitment(ctx, "test-org", contentHash)
	if err != types.ErrCommitmentNotFound {
		t.Fatalf("expected pending to be deleted, got: %v", err)
	}

	approved, err := k.GetApprovedMemory(ctx, "test-org", contentHash)
	if err != nil {
		t.Fatalf("GetApprovedMemory failed: %v", err)
	}
	if approved.LastActiveEpoch != commitment.Epoch {
		t.Errorf("LastActiveEpoch mismatch: got %d, want %d", approved.LastActiveEpoch, commitment.Epoch)
	}

	count, _ := k.GetApprovedCount(ctx, "test-org")
	if count != 1 {
		t.Errorf("count mismatch: got %d, want 1", count)
	}
}

func TestApproveMemory_NotLeader(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	contentHash := []byte("12345678901234567890123456789012")
	commitment := newPendingCommitment(
		"test-org",
		contentHash,
		[]string{"keyword1"},
		"contributor-pubkey",
		1,
		100,
	)
	_ = k.SubmitCommitment(ctx, commitment)

	err := approveMemory(k, ctx, "test-org", contentHash, []byte("blob"), "not-leader", nil)
	if err != types.ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestGetActiveMemoryCountByOrg_ExcludesArchivedAndDenied(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	storeMemoryWithKeywords(t, k, ctx, "test-org", []byte("11111111111111111111111111111111"), types.MemoryState_MEMORY_STATE_COMMITTED, 0)
	storeMemoryWithKeywords(t, k, ctx, "test-org", []byte("22222222222222222222222222222222"), types.MemoryState_MEMORY_STATE_ARCHIVED, 0)
	storeMemoryWithKeywords(t, k, ctx, "test-org", []byte("33333333333333333333333333333333"), types.MemoryState_MEMORY_STATE_DENIED, 0)
	storeMemoryWithKeywords(t, k, ctx, "other-org", []byte("44444444444444444444444444444444"), types.MemoryState_MEMORY_STATE_COMMITTED, 0)

	count, err := k.GetActiveMemoryCountByOrg(ctx, "test-org")
	if err != nil {
		t.Fatalf("GetActiveMemoryCountByOrg failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("active count mismatch: got %d want 1", count)
	}
}

func TestApproveMemory_NoPending(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	contentHash := []byte("12345678901234567890123456789012")
	err := approveMemory(k, ctx, "test-org", contentHash, []byte("blob"), "leader-pubkey", nil)
	if err != types.ErrCommitmentNotFound {
		t.Fatalf("expected ErrCommitmentNotFound, got: %v", err)
	}
}

func TestApproveMemory_BlobTooLarge(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	contentHash := []byte("12345678901234567890123456789012")
	commitment := newPendingCommitment(
		"test-org",
		contentHash,
		[]string{"keyword1"},
		"contributor-pubkey",
		1,
		100,
	)
	_ = k.SubmitCommitment(ctx, commitment)

	params, _ := k.GetParams(ctx)
	k.SetParams(ctx, types.Params{
		MaxPendingPerOrg:       params.MaxPendingPerOrg,
		PendingRetentionEpochs: params.PendingRetentionEpochs,
		MaxBlobSizeBytes:       10,
		MaxKeywordsPerMemory:   params.MaxKeywordsPerMemory,
	})

	largeBlob := []byte("this blob is definitely too large")
	err := approveMemory(k, ctx, "test-org", contentHash, largeBlob, "leader-pubkey", nil)
	if err != types.ErrBlobTooLarge {
		t.Fatalf("expected ErrBlobTooLarge, got: %v", err)
	}
}

func TestGetAllPendingForOrg(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	contentHash1 := []byte("11111111111111111111111111111111")
	commitment1 := newPendingCommitment(
		"test-org",
		contentHash1,
		[]string{"keyword1"},
		"contributor-pubkey",
		1,
		100,
	)
	_ = k.SubmitCommitment(ctx, commitment1)

	contentHash2 := []byte("22222222222222222222222222222222")
	commitment2 := newPendingCommitment(
		"test-org",
		contentHash2,
		[]string{"keyword2"},
		"contributor-pubkey",
		1,
		101,
	)
	_ = k.SubmitCommitment(ctx, commitment2)

	pending, err := k.GetAllPendingForOrg(ctx, "test-org")
	if err != nil {
		t.Fatalf("GetAllPendingForOrg failed: %v", err)
	}
	if len(pending) != 2 {
		t.Fatalf("expected 2 pending, got: %d", len(pending))
	}
}

func TestGetApprovedCount(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	contentHash := []byte("12345678901234567890123456789012")
	commitment := newPendingCommitment(
		"test-org",
		contentHash,
		[]string{"keyword1"},
		"contributor-pubkey",
		1,
		100,
	)
	_ = k.SubmitCommitment(ctx, commitment)

	initialCount, _ := k.GetApprovedCount(ctx, "test-org")
	if initialCount != 0 {
		t.Fatalf("expected initial count 0, got: %d", initialCount)
	}

	_ = approveMemory(k, ctx, "test-org", contentHash, []byte("blob"), "leader-pubkey", nil)

	count, _ := k.GetApprovedCount(ctx, "test-org")
	if count != 1 {
		t.Fatalf("expected count 1, got: %d", count)
	}
}

func TestComputeEpochMerkleRoot(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	hash1 := []byte("11111111111111111111111111111111")
	commitment1 := newPendingCommitment(
		"test-org",
		hash1,
		[]string{"keyword1"},
		"contributor-pubkey",
		1,
		100,
	)
	_ = k.SubmitCommitment(ctx, commitment1)
	_ = approveMemory(k, ctx, "test-org", hash1, []byte("blob1"), "leader-pubkey", nil)

	hash2 := []byte("22222222222222222222222222222222")
	commitment2 := newPendingCommitment(
		"test-org",
		hash2,
		[]string{"keyword2"},
		"contributor-pubkey",
		1,
		101,
	)
	_ = k.SubmitCommitment(ctx, commitment2)
	_ = approveMemory(k, ctx, "test-org", hash2, []byte("blob2"), "leader-pubkey", nil)

	err := k.ComputeAndStoreEpochMerkleRoot(ctx, "test-org", 1)
	if err != nil {
		t.Fatalf("ComputeAndStoreEpochMerkleRoot failed: %v", err)
	}

	merkle, err := k.GetEpochMerkleRoot(ctx, "test-org", 1)
	if err != nil {
		t.Fatalf("GetEpochMerkleRoot failed: %v", err)
	}
	if merkle.MemoryCount != 2 {
		t.Errorf("MemoryCount mismatch: got %d, want 2", merkle.MemoryCount)
	}
	if len(merkle.MerkleRoot) != 32 {
		t.Errorf("MerkleRoot length mismatch: got %d, want 32", len(merkle.MerkleRoot))
	}
}

func TestComputeEpochMerkleRoot_Empty(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	err := k.ComputeAndStoreEpochMerkleRoot(ctx, "test-org", 99)
	if err != nil {
		t.Fatalf("ComputeAndStoreEpochMerkleRoot failed: %v", err)
	}

	merkle, err := k.GetEpochMerkleRoot(ctx, "test-org", 99)
	if err != nil {
		t.Fatalf("GetEpochMerkleRoot failed: %v", err)
	}
	if merkle.MemoryCount != 0 {
		t.Errorf("MemoryCount mismatch: got %d, want 0", merkle.MemoryCount)
	}
}

func TestInitGenesisExportGenesis(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	contentHash := []byte("12345678901234567890123456789012")
	commitment := newPendingCommitment(
		"test-org",
		contentHash,
		[]string{"keyword1"},
		"contributor-pubkey",
		1,
		100,
	)
	_ = k.SubmitCommitment(ctx, commitment)
	_ = approveMemory(k, ctx, "test-org", contentHash, []byte("blob"), "leader-pubkey", nil)

	genesis, err := k.ExportGenesis(ctx)
	if err != nil {
		t.Fatalf("ExportGenesis failed: %v", err)
	}

	k2, _, _ := makeTestKeeper(t)
	ctx2 := context.Background()
	err = k2.InitGenesis(ctx2, genesis)
	if err != nil {
		t.Fatalf("InitGenesis failed: %v", err)
	}

	genesis2, err := k2.ExportGenesis(ctx2)
	if err != nil {
		t.Fatalf("ExportGenesis after InitGenesis failed: %v", err)
	}

	if len(genesis.PendingCommitments) != len(genesis2.PendingCommitments) {
		t.Errorf("pending count mismatch: got %d, want %d", len(genesis2.PendingCommitments), len(genesis.PendingCommitments))
	}
	if len(genesis.MemoryCommitments) != len(genesis2.MemoryCommitments) {
		t.Errorf("approved count mismatch: got %d, want %d", len(genesis2.MemoryCommitments), len(genesis.MemoryCommitments))
	}
}

func TestContentAddressedDedup(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	contentHash := []byte("12345678901234567890123456789012")
	commitment := newPendingCommitment(
		"test-org",
		contentHash,
		[]string{"keyword1"},
		"contributor-pubkey",
		1,
		100,
	)
	_ = k.SubmitCommitment(ctx, commitment)
	_ = approveMemory(k, ctx, "test-org", contentHash, []byte("blob"), "leader-pubkey", nil)

	commitment2 := newPendingCommitment(
		"test-org",
		contentHash,
		[]string{"keyword2"},
		"contributor-pubkey2",
		2,
		101,
	)
	_ = k.SubmitCommitment(ctx, commitment2)

	err := approveMemory(k, ctx, "test-org", contentHash, []byte("new-blob"), "leader-pubkey", nil)
	if err != types.ErrMemoryExists {
		t.Fatalf("expected ErrMemoryExists, got: %v", err)
	}
}

func TestStoreMemoryWithParameterizedHelpers(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	contentHash := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	cid := storeMemoryWithKeywords(
		t,
		k,
		ctx,
		"test-org",
		contentHash,
		types.MemoryState_MEMORY_STATE_COMMITTED,
		2,
		withMemoryServeTotal(7),
		withMemoryDenialTotal(3),
		withKeywords(
			&types.KeywordWeight{Keyword: "kw1", Weight: "0.50"},
			&types.KeywordWeight{Keyword: "kw2", Weight: "0.25"},
		),
		withKeywordWeight("kw1", "0.75"),
	)

	mock := attachMockServeKeeper(k, "test-org", cid,
		withServeCount(5, 2),
		withDenialCount(5, 1),
		withMatchedKeywords(5, "kw1"),
	)

	matches, err := mock.GetMatchedKeywordsForEpoch(ctx, "test-org", cid, 5)
	if err != nil {
		t.Fatalf("GetMatchedKeywordsForEpoch failed: %v", err)
	}
	if !matches["kw1"] {
		t.Fatalf("expected kw1 to be matched")
	}

	memory, err := k.GetApprovedMemory(ctx, "test-org", contentHash)
	if err != nil {
		t.Fatalf("GetApprovedMemory failed: %v", err)
	}
	if memory.ServeCountTotal != 7 {
		t.Fatalf("ServeCountTotal mismatch: got %d, want 7", memory.ServeCountTotal)
	}
	if memory.DenialCountTotal != 3 {
		t.Fatalf("DenialCountTotal mismatch: got %d, want 3", memory.DenialCountTotal)
	}
}
