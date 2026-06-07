package keeper

import (
	"context"

	"github.com/wevibe-network/wevibe-chain/x/emissions/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type queryServer struct {
	keeper *Keeper
}

var _ types.QueryServer = (*queryServer)(nil)

func NewQueryServerImpl(k *Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

func (q *queryServer) GetEmissionPool(ctx context.Context, req *types.QueryGetEmissionPoolRequest) (*types.QueryGetEmissionPoolResponse, error) {
	pool, err := q.keeper.GetEmissionPool(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetEmissionPoolResponse{
		TotalSupply:    pool.TotalSupply,
		DailyMint:      pool.DailyMint,
		OperatorShare:  pool.OperatorShare,
		ValidatorShare: pool.ValidatorShare,
		Epoch:          pool.Epoch,
	}, nil
}

func (q *queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := q.keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: &params}, nil
}

func (q *queryServer) ContributorReward(ctx context.Context, req *types.QueryContributorRewardRequest) (*types.QueryContributorRewardResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	pending, err := q.keeper.GetContributorReward(ctx, req.Pubkey)
	if err != nil {
		return nil, err
	}
	allTime, err := q.keeper.GetLifetimeContributorReward(ctx, req.Pubkey)
	if err != nil {
		return nil, err
	}
	return &types.QueryContributorRewardResponse{
		PendingWithdrawal: pending,
		AllTimeEarnings:   allTime,
	}, nil
}
