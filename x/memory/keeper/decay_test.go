package keeper

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/wevibe-network/wevibe-chain/x/memory/types"
)

func round4(value float64) float64 {
	return math.Round(value*10000) / 10000
}

func weightForKeyword(t *testing.T, memory *types.MemoryCommitment, keyword string) float64 {
	t.Helper()
	for _, kw := range memory.Keywords {
		if kw.Keyword == keyword {
			value, _ := parseWeight(kw.Weight).value.Float64()
			return value
		}
	}
	t.Fatalf("keyword %q not found", keyword)
	return 0
}

func newDecayMemory(epoch uint64, keywords ...*types.KeywordWeight) *types.MemoryCommitment {
	return &types.MemoryCommitment{
		OrgID:             defaultOrgID,
		ContentHash:       []byte("dddddddddddddddddddddddddddddddd"),
		EncryptedBlob:     []byte("blob"),
		Keywords:          keywords,
		Contributor:       "contrib",
		Epoch:             epoch,
		CommittedAtHeight: 1,
		CommittingLeader:  defaultLeader,
		State:             types.MemoryState_MEMORY_STATE_COMMITTED,
		MemoryType:        types.MemoryType_MEMORY_TYPE_MEMORY,
	}
}

func TestApplyDecay_ServeBoost_EarnedTrust_MatchedKeyword(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	params := types.DefaultParams()

	memory := newDecayMemory(1, &types.KeywordWeight{Keyword: "kw1", Weight: "0.5000"})
	memory.ServeCountTotal = 4
	memory.DenialCountTotal = 1

	err := k.applyDecay(memory, 30, 1, 0, map[string]bool{"kw1": true}, params, 1.0, false)
	if err != nil {
		t.Fatalf("applyDecay failed: %v", err)
	}

	denialRate := float64(memory.DenialCountTotal) / float64(memory.ServeCountTotal+memory.DenialCountTotal)
	trust := math.Max(0, 1.0-denialRate)
	delta := (float64(params.ServeDBps) / 10000.0) *
		(float64(params.ServeFloorBps)/10000.0 + (1.0-float64(params.ServeFloorBps)/10000.0)*trust*trust)
	expected := round4(0.5000 + delta)
	got := weightForKeyword(t, memory, "kw1")
	if math.Abs(got-expected) > 0.0001 {
		t.Fatalf("weight mismatch: got %.4f want %.4f", got, expected)
	}
}

func TestApplyDecay_ServeBoost_UnmatchedKeyword_NoBoost(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	params := types.DefaultParams()

	memory := newDecayMemory(1, &types.KeywordWeight{Keyword: "kw1", Weight: "0.5000"})
	memory.ServeCountTotal = 5
	memory.DenialCountTotal = 0

	err := k.applyDecay(memory, 30, 1, 0, map[string]bool{"other": true}, params, 1.0, false)
	if err != nil {
		t.Fatalf("applyDecay failed: %v", err)
	}

	idleDelta := (float64(params.IdleDBps) / 10000.0) * (float64(params.IdleProtectBps) / 10000.0)
	expected := round4(0.5000 - idleDelta)
	got := weightForKeyword(t, memory, "kw1")
	if math.Abs(got-expected) > 0.0001 {
		t.Fatalf("weight mismatch: got %.4f want %.4f", got, expected)
	}
}

func TestApplyDecay_DenialDecay_EarnedTrust(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	params := types.DefaultParams()

	memory := newDecayMemory(1, &types.KeywordWeight{Keyword: "kw1", Weight: "0.8000"})
	memory.ServeCountTotal = 4
	memory.DenialCountTotal = 1

	err := k.applyDecay(memory, 30, 0, 1, map[string]bool{"kw1": true}, params, 1.0, false)
	if err != nil {
		t.Fatalf("applyDecay failed: %v", err)
	}

	denialRate := float64(memory.DenialCountTotal) / float64(memory.ServeCountTotal+memory.DenialCountTotal)
	delta := (float64(params.DenialDBps) / 10000.0) *
		(float64(params.DenialFloorBps)/10000.0 + (1.0-float64(params.DenialFloorBps)/10000.0)*denialRate)
	expected := round4(0.8000 - delta)
	got := weightForKeyword(t, memory, "kw1")
	if math.Abs(got-expected) > 0.0001 {
		t.Fatalf("weight mismatch: got %.4f want %.4f", got, expected)
	}
}

