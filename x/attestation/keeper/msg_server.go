package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/wevibe-network/wevibe-chain/x/attestation/types"
)

type MsgServer struct {
	keeper *Keeper
}

var _ types.MsgServer = (*MsgServer)(nil)

func NewMsgServerImpl(k *Keeper) types.MsgServer {
	return &MsgServer{keeper: k}
}

func (s *MsgServer) SubmitSessionAttestation(ctx context.Context, msg *types.MsgSubmitSessionAttestation) (*types.MsgSubmitSessionAttestationResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	hasOrg, err := s.keeper.orgKeeper.HasOrg(ctx, msg.OrgId)
	if err != nil {
		return nil, err
	}
	if !hasOrg {
		return nil, types.ErrOrgNotFound
	}

	if s.keeper.HasSessionAttestation(ctx, msg.OrgId, msg.SessionHash) {
		return nil, types.ErrDuplicateAttestation
	}

	var verificationStatus string
	switch msg.ProviderType {
	case types.ProviderType_PROVIDER_TYPE_LOCAL:
		if len(msg.CommitllmReceiptHash) > 0 {
			_, verificationStatus = s.keeper.VerifyCommitLLMReceipt(ctx, msg.CommitllmReceiptHash)
		} else {
			verificationStatus = "unverified: no receipt provided"
		}
	case types.ProviderType_PROVIDER_TYPE_CLOUD:
		if len(msg.ProviderSignatureHash) > 0 {
			_, verificationStatus = s.keeper.VerifyCloudProviderSignature(ctx, msg.ProviderSignatureHash)
		} else {
			verificationStatus = "unverified: no provider signature"
		}
	default:
		verificationStatus = "unverified: unknown provider type"
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	att := &types.StoredSessionAttestation{
		OrgId:                 msg.OrgId,
		SessionHash:           msg.SessionHash,
		ModelId:               msg.ModelId,
		TurnCount:             msg.TurnCount,
		TokenCount:            msg.TokenCount,
		ProviderType:          msg.ProviderType,
		CommitllmReceiptHash:  msg.CommitllmReceiptHash,
		ProviderSignatureHash: msg.ProviderSignatureHash,
		ContributorId:         msg.ContributorId,
		Epoch:                 msg.Epoch,
		SubmittedAtHeight:     uint64(sdkCtx.BlockHeight()),
	}

	if err := s.keeper.SetSessionAttestation(ctx, att); err != nil {
		return nil, err
	}

	return &types.MsgSubmitSessionAttestationResponse{
		Accepted:            true,
		VerificationStatus:  verificationStatus,
	}, nil
}

func (s *MsgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if msg.Authority != s.keeper.authority {
		return nil, types.ErrUnauthorized
	}
	if err := s.keeper.SetParams(ctx, *msg.Params); err != nil {
		return nil, err
	}
	return &types.MsgUpdateParamsResponse{}, nil
}