package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/wevibe-network/wevibe-chain/x/memory/types"
)

type msgServer struct {
	types.UnimplementedMsgServer
	keeper *Keeper
}

var _ types.MsgServer = (*msgServer)(nil)

func NewMsgServerImpl(k *Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

func (m *msgServer) SubmitCommitment(ctx context.Context, msg *types.MsgSubmitCommitment) (*types.MsgSubmitCommitmentResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if !types.ValidMemoryType(msg.MemoryType) {
		return nil, types.ErrInvalidMemoryType
	}

	commitment := types.NewPendingCommitment(
		msg.OrgId,
		msg.ContentHash,
		msg.Keywords,
		msg.ContributorId,
		0,
		0,
		msg.MemoryType,
	)
	if err := m.keeper.SubmitCommitment(ctx, commitment); err != nil {
		return nil, err
	}

	// Emit commitment_submitted event — previously dead code, now wired (CO-016)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeCommitmentSubmitted,
		sdk.NewAttribute(types.AttributeKeyOrgID, msg.OrgId),
		sdk.NewAttribute(types.AttributeKeyContributor, msg.ContributorId),
		sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", sdkCtx.BlockHeight())),
	))

	return &types.MsgSubmitCommitmentResponse{}, nil
}

func (m *msgServer) ApproveMemory(ctx context.Context, msg *types.MsgApproveMemory) (*types.MsgApproveMemoryResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if !types.ValidMemoryType(msg.MemoryType) {
		return nil, types.ErrInvalidMemoryType
	}

	if err := m.keeper.ApproveMemory(ctx, msg.OrgId, msg.ContentHash, msg.EncryptedBlob, msg.Approvers, msg.CommittingLeader, msg.WrappedDekEnc, msg.MemoryType); err != nil {
		return nil, err
	}

	return &types.MsgApproveMemoryResponse{}, nil
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

func (m *msgServer) ReportMemory(ctx context.Context, msg *types.MsgReportMemory) (*types.MsgReportMemoryResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if err := m.keeper.ReportMemory(ctx, msg.OrgId, msg.ContentHash, msg.ContributorPubkey, msg.ApprovingModerators, msg.UpholdingModerators, msg.ReporterPubkey, msg.Signer, msg.Reason, msg.Plaintext, msg.Ciphertext, msg.Capsule, msg.PlaintextHash, msg.PlaintextOversized, 0); err != nil {
		return nil, err
	}

	return &types.MsgReportMemoryResponse{}, nil
}