func TestApplyDecay_IdleDecay_TrustEarned(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	params := types.DefaultParams()

	memory := newDecayMemory(1, &types.KeywordWeight{Keyword: "kw1", Weight: "0.9000"})
	memory.ServeCountTotal = 3
	memory.DenialCountTotal = 0

	err := k.applyDecay(memory, 30, 0, 0, map[string]bool{}, params, 1.0, false)
	if err != nil {
		t.Fatalf("applyDecay failed: %v", err)
	}

	idleDelta := (float64(params.IdleDBps) / 10000.0) * (float64(params.IdleProtectBps) / 10000.0)
	expected := round4(0.9000 - idleDelta)
	got := weightForKeyword(t, memory, "kw1")
	if math.Abs(got-expected) > 0.0001 {
		t.Fatalf("weight mismatch: got %.4f want %.4f", got, expected)
	}
}

func TestApplyDecay_IdleDecay_Untrusted(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	params := types.DefaultParams()

	memory := newDecayMemory(1, &types.KeywordWeight{Keyword: "kw1", Weight: "0.9000"})
	memory.ServeCountTotal = 0
	memory.DenialCountTotal = 0

	err := k.applyDecay(memory, 30, 0, 0, map[string]bool{}, params, 1.0, false)
	if err != nil {
		t.Fatalf("applyDecay failed: %v", err)
	}

	idleDelta := (float64(params.IdleDBps) / 10000.0) * (float64(params.IdleUntrustedBps) / 10000.0)
	expected := round4(0.9000 - idleDelta)
	got := weightForKeyword(t, memory, "kw1")
	if math.Abs(got-expected) > 0.0001 {
		t.Fatalf("weight mismatch: got %.4f want %.4f", got, expected)
	}
}

func TestApplyDecay_IdleDecay_PerKeywordGate_ActiveMemory(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	params := types.DefaultParams()

	memory := newDecayMemory(
		1,
		&types.KeywordWeight{Keyword: "kw1", Weight: "0.5000"},
		&types.KeywordWeight{Keyword: "kw2", Weight: "0.5000"},
	)
	memory.ServeCountTotal = 5
	memory.DenialCountTotal = 0

	err := k.applyDecay(memory, 30, 1, 0, map[string]bool{"kw1": true}, params, 1.0, false)
	if err != nil {
		t.Fatalf("applyDecay failed: %v", err)
	}

	kw1Expected := round4(0.5000 + (float64(params.ServeDBps) / 10000.0))
	kw2Expected := round4(0.5000 - (float64(params.IdleDBps)/10000.0)*(float64(params.IdleProtectBps)/10000.0))
	kw1Got := weightForKeyword(t, memory, "kw1")
	kw2Got := weightForKeyword(t, memory, "kw2")

	if math.Abs(kw1Got-kw1Expected) > 0.0001 {
		t.Fatalf("kw1 mismatch: got %.4f want %.4f", kw1Got, kw1Expected)
	}
	if math.Abs(kw2Got-kw2Expected) > 0.0001 {
		t.Fatalf("kw2 mismatch: got %.4f want %.4f", kw2Got, kw2Expected)
	}
}

func TestApplyDecay_GracePeriod_BlocksAllOperations(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	params := types.DefaultParams()

	memory := newDecayMemory(100, &types.KeywordWeight{Keyword: "kw1", Weight: "0.5000"})
	memory.LastActiveEpoch = 7

	err := k.applyDecay(memory, 119, 1, 1, map[string]bool{"kw1": true}, params, 1.0, false)
	if err != nil {
		t.Fatalf("applyDecay failed: %v", err)
	}

	if got := weightForKeyword(t, memory, "kw1"); math.Abs(got-0.5000) > 0.0001 {
		t.Fatalf("grace period should not change weight, got %.4f", got)
	}
	if memory.LastActiveEpoch != 7 {
		t.Fatalf("grace period should not update last_active_epoch, got %d", memory.LastActiveEpoch)
	}
	if memory.State == types.MemoryState_MEMORY_STATE_ARCHIVED {
		t.Fatalf("grace period should not archive memory")
	}
}

