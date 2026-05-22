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

func newTestKeeper(t *testing.T) (*keeper.Keeper, context.Context) {
	storeService, _ := testkeeper.NewTestStoreService(t, emissionsStoreKey)
	logger := testkeeper.NewTestLogger()
	k := keeper.NewKeeper(storeService, logger, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", nil, nil, nil, newMockReputationKeeper())
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

	pool := types.NewEmissionPool(1000000, 10000, 80, 20, 0)
	err := k.SetEmissionPool(ctx, pool)
	if err != nil {
		t.Fatalf("SetEmissionPool failed: %v", err)
	}

	emission, err := k.MintDailyEmission(ctx, 1)
	if err != nil {
		t.Fatalf("MintDailyEmission failed: %v", err)
	}
	if emission.Epoch != 1 {
		t.Errorf("expected Epoch 1, got %d", emission.Epoch)
	}
	if emission.TotalEmitted != 10000 {
		t.Errorf("expected TotalEmitted 10000, got %d", emission.TotalEmitted)
	}
	if emission.OperatorShare != 8000 {
		t.Errorf("expected OperatorShare 8000, got %d", emission.OperatorShare)
	}
	if emission.ValidatorShare != 2000 {
		t.Errorf("expected ValidatorShare 2000, got %d", emission.ValidatorShare)
	}

	retrievedPool, _ := k.GetEmissionPool(ctx)
	if retrievedPool.Epoch != 1 {
		t.Errorf("expected pool Epoch 1, got %d", retrievedPool.Epoch)
	}
	if retrievedPool.TotalSupply != 1010000 {
		t.Errorf("expected pool TotalSupply 1010000, got %d", retrievedPool.TotalSupply)
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

func TestComputeWorkScore(t *testing.T) {
	k, ctx := newTestKeeper(t)

	score, err := k.ComputeWorkScore(ctx, "op1", "org1", 1.5, 0.9, 100, 1)
	if err != nil {
		t.Fatalf("ComputeWorkScore failed: %v", err)
	}

	if score.OperatorID != "op1" {
		t.Errorf("expected OperatorID op1, got %s", score.OperatorID)
	}
	if score.OrgID != "org1" {
		t.Errorf("expected OrgID org1, got %s", score.OrgID)
	}
	if score.RarityMultiplier != 1.5 {
		t.Errorf("expected RarityMultiplier 1.5, got %f", score.RarityMultiplier)
	}
	if score.AvailabilityScore != 0.9 {
		t.Errorf("expected AvailabilityScore 0.9, got %f", score.AvailabilityScore)
	}
	if score.RetrievalVolume != 100 {
		t.Errorf("expected RetrievalVolume 100, got %d", score.RetrievalVolume)
	}

	expectedStorageScore := 0.3 * 0.9
	if score.StorageScore != expectedStorageScore {
		t.Errorf("expected StorageScore %f, got %f", expectedStorageScore, score.StorageScore)
	}

	expectedTotalScore := 1.5 * (0.3*0.9 + 0.7*100)
	if score.TotalScore != expectedTotalScore {
		t.Errorf("expected TotalScore %f, got %f", expectedTotalScore, score.TotalScore)
	}
}

func TestComputeWorkScore_InvalidInput(t *testing.T) {
	k, ctx := newTestKeeper(t)

	_, err := k.ComputeWorkScore(ctx, "", "org1", 1.5, 0.9, 100, 1)
	if err == nil {
		t.Fatal("expected error for empty operatorID")
	}

	_, err = k.ComputeWorkScore(ctx, "op1", "", 1.5, 0.9, 100, 1)
	if err == nil {
		t.Fatal("expected error for empty orgID")
	}
}

func TestGetWorkScore(t *testing.T) {
	k, ctx := newTestKeeper(t)

	_, err := k.ComputeWorkScore(ctx, "op1", "org1", 1.5, 0.9, 100, 1)
	if err != nil {
		t.Fatalf("ComputeWorkScore failed: %v", err)
	}

	score, err := k.GetWorkScore(ctx, "op1", "org1", 1)
	if err != nil {
		t.Fatalf("GetWorkScore failed: %v", err)
	}
	if score.OperatorID != "op1" {
		t.Errorf("expected OperatorID op1, got %s", score.OperatorID)
	}
	if score.OrgID != "org1" {
		t.Errorf("expected OrgID org1, got %s", score.OrgID)
	}
}

func TestGetWorkScore_NotFound(t *testing.T) {
	k, ctx := newTestKeeper(t)

	_, err := k.GetWorkScore(ctx, "op1", "org1", 1)
	if err == nil {
		t.Fatal("expected error for nonexistent work score")
	}
}

func TestComputeTotalWorkScore(t *testing.T) {
	k, ctx := newTestKeeper(t)

	_, err := k.ComputeWorkScore(ctx, "op1", "org1", 1.0, 1.0, 100, 1)
	if err != nil {
		t.Fatalf("ComputeWorkScore failed: %v", err)
	}

	_, err = k.ComputeWorkScore(ctx, "op1", "org2", 1.0, 1.0, 50, 1)
	if err != nil {
		t.Fatalf("ComputeWorkScore failed: %v", err)
	}

	total, err := k.ComputeTotalWorkScore(ctx, "op1", 1)
	if err != nil {
		t.Fatalf("ComputeTotalWorkScore failed: %v", err)
	}

	if total <= 0 {
		t.Errorf("expected positive total work score, got %f", total)
	}
}

func TestDistributeOperatorRewards(t *testing.T) {
	k, ctx := newTestKeeper(t)

	pool := types.NewEmissionPool(1000000, 10000, 80, 20, 0)
	err := k.SetEmissionPool(ctx, pool)
	if err != nil {
		t.Fatalf("SetEmissionPool failed: %v", err)
	}

	_, err = k.MintDailyEmission(ctx, 1)
	if err != nil {
		t.Fatalf("MintDailyEmission failed: %v", err)
	}

	rewards := map[string]uint64{
		"op1": 5000,
		"op2": 3000,
	}

	err = k.DistributeOperatorRewards(ctx, rewards, 1)
	if err != nil {
		t.Fatalf("DistributeOperatorRewards failed: %v", err)
	}

	reward1, err := k.GetOperatorReward(ctx, "op1")
	if err != nil {
		t.Fatalf("GetOperatorReward failed: %v", err)
	}
	if reward1 != 5000 {
		t.Errorf("expected reward1 5000, got %d", reward1)
	}

	reward2, err := k.GetOperatorReward(ctx, "op2")
	if err != nil {
		t.Fatalf("GetOperatorReward failed: %v", err)
	}
	if reward2 != 3000 {
		t.Errorf("expected reward2 3000, got %d", reward2)
	}

	emission, _ := k.GetDailyEmission(ctx, 1)
	if emission.OperatorRewards["op1"] != 5000 {
		t.Errorf("expected emission.OperatorRewards[op1] 5000, got %d", emission.OperatorRewards["op1"])
	}
	if emission.OperatorRewards["op2"] != 3000 {
		t.Errorf("expected emission.OperatorRewards[op2] 3000, got %d", emission.OperatorRewards["op2"])
	}
}

func TestDistributeValidatorRewards(t *testing.T) {
	k, ctx := newTestKeeper(t)

	pool := types.NewEmissionPool(1000000, 10000, 80, 20, 0)
	err := k.SetEmissionPool(ctx, pool)
	if err != nil {
		t.Fatalf("SetEmissionPool failed: %v", err)
	}

	_, err = k.MintDailyEmission(ctx, 1)
	if err != nil {
		t.Fatalf("MintDailyEmission failed: %v", err)
	}

	rewards := map[string]uint64{
		"val1": 1500,
		"val2": 500,
	}

	err = k.DistributeValidatorRewards(ctx, rewards, 1)
	if err != nil {
		t.Fatalf("DistributeValidatorRewards failed: %v", err)
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

func TestGetOperatorWorkScores(t *testing.T) {
	k, ctx := newTestKeeper(t)

	_, err := k.ComputeWorkScore(ctx, "op1", "org1", 1.0, 1.0, 100, 1)
	if err != nil {
		t.Fatalf("ComputeWorkScore failed: %v", err)
	}

	_, err = k.ComputeWorkScore(ctx, "op1", "org2", 1.0, 1.0, 50, 1)
	if err != nil {
		t.Fatalf("ComputeWorkScore failed: %v", err)
	}

	_, err = k.ComputeWorkScore(ctx, "op2", "org1", 1.0, 1.0, 75, 1)
	if err != nil {
		t.Fatalf("ComputeWorkScore failed: %v", err)
	}

	scores, err := k.GetOperatorWorkScores(ctx, "op1", 1)
	if err != nil {
		t.Fatalf("GetOperatorWorkScores failed: %v", err)
	}
	if len(scores) != 2 {
		t.Errorf("expected 2 scores, got %d", len(scores))
	}
}

func TestGetAllWorkScores(t *testing.T) {
	k, ctx := newTestKeeper(t)

	_, err := k.ComputeWorkScore(ctx, "op1", "org1", 1.0, 1.0, 100, 1)
	if err != nil {
		t.Fatalf("ComputeWorkScore failed: %v", err)
	}

	_, err = k.ComputeWorkScore(ctx, "op1", "org2", 1.0, 1.0, 50, 1)
	if err != nil {
		t.Fatalf("ComputeWorkScore failed: %v", err)
	}

	_, err = k.ComputeWorkScore(ctx, "op2", "org1", 1.0, 1.0, 75, 2)
	if err != nil {
		t.Fatalf("ComputeWorkScore failed: %v", err)
	}

	scores, err := k.GetAllWorkScores(ctx, 1)
	if err != nil {
		t.Fatalf("GetAllWorkScores failed: %v", err)
	}
	if len(scores) != 2 {
		t.Errorf("expected 2 scores for epoch 1, got %d", len(scores))
	}

	scores, err = k.GetAllWorkScores(ctx, 2)
	if err != nil {
		t.Fatalf("GetAllWorkScores failed: %v", err)
	}
	if len(scores) != 1 {
		t.Errorf("expected 1 score for epoch 2, got %d", len(scores))
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
}

func (m *mockMemoryKeeper) GetApprovedCountByContributor(ctx context.Context, orgID string, epoch uint64) (map[string]uint64, error) {
	key := fmt.Sprintf("%s_%d", orgID, epoch)
	counts, ok := m.approvedByContributor[key]
	if !ok {
		return map[string]uint64{}, nil
	}
	return counts, nil
}

type mockOrgKeeper struct {
	orgs           []*orgTypes.Org
	configs        map[string]*orgTypes.OrgConfig
	treasuryBal    map[string]math.Int
	repTiers       map[string]*orgTypes.RepTierConfig
	treasuryDebits map[string][]math.Int
}

func newMockOrgKeeper() *mockOrgKeeper {
	return &mockOrgKeeper{
		configs:        make(map[string]*orgTypes.OrgConfig),
		treasuryBal:    make(map[string]math.Int),
		repTiers:       make(map[string]*orgTypes.RepTierConfig),
		treasuryDebits: make(map[string][]math.Int),
	}
}

func (m *mockOrgKeeper) GetAllOrgs(ctx context.Context) ([]*orgTypes.Org, error) {
	return m.orgs, nil
}

func (m *mockOrgKeeper) GetTreasuryBalanceInt(ctx context.Context, orgID string) (math.Int, error) {
	bal, ok := m.treasuryBal[orgID]
	if !ok {
		return math.ZeroInt(), nil
	}
	return bal, nil
}

func (m *mockOrgKeeper) DebitTreasury(ctx context.Context, orgID string, amount math.Int) error {
	m.treasuryBal[orgID] = m.treasuryBal[orgID].Sub(amount)
	m.treasuryDebits[orgID] = append(m.treasuryDebits[orgID], amount)
	return nil
}

func (m *mockOrgKeeper) GetRepTiers(ctx context.Context, orgID string) (*orgTypes.RepTierConfig, error) {
	tiers, ok := m.repTiers[orgID]
	if !ok {
		return nil, fmt.Errorf("no rep tiers for org")
	}
	return tiers, nil
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
	orgKeeper.repTiers["org1"] = &orgTypes.RepTierConfig{
		OrgID: "org1",
		Tiers: []*orgTypes.RepTierRecord{
			{MinReputation: 0, MaxReputation: 100, PayoutPerMemory: "100"},
		},
	}

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
	expectedBal := math.NewInt(999900)
	if !finalBal.Equal(expectedBal) {
		t.Errorf("expected treasury balance %s, got %s", expectedBal, finalBal)
	}

	debits := orgKeeper.treasuryDebits["org1"]
	if len(debits) != 1 {
		t.Errorf("expected 1 treasury debit, got %d", len(debits))
	}
	if !debits[0].Equal(math.NewInt(100)) {
		t.Errorf("expected debit of 100, got %s", debits[0])
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
	orgKeeper.repTiers["org1"] = &orgTypes.RepTierConfig{
		OrgID: "org1",
		Tiers: []*orgTypes.RepTierRecord{
			{MinReputation: 0, MaxReputation: 100, PayoutPerMemory: "100"},
		},
	}

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
	if !finalBal.Equal(math.NewInt(50)) {
		t.Errorf("expected treasury balance to be 50 after paying contrib1, got %s", finalBal)
	}

	debits := orgKeeper.treasuryDebits["org1"]
	if len(debits) != 1 {
		t.Errorf("expected exactly 1 debit (contrib1 only), got %d", len(debits))
	}
	if !debits[0].Equal(math.NewInt(100)) {
		t.Errorf("expected debit of 100, got %s", debits[0])
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
	orgKeeper.repTiers["org1"] = &orgTypes.RepTierConfig{
		OrgID: "org1",
		Tiers: []*orgTypes.RepTierRecord{
			{MinReputation: 0, MaxReputation: 100, PayoutPerMemory: "100"},
		},
	}

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
