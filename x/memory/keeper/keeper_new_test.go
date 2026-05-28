package keeper

import (
	"context"
	"crypto/sha256"
	"testing"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/wevibe-network/wevibe-chain/testutil/keeper"
	"github.com/wevibe-network/wevibe-chain/x/memory/types"
)

func TestReportMemory_StoresTriplet(t *testing.T) {
	storeKey := storetypes.NewKVStoreKey("memory")
	storeService, _ := keeper.NewTestStoreService(t, storeKey)
	logger := keeper.NewTestLogger()
	sdk.DefaultBondDenom = "uvibe"

	mockOrg := &mockOrgKeeperServer{
		orgs:    map[string]bool{"test-org": true},
		leaders: map[string]string{"test-org": "wevibe1leader000000000000000000000000000000000"},
	}
	mockRep := &mockReputationKeeper{}

	k := NewKeeper(storeService, logger, "gov", mockOrg, mockRep)
	ctx := context.Background()

	contentHash := []byte("12345678901234567890123456789012")
	commitment := newPendingCommitment(
		"test-org",
		contentHash,
		[]string{"keyword1"},
		"wevibe1contrib",
		1,
		100,
	)
	_ = k.SubmitCommitment(ctx, commitment)

	_ = k.ApproveMemory(ctx, "test-org", contentHash, []byte("encrypted"), "wevibe1leader000000000000000000000000000000000", []byte("wrappedDek"), types.MemoryType_MEMORY_TYPE_CORRECT_IMPLEMENTATION)

	plaintext := []byte("revealed plaintext content")
	ciphertext := []byte("encrypted blob")
	capsule := []byte("umbral capsule")
	hash := sha256.Sum256(plaintext)

	err := k.ReportMemory(
		ctx, "test-org", contentHash,
		"wevibe1contrib", []string{"wevibe1mod1"}, []string{"wevibe1mod2"},
		"wevibe1reporter", "wevibe1leader000000000000000000000000000000000", "memory was incorrect",
		plaintext, ciphertext, capsule, hash[:], false,
		100,
	)
	if err != nil {
		t.Fatalf("ReportMemory failed: %v", err)
	}

	report, err := k.GetUpheldReport(ctx, "test-org", contentHash)
	if err != nil {
		t.Fatalf("GetUpheldReport failed: %v", err)
	}
	if report == nil {
		t.Fatal("GetUpheldReport returned nil")
	}
	if string(report.Plaintext) != string(plaintext) {
		t.Errorf("Plaintext mismatch")
	}
	if string(report.Ciphertext) != string(ciphertext) {
		t.Errorf("Ciphertext mismatch")
	}
	if string(report.Capsule) != string(capsule) {
		t.Errorf("Capsule mismatch")
	}
}

func TestReportMemory_OversizedStoresHashOnly(t *testing.T) {
	storeKey := storetypes.NewKVStoreKey("memory")
	storeService, _ := keeper.NewTestStoreService(t, storeKey)
	logger := keeper.NewTestLogger()
	sdk.DefaultBondDenom = "uvibe"

	mockOrg := &mockOrgKeeperServer{
		orgs:    map[string]bool{"test-org": true},
		leaders: map[string]string{"test-org": "wevibe1leader000000000000000000000000000000000"},
	}
	mockRep := &mockReputationKeeper{}

	k := NewKeeper(storeService, logger, "gov", mockOrg, mockRep)
	ctx := context.Background()

	contentHash := []byte("12345678901234567890123456789012")
	commitment := newPendingCommitment(
		"test-org",
		contentHash,
		[]string{"keyword1"},
		"wevibe1contrib",
		1,
		100,
	)
	_ = k.SubmitCommitment(ctx, commitment)

	_ = k.ApproveMemory(ctx, "test-org", contentHash, []byte("encrypted"), "wevibe1leader000000000000000000000000000000000", []byte("wrappedDek"), types.MemoryType_MEMORY_TYPE_CORRECT_IMPLEMENTATION)

	plaintext := make([]byte, 5000)
	hash := sha256.Sum256(plaintext)

	err := k.ReportMemory(
		ctx, "test-org", contentHash,
		"wevibe1contrib", []string{"wevibe1mod1"}, []string{"wevibe1mod2"},
		"wevibe1reporter", "wevibe1leader000000000000000000000000000000000", "too big",
		nil, nil, nil, hash[:], true,
		100,
	)
	if err != nil {
		t.Fatalf("ReportMemory failed: %v", err)
	}

	report, err := k.GetUpheldReport(ctx, "test-org", contentHash)
	if err != nil {
		t.Fatalf("GetUpheldReport failed: %v", err)
	}
	if report == nil {
		t.Fatal("GetUpheldReport returned nil")
	}
	if len(report.Plaintext) != 0 {
		t.Errorf("Plaintext should be empty for oversized")
	}
	if !report.PlaintextOversized {
		t.Error("PlaintextOversized should be true")
	}
}

