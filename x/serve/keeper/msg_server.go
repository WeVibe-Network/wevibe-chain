package keeper

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/wevibe-network/wevibe-chain/x/serve/types"
)

type MsgServer struct {
	keeper *Keeper
}

var _ types.MsgServer = (*MsgServer)(nil)

func NewMsgServerImpl(k *Keeper) types.MsgServer {
	return &MsgServer{keeper: k}
}

// requireServingKeySigner enforces D-S32-CO044-KEY-SEPARATION: only the org's
// registered hub serving key may submit serve/denial batches. The tx signer
// (msg.Signer) is the authenticated fee payer; it must equal the org's
// currently-registered serving address. A stolen serving key can therefore do
// nothing beyond submit serve/denial batches + drain its own gas; any other key
// is rejected. An org with no registered serving address can never serve.
func (s *MsgServer) requireServingKeySigner(ctx context.Context, orgID, signer string) error {
	servingAddr, err := s.keeper.orgKeeper.GetServingAddress(ctx, orgID)
	if err != nil {
		return err
	}
	if servingAddr == "" || signer != servingAddr {
		return types.ErrUnauthorized
	}
	return nil
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

	if err := s.requireServingKeySigner(ctx, msg.OrgId, msg.Signer); err != nil {
		return nil, err
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

	if err := s.requireServingKeySigner(ctx, msg.OrgId, msg.Signer); err != nil {
		return nil, err
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
	var rejectedDupFingerprint uint64
	var rejectedInvalidSignature uint64
	var rejectedNoReceipt uint64
	var rejectedHashMismatch uint64
	var rejectedServeKeyMismatch uint64
	var rejectedNoMemory uint64

	store := s.keeper.getStore(ctx)
	for _, entry := range msg.Entries {
		if len(entry.ServeKeyPubkey) != ed25519.PublicKeySize || len(entry.ServeSig) != ed25519.SignatureSize {
			rejected++
			rejectedInvalidSignature++
			continue
		}

		canonicalBody := types.CanonicalDenialBody(msg.OrgId, entry.MemoryHash, msg.Epoch, entry.ServeKeyPubkey, entry.ServeFingerprint, entry.Nonce)
		if !ed25519.Verify(ed25519.PublicKey(entry.ServeKeyPubkey), canonicalBody, entry.ServeSig) {
			rejected++
			rejectedInvalidSignature++
			continue
		}

		denialFingerprint := types.ComputeDenialFingerprint(msg.OrgId, entry.MemoryHash, msg.Epoch, entry.ServeKeyPubkey, entry.ServeFingerprint)
		if s.keeper.HasDenialFingerprint(ctx, denialFingerprint) {
			rejected++
			rejectedDupFingerprint++
			continue
		}

		originatingReceipt, found, err := s.keeper.GetServeReceiptByFingerprint(ctx, entry.ServeFingerprint)
		if err != nil {
			return nil, err
		}
		if !found {
			rejected++
			rejectedNoReceipt++
			continue
		}
		if !bytes.Equal(entry.MemoryHash, originatingReceipt.MemoryContentHash) {
			rejected++
			rejectedHashMismatch++
			continue
		}
		if !bytes.Equal(entry.ServeKeyPubkey, originatingReceipt.ServeKeyPubkey) {
			rejected++
			rejectedServeKeyMismatch++
			continue
		}

		if _, err := s.keeper.memoryKeeper.GetApprovedMemory(ctx, msg.OrgId, entry.MemoryHash); err != nil {
			rejected++
			rejectedNoMemory++
			continue
		}

		store.Set(denialFingerprintKey(denialFingerprint), []byte{1})
		s.keeper.IncrementDenialCount(ctx, msg.OrgId, originatingReceipt.MemoryContentHash, msg.Epoch)
		s.keeper.updateEpochDenialStats(ctx, msg.OrgId, msg.Epoch)
		if err := s.keeper.StoreDenialReceipt(ctx, msg.OrgId, msg.Epoch, denialFingerprint, entry); err != nil {
			return nil, err
		}
		accepted++

	}

	// Emit denial_batch_submitted event — follows MsgRemoveMember pattern (CO-016)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	s.keeper.logger.Info(fmt.Sprintf(
		"denial batch submitted: org=%s epoch=%d entries=%d accepted=%d rejected=%d duplicate_fingerprint=%d invalid_signature=%d no_receipt=%d hash_mismatch=%d serve_key_mismatch=%d no_memory=%d",
		msg.OrgId, msg.Epoch, len(msg.Entries), accepted, rejected,
		rejectedDupFingerprint, rejectedInvalidSignature, rejectedNoReceipt, rejectedHashMismatch, rejectedServeKeyMismatch, rejectedNoMemory,
	))
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

func (s *MsgServer) SubmitEventBatch(ctx context.Context, msg *types.MsgSubmitEventBatch) (*types.MsgSubmitEventBatchResponse, error) {
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

	if err := s.requireServingKeySigner(ctx, msg.OrgId, msg.Signer); err != nil {
		return nil, err
	}

	params, err := s.keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	if len(msg.Events) > int(params.MaxServesPerBatch) {
		return nil, types.ErrBatchTooLarge
	}

	accepted, rejectedDuplicate, rejectedInvalid, err := s.keeper.ProcessEventBatch(ctx, msg.OrgId, msg.Epoch, msg.Events)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeEventBatchSubmitted,
		sdk.NewAttribute(types.AttributeKeyOrgID, msg.OrgId),
		sdk.NewAttribute(types.AttributeKeySubmitter, msg.Signer),
		sdk.NewAttribute(types.AttributeKeyEpoch, fmt.Sprintf("%d", msg.Epoch)),
		sdk.NewAttribute(types.AttributeKeyAcceptedCount, fmt.Sprintf("%d", accepted)),
		sdk.NewAttribute(types.AttributeKeyRejectedDuplicateCount, fmt.Sprintf("%d", rejectedDuplicate)),
		sdk.NewAttribute(types.AttributeKeyRejectedInvalidCount, fmt.Sprintf("%d", rejectedInvalid)),
		sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", sdkCtx.BlockHeight())),
	))

	return &types.MsgSubmitEventBatchResponse{
		Accepted:          accepted,
		RejectedDuplicate: rejectedDuplicate,
		RejectedInvalid:   rejectedInvalid,
	}, nil
}

func (s *MsgServer) AnchorPolicyVersion(ctx context.Context, msg *types.MsgAnchorPolicyVersion) (*types.MsgAnchorPolicyVersionResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if msg.Authority != s.keeper.authority {
		return nil, types.ErrUnauthorized
	}

	if err := s.keeper.SetPolicyAnchor(ctx, msg.PolicyVersion, msg.PolicyHash); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypePolicyVersionAnchored,
		sdk.NewAttribute(types.AttributeKeyPolicyVersion, msg.PolicyVersion),
		sdk.NewAttribute(types.AttributeKeyPolicyHash, hex.EncodeToString(msg.PolicyHash)),
		sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", sdkCtx.BlockHeight())),
	))

	return &types.MsgAnchorPolicyVersionResponse{}, nil
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
