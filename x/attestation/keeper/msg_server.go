package keeper

import (
	"context"

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

	// D-ATTEST-ROADMAP: keep attestation fully wired but disabled until
	// verification infrastructure exists.
	return nil, types.ErrAttestationDisabled
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
