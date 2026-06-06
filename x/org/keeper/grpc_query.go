package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/wevibe-network/wevibe-chain/x/org/types"
)

type queryServer struct {
	keeper *Keeper
}

var _ types.QueryServer = (*queryServer)(nil)

func NewQueryServerImpl(k *Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

func (q *queryServer) GetOrg(ctx context.Context, req *types.QueryGetOrgRequest) (*types.QueryGetOrgResponse, error) {
	org, err := q.keeper.GetOrg(ctx, req.OrgId)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetOrgResponse{
		OrgId:             org.OrgID,
		Leader:            org.Leader,
		CreatedAt:         org.CreatedAt,
		RenewalHeight:     org.RenewalHeight,
		StorageQuota:      org.StorageQuota,
		RetrievalBudget:   org.RetrievalBudget,
		Status:            int32(org.Status),
		Domain:            org.Domain,
		HubServingAddress: org.HubServingAddress,
		HubEndpoints:      org.HubEndpoints,
	}, nil
}

func (q *queryServer) GetMembers(ctx context.Context, req *types.QueryGetMembersRequest) (*types.QueryGetMembersResponse, error) {
	members, err := q.keeper.GetAllMembers(ctx, req.OrgId)
	if err != nil {
		return nil, err
	}
	var memberInfos []*types.MemberInfo
	for _, m := range members {
		memberInfos = append(memberInfos, &types.MemberInfo{
			OrgId:  m.OrgID,
			Pubkey: m.Pubkey,
			Role:   m.Role,
		})
	}
	return &types.QueryGetMembersResponse{Members: memberInfos}, nil
}

func (q *queryServer) IsMember(ctx context.Context, req *types.QueryIsMemberRequest) (*types.QueryIsMemberResponse, error) {
	isMember, err := q.keeper.IsMember(ctx, req.OrgId, req.Pubkey)
	if err != nil {
		return nil, err
	}
	return &types.QueryIsMemberResponse{IsMember: isMember}, nil
}

func (q *queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := q.keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: &params}, nil
}

func (q *queryServer) GetOrgConfig(ctx context.Context, req *types.QueryGetOrgConfigRequest) (*types.QueryGetOrgConfigResponse, error) {
	cfg, err := q.keeper.GetOrgConfig(ctx, req.OrgId)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetOrgConfigResponse{
		ServeAttestationRequired: cfg.ServeAttestationRequired,
		MinContributionsPerEpoch: cfg.MinContributionsPerEpoch,
		ContestStakeVibe:         cfg.ContestStakeVibe,
	}, nil
}

func (q *queryServer) GetOrgAccount(ctx context.Context, req *types.QueryGetOrgAccountRequest) (*types.QueryGetOrgAccountResponse, error) {
	org, err := q.keeper.GetOrg(ctx, req.OrgId)
	if err != nil {
		return nil, err
	}

	accountAddr, err := sdk.AccAddressFromBech32(org.AccountAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid org account address: %w", err)
	}

	balance := q.keeper.bankKeeper.GetBalance(ctx, accountAddr, "uvibe")

	return &types.QueryGetOrgAccountResponse{
		AccountAddress: org.AccountAddress,
		Balance:        balance.Amount.String(),
	}, nil
}

func (q *queryServer) GetOrgProfile(ctx context.Context, req *types.QueryGetOrgProfileRequest) (*types.QueryGetOrgProfileResponse, error) {
	if req.OrgId == "" {
		return nil, types.ErrInvalidOrgID
	}

	org, err := q.keeper.GetOrg(ctx, req.OrgId)
	if err != nil {
		return nil, err
	}

	members, err := q.keeper.GetAllMembers(ctx, req.OrgId)
	if err != nil {
		return nil, err
	}

	var memberCount, moderatorCount uint64
	for _, m := range members {
		memberCount++
		if m.Role == "moderator" {
			moderatorCount++
		}
	}

	resp := &types.QueryGetOrgProfileResponse{
		OrgId:           org.OrgID,
		Leader:          org.Leader,
		AccountAddress:  org.AccountAddress,
		Domain:          org.Domain,
		Status:          int32(org.Status),
		CreatedAt:       org.CreatedAt,
		StorageQuota:    org.StorageQuota,
		RetrievalBudget: org.RetrievalBudget,
		MemberCount:     memberCount,
		ModeratorCount:  moderatorCount,
	}

	if q.keeper.memoryKeeper != nil {
		count, err := q.keeper.memoryKeeper.GetApprovedCount(ctx, req.OrgId)
		if err == nil {
			resp.ApprovedMemoryCount = count
		}
	}

	if q.keeper.serveKeeper != nil && req.Epoch > 0 {
		totalServes, uniqueMemories, selfServes, modelBreakdown, err := q.keeper.serveKeeper.GetEpochServeStatsRaw(ctx, req.OrgId, req.Epoch)
		if err == nil {
			resp.TotalServes = totalServes
			resp.UniqueMemoriesServed = uniqueMemories
			resp.SelfServes = selfServes
			resp.ModelBreakdown = modelBreakdown
		}
	}

	if q.keeper.bandwidthKeeper != nil && req.Epoch > 0 {
		memoryUsed, memoryCap, serveUsed, serveCap, err := q.keeper.bandwidthKeeper.GetOrInitBandwidthStateRaw(ctx, req.OrgId, req.Epoch)
		if err == nil {
			resp.MemoryBandwidthUsed = memoryUsed
			resp.MemoryBandwidthCap = memoryCap
			resp.ServeBandwidthUsed = serveUsed
			resp.ServeBandwidthCap = serveCap
		}
	}

	return resp, nil
}
