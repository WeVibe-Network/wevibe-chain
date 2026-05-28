package keeper

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/wevibe-network/wevibe-chain/testutil/keeper"
	"github.com/wevibe-network/wevibe-chain/x/memory/types"
	orgtypes "github.com/wevibe-network/wevibe-chain/x/org/types"
)

type mockOrgKeeperServer struct {
	orgs    map[string]bool
	leaders map[string]string
}

func (m *mockOrgKeeperServer) HasOrg(ctx context.Context, orgID string) (bool, error) {
	return m.orgs[orgID], nil
}

func (m *mockOrgKeeperServer) IsLeader(ctx context.Context, orgID string, memberPubkey string) (bool, error) {
	return m.leaders[orgID] == memberPubkey, nil
}

func (m *mockOrgKeeperServer) IsModerator(ctx context.Context, orgID string, memberPubkey string) (bool, error) {
	return false, nil
}

func (m *mockOrgKeeperServer) GetOrgConfig(ctx context.Context, orgID string) (*orgtypes.OrgConfig, error) {
	return &orgtypes.OrgConfig{OrgID: orgID}, nil
}

func makeTestMsgServer(t *testing.T) (types.MsgServer, *Keeper, *mockOrgKeeperServer, context.Context) {
	storeKey := storetypes.NewKVStoreKey("memory")
	storeService, cms := keeper.NewTestStoreService(t, storeKey)
	logger := keeper.NewTestLogger()
	sdk.DefaultBondDenom = "uvibe"
	mockOrg := &mockOrgKeeperServer{
		orgs:    map[string]bool{"test-org": true},
		leaders: map[string]string{"test-org": "leader-pubkey"},
	}
	k := NewKeeper(storeService, logger, "gov", mockOrg, &mockReputationKeeper{})
	sdkCtx := sdk.NewContext(cms, tmproto.Header{Height: 1, Time: time.Now().UTC()}, false, logger)
	return NewMsgServerImpl(k), k, mockOrg, sdk.WrapSDKContext(sdkCtx)
}

func keywordsToKeywordWeights(keywords []string) []*types.KeywordWeight {
	result := make([]*types.KeywordWeight, len(keywords))
	for i, kw := range keywords {
		result[i] = &types.KeywordWeight{Keyword: kw, Weight: "1.0", ServeCount: 0, DenialCount: 0}
	}
	return result
}

func makeApproveMemoryFixture(memoryType types.MemoryType) (*types.MsgSubmitCommitment, *types.MsgApproveMemory, []byte, []byte, []byte, []byte, []byte) {
	seed := []byte("12345678901234567890123456789012")
	privateKey := ed25519.NewKeyFromSeed(seed)
	contributorPubkeyHex := hex.EncodeToString(privateKey.Public().(ed25519.PublicKey))

	encryptedBlob := []byte("encrypted blob")
	wrappedDekEnc := []byte("wrapped dek")
	plaintextHash := sha256.Sum256([]byte("plaintext"))
	salt := []byte("test-salt")
	ciphertextHash := sha256.Sum256(encryptedBlob)
	wrappedDekHash := sha256.Sum256(wrappedDekEnc)

	submissionHasher := sha256.New()
	_, _ = submissionHasher.Write(encryptedBlob)
	_, _ = submissionHasher.Write(wrappedDekEnc)
	submissionHash := submissionHasher.Sum(nil)

	canonicalBody := buildSubmitMemoryCanonicalBody(
		ciphertextHash[:],
		contributorPubkeyHex,
		0,
		memoryType,
		"test-org",
		plaintextHash[:],
		salt,
		submissionHash,
		wrappedDekHash[:],
	)
	contributorSig := ed25519.Sign(privateKey, canonicalBody)

	submitMsg := &types.MsgSubmitCommitment{
		Signer:        "signer",
		OrgId:         "test-org",
		ContentHash:   submissionHash,
		Keywords:      keywordsToKeywordWeights([]string{"kw1"}),
		ContributorId: contributorPubkeyHex,
		MemoryType:    memoryType,
	}

	approveMsg := &types.MsgApproveMemory{
		Signer:           "leader-pubkey",
		OrgId:            "test-org",
		ContentHash:      submissionHash,
		EncryptedBlob:    encryptedBlob,
		CommittingLeader: "leader-pubkey",
		WrappedDekEnc:    wrappedDekEnc,
		MemoryType:       memoryType,
		PlaintextHash:    plaintextHash[:],
		Salt:             salt,
		CiphertextHash:   ciphertextHash[:],
		ContributorSig:   contributorSig,
	}

	return submitMsg, approveMsg, plaintextHash[:], salt, ciphertextHash[:], wrappedDekHash[:], contributorSig
}

