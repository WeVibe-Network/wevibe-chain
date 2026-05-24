package keeper

import (
	"context"

	"github.com/wevibe-network/wevibe-chain/x/serve/types"
)

type MsgServer struct {
	keeper *Keeper
}

var _ types.MsgServer = (*MsgServer)(nil)

func NewMsgServerImpl(k *Keeper) types.MsgServer {
	return &MsgServer{keeper: k}
}

func (s *MsgServer) SubmitServeBatch(ctx context.Context, msg *types.MsgSubmitServeBatch) (*types.MsgSubmitServeBatchResponse, error) {
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

	accepted, rejectedDuplicate, rejectedInvalid, err := s.keeper.ProcessServeBatch(ctx, msg.OrgId, msg.Epoch, msg.Serves)
	if err != nil {
		return nil, err
	}

	return &types.MsgSubmitServeBatchResponse{
		Accepted:          accepted,
		RejectedDuplicate: rejectedDuplicate,
		RejectedInvalid:   rejectedInvalid,
	}, nil
}

func (s *MsgServer) SubmitDenialBatch(ctx context.Context, msg *types.MsgSubmitDenialBatch) (*types.MsgSubmitDenialBatchResponse, error) {
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

	params, err := s.keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	if len(msg.Entries) > int(params.MaxServesPerBatch) {
		return nil, types.ErrBatchTooLarge
	}

	var accepted uint64
	var rejected uint64

	store := s.keeper.getStore(ctx)
	for _, entry := range msg.Entries {
		if s.keeper.HasDenialNullifier(ctx, entry.Nullifier) {
			rejected++
			continue
		}

		if _, err := s.keeper.memoryKeeper.GetApprovedMemory(ctx, msg.OrgId, entry.MemoryHash); err != nil {
			rejected++
			continue
		}

		store.Set(denialNullifierKey(entry.Nullifier), []byte{1})
		s.keeper.IncrementDenialCount(ctx, msg.OrgId, entry.MemoryHash, msg.Epoch)
		if err := s.keeper.StoreDenialAttestation(ctx, msg.OrgId, msg.Epoch, entry); err != nil {
			return nil, err
		}
		accepted++

		if err := s.keeper.memoryKeeper.ApplyDenialDecay(ctx, msg.OrgId, entry.MemoryHash); err != nil {
			// Non-fatal: the denial attestation is primary. The decay is a
			// secondary side effect that may fail if the memory is already
			// archived. Match the emissions pattern (payout failures do not
			// roll back the epoch).
			s.keeper.logger.Warn("ApplyDenialDecay failed",
				"org", msg.OrgId,
				"hash", types.ContentHashToHex(entry.MemoryHash),
				"err", err,
			)
		}
	}

	return &types.MsgSubmitDenialBatchResponse{
		Accepted: accepted,
		Rejected: rejected,
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
