package keeper

import (
	"context"

	"github.com/wevibe-network/wevibe-chain/x/emissions/types"
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
