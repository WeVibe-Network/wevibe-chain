package keeper

import (
	"context"

	"github.com/wevibe-network/wevibe-chain/x/identity/types"
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

func (q *queryServer) ResolveIdentity(ctx context.Context, req *types.QueryResolveIdentityRequest) (*types.QueryResolveIdentityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	alias, found, err := q.keeper.GetAlias(ctx, req.PasskeyPubkey)
	if err != nil {
		return nil, err
	}
	if !found {
		return &types.QueryResolveIdentityResponse{Found: false}, nil
	}

	return &types.QueryResolveIdentityResponse{
		WalletAddress: alias.WalletAddress,
		IsMigrated:    alias.IsMigrated,
		Found:         true,
	}, nil
}