func TestMsgSubmitCommitment_ValidateBasic(t *testing.T) {
	srv, _, _, _ := makeTestMsgServer(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		msg     *types.MsgSubmitCommitment
		wantErr bool
	}{
		{
			name: "empty signer",
			msg: &types.MsgSubmitCommitment{
				Signer:        "",
				OrgId:         "test-org",
				ContentHash:   []byte("12345678901234567890123456789012"),
				ContributorId: "contributor",
			},
			wantErr: true,
		},
		{
			name: "empty org",
			msg: &types.MsgSubmitCommitment{
				Signer:        "signer",
				OrgId:         "",
				ContentHash:   []byte("12345678901234567890123456789012"),
				ContributorId: "contributor",
			},
			wantErr: true,
		},
		{
			name: "bad hash",
			msg: &types.MsgSubmitCommitment{
				Signer:        "signer",
				OrgId:         "test-org",
				ContentHash:   []byte("bad"),
				ContributorId: "contributor",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := srv.SubmitCommitment(ctx, tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("SubmitCommitment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMsgSubmitCommitment_Success(t *testing.T) {
	srv, _, _, ctx := makeTestMsgServer(t)

	msg := &types.MsgSubmitCommitment{
		Signer:        "signer",
		OrgId:         "test-org",
		ContentHash:   []byte("12345678901234567890123456789012"),
		Keywords:      keywordsToKeywordWeights([]string{"kw1"}),
		ContributorId: "contributor",
		MemoryType:    types.MemoryType_MEMORY_TYPE_CORRECT_IMPLEMENTATION,
	}

	resp, err := srv.SubmitCommitment(ctx, msg)
	if err != nil {
		t.Fatalf("SubmitCommitment failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestMsgSubmitCommitment_InvalidMemoryType(t *testing.T) {
	srv, _, _, ctx := makeTestMsgServer(t)

	msg := &types.MsgSubmitCommitment{
		Signer:        "signer",
		OrgId:         "test-org",
		ContentHash:   []byte("12345678901234567890123456789012"),
		Keywords:      keywordsToKeywordWeights([]string{"kw1"}),
		ContributorId: "contributor",
		MemoryType:    types.MemoryType_MEMORY_TYPE_UNSPECIFIED,
	}

	_, err := srv.SubmitCommitment(ctx, msg)
	if err != types.ErrInvalidMemoryType {
		t.Fatalf("expected ErrInvalidMemoryType, got: %v", err)
	}
}

func TestMsgApproveMemory_Success(t *testing.T) {
	srv, k, _, ctx := makeTestMsgServer(t)
	submitMsg, approveMsg, plaintextHash, salt, ciphertextHash, wrappedDekHash, contributorSig := makeApproveMemoryFixture(types.MemoryType_MEMORY_TYPE_NEGATIVE_SIGNAL)
	_, _ = srv.SubmitCommitment(ctx, submitMsg)

	resp, err := srv.ApproveMemory(ctx, approveMsg)
	if err != nil {
		t.Fatalf("ApproveMemory failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	approved, err := k.GetApprovedMemory(ctx, "test-org", approveMsg.ContentHash)
	if err != nil {
		t.Fatalf("GetApprovedMemory failed: %v", err)
	}
	if approved.MemoryType != types.MemoryType_MEMORY_TYPE_NEGATIVE_SIGNAL {
		t.Fatalf("memory type mismatch: got %v, want %v", approved.MemoryType, types.MemoryType_MEMORY_TYPE_NEGATIVE_SIGNAL)
	}

	store := k.getStore(ctx)
	bz, err := store.Get(approvedKey("test-org", approveMsg.ContentHash))
	if err != nil {
		t.Fatalf("Get approved memory from store failed: %v", err)
	}

	var storedApproved types.StoredMemoryCommitment
	if err := proto.Unmarshal(bz, &storedApproved); err != nil {
		t.Fatalf("unmarshal approved memory failed: %v", err)
	}
	if string(storedApproved.PlaintextHash) != string(plaintextHash) {
		t.Fatalf("plaintext hash mismatch")
	}
	if string(storedApproved.Salt) != string(salt) {
		t.Fatalf("salt mismatch")
	}
	if string(storedApproved.CiphertextHash) != string(ciphertextHash) {
		t.Fatalf("ciphertext hash mismatch")
	}
	if string(storedApproved.WrappedDekHash) != string(wrappedDekHash) {
		t.Fatalf("wrapped DEK hash mismatch")
	}
	if string(storedApproved.ContributorSig) != string(contributorSig) {
		t.Fatalf("contributor signature mismatch")
	}
}

func TestMsgApproveMemory_Unauthorized(t *testing.T) {
	srv, _, _, ctx := makeTestMsgServer(t)
	submitMsg, approveMsg, _, _, _, _, _ := makeApproveMemoryFixture(types.MemoryType_MEMORY_TYPE_CORRECT_IMPLEMENTATION)
	_, _ = srv.SubmitCommitment(ctx, submitMsg)
	approveMsg.Signer = "not-leader"
	approveMsg.CommittingLeader = "not-leader"

	_, err := srv.ApproveMemory(ctx, approveMsg)
	if err != types.ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestMsgApproveMemory_InvalidMemoryType(t *testing.T) {
	srv, _, _, ctx := makeTestMsgServer(t)

	submitMsg := &types.MsgSubmitCommitment{
		Signer:        "signer",
		OrgId:         "test-org",
		ContentHash:   []byte("12345678901234567890123456789012"),
		Keywords:      keywordsToKeywordWeights([]string{"kw1"}),
		ContributorId: "contributor",
		MemoryType:    types.MemoryType_MEMORY_TYPE_CORRECT_IMPLEMENTATION,
	}
	_, _ = srv.SubmitCommitment(ctx, submitMsg)

	approveMsg := &types.MsgApproveMemory{
		Signer:        "leader-pubkey",
		OrgId:         "test-org",
		ContentHash:   []byte("12345678901234567890123456789012"),
		EncryptedBlob: []byte("encrypted blob"),
		MemoryType:    types.MemoryType_MEMORY_TYPE_UNSPECIFIED,
	}

	_, err := srv.ApproveMemory(ctx, approveMsg)
	if err != types.ErrInvalidMemoryType {
		t.Fatalf("expected ErrInvalidMemoryType, got: %v", err)
	}
}

func TestMsgApproveMemory_VerificationFailureReturnsSuccessAndKeepsPending(t *testing.T) {
	srv, k, _, ctx := makeTestMsgServer(t)
	submitMsg, approveMsg, _, _, _, _, _ := makeApproveMemoryFixture(types.MemoryType_MEMORY_TYPE_CORRECT_IMPLEMENTATION)
	_, _ = srv.SubmitCommitment(ctx, submitMsg)

	approveMsg.ContributorSig = []byte("invalid-signature")

	resp, err := srv.ApproveMemory(ctx, approveMsg)
	if err != nil {
		t.Fatalf("expected no error on verification failure, got: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	_, err = k.GetApprovedMemory(ctx, "test-org", approveMsg.ContentHash)
	if !errors.Is(err, types.ErrMemoryNotFound) {
		t.Fatalf("expected ErrMemoryNotFound, got: %v", err)
	}

	_, err = k.GetPendingCommitment(ctx, "test-org", approveMsg.ContentHash)
	if err != nil {
		t.Fatalf("expected pending commitment to remain, got: %v", err)
	}
}
