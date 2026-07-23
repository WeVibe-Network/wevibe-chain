package keeper

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
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

// requireLeaderWallet enforces D-S32-CO044-KEY-SEPARATION: org-decision messages
// (commit/approve/report) are authorized ONLY by the org's registered leader
// chain wallet, verified against the AUTHENTICATED tx signer (msg.Signer) — not
// a self-declared field. A stolen hub serving key (or any other key) is rejected
// here, so it can never forge a commit/approval/report; it can only submit
// serve/denial batches (enforced in x/serve) and drain its own gas.
func (m *msgServer) requireLeaderWallet(ctx context.Context, orgID, signer string) error {
	wallet, err := m.keeper.orgKeeper.GetLeaderWallet(ctx, orgID)
	if err != nil {
		return err
	}
	if wallet == "" || signer != wallet {
		return types.ErrUnauthorized
	}
	return nil
}

func (m *msgServer) SubmitCommitment(ctx context.Context, msg *types.MsgSubmitCommitment) (*types.MsgSubmitCommitmentResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if err := m.requireLeaderWallet(ctx, msg.OrgId, msg.Signer); err != nil {
		return nil, err
	}

	member, err := m.keeper.orgKeeper.GetMember(ctx, msg.OrgId, msg.ContributorId)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to load member record for contributor %s in org %s: %v", types.ErrNotContributor, msg.ContributorId, msg.OrgId, err)
	}
	if member == nil {
		return nil, types.ErrNotContributor
	}

	if member.Role != "leader" && !member.CanContribute {
		return nil, types.ErrNotContributor
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
	commitment.ContributorAddress = msg.ContributorWallet
	commitment.McVersion = msg.McVersion
	commitment.ProducerModelId = msg.ProducerModelId
	commitment.AttestationSessionHash = msg.AttestationSessionHash
	if err := m.keeper.SubmitCommitment(ctx, commitment); err != nil {
		return nil, err
	}

	producerModelIDAttr := msg.ProducerModelId
	if producerModelIDAttr == "" {
		producerModelIDAttr = "unattested"
	}
	attestationSessionHashAttr := "none"
	if len(msg.AttestationSessionHash) > 0 {
		attestationSessionHashAttr = types.ContentHashToHex(msg.AttestationSessionHash)
	}

	// Emit commitment_submitted event — previously dead code, now wired (CO-016)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeCommitmentSubmitted,
		sdk.NewAttribute(types.AttributeKeyOrgID, msg.OrgId),
		sdk.NewAttribute(types.AttributeKeyContributor, msg.ContributorId),
		sdk.NewAttribute(types.AttributeKeyProducerModelId, producerModelIDAttr),
		sdk.NewAttribute(types.AttributeKeyAttestationSessionHash, attestationSessionHashAttr),
		sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", sdkCtx.BlockHeight())),
	))

	return &types.MsgSubmitCommitmentResponse{}, nil
}

