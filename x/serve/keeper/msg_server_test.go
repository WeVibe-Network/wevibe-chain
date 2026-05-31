package keeper_test

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/wevibe-network/wevibe-chain/x/serve/keeper"
	"github.com/wevibe-network/wevibe-chain/x/serve/types"
)

// sdkCtx wraps the env in an SDK context with an event manager so that
// handlers which emit events do not panic.
func (env *serveTestEnv) sdkCtx() sdk.Context {
	return sdk.NewContext(env.cms, tmproto.Header{Height: 1, Time: time.Now().UTC()}, false, env.k.Logger(env.ctx)).
		WithEventManager(sdk.NewEventManager())
}

func TestMsgSubmitServeBatch_HappyPath(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	h := hash32(0x01)
	env.mem.approve("org-1", h)

	resp, err := srv.SubmitServeBatch(env.ctx, &types.MsgSubmitServeBatch{
		Signer: "signer",
		OrgId:  "org-1",
		Epoch:  1,
		Serves: []*types.ServeEntry{serveEntry(h, "sk", "c1", nullifier32(0x01))},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), resp.Accepted)
	require.Equal(t, uint64(0), resp.RejectedDuplicate)
	require.Equal(t, uint64(0), resp.RejectedInvalid)
}

func TestMsgSubmitServeBatch_ValidateBasicErrors(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	h := hash32(0x02)

	cases := []struct {
		name string
		msg  *types.MsgSubmitServeBatch
	}{
		{"empty signer", &types.MsgSubmitServeBatch{OrgId: "org-1", Serves: []*types.ServeEntry{serveEntry(h, "k", "c", nullifier32(1))}}},
		{"empty org", &types.MsgSubmitServeBatch{Signer: "s", Serves: []*types.ServeEntry{serveEntry(h, "k", "c", nullifier32(1))}}},
		{"empty batch", &types.MsgSubmitServeBatch{Signer: "s", OrgId: "org-1"}},
		{"bad hash len", &types.MsgSubmitServeBatch{Signer: "s", OrgId: "org-1", Serves: []*types.ServeEntry{serveEntry([]byte{1, 2}, "k", "c", nullifier32(1))}}},
		{"empty serve key", &types.MsgSubmitServeBatch{Signer: "s", OrgId: "org-1", Serves: []*types.ServeEntry{serveEntry(h, "", "c", nullifier32(1))}}},
		{"empty contributor", &types.MsgSubmitServeBatch{Signer: "s", OrgId: "org-1", Serves: []*types.ServeEntry{serveEntry(h, "k", "", nullifier32(1))}}},
		{"bad nullifier len", &types.MsgSubmitServeBatch{Signer: "s", OrgId: "org-1", Serves: []*types.ServeEntry{serveEntry(h, "k", "c", []byte{1})}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := srv.SubmitServeBatch(env.ctx, tc.msg)
			require.Error(t, err)
		})
	}
}

func TestMsgSubmitServeBatch_EmptyMatchedKeywords(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	h := hash32(0x03)
	bad := serveEntry(h, "k", "c", nullifier32(1))
	bad.MatchedKeywords = nil
	_, err := srv.SubmitServeBatch(env.ctx, &types.MsgSubmitServeBatch{
		Signer: "s", OrgId: "org-1", Epoch: 1, Serves: []*types.ServeEntry{bad},
	})
	require.Error(t, err)
}

func TestMsgSubmitServeBatch_OrgNotFound(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	h := hash32(0x04)
	env.mem.approve("ghost-org", h)
	_, err := srv.SubmitServeBatch(env.ctx, &types.MsgSubmitServeBatch{
		Signer: "s", OrgId: "ghost-org", Epoch: 1,
		Serves: []*types.ServeEntry{serveEntry(h, "k", "c", nullifier32(0x04))},
	})
	require.ErrorIs(t, err, types.ErrOrgNotFound)
}

func TestMsgSubmitDenialBatch_HappyPath(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	h := hash32(0x05)
	env.mem.approve("org-1", h)
	null := nullifier32(0x05)

	// First create an originating serve attestation (with matched keywords).
	_, err := srv.SubmitServeBatch(env.ctx, &types.MsgSubmitServeBatch{
		Signer: "s", OrgId: "org-1", Epoch: 1,
		Serves: []*types.ServeEntry{serveEntry(h, "sk", "c1", null)},
	})
	require.NoError(t, err)

	// Now deny it using the same nullifier and matching memory hash.
	resp, err := srv.SubmitDenialBatch(env.sdkCtx(), &types.MsgSubmitDenialBatch{
		Signer: "s", OrgId: "org-1", Epoch: 1,
		Entries: []*types.DenialEntry{{MemoryHash: h, Nullifier: null, DenyKey: "dk", Reason: "bad"}},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), resp.Accepted)
	require.Equal(t, uint64(0), resp.Rejected)
	require.True(t, env.k.HasDenialNullifier(env.ctx, null))
	require.Equal(t, uint64(1), env.k.GetMemoryDenialCount(env.ctx, "org-1", h, 1))
	require.Equal(t, 1, env.mem.decayCalls)
}

