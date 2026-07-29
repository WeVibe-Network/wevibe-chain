package keeper

import (
	"context"
	"fmt"

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
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	stats, err := q.keeper.GetEpochServeStats(ctx, req.OrgId, req.Epoch)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetEpochServeStatsResponse{
		Stats: &types.StoredEpochServeStats{
			OrgId:                stats.OrgID,
			Epoch:                stats.Epoch,
			TotalServes:          stats.TotalServes,
			UniqueMemoriesServed: stats.UniqueMemoriesServed,
			UniqueServeKeys:      stats.UniqueServeKeys,
			SelfServes:           stats.SelfServes,
			ModelBreakdown:       stats.ModelBreakdown,
		},
	}, nil
}

func (q *queryServer) GetContributorServes(ctx context.Context, req *types.QueryGetContributorServesRequest) (*types.QueryGetContributorServesResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
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
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	count := q.keeper.GetMemoryServeCount(ctx, req.OrgId, req.ContentHash, req.Epoch)
	return &types.QueryGetMemoryServeCountResponse{Count: count}, nil
}

func (q *queryServer) ListEvents(ctx context.Context, req *types.QueryListEventsRequest) (*types.QueryListEventsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	events, err := q.keeper.GetEventsForEpoch(ctx, req.OrgId, req.Epoch)
	if err != nil {
		return nil, err
	}
	return &types.QueryListEventsResponse{Events: events}, nil
}

func (q *queryServer) GetPolicyAnchor(ctx context.Context, req *types.QueryGetPolicyAnchorRequest) (*types.QueryGetPolicyAnchorResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	anchor, found, err := q.keeper.GetPolicyAnchor(ctx, req.PolicyVersion)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetPolicyAnchorResponse{Anchor: anchor, Found: found}, nil
}

func (q *queryServer) GetLatestPolicyAnchor(ctx context.Context, req *types.QueryGetLatestPolicyAnchorRequest) (*types.QueryGetLatestPolicyAnchorResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	anchor, found, err := q.keeper.GetLatestPolicyAnchor(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetLatestPolicyAnchorResponse{Anchor: anchor, Found: found}, nil
}

func (q *queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	params, err := q.keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: &params}, nil
}
