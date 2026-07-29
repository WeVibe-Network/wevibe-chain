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
		return &orgtypes.MemberRecord{OrgID: orgID, Pubkey: memberPubkey, Role: "leader", X25519Pubkey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"}, nil
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
	return types.NewPendingCommitment(orgID, contentHash, keywords, contributor, epoch, submittedAt, types.MemoryType_MEMORY_TYPE_MEMORY)
}

func approveMemory(k *Keeper, ctx context.Context, orgID string, contentHash, encryptedBlob []byte, leader string, wrappedDekEnc []byte) error {
	return k.ApproveMemory(ctx, orgID, contentHash, encryptedBlob, leader, wrappedDekEnc, nil, nil, nil, nil, nil, types.MemoryType_MEMORY_TYPE_MEMORY)
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

func TestApproveMemory_CarriesMcVersion(t *testing.T) {
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
	commitment.McVersion = 1
	if err := k.SubmitCommitment(ctx, commitment); err != nil {
		t.Fatalf("SubmitCommitment failed: %v", err)
	}

	// fields survive on the pending record
	pending, err := k.GetPendingCommitment(ctx, "test-org", contentHash)
	if err != nil {
		t.Fatalf("GetPendingCommitment failed: %v", err)
	}
	if pending.McVersion != 1 {
		t.Errorf("mc_version lost on pending: got %d, want 1", pending.McVersion)
	}

	encryptedBlob := []byte("encrypted blob data")
	if err := approveMemory(k, ctx, "test-org", contentHash, encryptedBlob, "leader-pubkey", nil); err != nil {
		t.Fatalf("ApproveMemory failed: %v", err)
	}

	// fields carry through to the committed record
	approved, err := k.GetApprovedMemory(ctx, "test-org", contentHash)
	if err != nil {
		t.Fatalf("GetApprovedMemory failed: %v", err)
	}
	if approved.McVersion != 1 {
		t.Errorf("McVersion mismatch on approved: got %d, want %d", approved.McVersion, 1)
	}
}

func TestApproveMemory_CarriesProvenance(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	contentHash := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	attestationSessionHash := []byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	commitment := newPendingCommitment(
		"test-org",
		contentHash,
		[]string{"keyword1", "keyword2"},
		"contributor-pubkey",
		1,
		100,
	)
	commitment.ProducerModelId = "openai/gpt-5.2"
	commitment.AttestationSessionHash = attestationSessionHash
	if err := k.SubmitCommitment(ctx, commitment); err != nil {
		t.Fatalf("SubmitCommitment failed: %v", err)
	}

	pending, err := k.GetPendingCommitment(ctx, "test-org", contentHash)
	if err != nil {
		t.Fatalf("GetPendingCommitment failed: %v", err)
	}
	if pending.ProducerModelId != commitment.ProducerModelId {
		t.Fatalf("pending producer_model_id mismatch: got %q, want %q", pending.ProducerModelId, commitment.ProducerModelId)
	}
	if string(pending.AttestationSessionHash) != string(attestationSessionHash) {
		t.Fatalf("pending attestation_session_hash mismatch: got %x, want %x", pending.AttestationSessionHash, attestationSessionHash)
	}

	encryptedBlob := []byte("encrypted blob data")
	if err := approveMemory(k, ctx, "test-org", contentHash, encryptedBlob, "leader-pubkey", nil); err != nil {
		t.Fatalf("ApproveMemory failed: %v", err)
	}

	approved, err := k.GetApprovedMemory(ctx, "test-org", contentHash)
	if err != nil {
		t.Fatalf("GetApprovedMemory failed: %v", err)
	}
	if approved.ProducerModelId != commitment.ProducerModelId {
		t.Fatalf("approved producer_model_id mismatch: got %q, want %q", approved.ProducerModelId, commitment.ProducerModelId)
	}
	if string(approved.AttestationSessionHash) != string(attestationSessionHash) {
		t.Fatalf("approved attestation_session_hash mismatch: got %x, want %x", approved.AttestationSessionHash, attestationSessionHash)
	}
	if approved.ProvenanceStatus() != types.ProvenanceSessionReferenced {
		t.Fatalf("approved provenance status mismatch: got %q, want %q", approved.ProvenanceStatus(), types.ProvenanceSessionReferenced)
	}
}

func TestApproveMemory_LegacyProvenanceUnattested(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	contentHash := []byte("cccccccccccccccccccccccccccccccc")
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

	if err := approveMemory(k, ctx, "test-org", contentHash, []byte("encrypted blob data"), "leader-pubkey", nil); err != nil {
		t.Fatalf("ApproveMemory failed: %v", err)
	}

	approved, err := k.GetApprovedMemory(ctx, "test-org", contentHash)
	if err != nil {
		t.Fatalf("GetApprovedMemory failed: %v", err)
	}
	if approved.ProducerModelId != "" {
		t.Fatalf("expected empty producer_model_id for legacy memory, got %q", approved.ProducerModelId)
	}
	if len(approved.AttestationSessionHash) != 0 {
		t.Fatalf("expected empty attestation_session_hash for legacy memory, got %x", approved.AttestationSessionHash)
	}
	if approved.ProvenanceStatus() != types.ProvenanceUnattested {
		t.Fatalf("legacy provenance status mismatch: got %q, want %q", approved.ProvenanceStatus(), types.ProvenanceUnattested)
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

func TestInitGenesisExportGenesis_CarriesProvenance(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	pendingHash := []byte("dddddddddddddddddddddddddddddddd")
	pendingCommitment := newPendingCommitment(
		"test-org",
		pendingHash,
		[]string{"pending-kw"},
		"contributor-pubkey",
		1,
		100,
	)
	pendingCommitment.ProducerModelId = "model-pending"
	pendingCommitment.AttestationSessionHash = []byte("eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee")
	if err := k.SubmitCommitment(ctx, pendingCommitment); err != nil {
		t.Fatalf("SubmitCommitment (pending) failed: %v", err)
	}

	approvedHash := []byte("ffffffffffffffffffffffffffffffff")
	approvedCommitment := newPendingCommitment(
		"test-org",
		approvedHash,
		[]string{"approved-kw"},
		"contributor-pubkey",
		2,
		101,
	)
	approvedCommitment.ProducerModelId = "model-approved"
	approvedCommitment.AttestationSessionHash = []byte("99999999999999999999999999999999")
	if err := k.SubmitCommitment(ctx, approvedCommitment); err != nil {
		t.Fatalf("SubmitCommitment (approved path) failed: %v", err)
	}
	if err := approveMemory(k, ctx, "test-org", approvedHash, []byte("blob"), "leader-pubkey", nil); err != nil {
		t.Fatalf("ApproveMemory failed: %v", err)
	}

	genesis, err := k.ExportGenesis(ctx)
	if err != nil {
		t.Fatalf("ExportGenesis failed: %v", err)
	}

	k2, _, _ := makeTestKeeper(t)
	ctx2 := context.Background()
	if err := k2.InitGenesis(ctx2, genesis); err != nil {
		t.Fatalf("InitGenesis failed: %v", err)
	}

	genesis2, err := k2.ExportGenesis(ctx2)
	if err != nil {
		t.Fatalf("ExportGenesis after InitGenesis failed: %v", err)
	}

	pendingHashHex := types.ContentHashToHex(pendingHash)
	pendingFound := false
	for _, pc := range genesis2.PendingCommitments {
		if types.ContentHashToHex(pc.ContentHash) != pendingHashHex {
			continue
		}
		pendingFound = true
		if pc.ProducerModelId != pendingCommitment.ProducerModelId {
			t.Fatalf("pending producer_model_id mismatch after genesis round-trip: got %q, want %q", pc.ProducerModelId, pendingCommitment.ProducerModelId)
		}
		if string(pc.AttestationSessionHash) != string(pendingCommitment.AttestationSessionHash) {
			t.Fatalf("pending attestation_session_hash mismatch after genesis round-trip: got %x, want %x", pc.AttestationSessionHash, pendingCommitment.AttestationSessionHash)
		}
		if types.DeriveProvenanceStatus(pc.ProducerModelId, pc.AttestationSessionHash) != types.ProvenanceSessionReferenced {
			t.Fatalf("pending provenance status mismatch after genesis round-trip: got %q, want %q", types.DeriveProvenanceStatus(pc.ProducerModelId, pc.AttestationSessionHash), types.ProvenanceSessionReferenced)
		}
		break
	}
	if !pendingFound {
		t.Fatalf("pending commitment not found after genesis round-trip")
	}

	approvedHashHex := types.ContentHashToHex(approvedHash)
	approvedFound := false
	for _, mc := range genesis2.MemoryCommitments {
		if types.ContentHashToHex(mc.ContentHash) != approvedHashHex {
			continue
		}
		approvedFound = true
		if mc.ProducerModelId != approvedCommitment.ProducerModelId {
			t.Fatalf("approved producer_model_id mismatch after genesis round-trip: got %q, want %q", mc.ProducerModelId, approvedCommitment.ProducerModelId)
		}
		if string(mc.AttestationSessionHash) != string(approvedCommitment.AttestationSessionHash) {
			t.Fatalf("approved attestation_session_hash mismatch after genesis round-trip: got %x, want %x", mc.AttestationSessionHash, approvedCommitment.AttestationSessionHash)
		}
		if mc.ProvenanceStatus() != types.ProvenanceSessionReferenced {
			t.Fatalf("approved provenance status mismatch after genesis round-trip: got %q, want %q", mc.ProvenanceStatus(), types.ProvenanceSessionReferenced)
		}
		break
	}
	if !approvedFound {
		t.Fatalf("approved commitment not found after genesis round-trip")
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
