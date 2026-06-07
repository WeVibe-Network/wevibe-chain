package keeper

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/wevibe-network/wevibe-chain/x/identity/types"
)

type msgServer struct {
	keeper *Keeper
}

var _ types.MsgServer = (*msgServer)(nil)

func NewMsgServerImpl(k *Keeper) types.MsgServer {
	return &msgServer{keeper: k}
}

func (m *msgServer) MigrateIdentity(ctx context.Context, msg *types.MsgMigrateIdentity) (*types.MsgMigrateIdentityResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if _, err := sdk.AccAddressFromBech32(msg.Signer); err != nil {
		return nil, types.ErrInvalidWalletAddress
	}

	canonical := []byte("wevibe.migrate_identity.v1\npasskey_pubkey:" + msg.PasskeyPubkey + "\nwallet:" + msg.Signer + "\nnonce:" + strconv.FormatUint(msg.Nonce, 10))

	passkeyPubkey, err := hex.DecodeString(msg.PasskeyPubkey)
	if err != nil || len(passkeyPubkey) != ed25519.PublicKeySize {
		return nil, types.ErrInvalidPasskeyPubkey
	}

	passkeySignature, err := hex.DecodeString(msg.PasskeySignature)
	if err != nil || len(passkeySignature) != ed25519.SignatureSize {
		return nil, types.ErrInvalidPasskeySignature
	}

	if !ed25519.Verify(ed25519.PublicKey(passkeyPubkey), canonical, passkeySignature) {
		return nil, types.ErrInvalidPasskeySignature
	}

	existingAlias, found, err := m.keeper.GetAlias(ctx, msg.PasskeyPubkey)
	if err != nil {
		return nil, err
	}
	if found && existingAlias.IsMigrated {
		return nil, types.ErrAliasAlreadyMigrated
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if err := m.keeper.SetAlias(ctx, &types.StoredIdentityAlias{
		PasskeyPubkey:   msg.PasskeyPubkey,
		WalletAddress:   msg.Signer,
		IsMigrated:      true,
		MigratedAtEpoch: uint64(sdkCtx.BlockHeight()),
	}); err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"identity.migrated",
		sdk.NewAttribute("passkey_pubkey", msg.PasskeyPubkey),
		sdk.NewAttribute("wallet", msg.Signer),
	))

	return &types.MsgMigrateIdentityResponse{}, nil
}
