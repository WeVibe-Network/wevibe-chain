package keeper

import (
	"context"
	"testing"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/wevibe-network/wevibe-chain/testutil/keeper"
	"github.com/wevibe-network/wevibe-chain/x/bandwidth/types"
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

func makeTestKeeper(t *testing.T) (*Keeper, *mockOrgKeeper) {
	storeKey := storetypes.NewKVStoreKey("bandwidth")
	storeService, _ := keeper.NewTestStoreService(t, storeKey)
	logger := keeper.NewTestLogger()
	mockOrg := &mockOrgKeeper{
		orgs:    map[string]bool{"test-org": true, "other-org": true},
		leaders: map[string]string{"test-org": "leader-pubkey", "other-org": "other-leader"},
	}
	return NewKeeper(storeService, logger, "gov", mockOrg), mockOrg
}

func TestConsumeMemoryBandwidth_HappyPath(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()

	err := k.ConsumeMemoryBandwidth(ctx, "test-org", 1)
	if err != nil {
		t.Fatalf("ConsumeMemoryBandwidth failed: %v", err)
	}

	state, _ := k.GetBandwidthState(ctx, "test-org", 1)
	if state.MemoryUsed != 1 {
		t.Errorf("MemoryUsed mismatch: got %d, want 1", state.MemoryUsed)
	}
	if state.MemoryCap != 10000 {
		t.Errorf("MemoryCap mismatch: got %d, want 10000", state.MemoryCap)
	}
}

func TestConsumeMemoryBandwidth_Exhausted(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()

	for i := uint64(0); i < 10000; i++ {
		err := k.ConsumeMemoryBandwidth(ctx, "test-org", 1)
		if err != nil {
			t.Fatalf("ConsumeMemoryBandwidth failed at iteration %d: %v", i, err)
		}
	}

	err := k.ConsumeMemoryBandwidth(ctx, "test-org", 1)
	if err != types.ErrMemoryBandwidthExhausted {
		t.Fatalf("expected ErrMemoryBandwidthExhausted, got: %v", err)
	}
}

func TestConsumeServeBandwidth_HappyPath(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()

	err := k.ConsumeServeBandwidth(ctx, "test-org", 1, 5)
	if err != nil {
		t.Fatalf("ConsumeServeBandwidth failed: %v", err)
	}

	state, _ := k.GetBandwidthState(ctx, "test-org", 1)
	if state.ServeUsed != 5 {
		t.Errorf("ServeUsed mismatch: got %d, want 5", state.ServeUsed)
	}
}

func TestConsumeServeBandwidth_Exhausted(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()

	k.ConsumeServeBandwidth(ctx, "test-org", 1, 49999)

	err := k.ConsumeServeBandwidth(ctx, "test-org", 1, 2)
	if err != types.ErrServeBandwidthExhausted {
		t.Fatalf("expected ErrServeBandwidthExhausted, got: %v", err)
	}
}

func TestConsumeServeBandwidth_ExactCap(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()

	err := k.ConsumeServeBandwidth(ctx, "test-org", 1, 50000)
	if err != nil {
		t.Fatalf("ConsumeServeBandwidth failed: %v", err)
	}

	state, _ := k.GetBandwidthState(ctx, "test-org", 1)
	if state.ServeUsed != 50000 {
		t.Errorf("ServeUsed mismatch: got %d, want 50000", state.ServeUsed)
	}
}

func TestLazyInit_DefaultParams(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()

	state, err := k.GetOrInitBandwidthState(ctx, "test-org", 1)
	if err != nil {
		t.Fatalf("GetOrInitBandwidthState failed: %v", err)
	}
	if state.MemoryCap != 10000 {
		t.Errorf("MemoryCap mismatch: got %d, want 10000", state.MemoryCap)
	}
	if state.ServeCap != 50000 {
		t.Errorf("ServeCap mismatch: got %d, want 50000", state.ServeCap)
	}
}

func TestLazyInit_WithOverride(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()

	k.SetBandwidthOverride(ctx, "test-org", 20000, 100000)

	state, err := k.GetOrInitBandwidthState(ctx, "test-org", 1)
	if err != nil {
		t.Fatalf("GetOrInitBandwidthState failed: %v", err)
	}
	if state.MemoryCap != 20000 {
		t.Errorf("MemoryCap mismatch: got %d, want 20000", state.MemoryCap)
	}
	if state.ServeCap != 100000 {
		t.Errorf("ServeCap mismatch: got %d, want 100000", state.ServeCap)
	}
}

func TestSetBandwidthOverride(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()

	err := k.SetBandwidthOverride(ctx, "test-org", 20000, 100000)
	if err != nil {
		t.Fatalf("SetBandwidthOverride failed: %v", err)
	}

	override, err := k.GetBandwidthOverride(ctx, "test-org")
	if err != nil {
		t.Fatalf("GetBandwidthOverride failed: %v", err)
	}
	if override.MemoryCap != 20000 {
		t.Errorf("MemoryCap mismatch: got %d, want 20000", override.MemoryCap)
	}
	if override.ServeCap != 100000 {
		t.Errorf("ServeCap mismatch: got %d, want 100000", override.ServeCap)
	}
}

func TestDeleteBandwidthOverride(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()

	k.SetBandwidthOverride(ctx, "test-org", 20000, 100000)
	k.DeleteBandwidthOverride(ctx, "test-org")

	_, err := k.GetBandwidthOverride(ctx, "test-org")
	if err != types.ErrOverrideNotFound {
		t.Fatalf("expected ErrOverrideNotFound, got: %v", err)
	}
}

func TestOverrideDoesNotAffectExistingEpoch(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()

	k.ConsumeMemoryBandwidth(ctx, "test-org", 1)
	state1, _ := k.GetBandwidthState(ctx, "test-org", 1)
	if state1.MemoryUsed != 1 {
		t.Errorf("State1 MemoryUsed mismatch: got %d, want 1", state1.MemoryUsed)
	}

	k.SetBandwidthOverride(ctx, "test-org", 20000, 100000)

	state2, _ := k.GetOrInitBandwidthState(ctx, "test-org", 1)
	if state2.MemoryUsed != 1 {
		t.Errorf("State2 MemoryUsed mismatch after override: got %d, want 1", state2.MemoryUsed)
	}
	if state2.MemoryCap != 10000 {
		t.Errorf("State2 MemoryCap mismatch after override: got %d, want 10000", state2.MemoryCap)
	}
}

func TestGetRemainingBandwidth(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()

	k.ConsumeMemoryBandwidth(ctx, "test-org", 1)
	k.ConsumeMemoryBandwidth(ctx, "test-org", 1)

	memRem, serveRem, err := k.GetRemainingBandwidth(ctx, "test-org", 1)
	if err != nil {
		t.Fatalf("GetRemainingBandwidth failed: %v", err)
	}
	if memRem != 9998 {
		t.Errorf("MemoryRemaining mismatch: got %d, want 9998", memRem)
	}
	if serveRem != 50000 {
		t.Errorf("ServeRemaining mismatch: got %d, want 50000", serveRem)
	}
}

func TestGetRemainingBandwidth_Fresh(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()

	memRem, serveRem, err := k.GetRemainingBandwidth(ctx, "test-org", 1)
	if err != nil {
		t.Fatalf("GetRemainingBandwidth failed: %v", err)
	}
	if memRem != 10000 {
		t.Errorf("MemoryRemaining mismatch: got %d, want 10000", memRem)
	}
	if serveRem != 50000 {
		t.Errorf("ServeRemaining mismatch: got %d, want 50000", serveRem)
	}
}

func TestMultipleOrgs(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()

	k.ConsumeMemoryBandwidth(ctx, "test-org", 1)
	k.ConsumeMemoryBandwidth(ctx, "other-org", 1)

	state1, _ := k.GetBandwidthState(ctx, "test-org", 1)
	state2, _ := k.GetBandwidthState(ctx, "other-org", 1)

	if state1.MemoryUsed != 1 {
		t.Errorf("test-org MemoryUsed mismatch: got %d, want 1", state1.MemoryUsed)
	}
	if state2.MemoryUsed != 1 {
		t.Errorf("other-org MemoryUsed mismatch: got %d, want 1", state2.MemoryUsed)
	}
}

func TestMultipleEpochs(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()

	k.ConsumeMemoryBandwidth(ctx, "test-org", 1)
	k.ConsumeMemoryBandwidth(ctx, "test-org", 2)
	k.ConsumeMemoryBandwidth(ctx, "test-org", 3)

	state1, _ := k.GetBandwidthState(ctx, "test-org", 1)
	state2, _ := k.GetBandwidthState(ctx, "test-org", 2)
	state3, _ := k.GetBandwidthState(ctx, "test-org", 3)

	if state1.MemoryUsed != 1 || state2.MemoryUsed != 1 || state3.MemoryUsed != 1 {
		t.Errorf("MemoryUsed mismatch across epochs")
	}
}

func TestInitGenesisExportGenesis(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()

	k.ConsumeMemoryBandwidth(ctx, "test-org", 1)
	k.ConsumeMemoryBandwidth(ctx, "test-org", 1)
	k.SetBandwidthOverride(ctx, "test-org", 20000, 100000)

	override, err := k.GetBandwidthOverride(ctx, "test-org")
	if err != nil {
		t.Fatalf("GetBandwidthOverride failed: %v", err)
	}
	t.Logf("Before export - Override: %+v", override)

	genesis, err := k.ExportGenesis(ctx)
	if err != nil {
		t.Fatalf("ExportGenesis failed: %v", err)
	}

	k2, _ := makeTestKeeper(t)
	ctx2 := context.Background()
	err = k2.InitGenesis(ctx2, genesis)
	if err != nil {
		t.Fatalf("InitGenesis failed: %v", err)
	}

	state, _ := k2.GetBandwidthState(ctx2, "test-org", 1)
	if state.MemoryUsed != 2 {
		t.Errorf("MemoryUsed mismatch after round-trip: got %d, want 2", state.MemoryUsed)
	}

	override2, err := k2.GetBandwidthOverride(ctx2, "test-org")
	if err != nil {
		t.Fatalf("GetBandwidthOverride after InitGenesis failed: %v", err)
	}
	t.Logf("After InitGenesis - Override: %+v", override2)
	if override2.MemoryCap != 20000 {
		t.Errorf("Override MemoryCap mismatch after round-trip: got %d, want 20000", override2.MemoryCap)
	}
}

func TestSetParams(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()

	err := k.SetParams(ctx, types.Params{
		DefaultMemoryCapPerEpoch: 50000,
		DefaultServeCapPerEpoch:  100000,
	})
	if err != nil {
		t.Fatalf("SetParams failed: %v", err)
	}

	params, _ := k.GetParams(ctx)
	if params.DefaultMemoryCapPerEpoch != 50000 {
		t.Errorf("DefaultMemoryCapPerEpoch mismatch: got %d, want 50000", params.DefaultMemoryCapPerEpoch)
	}
	if params.DefaultServeCapPerEpoch != 100000 {
		t.Errorf("DefaultServeCapPerEpoch mismatch: got %d, want 100000", params.DefaultServeCapPerEpoch)
	}
}

func TestGetBandwidthState_NotFound(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()

	state, err := k.GetBandwidthState(ctx, "test-org", 99)
	if err != nil {
		t.Fatalf("GetBandwidthState failed: %v", err)
	}
	if state != nil {
		t.Errorf("expected nil state for non-existent, got %+v", state)
	}
}
