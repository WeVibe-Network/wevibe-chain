package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wevibe-network/wevibe-chain/x/bandwidth/types"
)

// Note: the existing keeper tests live in package `keeper` (white-box) and
// provide makeTestKeeper / mockOrgKeeper, which these tests reuse. The query
// server is constructed via NewQueryServerImpl (see grpc_query.go).

// ---------------------------------------------------------------------------
// GetBandwidthState query
// ---------------------------------------------------------------------------

func TestQueryGetBandwidthState_Success(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()
	qs := NewQueryServerImpl(k)

	require.NoError(t, k.ConsumeMemoryBandwidth(ctx, "test-org", 1))

	resp, err := qs.GetBandwidthState(ctx, &types.QueryGetBandwidthStateRequest{
		OrgId: "test-org",
		Epoch: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.State)
	require.Equal(t, "test-org", resp.State.OrgId)
	require.Equal(t, uint64(1), resp.State.Epoch)
	require.Equal(t, uint64(1), resp.State.MemoryUsed)
	require.Equal(t, uint64(10000), resp.State.MemoryCap)
	require.Equal(t, uint64(50000), resp.State.ServeCap)
}

func TestQueryGetBandwidthState_NotFound(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()
	qs := NewQueryServerImpl(k)

	// No state stored for epoch 99 -> keeper returns nil state -> query maps
	// to ErrOverrideNotFound (see grpc_query.go).
	resp, err := qs.GetBandwidthState(ctx, &types.QueryGetBandwidthStateRequest{
		OrgId: "test-org",
		Epoch: 99,
	})
	require.ErrorIs(t, err, types.ErrOverrideNotFound)
	require.Nil(t, resp)
}

// ---------------------------------------------------------------------------
// GetBandwidthOverride query
// ---------------------------------------------------------------------------

func TestQueryGetBandwidthOverride_Success(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()
	qs := NewQueryServerImpl(k)

	require.NoError(t, k.SetBandwidthOverride(ctx, "test-org", 20000, 100000))

	resp, err := qs.GetBandwidthOverride(ctx, &types.QueryGetBandwidthOverrideRequest{
		OrgId: "test-org",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.HasOverride)
	require.NotNil(t, resp.Override)
	require.Equal(t, "test-org", resp.Override.OrgId)
	require.Equal(t, uint64(20000), resp.Override.MemoryCap)
	require.Equal(t, uint64(100000), resp.Override.ServeCap)
}

func TestQueryGetBandwidthOverride_NotFound(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()
	qs := NewQueryServerImpl(k)

	// No override set -> keeper returns ErrOverrideNotFound -> query returns a
	// non-error response with HasOverride=false.
	resp, err := qs.GetBandwidthOverride(ctx, &types.QueryGetBandwidthOverrideRequest{
		OrgId: "test-org",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.False(t, resp.HasOverride)
	require.Nil(t, resp.Override)
}

// ---------------------------------------------------------------------------
// GetRemainingBandwidth query
// ---------------------------------------------------------------------------

func TestQueryGetRemainingBandwidth_Success(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()
	qs := NewQueryServerImpl(k)

	require.NoError(t, k.ConsumeMemoryBandwidth(ctx, "test-org", 1))
	require.NoError(t, k.ConsumeMemoryBandwidth(ctx, "test-org", 1))

	resp, err := qs.GetRemainingBandwidth(ctx, &types.QueryGetRemainingBandwidthRequest{
		OrgId: "test-org",
		Epoch: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(9998), resp.MemoryRemaining)
	require.Equal(t, uint64(50000), resp.ServeRemaining)
}

func TestQueryGetRemainingBandwidth_FreshEpoch(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()
	qs := NewQueryServerImpl(k)

	// Fresh org/epoch lazily initializes from default params; nothing consumed.
	resp, err := qs.GetRemainingBandwidth(ctx, &types.QueryGetRemainingBandwidthRequest{
		OrgId: "test-org",
		Epoch: 42,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(10000), resp.MemoryRemaining)
	require.Equal(t, uint64(50000), resp.ServeRemaining)
}

// ---------------------------------------------------------------------------
// Params query
// ---------------------------------------------------------------------------

func TestQueryParams_Default(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()
	qs := NewQueryServerImpl(k)

	resp, err := qs.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Params)
	require.Equal(t, uint64(10000), resp.Params.DefaultMemoryCapPerEpoch)
	require.Equal(t, uint64(50000), resp.Params.DefaultServeCapPerEpoch)
}

func TestQueryParams_AfterSet(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()
	qs := NewQueryServerImpl(k)

	require.NoError(t, k.SetParams(ctx, types.Params{
		DefaultMemoryCapPerEpoch: 12345,
		DefaultServeCapPerEpoch:  67890,
	}))

	resp, err := qs.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Params)
	require.Equal(t, uint64(12345), resp.Params.DefaultMemoryCapPerEpoch)
	require.Equal(t, uint64(67890), resp.Params.DefaultServeCapPerEpoch)
}
