package emissions_test

import (
	"encoding/json"
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	testkeeper "github.com/wevibe-network/wevibe-chain/testutil/keeper"
	"github.com/wevibe-network/wevibe-chain/x/emissions/keeper"
	emissions "github.com/wevibe-network/wevibe-chain/x/emissions/module"
	"github.com/wevibe-network/wevibe-chain/x/emissions/types"
)

func newGenesisTestModule(t *testing.T) (*emissions.Module, *keeper.Keeper, sdk.Context) {
	t.Helper()
	storeKey := storetypes.NewKVStoreKey("emissions")
	storeService, cms := testkeeper.NewTestStoreService(t, storeKey)
	logger := testkeeper.NewTestLogger()
	k := keeper.NewKeeper(storeService, logger, "gov", nil, nil, nil, nil)
	ctx := sdk.NewContext(cms, tmproto.Header{Height: 1, Time: time.Now().UTC()}, false, logger)
	return emissions.NewModule(k), k, ctx
}

// TestDefaultGenesis_SeedsPoolFromDefaultParams verifies the default genesis
// JSON contains an emission pool whose values come from DefaultParams().
func TestDefaultGenesis_SeedsPoolFromDefaultParams(t *testing.T) {
	mod, _, _ := newGenesisTestModule(t)

	bz := mod.DefaultGenesis(nil)
	require.NotEmpty(t, bz)

	var state types.GenesisState
	require.NoError(t, json.Unmarshal(bz, &state))
	require.NotNil(t, state.EmissionPool, "default genesis must seed an emission pool")

	p := types.DefaultParams()
	require.Equal(t, p.DailyMintAmount, state.EmissionPool.DailyMint)
	require.Equal(t, p.OperatorSharePercent, state.EmissionPool.OperatorShare)
	require.Equal(t, p.ValidatorSharePercent, state.EmissionPool.ValidatorShare)
}

// TestInitGenesis_SeedsPool verifies that running InitGenesis with the default
// genesis JSON persists a usable emission pool — the fix for the
// "no emission pool found" epoch-hook log.
func TestInitGenesis_SeedsPool(t *testing.T) {
	mod, k, ctx := newGenesisTestModule(t)

	mod.InitGenesis(ctx, nil, mod.DefaultGenesis(nil))

	pool, err := k.GetEmissionPool(ctx)
	require.NoError(t, err)
	require.NotNil(t, pool)
	require.Equal(t, types.DefaultParams().DailyMintAmount, pool.DailyMint)

	// Params must also be persisted.
	params, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), params)
}

// TestInitGenesis_EmptyObjectFallsBackToDefaultPool verifies the robustness
// path used by init-chain.sh, which seeds app_state.emissions = {}. The module
// must derive the pool from DefaultParams() in that case.
func TestInitGenesis_EmptyObjectFallsBackToDefaultPool(t *testing.T) {
	mod, k, ctx := newGenesisTestModule(t)

	mod.InitGenesis(ctx, nil, json.RawMessage(`{}`))

	pool, err := k.GetEmissionPool(ctx)
	require.NoError(t, err)
	require.NotNil(t, pool)
	require.Equal(t, types.DefaultParams().DailyMintAmount, pool.DailyMint)
	require.Equal(t, uint64(100), pool.OperatorShare+pool.ValidatorShare)
}

// TestInitGenesis_NilBytesFallsBackToDefaultPool covers the case where the
// genesis blob is absent entirely.
func TestInitGenesis_NilBytesFallsBackToDefaultPool(t *testing.T) {
	mod, k, ctx := newGenesisTestModule(t)

	mod.InitGenesis(ctx, nil, nil)

	pool, err := k.GetEmissionPool(ctx)
	require.NoError(t, err)
	require.NotNil(t, pool)
	require.Equal(t, types.DefaultParams().DailyMintAmount, pool.DailyMint)
}

// TestGenesisRoundTrip verifies Export after Init reproduces the seeded pool.
func TestGenesisRoundTrip(t *testing.T) {
	mod, _, ctx := newGenesisTestModule(t)

	mod.InitGenesis(ctx, nil, mod.DefaultGenesis(nil))
	exported := mod.ExportGenesis(ctx, nil)

	var state types.GenesisState
	require.NoError(t, json.Unmarshal(exported, &state))
	require.NotNil(t, state.EmissionPool)
	require.Equal(t, types.DefaultParams().DailyMintAmount, state.EmissionPool.DailyMint)
}

func TestValidateGenesis(t *testing.T) {
	mod, _, _ := newGenesisTestModule(t)

	// Default genesis is valid.
	require.NoError(t, mod.ValidateGenesis(nil, nil, mod.DefaultGenesis(nil)))

	// Empty bytes are valid (treated as default).
	require.NoError(t, mod.ValidateGenesis(nil, nil, nil))

	// A pool whose shares do not sum to 100 is invalid.
	bad := types.GenesisState{EmissionPool: &types.EmissionPool{
		DailyMint:      1000000000,
		OperatorShare:  70,
		ValidatorShare: 20,
	}}
	badBz, err := json.Marshal(bad)
	require.NoError(t, err)
	require.Error(t, mod.ValidateGenesis(nil, nil, badBz))
}