func TestMsgSubmitDenialBatch_RejectsUnknownNullifier(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	h := hash32(0x06)
	env.mem.approve("org-1", h)

	resp, err := srv.SubmitDenialBatch(env.sdkCtx(), &types.MsgSubmitDenialBatch{
		Signer: "s", OrgId: "org-1", Epoch: 1,
		Entries: []*types.DenialEntry{{MemoryHash: h, Nullifier: nullifier32(0xAB), DenyKey: "dk"}},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(0), resp.Accepted)
	require.Equal(t, uint64(1), resp.Rejected)
}

func TestMsgSubmitDenialBatch_RejectsMismatchedMemoryHash(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	h := hash32(0x07)
	other := hash32(0x08)
	env.mem.approve("org-1", h)
	env.mem.approve("org-1", other)
	null := nullifier32(0x07)

	_, err := srv.SubmitServeBatch(env.ctx, &types.MsgSubmitServeBatch{
		Signer: "s", OrgId: "org-1", Epoch: 1,
		Serves: []*types.ServeEntry{serveEntry(h, "sk", "c1", null)},
	})
	require.NoError(t, err)

	// Deny with a different memory hash than the originating attestation.
	resp, err := srv.SubmitDenialBatch(env.sdkCtx(), &types.MsgSubmitDenialBatch{
		Signer: "s", OrgId: "org-1", Epoch: 1,
		Entries: []*types.DenialEntry{{MemoryHash: other, Nullifier: null, DenyKey: "dk"}},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(0), resp.Accepted)
	require.Equal(t, uint64(1), resp.Rejected)
}

func TestMsgSubmitDenialBatch_DuplicateDenialRejected(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	h := hash32(0x09)
	env.mem.approve("org-1", h)
	null := nullifier32(0x09)

	_, err := srv.SubmitServeBatch(env.ctx, &types.MsgSubmitServeBatch{
		Signer: "s", OrgId: "org-1", Epoch: 1,
		Serves: []*types.ServeEntry{serveEntry(h, "sk", "c1", null)},
	})
	require.NoError(t, err)

	entry := &types.DenialEntry{MemoryHash: h, Nullifier: null, DenyKey: "dk"}
	r1, err := srv.SubmitDenialBatch(env.sdkCtx(), &types.MsgSubmitDenialBatch{
		Signer: "s", OrgId: "org-1", Epoch: 1, Entries: []*types.DenialEntry{entry},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), r1.Accepted)

	// Second denial with same nullifier is rejected as duplicate.
	r2, err := srv.SubmitDenialBatch(env.sdkCtx(), &types.MsgSubmitDenialBatch{
		Signer: "s", OrgId: "org-1", Epoch: 1, Entries: []*types.DenialEntry{entry},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(0), r2.Accepted)
	require.Equal(t, uint64(1), r2.Rejected)
}

func TestMsgSubmitDenialBatch_OrgNotFound(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	_, err := srv.SubmitDenialBatch(env.sdkCtx(), &types.MsgSubmitDenialBatch{
		Signer: "s", OrgId: "ghost", Epoch: 1,
		Entries: []*types.DenialEntry{{MemoryHash: hash32(1), Nullifier: nullifier32(1), DenyKey: "dk"}},
	})
	require.ErrorIs(t, err, types.ErrOrgNotFound)
}

func TestMsgSubmitDenialBatch_ValidateBasic(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	_, err := srv.SubmitDenialBatch(env.sdkCtx(), &types.MsgSubmitDenialBatch{Signer: "", OrgId: "org-1"})
	require.Error(t, err)
}

func TestMsgUpdateParams_AuthorizedAndUnauthorized(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	newParams := types.DefaultParams()
	newParams.MaxServesPerBatch = 42

	// Unauthorized authority rejected.
	_, err := srv.UpdateParams(env.ctx, &types.MsgUpdateParams{Authority: "intruder", Params: &newParams})
	require.ErrorIs(t, err, types.ErrUnauthorized)

	// Empty authority fails ValidateBasic.
	_, err = srv.UpdateParams(env.ctx, &types.MsgUpdateParams{Authority: "", Params: &newParams})
	require.Error(t, err)

	// Correct authority accepted and persisted.
	_, err = srv.UpdateParams(env.ctx, &types.MsgUpdateParams{Authority: govAuthority, Params: &newParams})
	require.NoError(t, err)
	got, err := env.k.GetParams(env.ctx)
	require.NoError(t, err)
	require.Equal(t, newParams.MaxServesPerBatch, got.MaxServesPerBatch)
}
