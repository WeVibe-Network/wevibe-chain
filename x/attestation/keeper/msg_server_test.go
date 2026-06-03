package keeper_test

import (
	"testing"
	"time"

	logv2 "cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/wevibe-network/wevibe-chain/testutil/keeper"
	"github.com/wevibe-network/wevibe-chain/x/attestation/keeper"
	"github.com/wevibe-network/wevibe-chain/x/attestation/types"
)

func setupMsgServer(t *testing.T) (types.MsgServer, *keeper.Keeper, sdk.Context) {
	t.Helper()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService, cms := testkeeper.NewTestStoreService(t, key)
	logger := logv2.NewNopLogger()
	orgKeeper := &mockOrgKeeper{orgs: map[string]bool{"org-1": true}}
	k := keeper.NewKeeper(storeService, logger, "gov-authority", orgKeeper)
	sdkCtx := sdk.NewContext(cms, tmproto.Header{Height: 42, Time: time.Now().UTC()}, false, logger).
		WithEventManager(sdk.NewEventManager())
	return keeper.NewMsgServerImpl(k), k, sdkCtx
}

func validSessionHash(seed byte) []byte {
	h := make([]byte, types.SessionHashLen)
	for i := range h {
		h[i] = seed
	}
	return h
}

func baseAttestationMsg(seed byte) *types.MsgSubmitSessionAttestation {
	return &types.MsgSubmitSessionAttestation{
		Signer:        "signer",
		OrgId:         "org-1",
		SessionHash:   validSessionHash(seed),
		ModelId:       "qwen3:4b",
		TurnCount:     4,
		TokenCount:    800,
		ProviderType:  types.ProviderType_PROVIDER_TYPE_LOCAL,
		ContributorId: "contributor-1",
		Epoch:         1,
	}
}

func TestSubmitSessionAttestation_DisabledRejectsAndStoresNothing(t *testing.T) {
	srv, k, ctx := setupMsgServer(t)
	msg := baseAttestationMsg(0x01)
	msg.CommitllmReceiptHash = make([]byte, 32)

	resp, err := srv.SubmitSessionAttestation(ctx, msg)
	require.ErrorIs(t, err, types.ErrAttestationDisabled)
	require.Nil(t, resp)

	_, err = k.GetSessionAttestation(ctx, "org-1", msg.SessionHash)
	require.ErrorIs(t, err, types.ErrAttestationNotFound)
	require.Len(t, ctx.EventManager().Events(), 0)
}

func TestSubmitSessionAttestation_LocalNoReceipt(t *testing.T) {
	srv, _, ctx := setupMsgServer(t)
	msg := baseAttestationMsg(0x02) // no receipt hash
	resp, err := srv.SubmitSessionAttestation(ctx, msg)
	require.ErrorIs(t, err, types.ErrAttestationDisabled)
	require.Nil(t, resp)
}

func TestSubmitSessionAttestation_CloudNoSignature(t *testing.T) {
	srv, _, ctx := setupMsgServer(t)
	msg := baseAttestationMsg(0x03)
	msg.ProviderType = types.ProviderType_PROVIDER_TYPE_CLOUD
	resp, err := srv.SubmitSessionAttestation(ctx, msg)
	require.ErrorIs(t, err, types.ErrAttestationDisabled)
	require.Nil(t, resp)
}

func TestSubmitSessionAttestation_CloudWithSignature(t *testing.T) {
	srv, _, ctx := setupMsgServer(t)
	msg := baseAttestationMsg(0x04)
	msg.ProviderType = types.ProviderType_PROVIDER_TYPE_CLOUD
	msg.ProviderSignatureHash = make([]byte, 32)
	resp, err := srv.SubmitSessionAttestation(ctx, msg)
	require.ErrorIs(t, err, types.ErrAttestationDisabled)
	require.Nil(t, resp)
}

func TestSubmitSessionAttestation_DuplicateRejected(t *testing.T) {
	srv, _, ctx := setupMsgServer(t)
	msg := baseAttestationMsg(0x05)

	_, err := srv.SubmitSessionAttestation(ctx, msg)
	require.ErrorIs(t, err, types.ErrAttestationDisabled)

	// Disabled path rejects every submission, including retries.
	_, err = srv.SubmitSessionAttestation(ctx, baseAttestationMsg(0x05))
	require.ErrorIs(t, err, types.ErrAttestationDisabled)
}

func TestSubmitSessionAttestation_OrgNotFound(t *testing.T) {
	srv, _, ctx := setupMsgServer(t)
	msg := baseAttestationMsg(0x06)
	msg.OrgId = "ghost-org"
	_, err := srv.SubmitSessionAttestation(ctx, msg)
	require.ErrorIs(t, err, types.ErrAttestationDisabled)
}

func TestSubmitSessionAttestation_ValidateBasic(t *testing.T) {
	srv, _, ctx := setupMsgServer(t)
	cases := []struct {
		name   string
		mutate func(*types.MsgSubmitSessionAttestation)
	}{
		{"empty signer", func(m *types.MsgSubmitSessionAttestation) { m.Signer = "" }},
		{"empty org", func(m *types.MsgSubmitSessionAttestation) { m.OrgId = "" }},
		{"bad session hash", func(m *types.MsgSubmitSessionAttestation) { m.SessionHash = []byte{1, 2, 3} }},
		{"empty model", func(m *types.MsgSubmitSessionAttestation) { m.ModelId = "" }},
		{"empty contributor", func(m *types.MsgSubmitSessionAttestation) { m.ContributorId = "" }},
		{"unspecified provider", func(m *types.MsgSubmitSessionAttestation) {
			m.ProviderType = types.ProviderType_PROVIDER_TYPE_UNSPECIFIED
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := baseAttestationMsg(0x07)
			tc.mutate(msg)
			_, err := srv.SubmitSessionAttestation(ctx, msg)
			require.Error(t, err)
		})
	}
}

func TestAttestationUpdateParams(t *testing.T) {
	srv, k, ctx := setupMsgServer(t)
	params := types.Params{MaxAttestationsPerEpoch: 123, RequireAttestationForServe: true}

	// Unauthorized.
	_, err := srv.UpdateParams(ctx, &types.MsgUpdateParams{Authority: "intruder", Params: &params})
	require.ErrorIs(t, err, types.ErrUnauthorized)

	// Empty authority => ValidateBasic failure.
	_, err = srv.UpdateParams(ctx, &types.MsgUpdateParams{Authority: "", Params: &params})
	require.Error(t, err)

	// Authorized.
	_, err = srv.UpdateParams(ctx, &types.MsgUpdateParams{Authority: "gov-authority", Params: &params})
	require.NoError(t, err)
	got, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(123), got.MaxAttestationsPerEpoch)
	require.True(t, got.RequireAttestationForServe)
}
