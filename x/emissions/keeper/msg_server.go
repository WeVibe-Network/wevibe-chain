package keeper

import (
	"context"
	"strconv"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		OperatorShare:  0,
		ValidatorShare: emission.ValidatorShare,
	}, nil
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

func (m *msgServer) ClaimContributorReward(ctx context.Context, msg *types.MsgClaimContributorReward) (*types.MsgClaimContributorRewardResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	wallet, isMigrated, found, err := m.keeper.identityKeeper.ResolveIdentity(ctx, msg.PasskeyPubkey)
	if err != nil {
		return nil, err
	}
	if !found || !isMigrated {
		return nil, types.ErrNotMigrated
	}

	if wallet != msg.Signer {
		return nil, types.ErrUnauthorizedClaim
	}

	balance, err := m.keeper.GetContributorReward(ctx, msg.PasskeyPubkey)
	if err != nil {
		return nil, err
	}
	if balance == 0 {
		return nil, types.ErrNothingToClaim
	}

	recipient, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, types.ErrInvalidWalletAddress
	}

	if err := m.keeper.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.EmissionsModuleName,
		recipient,
		sdk.NewCoins(sdk.NewCoin("uvibe", math.NewIntFromUint64(balance))), // uvibe is chain bond denom
	); err != nil {
		return nil, err
	}

	if err := m.keeper.SetContributorReward(ctx, msg.PasskeyPubkey, 0); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"emissions.contributor_reward_claimed",
		sdk.NewAttribute("passkey_pubkey", msg.PasskeyPubkey),
		sdk.NewAttribute("wallet", msg.Signer),
		sdk.NewAttribute("amount", strconv.FormatUint(balance, 10)),
	))

	return &types.MsgClaimContributorRewardResponse{AmountClaimed: balance}, nil
}