func (m *msgServer) ApproveMemory(ctx context.Context, msg *types.MsgApproveMemory) (*types.MsgApproveMemoryResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	if err := m.requireLeaderWallet(ctx, msg.OrgId, msg.Signer); err != nil {
		return nil, err
	}
	if !types.ValidMemoryType(msg.MemoryType) {
		return nil, types.ErrInvalidMemoryType
	}

	pending, err := m.keeper.GetPendingCommitment(ctx, msg.OrgId, msg.ContentHash)
	if err != nil {
		return nil, err
	}

	wrappedDekHash := sha256.Sum256(msg.WrappedDekEnc)

	submissionHasher := sha256.New()
	_, _ = submissionHasher.Write(msg.EncryptedBlob)
	_, _ = submissionHasher.Write(msg.WrappedDekEnc)
	submissionHash := submissionHasher.Sum(nil)

	canonicalBody := buildSubmitMemoryCanonicalBody(
		msg.CiphertextHash,
		pending.Contributor,
		pending.Epoch,
		msg.MemoryType,
		msg.OrgId,
		msg.PlaintextHash,
		msg.Salt,
		submissionHash,
		wrappedDekHash[:],
	)

	contributorPubkeyBytes, decodeErr := hex.DecodeString(pending.Contributor)
	if decodeErr != nil || len(contributorPubkeyBytes) != ed25519.PublicKeySize {
		m.keeper.logger.Info("memory approval rejected: invalid contributor pubkey",
			"org_id", msg.OrgId,
			"content_hash", types.ContentHashToHex(msg.ContentHash),
		)
		return nil, errorsmod.Wrapf(
			types.ErrInvalidContributor,
			"memory approval rejected: invalid contributor pubkey (org=%s content_hash=%s)",
			msg.OrgId,
			types.ContentHashToHex(msg.ContentHash),
		)
	}

	if !ed25519.Verify(ed25519.PublicKey(contributorPubkeyBytes), canonicalBody, msg.ContributorSig) {
		m.keeper.logger.Info("memory approval rejected: contributor signature verification failed",
			"org_id", msg.OrgId,
			"content_hash", types.ContentHashToHex(msg.ContentHash),
		)
		return nil, errorsmod.Wrapf(
			types.ErrInvalidContributorSignature,
			"memory approval rejected: contributor signature verification failed (org=%s content_hash=%s)",
			msg.OrgId,
			types.ContentHashToHex(msg.ContentHash),
		)
	}

	derivedCiphertextHash := sha256.Sum256(msg.EncryptedBlob)
	if !bytesEqual(msg.CiphertextHash, derivedCiphertextHash[:]) {
		m.keeper.logger.Info("memory approval rejected: ciphertext hash mismatch",
			"org_id", msg.OrgId,
			"content_hash", types.ContentHashToHex(msg.ContentHash),
		)
		return nil, errorsmod.Wrapf(
			types.ErrInvalidBlob,
			"memory approval rejected: ciphertext hash mismatch (org=%s content_hash=%s)",
			msg.OrgId,
			types.ContentHashToHex(msg.ContentHash),
		)
	}

	if !bytesEqual(msg.ContentHash, submissionHash) {
		m.keeper.logger.Info("memory approval rejected: content hash mismatch",
			"org_id", msg.OrgId,
			"content_hash", types.ContentHashToHex(msg.ContentHash),
		)
		return nil, errorsmod.Wrapf(
			types.ErrInvalidContentHash,
			"memory approval rejected: content hash mismatch (org=%s content_hash=%s)",
			msg.OrgId,
			types.ContentHashToHex(msg.ContentHash),
		)
	}

	if err := m.keeper.ApproveMemory(
		ctx,
		msg.OrgId,
		msg.ContentHash,
		msg.EncryptedBlob,
		msg.CommittingLeader,
		msg.WrappedDekEnc,
		msg.PlaintextHash,
		msg.Salt,
		msg.CiphertextHash,
		wrappedDekHash[:],
		msg.ContributorSig,
		msg.MemoryType,
	); err != nil {
		return nil, err
	}

	return &types.MsgApproveMemoryResponse{}, nil
}

func buildSubmitMemoryCanonicalBody(ciphertextHash []byte, contributorPubkey string, epochID uint64, memoryType types.MemoryType, orgID string, plaintextHash, salt, submissionHash, wrappedDekHash []byte) []byte {
	canonicalLines := []string{
		"wevibe.submit_memory.v1",
		"ciphertext_hash:" + hex.EncodeToString(ciphertextHash),
		"contributor_pubkey:" + contributorPubkey,
		fmt.Sprintf("epoch_id:%d", epochID),
		"memory_type:" + types.CanonicalMemoryType(memoryType),
		"org_id:" + orgID,
		"plaintext_hash:" + hex.EncodeToString(plaintextHash),
		"salt:" + hex.EncodeToString(salt),
		"submission_hash:" + hex.EncodeToString(submissionHash),
		"wrapped_dek_hash:" + hex.EncodeToString(wrappedDekHash),
	}

	return []byte(strings.Join(canonicalLines, "\n"))
}

func bytesEqual(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
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

	if err := m.requireLeaderWallet(ctx, msg.OrgId, msg.Signer); err != nil {
		return nil, err
	}

	if err := m.keeper.ReportMemory(ctx, msg.OrgId, msg.ContentHash, msg.ContributorPubkey, msg.ApprovingModerators, msg.UpholdingModerators, msg.ReporterPubkey, msg.Signer, msg.Reason, msg.Plaintext, msg.Ciphertext, msg.Capsule, msg.PlaintextHash, msg.PlaintextOversized, 0); err != nil {
		return nil, err
	}

	return &types.MsgReportMemoryResponse{}, nil
}
