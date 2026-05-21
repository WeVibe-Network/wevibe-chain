package keeper

import (
	"context"

	"github.com/wevibe-network/wevibe-chain/x/bandwidth/types"
)

type queryServer struct {
	keeper *Keeper
}

var _ types.QueryServer = (*queryServer)(nil)

func NewQueryServerImpl(k *Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

func (q *queryServer) GetBandwidthState(ctx context.Context, req *types.QueryGetBandwidthStateRequest) (*types.QueryGetBandwidthStateResponse, error) {
	state, err := q.keeper.GetBandwidthState(ctx, req.OrgId, req.Epoch)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, types.ErrOverrideNotFound
	}
	return &types.QueryGetBandwidthStateResponse{
		State: &types.StoredBandwidthState{
			OrgId:       state.OrgID,
			Epoch:       state.Epoch,
			MemoryUsed:  state.MemoryUsed,
			MemoryCap:   state.MemoryCap,
			ServeUsed:   state.ServeUsed,
			ServeCap:    state.ServeCap,
		},
	}, nil
}

func (q *queryServer) GetBandwidthOverride(ctx context.Context, req *types.QueryGetBandwidthOverrideRequest) (*types.QueryGetBandwidthOverrideResponse, error) {
	override, err := q.keeper.GetBandwidthOverride(ctx, req.OrgId)
	if err != nil {
		if err == types.ErrOverrideNotFound {
			return &types.QueryGetBandwidthOverrideResponse{
				HasOverride: false,
			}, nil
		}
		return nil, err
	}
	return &types.QueryGetBandwidthOverrideResponse{
		Override: &types.StoredBandwidthOverride{
			OrgId:     override.OrgID,
			MemoryCap: override.MemoryCap,
			ServeCap:  override.ServeCap,
		},
		HasOverride: true,
	}, nil
}

func (q *queryServer) GetRemainingBandwidth(ctx context.Context, req *types.QueryGetRemainingBandwidthRequest) (*types.QueryGetRemainingBandwidthResponse, error) {
	memoryRemaining, serveRemaining, err := q.keeper.GetRemainingBandwidth(ctx, req.OrgId, req.Epoch)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetRemainingBandwidthResponse{
		MemoryRemaining: memoryRemaining,
		ServeRemaining:  serveRemaining,
	}, nil
}

func (q *queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := q.keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: &params}, nil
}