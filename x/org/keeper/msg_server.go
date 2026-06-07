package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/math"

	"cosmossdk.io/x/feegrant"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/wevibe-network/wevibe-chain/x/org/types"
)

type msgServer struct {
	keeper *Keeper
}

var _ types.MsgServer = (*msgServer)(nil)

func NewMsgServerImpl(k *Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

func (m *msgServer) RegisterOrg(ctx context.Context, msg *types.MsgRegisterOrg) (*types.MsgRegisterOrgResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	creator, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, fmt.Errorf("invalid signer address: %w", err)
	}

	org := types.NewOrg("", msg.Leader, msg.Domain, msg.StorageQuota, msg.RetrievalBudget)
	org.HubServingAddress = msg.HubServingKey
	org.LeaderWalletAddress = msg.LeaderWallet
	if err := m.keeper.RegisterOrg(ctx, org, creator); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeOrgRegistered,
		sdk.NewAttribute(types.AttributeKeyOrgID, org.OrgID),
		sdk.NewAttribute(types.AttributeKeyLeader, msg.Leader),
	))

	return &types.MsgRegisterOrgResponse{OrgId: org.OrgID}, nil
}

func (m *msgServer) AddMember(ctx context.Context, msg *types.MsgAddMember) (*types.MsgAddMemberResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	has, err := m.keeper.HasOrg(ctx, msg.OrgId)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.ErrOrgNotFound
	}

	org, err := m.keeper.GetOrg(ctx, msg.OrgId)
	if err != nil {
		return nil, err
	}
	if org.LeaderWalletAddress == "" || msg.Signer != org.LeaderWalletAddress {
		return nil, types.ErrNotLeader
	}

	if msg.Role == "leader" {
		return nil, types.ErrInvalidRole
	}

	member := types.NewMemberRecord(msg.OrgId, msg.Pubkey, msg.Role)
	if err := m.keeper.AddMember(ctx, member); err != nil {
		return nil, err
	}

	return &types.MsgAddMemberResponse{}, nil
}

func (m *msgServer) SetServingKey(ctx context.Context, msg *types.MsgSetServingKey) (*types.MsgSetServingKeyResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	has, err := m.keeper.HasOrg(ctx, msg.OrgId)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.ErrOrgNotFound
	}

	if err := m.keeper.SetServingKey(ctx, msg.OrgId, msg.NewServingKey, msg.Signer); err != nil {
		return nil, err
	}

	return &types.MsgSetServingKeyResponse{}, nil
}

func (m *msgServer) SetServingInfo(ctx context.Context, msg *types.MsgSetServingInfo) (*types.MsgSetServingInfoResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	has, err := m.keeper.HasOrg(ctx, msg.OrgId)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.ErrOrgNotFound
	}

	if err := m.keeper.SetServingInfo(ctx, msg.OrgId, msg.HubEndpoints, msg.HubResponsePubkey, msg.Signer); err != nil {
		return nil, err
	}

	return &types.MsgSetServingInfoResponse{}, nil
}

func (m *msgServer) SetExtractionProfile(ctx context.Context, msg *types.MsgSetExtractionProfile) (*types.MsgSetExtractionProfileResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	has, err := m.keeper.HasOrg(ctx, msg.OrgId)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.ErrOrgNotFound
	}

	profileVersion, err := m.keeper.SetExtractionProfile(
		ctx,
		msg.Signer,
		msg.OrgId,
		msg.ExtractionModel,
		msg.NumCtx,
		msg.SystemPrompt,
		msg.OutputSchema,
		msg.DomainFraming,
		msg.Exemplars,
		msg.Constraints,
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgSetExtractionProfileResponse{ProfileVersion: profileVersion}, nil
}

func (m *msgServer) RemoveMember(ctx context.Context, msg *types.MsgRemoveMember) (*types.MsgRemoveMemberResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	has, err := m.keeper.HasOrg(ctx, msg.OrgId)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.ErrOrgNotFound
	}

	org, err := m.keeper.GetOrg(ctx, msg.OrgId)
	if err != nil {
		return nil, err
	}
	if org.LeaderWalletAddress == "" || msg.Signer != org.LeaderWalletAddress {
		return nil, types.ErrNotLeader
	}

	if err := m.keeper.RemoveMember(ctx, msg.OrgId, msg.Pubkey); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeMemberRemoved,
		sdk.NewAttribute(types.AttributeKeyOrgID, msg.OrgId),
		sdk.NewAttribute(types.AttributeKeyMemberPubkey, msg.Pubkey),
		sdk.NewAttribute(types.AttributeKeyRemovedBy, msg.Signer),
		sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", sdkCtx.BlockHeight())),
	))

	return &types.MsgRemoveMemberResponse{}, nil
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