func TestApplyDecay_ArchiveTrigger_AllKeywordsBelowThreshold(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	params := types.DefaultParams()

	memory := newDecayMemory(
		1,
		&types.KeywordWeight{Keyword: "k1", Weight: "0.1000"},
		&types.KeywordWeight{Keyword: "k2", Weight: "0.1000"},
		&types.KeywordWeight{Keyword: "k3", Weight: "0.1000"},
	)

	err := k.applyDecay(memory, 50, 0, 0, nil, params, 1.0, false)
	if err != nil {
		t.Fatalf("applyDecay failed: %v", err)
	}

	if memory.State != types.MemoryState_MEMORY_STATE_ARCHIVED {
		t.Fatalf("expected ARCHIVED state, got %v", memory.State)
	}
	if memory.ArchivedEpoch != 50 {
		t.Fatalf("archived_epoch mismatch: got %d want 50", memory.ArchivedEpoch)
	}
}

func TestApplyDecay_ArchiveTrigger_NotAllKeywordsBelow_NoArchive(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	params := types.DefaultParams()

	memory := newDecayMemory(
		1,
		&types.KeywordWeight{Keyword: "k1", Weight: "0.1000"},
		&types.KeywordWeight{Keyword: "k2", Weight: "0.1000"},
		&types.KeywordWeight{Keyword: "k3", Weight: "0.5000"},
	)

	err := k.applyDecay(memory, 50, 0, 0, nil, params, 1.0, false)
	if err != nil {
		t.Fatalf("applyDecay failed: %v", err)
	}

	if memory.State == types.MemoryState_MEMORY_STATE_ARCHIVED {
		t.Fatalf("memory should remain non-archived")
	}
	if memory.ArchivedEpoch != 0 {
		t.Fatalf("archived_epoch should remain 0, got %d", memory.ArchivedEpoch)
	}
}

func TestApplyDecay_ZeroEventsZeroDenials_EdgeCase(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	params := types.DefaultParams()

	memory := newDecayMemory(1, &types.KeywordWeight{Keyword: "kw1", Weight: "0.8000"})
	memory.ServeCountTotal = 0
	memory.DenialCountTotal = 0

	err := k.applyDecay(memory, 30, 0, 0, nil, params, 1.0, false)
	if err != nil {
		t.Fatalf("applyDecay failed: %v", err)
	}

	idleDelta := (float64(params.IdleDBps) / 10000.0) * (float64(params.IdleUntrustedBps) / 10000.0)
	expected := round4(0.8000 - idleDelta)
	got := weightForKeyword(t, memory, "kw1")
	if math.Abs(got-expected) > 0.0001 {
		t.Fatalf("weight mismatch: got %.4f want %.4f", got, expected)
	}
}

func TestApplyDecay_Clamps_WeightAtZero(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	params := types.DefaultParams()

	memory := newDecayMemory(1, &types.KeywordWeight{Keyword: "kw1", Weight: "0.0100"})

	err := k.applyDecay(memory, 30, 0, 0, nil, params, 1.0, false)
	if err != nil {
		t.Fatalf("applyDecay failed: %v", err)
	}

	if got := weightForKeyword(t, memory, "kw1"); got != 0 {
		t.Fatalf("weight should clamp to zero, got %.4f", got)
	}
}

func TestApplyDecay_Clamps_WeightAtOne(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	params := types.DefaultParams()

	memory := newDecayMemory(1, &types.KeywordWeight{Keyword: "kw1", Weight: "0.9900"})
	memory.ServeCountTotal = 10
	memory.DenialCountTotal = 0

	err := k.applyDecay(memory, 30, 3, 0, map[string]bool{"kw1": true}, params, 1.0, false)
	if err != nil {
		t.Fatalf("applyDecay failed: %v", err)
	}

	if got := weightForKeyword(t, memory, "kw1"); got != 1.0 {
		t.Fatalf("weight should clamp to one, got %.4f", got)
	}
}

