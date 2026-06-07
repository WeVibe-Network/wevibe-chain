package keeper_test

import (
	"context"
	"testing"

	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/wevibe-network/wevibe-chain/testutil/keeper"
	"github.com/wevibe-network/wevibe-chain/x/emissions/keeper"
	"github.com/wevibe-network/wevibe-chain/x/emissions/types"
)

func setupMsgServer(t *testing.T) (types.MsgServer, context.Context) {
	srv, ctx, _, _, _ := setupMsgServerWithMocks(t)
	return srv, ctx
}

func setupMsgServerWithMocks(t *testing.T) (types.MsgServer, context.Context, *keeper.Keeper, *mockBankKeeper, *mockIdentityKeeper) {
	storeKey := storetypes.NewKVStoreKey("emissions")
	storeService, cms := testkeeper.NewTestStoreService(t, storeKey)
	logger := testkeeper.NewTestLogger()
	k := keeper.NewKeeper(storeService, logger, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", nil, nil, nil, newMockReputationKeeper())
	bankKeeper := newMockBankKeeper()
	identityKeeper := newMockIdentityKeeper()
	k.SetBankKeeper(bankKeeper)
	k.SetIdentityKeeper(identityKeeper)
	sdkCtx := sdk.NewContext(cms, tmproto.Header{}, false, logger)
	return keeper.NewMsgServerImpl(k), sdk.WrapSDKContext(sdkCtx), k, bankKeeper, identityKeeper
}

func newTestWalletAddress() string {
	return sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()
}

func TestMsgMintDailyEmission_ValidateBasic(t *testing.T) {
	msg := &types.MsgMintDailyEmission{}
	require.Error(t, msg.ValidateBasic())

	msg.Authority = "cosmos1abc"
	msg.Epoch = 1
	require.NoError(t, msg.ValidateBasic())
}

func TestMsgMintDailyEmission_NoPool(t *testing.T) {
	srv, ctx := setupMsgServer(t)

	msg := &types.MsgMintDailyEmission{
		Authority: "cosmos1abc",
		Epoch:     1,
	}

	_, err := srv.MintDailyEmission(ctx, msg)
	require.Error(t, err)
}

func TestMsgClaimContributorReward_HappyPath(t *testing.T) {
	srv, ctx, k, bankKeeper, identityKeeper := setupMsgServerWithMocks(t)
	signer := newTestWalletAddress()

	const (
		passkeyPubkey = "passkey_pubkey_1"
	)

	identityKeeper.walletAddress = signer
	identityKeeper.isMigrated = true
	identityKeeper.found = true

	require.NoError(t, k.SetContributorReward(ctx, passkeyPubkey, 42))

	resp, err := srv.ClaimContributorReward(ctx, &types.MsgClaimContributorReward{
		Signer:        signer,
		PasskeyPubkey: passkeyPubkey,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(42), resp.AmountClaimed)

	require.Len(t, bankKeeper.sendCalls, 1)
	sendCall := bankKeeper.sendCalls[0]
	require.Equal(t, types.EmissionsModuleName, sendCall.senderModule)
	require.Equal(t, signer, sendCall.recipientAddr.String())
	require.Equal(t, uint64(42), sendCall.amt.AmountOf("uvibe").Uint64())

	balance, err := k.GetContributorReward(ctx, passkeyPubkey)
	require.NoError(t, err)
	require.Zero(t, balance)
}

func TestMsgClaimContributorReward_NotMigrated(t *testing.T) {
	srv, ctx, _, bankKeeper, identityKeeper := setupMsgServerWithMocks(t)
	signer := newTestWalletAddress()

	identityKeeper.walletAddress = signer
	identityKeeper.isMigrated = false
	identityKeeper.found = true

	_, err := srv.ClaimContributorReward(ctx, &types.MsgClaimContributorReward{
		Signer:        signer,
		PasskeyPubkey: "passkey_pubkey_2",
	})
	require.ErrorIs(t, err, types.ErrNotMigrated)
	require.Len(t, bankKeeper.sendCalls, 0)
}

func TestMsgClaimContributorReward_UnauthorizedSigner(t *testing.T) {
	srv, ctx, k, bankKeeper, identityKeeper := setupMsgServerWithMocks(t)
	signer := newTestWalletAddress()
	walletAddress := newTestWalletAddress()

	const (
		passkeyPubkey = "passkey_pubkey_3"
	)

	identityKeeper.walletAddress = walletAddress
	identityKeeper.isMigrated = true
	identityKeeper.found = true

	require.NoError(t, k.SetContributorReward(ctx, passkeyPubkey, 11))

	_, err := srv.ClaimContributorReward(ctx, &types.MsgClaimContributorReward{
		Signer:        signer,
		PasskeyPubkey: passkeyPubkey,
	})
	require.ErrorIs(t, err, types.ErrUnauthorizedClaim)
	require.Len(t, bankKeeper.sendCalls, 0)
}

func TestMsgClaimContributorReward_ZeroBalance(t *testing.T) {
	srv, ctx, _, bankKeeper, identityKeeper := setupMsgServerWithMocks(t)
	signer := newTestWalletAddress()

	identityKeeper.walletAddress = signer
	identityKeeper.isMigrated = true
	identityKeeper.found = true

	_, err := srv.ClaimContributorReward(ctx, &types.MsgClaimContributorReward{
		Signer:        signer,
		PasskeyPubkey: "passkey_pubkey_4",
	})
	require.ErrorIs(t, err, types.ErrNothingToClaim)
	require.Len(t, bankKeeper.sendCalls, 0)
}