func TestReportMemory_RejectsOversizedPlaintextWithoutFlag(t *testing.T) {
	storeKey := storetypes.NewKVStoreKey("memory")
	storeService, _ := keeper.NewTestStoreService(t, storeKey)
	logger := keeper.NewTestLogger()
	sdk.DefaultBondDenom = "uvibe"

	mockOrg := &mockOrgKeeperServer{
		orgs:    map[string]bool{"test-org": true},
		leaders: map[string]string{"test-org": "wevibe1leader000000000000000000000000000000000"},
	}
	mockRep := &mockReputationKeeper{}

	k := NewKeeper(storeService, logger, "gov", mockOrg, mockRep)
	ctx := context.Background()

	contentHash := []byte("12345678901234567890123456789012")
	commitment := newPendingCommitment(
		"test-org",
		contentHash,
		[]string{"keyword1"},
		"wevibe1contrib",
		1,
		100,
	)
	_ = k.SubmitCommitment(ctx, commitment)

	_ = k.ApproveMemory(ctx, "test-org", contentHash, []byte("encrypted"), "wevibe1leader000000000000000000000000000000000", []byte("wrappedDek"), types.MemoryType_MEMORY_TYPE_CORRECT_IMPLEMENTATION)

	plaintext := make([]byte, 5000)
	hash := sha256.Sum256(plaintext)

	err := k.ReportMemory(
		ctx, "test-org", contentHash,
		"wevibe1contrib", []string{"wevibe1mod1"}, []string{"wevibe1mod2"},
		"wevibe1reporter", "wevibe1leader000000000000000000000000000000000", "tries to lie",
		plaintext, []byte("ct"), []byte("cap"), hash[:], false,
		100,
	)
	if err == nil {
		t.Fatal("ReportMemory should fail when plaintext > 4KB but flag is false")
	}
}

func TestEndToEnd_MemoryLifecycle_WithSocialGraph(t *testing.T) {
	storeKey := storetypes.NewKVStoreKey("memory")
	storeService, _ := keeper.NewTestStoreService(t, storeKey)
	logger := keeper.NewTestLogger()
	sdk.DefaultBondDenom = "uvibe"

	mockOrg := &mockOrgKeeperServer{
		orgs:    map[string]bool{"test-org": true},
		leaders: map[string]string{"test-org": "wevibe1leader000000000000000000000000000000000"},
	}
	mockRep := &mockReputationKeeper{}

	k := NewKeeper(storeService, logger, "gov", mockOrg, mockRep)
	ctx := context.Background()

	contentHash := []byte("12345678901234567890123456789012")
	commitment := newPendingCommitment(
		"test-org",
		contentHash,
		[]string{"keyword1", "keyword2"},
		"wevibe1contrib",
		1,
		100,
	)
	_ = k.SubmitCommitment(ctx, commitment)

	err := k.ApproveMemory(ctx, "test-org", contentHash, []byte("encrypted"), "wevibe1leader000000000000000000000000000000000", []byte("wrappedDek"), types.MemoryType_MEMORY_TYPE_CORRECT_IMPLEMENTATION)
	if err != nil {
		t.Fatalf("ApproveMemory failed: %v", err)
	}

	_, err = k.GetApprovedMemory(ctx, "test-org", contentHash)
	if err != nil {
		t.Fatalf("GetApprovedMemory failed: %v", err)
	}

	plaintext := []byte("revealed plaintext")
	ciphertext := []byte("the ciphertext")
	capsule := []byte("the capsule")
	hash := sha256.Sum256(plaintext)

	err = k.ReportMemory(
		ctx, "test-org", contentHash,
		"wevibe1contrib", []string{"wevibe1mod1", "wevibe1mod2"}, []string{"wevibe1mod3"},
		"wevibe1reporter", "wevibe1leader000000000000000000000000000000000", "incorrect content",
		plaintext, ciphertext, capsule, hash[:], false,
		100,
	)
	if err != nil {
		t.Fatalf("ReportMemory failed: %v", err)
	}

	report, err := k.GetUpheldReport(ctx, "test-org", contentHash)
	if err != nil {
		t.Fatalf("GetUpheldReport failed: %v", err)
	}
	if report == nil {
		t.Fatal("GetUpheldReport returned nil after upheld report")
	}
	if string(report.Plaintext) != string(plaintext) {
		t.Errorf("Upheld report plaintext mismatch")
	}

	computed := sha256.Sum256(report.Plaintext)
	if string(computed[:]) != string(report.PlaintextHash) {
		t.Errorf("PlaintextHash verification failed")
	}
}
