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
		serveEntry("org-1", 1, h, "sk", "c1", nonce32(0x01)),
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
		serveEntry("org-1", 1, h, "sk", "contrib-x", nonce32(0x02)),
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
		serveEntry("org-1", 1, h, "sk", "c1", nonce32(0x03)),
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

func TestQueryListEventsAndPolicyAnchors(t *testing.T) {
	env := setupKeeper(t)
	q := keeper.NewQueryServerImpl(env.k)
	h := hash32(0x71)
	env.mem.approve("org-1", h)
	entry, fp := outcomeEventEntry(t, "org-1", 7, h, 0x71)
	accepted, _, _, err := env.k.ProcessEventBatch(env.ctx, "org-1", 7, []*types.EventEntry{entry})
	require.NoError(t, err)
	require.Equal(t, uint64(1), accepted)

	eventsResp, err := q.ListEvents(env.ctx, &types.QueryListEventsRequest{OrgId: "org-1", Epoch: 7})
	require.NoError(t, err)
	require.Len(t, eventsResp.Events, 1)
	require.Equal(t, fp, eventsResp.Events[0].Fingerprint)

	ctx := env.sdkCtx()
	policyHash := hash32(0x72)
	require.NoError(t, env.k.SetPolicyAnchor(ctx, "policy-v1", policyHash))
	anchorResp, err := q.GetPolicyAnchor(ctx, &types.QueryGetPolicyAnchorRequest{PolicyVersion: "policy-v1"})
	require.NoError(t, err)
	require.True(t, anchorResp.Found)
	require.Equal(t, policyHash, anchorResp.Anchor.PolicyHash)
	latestResp, err := q.GetLatestPolicyAnchor(ctx, &types.QueryGetLatestPolicyAnchorRequest{})
	require.NoError(t, err)
	require.True(t, latestResp.Found)
	require.Equal(t, "policy-v1", latestResp.Anchor.PolicyVersion)
}
