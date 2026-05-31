package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wevibe-network/wevibe-chain/x/serve/keeper"
	"github.com/wevibe-network/wevibe-chain/x/serve/types"
)

func TestQueryParams(t *testing.T) {
	env := setupKeeper(t)
	q := keeper.NewQueryServerImpl(env.k)

	resp, err := q.Params(env.ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp.Params)
	require.Equal(t, types.DefaultParams().MaxServesPerBatch, resp.Params.MaxServesPerBatch)
}

func TestQueryGetEpochServeStats(t *testing.T) {
	env := setupKeeper(t)
	q := keeper.NewQueryServerImpl(env.k)
	h := hash32(0x01)
	env.mem.approve("org-1", h)

	_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry(h, "sk", "c1", nullifier32(0x01)),
	})
	require.NoError(t, err)

	resp, err := q.GetEpochServeStats(env.ctx, &types.QueryGetEpochServeStatsRequest{OrgId: "org-1", Epoch: 1})
	require.NoError(t, err)
	require.NotNil(t, resp.Stats)
	require.Equal(t, uint64(1), resp.Stats.TotalServes)
	require.Equal(t, "org-1", resp.Stats.OrgId)
}

func TestQueryGetEpochServeStats_NotFound(t *testing.T) {
	env := setupKeeper(t)
	q := keeper.NewQueryServerImpl(env.k)
	_, err := q.GetEpochServeStats(env.ctx, &types.QueryGetEpochServeStatsRequest{OrgId: "org-1", Epoch: 77})
	require.Error(t, err)
}

func TestQueryGetContributorServes(t *testing.T) {
	env := setupKeeper(t)
	q := keeper.NewQueryServerImpl(env.k)
	h := hash32(0x02)
	env.mem.approve("org-1", h)

	_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry(h, "sk", "contrib-x", nullifier32(0x02)),
	})
	require.NoError(t, err)

	resp, err := q.GetContributorServes(env.ctx, &types.QueryGetContributorServesRequest{ContributorId: "contrib-x", Epoch: 1})
	require.NoError(t, err)
	require.NotNil(t, resp.Serves)
	require.Equal(t, uint64(1), resp.Serves.ServeCount)
	require.Equal(t, "contrib-x", resp.Serves.ContributorId)
}

func TestQueryGetContributorServes_NotFound(t *testing.T) {
	env := setupKeeper(t)
	q := keeper.NewQueryServerImpl(env.k)
	_, err := q.GetContributorServes(env.ctx, &types.QueryGetContributorServesRequest{ContributorId: "ghost", Epoch: 1})
	require.Error(t, err)
}

func TestQueryGetMemoryServeCount(t *testing.T) {
	env := setupKeeper(t)
	q := keeper.NewQueryServerImpl(env.k)
	h := hash32(0x03)
	env.mem.approve("org-1", h)

	_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry(h, "sk", "c1", nullifier32(0x03)),
	})
	require.NoError(t, err)

	resp, err := q.GetMemoryServeCount(env.ctx, &types.QueryGetMemoryServeCountRequest{OrgId: "org-1", ContentHash: h, Epoch: 1})
	require.NoError(t, err)
	require.Equal(t, uint64(1), resp.Count)

	// Unknown memory => zero count, no error.
	resp0, err := q.GetMemoryServeCount(env.ctx, &types.QueryGetMemoryServeCountRequest{OrgId: "org-1", ContentHash: hash32(0xFF), Epoch: 1})
	require.NoError(t, err)
	require.Equal(t, uint64(0), resp0.Count)
}
