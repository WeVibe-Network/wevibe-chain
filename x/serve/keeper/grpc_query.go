package keeper

import (
	"context"

	"github.com/wevibe-network/wevibe-chain/x/serve/types"
)

type queryServer struct {
	keeper *Keeper
}

var _ types.QueryServer = (*queryServer)(nil)

func NewQueryServerImpl(k *Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

func (q *queryServer) GetEpochServeStats(ctx context.Context, req *types.QueryGetEpochServeStatsRequest) (*types.QueryGetEpochServeStatsResponse, error) {
	stats, err := q.keeper.GetEpochServeStats(ctx, req.OrgId, req.Epoch)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetEpochServeStatsResponse{
		Stats: &types.StoredEpochServeStats{
			OrgId:               stats.OrgID,
			Epoch:               stats.Epoch,
			TotalServes:         stats.TotalServes,
			UniqueMemoriesServed: stats.UniqueMemoriesServed,
			UniqueServeKeys:     stats.UniqueServeKeys,
			SelfServes:         stats.SelfServes,
			ModelBreakdown:      stats.ModelBreakdown,
		},
	}, nil
}

func (q *queryServer) GetContributorServes(ctx context.Context, req *types.QueryGetContributorServesRequest) (*types.QueryGetContributorServesResponse, error) {
	cs, err := q.keeper.GetContributorEpochServes(ctx, req.ContributorId, req.Epoch)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetContributorServesResponse{
		Serves: &types.StoredContributorEpochServes{
			ContributorId:  cs.ContributorID,
			Epoch:          cs.Epoch,
			ServeCount:     cs.ServeCount,
			SelfServeCount: cs.SelfServeCount,
			OrgIds:         cs.OrgIDs,
			TotalTurns:     cs.TotalTurns,
		},
	}, nil
}

func (q *queryServer) GetMemoryServeCount(ctx context.Context, req *types.QueryGetMemoryServeCountRequest) (*types.QueryGetMemoryServeCountResponse, error) {
	count := q.keeper.GetMemoryServeCount(ctx, req.OrgId, req.ContentHash, req.Epoch)
	return &types.QueryGetMemoryServeCountResponse{Count: count}, nil
}

func (q *queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := q.keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: &params}, nil
}