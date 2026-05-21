package keeper

import (
	"context"

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

func (m *msgServer) RejectMemory(ctx context.Context, msg *types.MsgRejectMemory) (*types.MsgRejectMemoryResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if err := m.keeper.RejectMemory(ctx, msg.OrgId, msg.ContentHash, msg.Signer); err != nil {
		return nil, err
	}

	return &types.MsgRejectMemoryResponse{}, nil
}

func (m *msgServer) PurgeExpired(ctx context.Context, msg *types.MsgPurgeExpired) (*types.MsgPurgeExpiredResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	count, err := m.keeper.PurgeExpired(ctx, msg.OrgId, 0)
	if err != nil {
		return nil, err
	}

	return &types.MsgPurgeExpiredResponse{PurgedCount: count}, nil
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

func (m *msgServer) RelateMemories(ctx context.Context, msg *types.MsgRelateMemories) (*types.MsgRelateMemoriesResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if err := m.keeper.ProposeRelationship(ctx, msg); err != nil {
		return nil, err
	}
	return &types.MsgRelateMemoriesResponse{}, nil
}

func (m *msgServer) ApproveRelationship(ctx context.Context, msg *types.MsgApproveRelationship) (*types.MsgApproveRelationshipResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if err := m.keeper.ApproveRelationship(ctx, msg); err != nil {
		return nil, err
	}
	return &types.MsgApproveRelationshipResponse{}, nil
}

func (m *msgServer) SetValidityBounds(ctx context.Context, msg *types.MsgSetValidityBounds) (*types.MsgSetValidityBoundsResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if err := m.keeper.SetValidityBounds(ctx, msg); err != nil {
		return nil, err
	}
	return &types.MsgSetValidityBoundsResponse{}, nil
}

func (m *msgServer) ArchiveMemory(ctx context.Context, msg *types.MsgArchiveMemory) (*types.MsgArchiveMemoryResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if err := m.keeper.ArchiveMemory(ctx, msg.OrgId, msg.MemoryCid, msg.Sender); err != nil {
		return nil, err
	}
	return &types.MsgArchiveMemoryResponse{}, nil
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
