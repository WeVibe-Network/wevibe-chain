package reputation_test

import (
	"encoding/json"
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	testkeeper "github.com/wevibe-network/wevibe-chain/testutil/keeper"
	"github.com/wevibe-network/wevibe-chain/x/reputation/keeper"
	reputation "github.com/wevibe-network/wevibe-chain/x/reputation/module"
	"github.com/wevibe-network/wevibe-chain/x/reputation/types"
)

func newGenesisTestModule(t *testing.T) (*reputation.Module, *keeper.Keeper, sdk.Context) {
	t.Helper()
	storeKey := storetypes.NewKVStoreKey("reputation")
	storeService, cms := testkeeper.NewTestStoreService(t, storeKey)
	logger := testkeeper.NewTestLogger()
	k := keeper.NewKeeper(storeService, logger, "gov")
	ctx := sdk.NewContext(cms, tmproto.Header{Height: 1, Time: time.Now().UTC()}, false, logger)
	return reputation.NewModule(k), k, ctx
}

// TestDefaultGenesis_ActiveTrue verifies the default genesis activates the
// module (the GAP-REP-1 fix: previously DefaultGenesis returned Active:false).
func TestDefaultGenesis_ActiveTrue(t *testing.T) {
	mod, _, _ := newGenesisTestModule(t)

	bz := mod.DefaultGenesis(nil)
	require.NotEmpty(t, bz)

	var state types.GenesisState
	require.NoError(t, json.Unmarshal(bz, &state))
	require.True(t, state.Active, "default genesis must activate the reputation module")
}

// TestInitGenesis_PersistsActiveAndParams verifies InitGenesis activates the
// module and persists default params.
func TestInitGenesis_PersistsActiveAndParams(t *testing.T) {
	mod, k, ctx := newGenesisTestModule(t)

	mod.InitGenesis(ctx, nil, mod.DefaultGenesis(nil))

	require.True(t, k.IsActive(ctx), "module must be active after InitGenesis with default genesis")

	params, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), params)
}

// TestInitGenesis_RespectsSuppliedActiveFlag verifies an explicit active flag
// in the genesis blob is honored.
func TestInitGenesis_RespectsSuppliedActiveFlag(t *testing.T) {
	mod, k, ctx := newGenesisTestModule(t)

	mod.InitGenesis(ctx, nil, json.RawMessage(`{"active": true}`))
	require.True(t, k.IsActive(ctx))
}

// TestGenesisRoundTrip verifies Export after Init reports the module as active.
func TestGenesisRoundTrip(t *testing.T) {
	mod, _, ctx := newGenesisTestModule(t)

	mod.InitGenesis(ctx, nil, mod.DefaultGenesis(nil))
	exported := mod.ExportGenesis(ctx, nil)

	var state types.GenesisState
	require.NoError(t, state.UnmarshalJSON(exported))
	require.True(t, state.Active)
}

func TestValidateGenesis(t *testing.T) {
	mod, _, _ := newGenesisTestModule(t)

	require.NoError(t, mod.ValidateGenesis(nil, nil, mod.DefaultGenesis(nil)))
	require.NoError(t, mod.ValidateGenesis(nil, nil, nil))
}
