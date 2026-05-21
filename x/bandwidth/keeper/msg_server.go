package keeper

import (
	"context"

	"github.com/wevibe-network/wevibe-chain/x/bandwidth/types"
)

type msgServer struct {
	keeper *Keeper
}

var _ types.MsgServer = (*msgServer)(nil)

func NewMsgServerImpl(k *Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

func (m *msgServer) SetBandwidthOverride(ctx context.Context, msg *types.MsgSetBandwidthOverride) (*types.MsgSetBandwidthOverrideResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	hasOrg, err := m.keeper.orgKeeper.HasOrg(ctx, msg.OrgId)
	if err != nil {
		return nil, err
	}
	if !hasOrg {
		return nil, types.ErrOrgNotFound
	}

	isLeader, err := m.keeper.orgKeeper.IsLeader(ctx, msg.OrgId, msg.Signer)
	if err != nil {
		return nil, err
	}
	if !isLeader {
		return nil, types.ErrUnauthorized
	}

	if err := m.keeper.SetBandwidthOverride(ctx, msg.OrgId, msg.MemoryCap, msg.ServeCap); err != nil {
		return nil, err
	}

	return &types.MsgSetBandwidthOverrideResponse{}, nil
}

func (m *msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if msg.Authority != m.keeper.authority {
		return nil, types.ErrUnauthorized
	}
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if err := m.keeper.SetParams(ctx, *msg.Params); err != nil {
		return nil, err
	}
	return &types.MsgUpdateParamsResponse{}, nil
}