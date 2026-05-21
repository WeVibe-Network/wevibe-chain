package keeper

import (
	"context"

	"github.com/wevibe-network/wevibe-chain/x/attestation/types"
)

type queryServer struct {
	keeper *Keeper
}

var _ types.QueryServer = (*queryServer)(nil)

func NewQueryServerImpl(k *Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

func (q *queryServer) GetSessionAttestation(ctx context.Context, req *types.QueryGetSessionAttestationRequest) (*types.QueryGetSessionAttestationResponse, error) {
	att, err := q.keeper.GetSessionAttestation(ctx, req.OrgId, req.SessionHash)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetSessionAttestationResponse{Attestation: att}, nil
}

func (q *queryServer) ListSessionAttestations(ctx context.Context, req *types.QueryListSessionAttestationsRequest) (*types.QueryListSessionAttestationsResponse, error) {
	attestations, err := q.keeper.ListSessionAttestations(ctx, req.OrgId, req.Epoch)
	if err != nil {
		return nil, err
	}
	return &types.QueryListSessionAttestationsResponse{Attestations: attestations}, nil
}

func (q *queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := q.keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: &params}, nil
}