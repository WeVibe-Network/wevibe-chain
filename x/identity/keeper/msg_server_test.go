package keeper_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/wevibe-network/wevibe-chain/testutil/keeper"
	"github.com/wevibe-network/wevibe-chain/x/identity/keeper"
	"github.com/wevibe-network/wevibe-chain/x/identity/types"
)

var testSigner = sdk.AccAddress([]byte("identity-test-signer")).String()

func setupMsgServer(t *testing.T) (types.MsgServer, *keeper.Keeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	storeService, cms := testkeeper.NewTestStoreService(t, storeKey)
	logger := testkeeper.NewTestLogger()
	k := keeper.NewKeeper(storeService, logger, "gov")
	ctx := sdk.NewContext(cms, tmproto.Header{Height: 42, Time: time.Now().UTC()}, false, logger)

	return keeper.NewMsgServerImpl(k), k, ctx
}

func canonicalMigrateIdentityMessage(passkeyPubkey, signer string, nonce uint64) []byte {
	return []byte("wevibe.migrate_identity.v1\npasskey_pubkey:" + passkeyPubkey + "\nwallet:" + signer + "\nnonce:" + strconv.FormatUint(nonce, 10))
}

func TestMigrateIdentity_Success(t *testing.T) {
	msgSrv, k, ctx := setupMsgServer(t)

	passkeyPubkey, passkeyPrivkey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	passkeyPubkeyHex := hex.EncodeToString(passkeyPubkey)
	nonce := uint64(7)
	signature := ed25519.Sign(passkeyPrivkey, canonicalMigrateIdentityMessage(passkeyPubkeyHex, testSigner, nonce))

	msg := &types.MsgMigrateIdentity{
		Signer:           testSigner,
		PasskeyPubkey:    passkeyPubkeyHex,
		PasskeySignature: hex.EncodeToString(signature),
		Nonce:            nonce,
	}

	resp, err := msgSrv.MigrateIdentity(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	walletAddress, isMigrated, found, err := k.ResolveIdentity(ctx, passkeyPubkeyHex)
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, isMigrated)
	require.Equal(t, testSigner, walletAddress)

	alias, found, err := k.GetAlias(ctx, passkeyPubkeyHex)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, uint64(ctx.BlockHeight()), alias.MigratedAtEpoch)
}

func TestMigrateIdentity_InvalidPasskeySignature(t *testing.T) {
	msgSrv, _, ctx := setupMsgServer(t)

	passkeyPubkey, passkeyPrivkey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	passkeyPubkeyHex := hex.EncodeToString(passkeyPubkey)
	nonce := uint64(11)
	invalidSignature := ed25519.Sign(passkeyPrivkey, []byte("not-the-canonical-message"))

	msg := &types.MsgMigrateIdentity{
		Signer:           testSigner,
		PasskeyPubkey:    passkeyPubkeyHex,
		PasskeySignature: hex.EncodeToString(invalidSignature),
		Nonce:            nonce,
	}

	_, err = msgSrv.MigrateIdentity(ctx, msg)
	require.ErrorIs(t, err, types.ErrInvalidPasskeySignature)
}

func TestMigrateIdentity_AlreadyMigrated(t *testing.T) {
	msgSrv, _, ctx := setupMsgServer(t)

	passkeyPubkey, passkeyPrivkey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	passkeyPubkeyHex := hex.EncodeToString(passkeyPubkey)
	firstNonce := uint64(21)
	firstSignature := ed25519.Sign(passkeyPrivkey, canonicalMigrateIdentityMessage(passkeyPubkeyHex, testSigner, firstNonce))

	firstMsg := &types.MsgMigrateIdentity{
		Signer:           testSigner,
		PasskeyPubkey:    passkeyPubkeyHex,
		PasskeySignature: hex.EncodeToString(firstSignature),
		Nonce:            firstNonce,
	}

	_, err = msgSrv.MigrateIdentity(ctx, firstMsg)
	require.NoError(t, err)

	secondNonce := uint64(22)
	secondSignature := ed25519.Sign(passkeyPrivkey, canonicalMigrateIdentityMessage(passkeyPubkeyHex, testSigner, secondNonce))
	secondMsg := &types.MsgMigrateIdentity{
		Signer:           testSigner,
		PasskeyPubkey:    passkeyPubkeyHex,
		PasskeySignature: hex.EncodeToString(secondSignature),
		Nonce:            secondNonce,
	}

	_, err = msgSrv.MigrateIdentity(ctx, secondMsg)
	require.ErrorIs(t, err, types.ErrAliasAlreadyMigrated)
}
