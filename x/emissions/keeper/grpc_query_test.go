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
// GetWorkScore query
// ---------------------------------------------------------------------------

func TestQueryGetWorkScore_Success(t *testing.T) {
	k, ctx := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	_, err := k.ComputeWorkScore(ctx, "op1", "org1", 1.5, 0.9, 100, 1)
	require.NoError(t, err)

	resp, err := qs.GetWorkScore(ctx, &types.QueryGetWorkScoreRequest{
		OperatorId: "op1",
		OrgId:      "org1",
		Epoch:      1,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "op1", resp.OperatorId)
	require.Equal(t, "org1", resp.OrgId)
	require.Equal(t, 1.5, resp.RarityMultiplier)
	require.Equal(t, 0.9, resp.AvailabilityScore)
	require.Equal(t, uint64(100), resp.RetrievalVolume)
	require.Equal(t, uint64(1), resp.Epoch)
}

func TestQueryGetWorkScore_NotFound(t *testing.T) {
	k, ctx := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	resp, err := qs.GetWorkScore(ctx, &types.QueryGetWorkScoreRequest{
		OperatorId: "missing",
		OrgId:      "org1",
		Epoch:      1,
	})
	require.Error(t, err)
	require.Nil(t, resp)
}

// ---------------------------------------------------------------------------
// GetOperatorReward query
// ---------------------------------------------------------------------------

func TestQueryGetOperatorReward_Success(t *testing.T) {
	k, ctx := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	pool := types.NewEmissionPool(1000000, 10000, 80, 20, 0)
	require.NoError(t, k.SetEmissionPool(ctx, pool))
	_, err := k.MintDailyEmission(ctx, 1)
	require.NoError(t, err)
	require.NoError(t, k.DistributeOperatorRewards(ctx, map[string]uint64{"op1": 5000}, 1))

	resp, err := qs.GetOperatorReward(ctx, &types.QueryGetOperatorRewardRequest{OperatorId: "op1"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(5000), resp.Amount)
}

func TestQueryGetOperatorReward_NotFound(t *testing.T) {
	k, ctx := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	resp, err := qs.GetOperatorReward(ctx, &types.QueryGetOperatorRewardRequest{OperatorId: "missing"})
	require.Error(t, err)
	require.Nil(t, resp)
	require.ErrorIs(t, err, types.ErrNoPendingReward)
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