func TestResolveOrgIdleDecayConfig_AtReferenceTrafficScaleIsOne(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()
	params := types.DefaultParams()

	for i := 0; i < 50; i++ {
		hash := []byte(fmt.Sprintf("%032d", i))
		storeMemoryWithKeywords(t, k, ctx, defaultOrgID, hash, types.MemoryState_MEMORY_STATE_COMMITTED, 0)
	}

	mock := newMockServeKeeper()
	cid := types.ContentHashToHex([]byte(fmt.Sprintf("%032d", 0)))
	mock.servesByEpoch[makeServeEpochKey(defaultOrgID, cid, 40)] = 11
	k.SetServeKeeper(mock)

	config, err := k.resolveOrgIdleDecayConfig(ctx, defaultOrgID, 40, params)
	if err != nil {
		t.Fatalf("resolveOrgIdleDecayConfig failed: %v", err)
	}
	if config.suppressIdle {
		t.Fatalf("expected suppressIdle=false when org has traffic")
	}
	if math.Abs(config.idleScale-1.0) > 0.000001 {
		t.Fatalf("idle scale mismatch at T_ref: got %.6f want 1.0", config.idleScale)
	}
}

func TestApplyEpochDecay_ZeroSignalGuardSkipsIdleDecay(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	hash := []byte("quiet_epoch_good_memory_hash_000001")
	storeMemoryWithKeywords(
		t,
		k,
		ctx,
		defaultOrgID,
		hash,
		types.MemoryState_MEMORY_STATE_COMMITTED,
		0,
		withKeywords(&types.KeywordWeight{Keyword: "kw", Weight: "0.1600"}),
	)

	k.SetServeKeeper(newMockServeKeeper())

	if err := k.ApplyEpochDecay(ctx, 40); err != nil {
		t.Fatalf("ApplyEpochDecay failed: %v", err)
	}

	memory, err := k.GetApprovedMemory(ctx, defaultOrgID, hash)
	if err != nil {
		t.Fatalf("GetApprovedMemory failed: %v", err)
	}
	if memory.State != types.MemoryState_MEMORY_STATE_COMMITTED {
		t.Fatalf("memory should stay committed in quiet epoch, got %v", memory.State)
	}
	if got := weightForKeyword(t, memory, "kw"); math.Abs(got-0.1600) > 0.0001 {
		t.Fatalf("weight changed during quiet epoch: got %.4f want 0.1600", got)
	}
}

func TestApplyEpochDecay_BadMemoryWithDenialsStillDecays(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	hash := []byte("bad_memory_denial_decay_hash_000000")
	cid := storeMemoryWithKeywords(
		t,
		k,
		ctx,
		defaultOrgID,
		hash,
		types.MemoryState_MEMORY_STATE_COMMITTED,
		0,
		withMemoryType(types.MemoryType_MEMORY_TYPE_MEMORY),
		withKeywords(&types.KeywordWeight{Keyword: "kw", Weight: "0.8000"}),
		withMemoryServeTotal(3),
		withMemoryDenialTotal(2),
	)

	mock := newMockServeKeeper()
	mock.denialsByEpoch[makeServeEpochKey(defaultOrgID, cid, 40)] = 1
	mock.matchedByEpoch[makeServeEpochKey(defaultOrgID, cid, 40)] = map[string]bool{"kw": true}
	k.SetServeKeeper(mock)

	before, err := k.GetApprovedMemory(ctx, defaultOrgID, hash)
	if err != nil {
		t.Fatalf("GetApprovedMemory(before) failed: %v", err)
	}
	beforeWeight := weightForKeyword(t, before, "kw")

	if err := k.ApplyEpochDecay(ctx, 40); err != nil {
		t.Fatalf("ApplyEpochDecay failed: %v", err)
	}

	after, err := k.GetApprovedMemory(ctx, defaultOrgID, hash)
	if err != nil {
		t.Fatalf("GetApprovedMemory(after) failed: %v", err)
	}
	afterWeight := weightForKeyword(t, after, "kw")
	if !(afterWeight < beforeWeight) {
		t.Fatalf("bad memory weight should decay on denial: before %.4f after %.4f", beforeWeight, afterWeight)
	}
}

