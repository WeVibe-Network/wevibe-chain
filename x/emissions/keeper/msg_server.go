package keeper

import (
	"context"

	"github.com/wevibe-network/wevibe-chain/x/emissions/types"
)

type msgServer struct {
	keeper *Keeper
}

var _ types.MsgServer = (*msgServer)(nil)

func NewMsgServerImpl(k *Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

func (m *msgServer) MintDailyEmission(ctx context.Context, msg *types.MsgMintDailyEmission) (*types.MsgMintDailyEmissionResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if msg.Authority != m.keeper.authority {
		return nil, types.ErrUnauthorized
	}

	emission, err := m.keeper.MintDailyEmission(ctx, msg.Epoch)
	if err != nil {
		return nil, err
	}

	return &types.MsgMintDailyEmissionResponse{
		TotalEmitted:   emission.TotalEmitted,
		OperatorShare:  emission.OperatorShare,
		ValidatorShare: emission.ValidatorShare,
	}, nil
}

func (m *msgServer) DistributeOperatorRewards(ctx context.Context, msg *types.MsgDistributeOperatorRewards) (*types.MsgDistributeOperatorRewardsResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	rewards := make(map[string]uint64)
	for _, entry := range msg.Rewards {
		rewards[entry.OperatorId] = entry.Amount
	}

	if err := m.keeper.DistributeOperatorRewards(ctx, rewards, msg.Epoch); err != nil {
		return nil, err
	}

	return &types.MsgDistributeOperatorRewardsResponse{}, nil
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