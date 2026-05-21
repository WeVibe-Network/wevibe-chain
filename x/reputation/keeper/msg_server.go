package keeper

import (
	"context"

	"github.com/wevibe-network/wevibe-chain/x/reputation/types"
)

type msgServer struct {
	keeper *Keeper
}

var _ types.MsgServer = (*msgServer)(nil)

func NewMsgServerImpl(k *Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

func (m *msgServer) UpdateReputation(ctx context.Context, msg *types.MsgUpdateReputation) (*types.MsgUpdateReputationResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	memory := &types.AttestedMemory{
		Developer:  msg.Developer,
		MemoryCID:  msg.MemoryCid,
		Difficulty: uint8(msg.Difficulty),
		Quality:    uint8(msg.Quality),
		DomainTags: msg.DomainTags,
		Provenance: msg.Provenance,
	}

	if err := m.keeper.UpdateReputation(ctx, msg.Developer, memory); err != nil {
		return nil, err
	}

	xp, err := m.keeper.GetXP(ctx, msg.Developer)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateReputationResponse{Xp: xp}, nil
}

func (m *msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if msg.Authority != m.keeper.authority {
		return nil, types.ErrUnauthorized
	}
	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}
	if err := m.keeper.SetParams(ctx, *msg.Params); err != nil {
		return nil, err
	}
	return &types.MsgUpdateParamsResponse{}, nil
}

func (m *msgServer) IncrementContribution(ctx context.Context, msg *types.MsgIncrementContribution) (*types.MsgIncrementContributionResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if msg.Authority != m.keeper.authority {
		return nil, types.ErrUnauthorized
	}
	if err := m.keeper.IncrementContribution(ctx, msg.ContributorId, msg.OrgId, msg.MemoryCid); err != nil {
		return nil, err
	}
	return &types.MsgIncrementContributionResponse{}, nil
}

func (m *msgServer) IncrementServe(ctx context.Context, msg *types.MsgIncrementServe) (*types.MsgIncrementServeResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if msg.Authority != m.keeper.authority {
		return nil, types.ErrUnauthorized
	}
	if err := m.keeper.IncrementServe(ctx, msg.ContributorId, msg.OrgId, msg.MemoryCid, msg.ServeCount); err != nil {
		return nil, err
	}
	return &types.MsgIncrementServeResponse{}, nil
}

func (m *msgServer) RecordBan(ctx context.Context, msg *types.MsgRecordBan) (*types.MsgRecordBanResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if msg.Authority != m.keeper.authority {
		return nil, types.ErrUnauthorized
	}
	if err := m.keeper.RecordBan(ctx, msg.ContributorId, msg.OrgId, msg.MemoryCid); err != nil {
		return nil, err
	}
	return &types.MsgRecordBanResponse{}, nil
}