func TestIntegration_SteadyStateScenario(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()
	params := types.DefaultParams()

	goodHash := []byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	badHash := []byte("cccccccccccccccccccccccccccccccc")

	goodCID := storeMemoryWithKeywords(
		t,
		k,
		ctx,
		defaultOrgID,
		goodHash,
		types.MemoryState_MEMORY_STATE_COMMITTED,
		0,
		withKeywords(
			&types.KeywordWeight{Keyword: "kw1", Weight: "0.6000"},
			&types.KeywordWeight{Keyword: "kw2", Weight: "0.6000"},
		),
	)
	badCID := storeMemoryWithKeywords(
		t,
		k,
		ctx,
		defaultOrgID,
		badHash,
		types.MemoryState_MEMORY_STATE_COMMITTED,
		0,
		withKeywords(
			&types.KeywordWeight{Keyword: "bad1", Weight: "0.6000"},
			&types.KeywordWeight{Keyword: "bad2", Weight: "0.6000"},
		),
	)

	mock := newMockServeKeeper()
	k.SetServeKeeper(mock)

	setEpochData := func(cid string, epoch uint64, serves, denials uint64, keywords ...string) {
		key := makeServeEpochKey(defaultOrgID, cid, epoch)
		mock.servesByEpoch[key] = serves
		mock.denialsByEpoch[key] = denials
		matched := make(map[string]bool, len(keywords))
		for _, keyword := range keywords {
			matched[keyword] = true
		}
		mock.matchedByEpoch[key] = matched
	}

	for epoch := uint64(1); epoch <= 50; epoch++ {
		if err := k.setCurrentEpoch(ctx, epoch); err != nil {
			t.Fatalf("setCurrentEpoch failed: %v", err)
		}

		setEpochData(goodCID, epoch, 1, 0, "kw1")
		setEpochData(badCID, epoch, 0, 1, "bad1", "bad2")

		if err := k.ApplyServeBoost(ctx, defaultOrgID, goodHash, epoch); err != nil {
			t.Fatalf("ApplyServeBoost failed at epoch %d: %v", epoch, err)
		}
		if err := k.ApplyDenialDecay(ctx, defaultOrgID, badHash, epoch); err != nil {
			t.Fatalf("ApplyDenialDecay failed at epoch %d: %v", epoch, err)
		}
		if err := k.ApplyEpochDecay(ctx, epoch); err != nil {
			t.Fatalf("ApplyEpochDecay failed at epoch %d: %v", epoch, err)
		}
	}

	goodMemory, err := k.GetApprovedMemory(ctx, defaultOrgID, goodHash)
	if err != nil {
		t.Fatalf("GetApprovedMemory(good) failed: %v", err)
	}
	badMemory, err := k.GetApprovedMemory(ctx, defaultOrgID, badHash)
	if err != nil {
		t.Fatalf("GetApprovedMemory(bad) failed: %v", err)
	}

	retrievalThreshold := float64(params.RetrievalThresholdBps) / 10000.0
	goodAboveThreshold := false
	for _, kw := range goodMemory.Keywords {
		if weightForKeyword(t, goodMemory, kw.Keyword) > retrievalThreshold {
			goodAboveThreshold = true
			break
		}
	}

	if !goodAboveThreshold {
		t.Fatalf("good memory should keep at least one keyword above retrieval threshold")
	}
	if goodMemory.State == types.MemoryState_MEMORY_STATE_ARCHIVED {
		t.Fatalf("good memory should not be archived")
	}
	badWeights := make([]float64, 0, len(badMemory.Keywords))
	for _, kw := range badMemory.Keywords {
		badWeights = append(badWeights, weightForKeyword(t, badMemory, kw.Keyword))
	}
	if badMemory.State != types.MemoryState_MEMORY_STATE_ARCHIVED {
		t.Fatalf("bad memory should be archived, got %v weights=%v", badMemory.State, badWeights)
	}
	if badMemory.ArchivedEpoch == 0 || badMemory.ArchivedEpoch > 50 {
		t.Fatalf("bad memory archived_epoch out of range: %d weights=%v", badMemory.ArchivedEpoch, badWeights)
	}
}
