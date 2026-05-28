package keeper

import (
	"bytes"
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/wevibe-network/wevibe-chain/x/serve/types"
)

type MsgServer struct {
	keeper *Keeper
}

func matchedKeywordKey(orgID, cidHex string, epoch uint64, keyword string) []byte {
	return []byte(fmt.Sprintf("%s%s/%s/%d/%s", types.MatchedKeywordsPrefix, orgID, cidHex, epoch, keyword))
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

		originatingAttestation, found, err := s.keeper.GetServeAttestationByNullifier(ctx, entry.Nullifier)
		if err != nil {
			return nil, err
		}
		if !found {
			rejected++
			continue
		}
		if len(originatingAttestation.MatchedKeywords) == 0 {
			rejected++
			continue
		}
		if !bytes.Equal(entry.MemoryHash, originatingAttestation.MemoryContentHash) {
			rejected++
			continue
		}

		if _, err := s.keeper.memoryKeeper.GetApprovedMemory(ctx, msg.OrgId, entry.MemoryHash); err != nil {
			rejected++
			continue
		}

		store.Set(denialNullifierKey(entry.Nullifier), []byte{1})
		s.keeper.IncrementDenialCount(ctx, msg.OrgId, originatingAttestation.MemoryContentHash, msg.Epoch)
		if err := s.keeper.StoreDenialAttestation(ctx, msg.OrgId, msg.Epoch, entry); err != nil {
			return nil, err
		}
		cidHex := types.ContentHashToHex(originatingAttestation.MemoryContentHash)
		for _, keyword := range originatingAttestation.MatchedKeywords {
			if keyword == "" {
				return nil, fmt.Errorf("originating attestation has empty matched keyword")
			}
			store.Set(matchedKeywordKey(msg.OrgId, cidHex, msg.Epoch, keyword), []byte{0x01})
		}
		accepted++

		if err := s.keeper.memoryKeeper.ApplyDenialDecay(ctx, msg.OrgId, originatingAttestation.MemoryContentHash); err != nil {
			// Non-fatal: the denial attestation is primary. The decay is a
			// secondary side effect that may fail if the memory is already
			// archived. Match the emissions pattern (payout failures do not
			// roll back the epoch).
			s.keeper.logger.Warn("ApplyDenialDecay failed",
				"org", msg.OrgId,
				"hash", types.ContentHashToHex(originatingAttestation.MemoryContentHash),
				"err", err,
			)
		}
	}

	// Emit denial_batch_submitted event — follows MsgRemoveMember pattern (CO-016)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeDenialBatchSubmitted,
		sdk.NewAttribute(types.AttributeKeyOrgID, msg.OrgId),
		sdk.NewAttribute(types.AttributeKeySubmitter, msg.Signer),
		sdk.NewAttribute(types.AttributeKeyEpoch, fmt.Sprintf("%d", msg.Epoch)),
		sdk.NewAttribute(types.AttributeKeyAcceptedCount, fmt.Sprintf("%d", accepted)),
		sdk.NewAttribute(types.AttributeKeyRejectedCount, fmt.Sprintf("%d", rejected)),
		sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", sdkCtx.BlockHeight())),
	))

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
