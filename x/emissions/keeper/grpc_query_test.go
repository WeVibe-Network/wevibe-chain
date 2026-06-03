package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wevibe-network/wevibe-chain/x/emissions/keeper"
	"github.com/wevibe-network/wevibe-chain/x/emissions/types"
)

// ---------------------------------------------------------------------------
// GetEmissionPool query
// ---------------------------------------------------------------------------

func TestQueryGetEmissionPool_Success(t *testing.T) {
	k, ctx := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	pool := types.NewEmissionPool(1000000, 10000, 80, 20, 3)
	require.NoError(t, k.SetEmissionPool(ctx, pool))

	resp, err := qs.GetEmissionPool(ctx, &types.QueryGetEmissionPoolRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(1000000), resp.TotalSupply)
	require.Equal(t, uint64(10000), resp.DailyMint)
	require.Equal(t, uint64(80), resp.OperatorShare)
	require.Equal(t, uint64(20), resp.ValidatorShare)
	require.Equal(t, uint64(3), resp.Epoch)
}

func TestQueryGetEmissionPool_NotFound(t *testing.T) {
	k, ctx := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	resp, err := qs.GetEmissionPool(ctx, &types.QueryGetEmissionPoolRequest{})
	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorIs(t, err, types.ErrNoEmissionPool)
}

// ---------------------------------------------------------------------------
// Params query
// ---------------------------------------------------------------------------

func TestQueryParams_DefaultWhenUnset(t *testing.T) {
	k, ctx := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	// GetParams returns DefaultParams when the store has no params set.
	resp, err := qs.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Params)

	def := types.DefaultParams()
	require.Equal(t, def.DailyMintAmount, resp.Params.DailyMintAmount)
	require.Equal(t, def.OperatorSharePercent, resp.Params.OperatorSharePercent)
	require.Equal(t, def.ValidatorSharePercent, resp.Params.ValidatorSharePercent)
	require.Equal(t, def.StorageWeightPercent, resp.Params.StorageWeightPercent)
	require.Equal(t, def.RetrievalWeightPercent, resp.Params.RetrievalWeightPercent)
	require.Equal(t, def.RarityMultiplierCap, resp.Params.RarityMultiplierCap)
	require.Equal(t, def.BootstrapDurationEpochs, resp.Params.BootstrapDurationEpochs)
}
