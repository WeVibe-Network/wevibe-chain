package keeper_test

import (
	"context"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	testkeeper "github.com/wevibe-network/wevibe-chain/testutil/keeper"
	"github.com/wevibe-network/wevibe-chain/x/emissions/keeper"
	"github.com/wevibe-network/wevibe-chain/x/emissions/types"
	orgTypes "github.com/wevibe-network/wevibe-chain/x/org/types"
	serveTypes "github.com/wevibe-network/wevibe-chain/x/serve/types"
)

var (
	emissionsStoreKey = storetypes.NewKVStoreKey("emissions")
)

// contributorPool32yr is the locked 32-year contributor pool size (320,000,000 VIBE in uvibe).
const contributorPool32yr = uint64(320_000_000_000_000)

func newTestKeeper(t *testing.T) (*keeper.Keeper, context.Context) {
	storeService, _ := testkeeper.NewTestStoreService(t, emissionsStoreKey)
	logger := testkeeper.NewTestLogger()
	memoryKeeper := &mockMemoryKeeper{
		approvedByContributor: make(map[string]map[string]uint64),
		contributorsByEpoch:   make(map[uint64]map[string]uint64),
	}
	k := keeper.NewKeeper(storeService, logger, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", nil, memoryKeeper, nil, newMockReputationKeeper())
	ctx := context.Background()
	return k, ctx
}

func TestSetEmissionPool(t *testing.T) {
	k, ctx := newTestKeeper(t)

	pool := types.NewEmissionPool(1000000, 10000, 80, 20, 0)
	err := k.SetEmissionPool(ctx, pool)
	if err != nil {
		t.Fatalf("SetEmissionPool failed: %v", err)
	}

	retrieved, err := k.GetEmissionPool(ctx)
	if err != nil {
		t.Fatalf("GetEmissionPool failed: %v", err)
	}
	if retrieved.TotalSupply != pool.TotalSupply {
		t.Errorf("expected TotalSupply %d, got %d", pool.TotalSupply, retrieved.TotalSupply)
	}
	if retrieved.DailyMint != pool.DailyMint {
		t.Errorf("expected DailyMint %d, got %d", pool.DailyMint, retrieved.DailyMint)
	}
	if retrieved.OperatorShare != pool.OperatorShare {
		t.Errorf("expected OperatorShare %d, got %d", pool.OperatorShare, retrieved.OperatorShare)
	}
	if retrieved.ValidatorShare != pool.ValidatorShare {
		t.Errorf("expected ValidatorShare %d, got %d", pool.ValidatorShare, retrieved.ValidatorShare)
	}
}

func TestMintDailyEmission(t *testing.T) {
	k, ctx := newTestKeeper(t)

	if err := k.SetParams(ctx, types.DefaultParams()); err != nil {
		t.Fatalf("SetParams failed: %v", err)
	}

	pool := types.NewEmissionPool(0, 0, 80, 20, 0)
	pool.ValidatorPoolRemainingUvibe = 570_000_000_000_000
	pool.ContributorPoolRemainingUvibe = 320_000_000_000_000
	err := k.SetEmissionPool(ctx, pool)
	if err != nil {
		t.Fatalf("SetEmissionPool failed: %v", err)
	}

	params := types.DefaultParams()
	expectedValidator := pool.ValidatorPoolRemainingUvibe / params.ScheduleDurationDays

	emission, err := k.MintDailyEmission(ctx, 1)
	if err != nil {
		t.Fatalf("MintDailyEmission failed: %v", err)
	}
	if emission.Epoch != 1 {
		t.Errorf("expected Epoch 1, got %d", emission.Epoch)
	}
	if emission.ValidatorShare != expectedValidator {
		t.Errorf("expected ValidatorShare %d, got %d", expectedValidator, emission.ValidatorShare)
	}

	retrievedPool, _ := k.GetEmissionPool(ctx)
	if retrievedPool.Epoch != 1 {
		t.Errorf("expected pool Epoch 1, got %d", retrievedPool.Epoch)
	}
	if retrievedPool.ValidatorPoolRemainingUvibe != pool.ValidatorPoolRemainingUvibe-expectedValidator {
		t.Errorf("expected validator pool remaining %d, got %d", pool.ValidatorPoolRemainingUvibe-expectedValidator, retrievedPool.ValidatorPoolRemainingUvibe)
	}
	if retrievedPool.StartEpoch != 1 {
		t.Errorf("expected StartEpoch 1, got %d", retrievedPool.StartEpoch)
	}
	if retrievedPool.TotalEpochsElapsed != 1 {
		t.Errorf("expected TotalEpochsElapsed 1, got %d", retrievedPool.TotalEpochsElapsed)
	}
}

func TestMintDailyEmission_InvalidEpoch(t *testing.T) {
	k, ctx := newTestKeeper(t)

	pool := types.NewEmissionPool(1000000, 10000, 80, 20, 0)
	err := k.SetEmissionPool(ctx, pool)
	if err != nil {
		t.Fatalf("SetEmissionPool failed: %v", err)
	}

	_, err = k.MintDailyEmission(ctx, 0)
	if err == nil {
		t.Fatal("expected error for invalid epoch 0")
	}
}

// newScheduleKeeper builds a keeper with a controllable memory keeper for
// 32-year schedule emission tests.
func newScheduleKeeper(t *testing.T, contributorsByEpoch map[uint64]map[string]uint64) (*keeper.Keeper, context.Context, *mockMemoryKeeper) {
	storeService, _ := testkeeper.NewTestStoreService(t, emissionsStoreKey)
	logger := testkeeper.NewTestLogger()
	if contributorsByEpoch == nil {
		contributorsByEpoch = make(map[uint64]map[string]uint64)
	}
	memoryKeeper := &mockMemoryKeeper{
		approvedByContributor: make(map[string]map[string]uint64),
		contributorsByEpoch:   contributorsByEpoch,
	}
	k := keeper.NewKeeper(storeService, logger, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", nil, memoryKeeper, nil, newMockReputationKeeper())
	ctx := context.Background()
	if err := k.SetParams(ctx, types.DefaultParams()); err != nil {
		t.Fatalf("SetParams failed: %v", err)
	}
	return k, ctx, memoryKeeper
}

func seedSchedulePool(t *testing.T, k *keeper.Keeper, ctx context.Context, validatorRemaining, contributorRemaining uint64) {
	pool := types.NewEmissionPool(0, 0, 80, 20, 0)
	pool.ValidatorPoolRemainingUvibe = validatorRemaining
	pool.ContributorPoolRemainingUvibe = contributorRemaining
	if err := k.SetEmissionPool(ctx, pool); err != nil {
		t.Fatalf("SetEmissionPool failed: %v", err)
	}
}

// (a) validator per-epoch emission == ValidatorPoolRemaining/remainingEpochs and pool deducted.
func TestMintDailyEmission_ValidatorPerEpoch(t *testing.T) {
	params := types.DefaultParams()
	k, ctx, _ := newScheduleKeeper(t, nil)
	seedSchedulePool(t, k, ctx, params.ValidatorEmissionPoolUvibe, 0)

	// remainingEpochs == ScheduleDurationDays - TotalEpochsElapsed(0) == ScheduleDurationDays
	expected := params.ValidatorEmissionPoolUvibe / params.ScheduleDurationDays

	emission, err := k.MintDailyEmission(ctx, 1)
	if err != nil {
		t.Fatalf("MintDailyEmission failed: %v", err)
	}
	if emission.ValidatorShare != expected {
		t.Errorf("expected validator emission %d, got %d", expected, emission.ValidatorShare)
	}
	if emission.TotalEmitted != expected {
		t.Errorf("expected total emitted %d (no contributors), got %d", expected, emission.TotalEmitted)
	}

	pool, _ := k.GetEmissionPool(ctx)
	if pool.ValidatorPoolRemainingUvibe != params.ValidatorEmissionPoolUvibe-expected {
		t.Errorf("expected validator pool remaining %d, got %d", params.ValidatorEmissionPoolUvibe-expected, pool.ValidatorPoolRemainingUvibe)
	}
}

// (b) contributor budget capped at ContributorAnnualCapUvibe/EpochsPerYear.
func TestMintDailyEmission_ContributorBudgetCapped(t *testing.T) {
	params := types.DefaultParams()
	// Contributor pool large enough that pool/remainingEpochs strictly exceeds
	// the daily cap, so the annual-cap clamp is exercised.
	largePool := contributorPool32yr * 2
	k, ctx, _ := newScheduleKeeper(t, map[uint64]map[string]uint64{
		1: {"contrib1": 1},
	})
	seedSchedulePool(t, k, ctx, 0, largePool)

	uncapped := largePool / params.ScheduleDurationDays
	epochCap := params.ContributorAnnualCapUvibe / types.EpochsPerYear
	if uncapped <= epochCap {
		t.Fatalf("test precondition failed: uncapped budget %d should exceed epoch cap %d", uncapped, epochCap)
	}

	emission, err := k.MintDailyEmission(ctx, 1)
	if err != nil {
		t.Fatalf("MintDailyEmission failed: %v", err)
	}

	// One qualifying contributor receives the full capped budget (rollover was 0).
	reward, err := k.GetContributorReward(ctx, "contrib1")
	if err != nil {
		t.Fatalf("GetContributorReward failed: %v", err)
	}
	if reward != epochCap {
		t.Errorf("expected contributor reward capped at %d, got %d", epochCap, reward)
	}
	if emission.TotalEmitted != epochCap {
		t.Errorf("expected total emitted %d, got %d", epochCap, emission.TotalEmitted)
	}

	pool, _ := k.GetEmissionPool(ctx)
	if pool.ContributorPoolRemainingUvibe != largePool-epochCap {
		t.Errorf("expected contributor pool remaining %d, got %d", largePool-epochCap, pool.ContributorPoolRemainingUvibe)
	}
}

// (c) rollover when zero qualifying contributors (ContributorRolloverUvibe grows by the budget).
func TestMintDailyEmission_RolloverWhenNoContributors(t *testing.T) {
	params := types.DefaultParams()
	// No contributors for epoch 1.
	k, ctx, _ := newScheduleKeeper(t, nil)
	seedSchedulePool(t, k, ctx, 0, contributorPool32yr)

	epochCap := params.ContributorAnnualCapUvibe / types.EpochsPerYear

	emission, err := k.MintDailyEmission(ctx, 1)
	if err != nil {
		t.Fatalf("MintDailyEmission failed: %v", err)
	}
	if emission.TotalEmitted != 0 {
		t.Errorf("expected 0 emitted with no qualifying contributors, got %d", emission.TotalEmitted)
	}

	pool, _ := k.GetEmissionPool(ctx)
	// budget for epoch 1 == epochCap (capped); all of it rolls over.
	if pool.ContributorRolloverUvibe != epochCap {
		t.Errorf("expected rollover %d, got %d", epochCap, pool.ContributorRolloverUvibe)
	}
}

// (d) even split among N qualifying contributors with integer remainder carried into rollover.
func TestMintDailyEmission_EvenSplitWithRemainder(t *testing.T) {
	// Use 3 contributors and a small pool so the per-epoch budget does NOT
	// divide evenly, exercising the integer-remainder carry-forward.
	k, ctx, _ := newScheduleKeeper(t, map[uint64]map[string]uint64{
		1: {"a": 1, "b": 1, "c": 1},
	})

	// Short schedule + small pool so the annual cap does not bind and the
	// per-epoch budget has a known, non-zero remainder when split 3 ways.
	p := types.DefaultParams()
	p.ScheduleDurationDays = 10
	if err := k.SetParams(ctx, p); err != nil {
		t.Fatalf("SetParams failed: %v", err)
	}
	// pool 100 / 10 epochs = budget 10; 10 / 3 = 3 each, remainder 1.
	seedSchedulePool(t, k, ctx, 0, 100)

	n := uint64(3)
	expectedBudget := uint64(100) / p.ScheduleDurationDays // 10
	expectedPer := expectedBudget / n                      // 3
	expectedRemainder := expectedBudget % n                // 1

	emission, err := k.MintDailyEmission(ctx, 1)
	if err != nil {
		t.Fatalf("MintDailyEmission failed: %v", err)
	}

	for _, addr := range []string{"a", "b", "c"} {
		reward, err := k.GetContributorReward(ctx, addr)
		if err != nil {
			t.Fatalf("GetContributorReward(%s) failed: %v", addr, err)
		}
		if reward != expectedPer {
			t.Errorf("contributor %s: expected %d, got %d", addr, expectedPer, reward)
		}
	}

	if emission.TotalEmitted != expectedPer*n {
		t.Errorf("expected total emitted %d, got %d", expectedPer*n, emission.TotalEmitted)
	}

	pool, _ := k.GetEmissionPool(ctx)
	if pool.ContributorRolloverUvibe != expectedRemainder {
		t.Errorf("expected rollover remainder %d, got %d", expectedRemainder, pool.ContributorRolloverUvibe)
	}
}

// (e) pool depletion: after many epochs pools never go negative and trend toward 0.
func TestMintDailyEmission_PoolDepletion(t *testing.T) {
	params := types.DefaultParams()
	// One qualifying contributor every epoch so contributor budget is distributed.
	contributors := make(map[uint64]map[string]uint64)
	k, ctx, mem := newScheduleKeeper(t, contributors)

	// Small pools and a short schedule so depletion completes within the test loop.
	const shortSchedule = 50
	p := params
	p.ScheduleDurationDays = shortSchedule
	if err := k.SetParams(ctx, p); err != nil {
		t.Fatalf("SetParams failed: %v", err)
	}

	validatorPool := uint64(1_000_000)
	contributorPool := uint64(500_000)
	seedSchedulePool(t, k, ctx, validatorPool, contributorPool)

	prevValidator := validatorPool
	prevContributor := contributorPool
	for epoch := uint64(1); epoch <= shortSchedule; epoch++ {
		mem.contributorsByEpoch[epoch] = map[string]uint64{"contrib1": 1}
		if _, err := k.MintDailyEmission(ctx, epoch); err != nil {
			t.Fatalf("MintDailyEmission(%d) failed: %v", epoch, err)
		}
		pool, _ := k.GetEmissionPool(ctx)
		// Never negative (uint64 underflow would surface as a huge value > original).
		if pool.ValidatorPoolRemainingUvibe > validatorPool {
			t.Fatalf("epoch %d: validator pool underflowed: %d", epoch, pool.ValidatorPoolRemainingUvibe)
		}
		if pool.ContributorPoolRemainingUvibe > contributorPool {
			t.Fatalf("epoch %d: contributor pool underflowed: %d", epoch, pool.ContributorPoolRemainingUvibe)
		}
		// Monotonically non-increasing.
		if pool.ValidatorPoolRemainingUvibe > prevValidator {
			t.Fatalf("epoch %d: validator pool increased", epoch)
		}
		if pool.ContributorPoolRemainingUvibe > prevContributor {
			t.Fatalf("epoch %d: contributor pool increased", epoch)
		}
		prevValidator = pool.ValidatorPoolRemainingUvibe
		prevContributor = pool.ContributorPoolRemainingUvibe
	}

	finalPool, _ := k.GetEmissionPool(ctx)
	if finalPool.ValidatorPoolRemainingUvibe != 0 {
		t.Errorf("expected validator pool fully depleted, got %d", finalPool.ValidatorPoolRemainingUvibe)
	}
	if finalPool.ContributorPoolRemainingUvibe != 0 {
		t.Errorf("expected contributor pool fully depleted, got %d", finalPool.ContributorPoolRemainingUvibe)
	}
}

func TestAsymmetricGate(t *testing.T) {
	k, ctx := newTestKeeper(t)

	gate := types.NewAsymmetricGate("op1", "org1", true, 1)
	err := k.SetAsymmetricGate(ctx, gate)
	if err != nil {
		t.Fatalf("SetAsymmetricGate failed: %v", err)
	}

	retrieved, err := k.GetAsymmetricGate(ctx, "op1", "org1", 1)
	if err != nil {
		t.Fatalf("GetAsymmetricGate failed: %v", err)
	}
	if !retrieved.StoragePassed {
		t.Error("expected StoragePassed to be true")
	}
	if !retrieved.RetrievalAllowed {
		t.Error("expected RetrievalAllowed to be true")
	}

	allowed, err := k.CheckRetrievalAllowed(ctx, "op1", "org1", 1)
	if err != nil {
		t.Fatalf("CheckRetrievalAllowed failed: %v", err)
	}
	if !allowed {
		t.Error("expected retrieval to be allowed")
	}
}

func TestAsymmetricGate_StorageFailed(t *testing.T) {
	k, ctx := newTestKeeper(t)

	gate := types.NewAsymmetricGate("op1", "org1", false, 1)
	err := k.SetAsymmetricGate(ctx, gate)
	if err != nil {
		t.Fatalf("SetAsymmetricGate failed: %v", err)
	}

	allowed, err := k.CheckRetrievalAllowed(ctx, "op1", "org1", 1)
	if err != nil {
		t.Fatalf("CheckRetrievalAllowed failed: %v", err)
	}
	if allowed {
		t.Error("expected retrieval to be disallowed")
	}
}

func TestAsymmetricGate_DefaultPassed(t *testing.T) {
	k, ctx := newTestKeeper(t)

	gate, err := k.GetAsymmetricGate(ctx, "op1", "org1", 1)
	if err != nil {
		t.Fatalf("GetAsymmetricGate failed: %v", err)
	}
	if !gate.StoragePassed {
		t.Error("expected default gate StoragePassed to be true")
	}
	if !gate.RetrievalAllowed {
		t.Error("expected default gate RetrievalAllowed to be true")
	}
}

func TestBootstrapCredits(t *testing.T) {
	k, ctx := newTestKeeper(t)

	k.SetBootstrapCredits(ctx, 100000)
	k.SetBootstrapExpiry(ctx, 100)

	credits, err := k.GetBootstrapCredits(ctx)
	if err != nil {
		t.Fatalf("GetBootstrapCredits failed: %v", err)
	}
	if credits != 100000 {
		t.Errorf("expected 100000 credits, got %d", credits)
	}

	k.AddBootstrapCredit(ctx, "op1", 5000)

	err = k.RedeemBootstrapCredit(ctx, "op1", 1000)
	if err != nil {
		t.Fatalf("RedeemBootstrapCredit failed: %v", err)
	}

	credit, err := k.GetBootstrapCredit(ctx, "op1")
	if err != nil {
		t.Fatalf("GetBootstrapCredit failed: %v", err)
	}
	if credit.Redeemed != 1000 {
		t.Errorf("expected Redeemed 1000, got %d", credit.Redeemed)
	}
	if credit.Credits != 5000 {
		t.Errorf("expected Credits 5000, got %d", credit.Credits)
	}

	remaining, err := k.GetBootstrapCredits(ctx)
	if err != nil {
		t.Fatalf("GetBootstrapCredits failed: %v", err)
	}
	if remaining != 99000 {
		t.Errorf("expected remaining 99000, got %d", remaining)
	}
}

func TestBootstrapCredit_CannotRedeem(t *testing.T) {
	k, ctx := newTestKeeper(t)

	k.SetBootstrapCredits(ctx, 500)
	k.SetBootstrapExpiry(ctx, 100)

	err := k.RedeemBootstrapCredit(ctx, "op1", 1000)
	if err == nil {
		t.Fatal("expected error when redeeming more than available")
	}
}

func TestComputeRarityMultiplier(t *testing.T) {
	k, ctx := newTestKeeper(t)

	multiplier := k.ComputeRarityMultiplier(ctx, "org1", 10, 10)
	if multiplier != 1.0 {
		t.Errorf("expected 1.0, got %f", multiplier)
	}

	multiplier = k.ComputeRarityMultiplier(ctx, "org1", 10, 5)
	if multiplier <= 1.0 {
		t.Errorf("expected > 1.0, got %f", multiplier)
	}
	if multiplier > 3.0 {
		t.Errorf("expected <= 3.0, got %f", multiplier)
	}

	multiplier = k.ComputeRarityMultiplier(ctx, "org1", 10, 1)
	if multiplier != 3.0 {
		t.Errorf("expected 3.0, got %f", multiplier)
	}

	multiplier = k.ComputeRarityMultiplier(ctx, "org1", 0, 0)
	if multiplier != 1.0 {
		t.Errorf("expected 1.0 for zero operators, got %f", multiplier)
	}
}

func TestInitGenesisAndExportGenesis(t *testing.T) {
	k, ctx := newTestKeeper(t)

	pool := types.NewEmissionPool(1000000, 10000, 80, 20, 0)
	k.SetEmissionPool(ctx, pool)

	_, err := k.MintDailyEmission(ctx, 1)
	if err != nil {
		t.Fatalf("MintDailyEmission failed: %v", err)
	}

	_, err = k.MintDailyEmission(ctx, 2)
	if err != nil {
		t.Fatalf("MintDailyEmission failed: %v", err)
	}

	k.AddBootstrapCredit(ctx, "op1", 5000)
	k.SetBootstrapExpiry(ctx, 100)

	state := &types.GenesisState{
		EmissionPool:    pool,
		BootstrapExpiry: 100,
	}

	storeService2, _ := testkeeper.NewTestStoreService(t, emissionsStoreKey)
	logger2 := testkeeper.NewTestLogger()
	k2 := keeper.NewKeeper(storeService2, logger2, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", nil, nil, nil, newMockReputationKeeper())
	ctx2 := context.Background()

	err = k2.InitGenesis(ctx2, state)
	if err != nil {
		t.Fatalf("InitGenesis failed: %v", err)
	}

	exported, err := k2.ExportGenesis(ctx2)
	if err != nil {
		t.Fatalf("ExportGenesis failed: %v", err)
	}
	if exported.EmissionPool.TotalSupply != pool.TotalSupply {
		t.Errorf("expected TotalSupply %d, got %d", pool.TotalSupply, exported.EmissionPool.TotalSupply)
	}
	if exported.EmissionPool.DailyMint != pool.DailyMint {
		t.Errorf("expected DailyMint %d, got %d", pool.DailyMint, exported.EmissionPool.DailyMint)
	}
}

// TestInitGenesisExportGenesis_RoundTrips32YearPool verifies that seeding the
// keeper with the full 32-year DefaultGenesis pool and exporting it preserves
// all five new pool fields (validator/contributor remaining, rollover, start
// epoch, total epochs elapsed). This exercises the keeper's reliance on the
// EmissionPoolToStored/StoredToEmissionPool helpers.
func TestInitGenesisExportGenesis_RoundTrips32YearPool(t *testing.T) {
	storeService, _ := testkeeper.NewTestStoreService(t, emissionsStoreKey)
	logger := testkeeper.NewTestLogger()
	k := keeper.NewKeeper(storeService, logger, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", nil, nil, nil, newMockReputationKeeper())
	ctx := context.Background()

	state := types.DefaultGenesis()

	// Sanity: the seeded genesis already carries the locked 32-year amounts.
	if state.EmissionPool.ValidatorPoolRemainingUvibe != 570_000_000_000_000 {
		t.Fatalf("seeded ValidatorPoolRemainingUvibe = %d, want 570000000000000", state.EmissionPool.ValidatorPoolRemainingUvibe)
	}
	if state.EmissionPool.ContributorPoolRemainingUvibe != contributorPool32yr {
		t.Fatalf("seeded ContributorPoolRemainingUvibe = %d, want %d", state.EmissionPool.ContributorPoolRemainingUvibe, contributorPool32yr)
	}

	if err := k.InitGenesis(ctx, state); err != nil {
		t.Fatalf("InitGenesis failed: %v", err)
	}

	exported, err := k.ExportGenesis(ctx)
	if err != nil {
		t.Fatalf("ExportGenesis failed: %v", err)
	}
	if exported.EmissionPool == nil {
		t.Fatalf("exported EmissionPool is nil")
	}

	got := exported.EmissionPool
	want := state.EmissionPool

	if got.ValidatorPoolRemainingUvibe != want.ValidatorPoolRemainingUvibe {
		t.Errorf("ValidatorPoolRemainingUvibe: got %d, want %d", got.ValidatorPoolRemainingUvibe, want.ValidatorPoolRemainingUvibe)
	}
	if got.ValidatorPoolRemainingUvibe != 570_000_000_000_000 {
		t.Errorf("ValidatorPoolRemainingUvibe: got %d, want 570000000000000", got.ValidatorPoolRemainingUvibe)
	}
	if got.ContributorPoolRemainingUvibe != want.ContributorPoolRemainingUvibe {
		t.Errorf("ContributorPoolRemainingUvibe: got %d, want %d", got.ContributorPoolRemainingUvibe, want.ContributorPoolRemainingUvibe)
	}
	if got.ContributorPoolRemainingUvibe != contributorPool32yr {
		t.Errorf("ContributorPoolRemainingUvibe: got %d, want %d", got.ContributorPoolRemainingUvibe, contributorPool32yr)
	}
	if got.ContributorRolloverUvibe != want.ContributorRolloverUvibe {
		t.Errorf("ContributorRolloverUvibe: got %d, want %d", got.ContributorRolloverUvibe, want.ContributorRolloverUvibe)
	}
	if got.StartEpoch != want.StartEpoch {
		t.Errorf("StartEpoch: got %d, want %d", got.StartEpoch, want.StartEpoch)
	}
	if got.TotalEpochsElapsed != want.TotalEpochsElapsed {
		t.Errorf("TotalEpochsElapsed: got %d, want %d", got.TotalEpochsElapsed, want.TotalEpochsElapsed)
	}

	// The pre-existing fields must also round-trip.
	if got.DailyMint != want.DailyMint {
		t.Errorf("DailyMint: got %d, want %d", got.DailyMint, want.DailyMint)
	}
	if got.OperatorShare != want.OperatorShare || got.OperatorShare != 80 {
		t.Errorf("OperatorShare: got %d, want 80", got.OperatorShare)
	}
	if got.ValidatorShare != want.ValidatorShare || got.ValidatorShare != 20 {
		t.Errorf("ValidatorShare: got %d, want 20", got.ValidatorShare)
	}
}

func TestSDKTransactionIsolation(t *testing.T) {
	storeService1, cms1 := testkeeper.NewTestStoreService(t, emissionsStoreKey)
	logger1 := testkeeper.NewTestLogger()
	k1 := keeper.NewKeeper(storeService1, logger1, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", nil, nil, nil, newMockReputationKeeper())
	ctx1 := context.Background()

	pool := types.NewEmissionPool(1000000, 10000, 80, 20, 0)
	err := k1.SetEmissionPool(ctx1, pool)
	if err != nil {
		t.Fatalf("SetEmissionPool failed: %v", err)
	}

	retrieved1, err := k1.GetEmissionPool(ctx1)
	if err != nil {
		t.Fatalf("GetEmissionPool failed: %v", err)
	}
	if retrieved1.TotalSupply != pool.TotalSupply {
		t.Fatal("tx1 should see its own emission pool")
	}

	cms1.Commit()

	storeService2 := testkeeper.NewTestStoreServiceWithCMS(t, emissionsStoreKey, cms1)
	logger2 := testkeeper.NewTestLogger()
	k2 := keeper.NewKeeper(storeService2, logger2, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", nil, nil, nil, newMockReputationKeeper())
	ctx2 := context.Background()

	retrieved2, err := k2.GetEmissionPool(ctx2)
	if err != nil {
		t.Fatalf("GetEmissionPool failed: %v", err)
	}
	if retrieved2.TotalSupply != pool.TotalSupply {
		t.Fatal("tx2 should see tx1's emission pool after commit")
	}
}

func TestSDKStoreCommitPersist(t *testing.T) {
	storeService1, cms1 := testkeeper.NewTestStoreService(t, emissionsStoreKey)
	logger1 := testkeeper.NewTestLogger()
	k1 := keeper.NewKeeper(storeService1, logger1, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", nil, nil, nil, newMockReputationKeeper())
	ctx1 := context.Background()

	pool := types.NewEmissionPool(1000000, 10000, 80, 20, 0)
	err := k1.SetEmissionPool(ctx1, pool)
	if err != nil {
		t.Fatalf("SetEmissionPool failed: %v", err)
	}

	cms1.Commit()

	storeService2 := testkeeper.NewTestStoreServiceWithCMS(t, emissionsStoreKey, cms1)
	logger2 := testkeeper.NewTestLogger()
	k2 := keeper.NewKeeper(storeService2, logger2, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", nil, nil, nil, newMockReputationKeeper())
	ctx2 := context.Background()

	retrieved, err := k2.GetEmissionPool(ctx2)
	if err != nil {
		t.Fatalf("GetEmissionPool failed: %v", err)
	}
	if retrieved.TotalSupply != pool.TotalSupply {
		t.Errorf("expected TotalSupply %d, got %d", pool.TotalSupply, retrieved.TotalSupply)
	}
}

type mockServeKeeper struct {
	attestations map[string][]*serveTypes.ServeAttestation
}

func (m *mockServeKeeper) GetEpochServeStats(ctx context.Context, orgID string, epoch uint64) (*serveTypes.EpochServeStats, error) {
	return nil, nil
}

func (m *mockServeKeeper) GetServeAttestations(ctx context.Context, orgID string, epoch uint64) ([]*serveTypes.ServeAttestation, error) {
	key := fmt.Sprintf("%s_%d", orgID, epoch)
	return m.attestations[key], nil
}

type mockMemoryKeeper struct {
	approvedByContributor map[string]map[string]uint64
	contributorsByEpoch   map[uint64]map[string]uint64
}

func (m *mockMemoryKeeper) GetApprovedCountByContributor(ctx context.Context, orgID string, epoch uint64) (map[string]uint64, error) {
	key := fmt.Sprintf("%s_%d", orgID, epoch)
	counts, ok := m.approvedByContributor[key]
	if !ok {
		return map[string]uint64{}, nil
	}
	return counts, nil
}

func (m *mockMemoryKeeper) GetContributorsWithApprovalsInEpoch(ctx context.Context, epoch uint64) (map[string]uint64, error) {
	counts, ok := m.contributorsByEpoch[epoch]
	if !ok {
		return map[string]uint64{}, nil
	}
	return counts, nil
}

type mockOrgKeeper struct {
	orgs           []*orgTypes.Org
	configs        map[string]*orgTypes.OrgConfig
	treasuryBal    map[string]math.Int
	treasuryDebits map[string][]math.Int
}

func newMockOrgKeeper() *mockOrgKeeper {
	return &mockOrgKeeper{
		configs:        make(map[string]*orgTypes.OrgConfig),
		treasuryBal:    make(map[string]math.Int),
		treasuryDebits: make(map[string][]math.Int),
	}
}

func (m *mockOrgKeeper) GetAllOrgs(ctx context.Context) ([]*orgTypes.Org, error) {
	return m.orgs, nil
}

func (m *mockOrgKeeper) GetOrgConfig(ctx context.Context, orgID string) (*orgTypes.OrgConfig, error) {
	cfg, ok := m.configs[orgID]
	if !ok {
		return &orgTypes.OrgConfig{OrgID: orgID, ServeAttestationRequired: false}, nil
	}
	return cfg, nil
}

func TestDistributePayout_NoOrgs(t *testing.T) {
	storeService, _ := testkeeper.NewTestStoreService(t, emissionsStoreKey)
	logger := testkeeper.NewTestLogger()
	serveKeeper := &mockServeKeeper{attestations: make(map[string][]*serveTypes.ServeAttestation)}
	memoryKeeper := &mockMemoryKeeper{approvedByContributor: make(map[string]map[string]uint64)}
	orgKeeper := newMockOrgKeeper()
	reputationKeeper := newMockReputationKeeper()
	k := keeper.NewKeeper(storeService, logger, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", serveKeeper, memoryKeeper, orgKeeper, reputationKeeper)
	ctx := context.Background()

	pool := types.NewEmissionPool(1000000, 10000, 80, 20, 0)
	err := k.SetEmissionPool(ctx, pool)
	if err != nil {
		t.Fatalf("SetEmissionPool failed: %v", err)
	}

	err = k.AfterEpochEnd(ctx, keeper.WeVibeEpochIdentifier, 1)
	if err != nil {
		t.Fatalf("AfterEpochEnd failed: %v", err)
	}
}

func TestDistributePayout_OneContributor(t *testing.T) {
	storeService, _ := testkeeper.NewTestStoreService(t, emissionsStoreKey)
	logger := testkeeper.NewTestLogger()
	serveKeeper := &mockServeKeeper{attestations: make(map[string][]*serveTypes.ServeAttestation)}
	memoryKeeper := &mockMemoryKeeper{
		approvedByContributor: map[string]map[string]uint64{
			"org1_1": {"contrib1": 1},
		},
	}
	orgKeeper := newMockOrgKeeper()
	orgKeeper.orgs = []*orgTypes.Org{{OrgID: "org1"}}
	orgKeeper.configs["org1"] = &orgTypes.OrgConfig{OrgID: "org1", ServeAttestationRequired: true, MinContributionsPerEpoch: 1}
	orgKeeper.treasuryBal["org1"] = math.NewInt(1000000)

	k := keeper.NewKeeper(storeService, logger, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", serveKeeper, memoryKeeper, orgKeeper, newMockReputationKeeper())
	ctx := context.Background()

	pool := types.NewEmissionPool(1000000, 10000, 80, 20, 0)
	err := k.SetEmissionPool(ctx, pool)
	if err != nil {
		t.Fatalf("SetEmissionPool failed: %v", err)
	}

	err = k.AfterEpochEnd(ctx, keeper.WeVibeEpochIdentifier, 1)
	if err != nil {
		t.Fatalf("AfterEpochEnd failed: %v", err)
	}

	finalBal := orgKeeper.treasuryBal["org1"]
	expectedBal := math.NewInt(1000000)
	if !finalBal.Equal(expectedBal) {
		t.Errorf("expected treasury balance %s, got %s", expectedBal, finalBal)
	}

	debits := orgKeeper.treasuryDebits["org1"]
	if len(debits) != 0 {
		t.Errorf("expected 0 treasury debits, got %d", len(debits))
	}
}

func TestDistributePayout_TreasuryExhausted(t *testing.T) {
	storeService, _ := testkeeper.NewTestStoreService(t, emissionsStoreKey)
	logger := testkeeper.NewTestLogger()
	serveKeeper := &mockServeKeeper{attestations: make(map[string][]*serveTypes.ServeAttestation)}
	memoryKeeper := &mockMemoryKeeper{
		approvedByContributor: map[string]map[string]uint64{
			"org1_1": {
				"contrib1": 1,
				"contrib2": 1,
			},
		},
	}
	orgKeeper := newMockOrgKeeper()
	orgKeeper.orgs = []*orgTypes.Org{{OrgID: "org1"}}
	orgKeeper.configs["org1"] = &orgTypes.OrgConfig{OrgID: "org1", ServeAttestationRequired: true, MinContributionsPerEpoch: 1}
	orgKeeper.treasuryBal["org1"] = math.NewInt(150)

	k := keeper.NewKeeper(storeService, logger, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", serveKeeper, memoryKeeper, orgKeeper, newMockReputationKeeper())
	ctx := context.Background()

	pool := types.NewEmissionPool(1000000, 10000, 80, 20, 0)
	err := k.SetEmissionPool(ctx, pool)
	if err != nil {
		t.Fatalf("SetEmissionPool failed: %v", err)
	}

	err = k.AfterEpochEnd(ctx, keeper.WeVibeEpochIdentifier, 1)
	if err != nil {
		t.Fatalf("AfterEpochEnd failed: %v", err)
	}

	finalBal := orgKeeper.treasuryBal["org1"]
	if !finalBal.Equal(math.NewInt(150)) {
		t.Errorf("expected treasury balance to remain 150, got %s", finalBal)
	}

	debits := orgKeeper.treasuryDebits["org1"]
	if len(debits) != 0 {
		t.Errorf("expected 0 debits, got %d", len(debits))
	}
}

func TestDistributePayout_ServeAttestationNotRequired(t *testing.T) {
	storeService, _ := testkeeper.NewTestStoreService(t, emissionsStoreKey)
	logger := testkeeper.NewTestLogger()
	serveKeeper := &mockServeKeeper{attestations: make(map[string][]*serveTypes.ServeAttestation)}
	memoryKeeper := &mockMemoryKeeper{
		approvedByContributor: map[string]map[string]uint64{
			"org1_1": {"contrib1": 1},
		},
	}
	orgKeeper := newMockOrgKeeper()
	orgKeeper.orgs = []*orgTypes.Org{{OrgID: "org1"}}
	orgKeeper.configs["org1"] = &orgTypes.OrgConfig{OrgID: "org1", ServeAttestationRequired: false}
	orgKeeper.treasuryBal["org1"] = math.NewInt(1000000)

	k := keeper.NewKeeper(storeService, logger, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", serveKeeper, memoryKeeper, orgKeeper, newMockReputationKeeper())
	ctx := context.Background()

	pool := types.NewEmissionPool(1000000, 10000, 80, 20, 0)
	err := k.SetEmissionPool(ctx, pool)
	if err != nil {
		t.Fatalf("SetEmissionPool failed: %v", err)
	}

	err = k.AfterEpochEnd(ctx, keeper.WeVibeEpochIdentifier, 1)
	if err != nil {
		t.Fatalf("AfterEpochEnd failed: %v", err)
	}

	finalBal := orgKeeper.treasuryBal["org1"]
	if !finalBal.Equal(math.NewInt(1000000)) {
		t.Errorf("expected treasury balance to remain 1000000 (org skipped), got %s", finalBal)
	}

	debits := orgKeeper.treasuryDebits["org1"]
	if len(debits) != 0 {
		t.Errorf("expected 0 debits (org skipped), got %d", len(debits))
	}
}