func (m *msgServer) SetOrgConfig(ctx context.Context, msg *types.MsgSetOrgConfig) (*types.MsgSetOrgConfigResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	has, err := m.keeper.HasOrg(ctx, msg.OrgId)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.ErrOrgNotFound
	}

	org, err := m.keeper.GetOrg(ctx, msg.OrgId)
	if err != nil {
		return nil, err
	}
	if org.LeaderWalletAddress == "" || msg.Signer != org.LeaderWalletAddress {
		return nil, types.ErrNotLeader
	}
	if msg.MinContributionsPerEpoch > 100 {
		return nil, fmt.Errorf("min_contributions_per_epoch must be <= 100")
	}

	cfg := &types.OrgConfig{
		OrgID:                    msg.OrgId,
		ServeAttestationRequired: msg.ServeAttestationRequired,
		ContestStakeVibe:         msg.ContestStakeVibe,
		MinContributionsPerEpoch: msg.MinContributionsPerEpoch,
	}

	if err := m.keeper.SetOrgConfig(ctx, msg.OrgId, cfg); err != nil {
		return nil, err
	}

	return &types.MsgSetOrgConfigResponse{}, nil
}

func (m *msgServer) GrantTrialAllowance(ctx context.Context, msg *types.MsgGrantTrialAllowance) (*types.MsgGrantTrialAllowanceResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if m.keeper.feegrantKeeper == nil {
		return nil, types.ErrFeegrantUnavailable
	}

	hasOrg, err := m.keeper.HasOrg(ctx, msg.OrgId)
	if err != nil {
		return nil, err
	}
	if !hasOrg {
		return nil, types.ErrOrgNotFound
	}

	org, err := m.keeper.GetOrg(ctx, msg.OrgId)
	if err != nil {
		return nil, err
	}
	if org.LeaderWalletAddress == "" || msg.Signer != org.LeaderWalletAddress {
		return nil, types.ErrNotLeader
	}

	granter, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, fmt.Errorf("invalid signer address: %w", err)
	}

	grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
	if err != nil {
		return nil, fmt.Errorf("invalid grantee address: %w", err)
	}

	gasPerSubmission := math.NewInt(2000)
	dailyLimit := gasPerSubmission.Mul(math.NewIntFromUint64(msg.DailySubmissions))
	totalLimit := dailyLimit.Mul(math.NewIntFromUint64(msg.TrialDays))

	dailyCoin := sdk.NewCoin("uvibe", dailyLimit)
	totalCoin := sdk.NewCoin("uvibe", totalLimit)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	expires := sdkCtx.BlockTime().Add(time.Duration(msg.TrialDays) * 24 * time.Hour)

	basic := feegrant.BasicAllowance{
		SpendLimit: sdk.NewCoins(totalCoin),
		Expiration: &expires,
	}
	periodic := &feegrant.PeriodicAllowance{
		Basic:            basic,
		Period:           24 * time.Hour,
		PeriodSpendLimit: sdk.NewCoins(dailyCoin),
		PeriodCanSpend:   sdk.NewCoins(dailyCoin),
	}

	if err := m.keeper.feegrantKeeper.GrantAllowance(ctx, granter, grantee, periodic); err != nil {
		return nil, fmt.Errorf("grant allowance: %w", err)
	}

	return &types.MsgGrantTrialAllowanceResponse{}, nil
}

func (m *msgServer) UpdateMemberRole(ctx context.Context, msg *types.MsgUpdateMemberRole) (*types.MsgUpdateMemberRoleResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	has, err := m.keeper.HasOrg(ctx, msg.OrgId)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.ErrOrgNotFound
	}

	if err := m.keeper.UpdateMemberRole(ctx, msg.OrgId, msg.Pubkey, msg.NewRole, msg.Signer); err != nil {
		return nil, err
	}

	return &types.MsgUpdateMemberRoleResponse{}, nil
}

func (m *msgServer) RotateEpoch(ctx context.Context, msg *types.MsgRotateEpoch) (*types.MsgRotateEpochResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	has, err := m.keeper.HasOrg(ctx, msg.OrgId)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.ErrOrgNotFound
	}

	newEpoch, err := m.keeper.RotateEpoch(ctx, msg.OrgId, msg.Signer)
	if err != nil {
		return nil, err
	}

	return &types.MsgRotateEpochResponse{NewEpoch: newEpoch}, nil
}

func (m *msgServer) TransferLeadership(ctx context.Context, msg *types.MsgTransferLeadership) (*types.MsgTransferLeadershipResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	has, err := m.keeper.HasOrg(ctx, msg.OrgId)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.ErrOrgNotFound
	}

	if err := m.keeper.TransferLeadership(ctx, msg.OrgId, msg.NewLeader, msg.Signer); err != nil {
		return nil, err
	}

	return &types.MsgTransferLeadershipResponse{}, nil
}

func (m *msgServer) CloseOrg(ctx context.Context, msg *types.MsgCloseOrg) (*types.MsgCloseOrgResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	has, err := m.keeper.HasOrg(ctx, msg.OrgId)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.ErrOrgNotFound
	}

	if err := m.keeper.CloseOrg(ctx, msg.OrgId, msg.Signer); err != nil {
		return nil, err
	}

	return &types.MsgCloseOrgResponse{}, nil
}
