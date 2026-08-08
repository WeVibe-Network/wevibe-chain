package keeper_test

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"strings"
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
		Signer: "s",
		OrgId:  "org-1",
		Epoch:  1,
		Serves: []*types.ServeEntry{serveEntry("org-1", 1, h, "sk", "c1", nonce32(0x01))},
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
		{"empty signer", &types.MsgSubmitServeBatch{OrgId: "org-1", Epoch: 1, Serves: []*types.ServeEntry{serveEntry("org-1", 1, h, "k", "c", nonce32(1))}}},
		{"empty org", &types.MsgSubmitServeBatch{Signer: "s", Epoch: 1, Serves: []*types.ServeEntry{serveEntry("org-1", 1, h, "k", "c", nonce32(1))}}},
		{"empty batch", &types.MsgSubmitServeBatch{Signer: "s", OrgId: "org-1"}},
		{"bad hash len", &types.MsgSubmitServeBatch{Signer: "s", OrgId: "org-1", Epoch: 1, Serves: []*types.ServeEntry{serveEntry("org-1", 1, []byte{1, 2}, "k", "c", nonce32(1))}}},
		{"empty contributor", &types.MsgSubmitServeBatch{Signer: "s", OrgId: "org-1", Epoch: 1, Serves: []*types.ServeEntry{serveEntry("org-1", 1, h, "k", "", nonce32(1))}}},
		{"bad serve pubkey len", func() *types.MsgSubmitServeBatch {
			entry := serveEntry("org-1", 1, h, "k", "c", nonce32(1))
			entry.ServeKeyPubkey = []byte{1}
			return &types.MsgSubmitServeBatch{Signer: "s", OrgId: "org-1", Epoch: 1, Serves: []*types.ServeEntry{entry}}
		}()},
		{"bad serve signature len", func() *types.MsgSubmitServeBatch {
			entry := serveEntry("org-1", 1, h, "k", "c", nonce32(1))
			entry.ServeSig = []byte{1}
			return &types.MsgSubmitServeBatch{Signer: "s", OrgId: "org-1", Epoch: 1, Serves: []*types.ServeEntry{entry}}
		}()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := srv.SubmitServeBatch(env.ctx, tc.msg)
			require.Error(t, err)
		})
	}
}

func TestMsgSubmitServeBatch_OrgNotFound(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	h := hash32(0x04)
	env.mem.approve("ghost-org", h)
	_, err := srv.SubmitServeBatch(env.ctx, &types.MsgSubmitServeBatch{
		Signer: "s", OrgId: "ghost-org", Epoch: 1,
		Serves: []*types.ServeEntry{serveEntry("ghost-org", 1, h, "k", "c", nonce32(0x04))},
	})
	require.ErrorIs(t, err, types.ErrOrgNotFound)
}

func TestMsgSubmitDenialBatch_HappyPath(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	h := hash32(0x05)
	env.mem.approve("org-1", h)
	originatingServe := serveEntry("org-1", 1, h, "sk", "c1", nonce32(0x05))

	// First create an originating v2 serve receipt.
	_, err := srv.SubmitServeBatch(env.ctx, &types.MsgSubmitServeBatch{
		Signer: "s", OrgId: "org-1", Epoch: 1,
		Serves: []*types.ServeEntry{originatingServe},
	})
	require.NoError(t, err)
	originatingFingerprint := serveFingerprint(originatingServe, 1)
	denial := denialEntry("org-1", 1, h, "sk", originatingFingerprint, nonce32(0x06), "bad")

	// Now deny it using the originating serve fingerprint and matching memory hash.
	resp, err := srv.SubmitDenialBatch(env.sdkCtx(), &types.MsgSubmitDenialBatch{
		Signer: "s", OrgId: "org-1", Epoch: 1,
		Entries: []*types.DenialEntry{denial},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), resp.Accepted)
	require.Equal(t, uint64(0), resp.Rejected)
	denialFingerprint := types.ComputeDenialFingerprint("org-1", h, 1, denial.ServeKeyPubkey, denial.ServeFingerprint)
	require.True(t, env.k.HasDenialFingerprint(env.ctx, denialFingerprint))
	require.Equal(t, uint64(1), env.k.GetMemoryDenialCount(env.ctx, "org-1", h, 1))
}

func TestMsgSubmitDenialBatch_RejectsUnknownServeFingerprint(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	h := hash32(0x06)
	env.mem.approve("org-1", h)
	unknown := denialEntry("org-1", 1, h, "dk", hash32(0xAB), nonce32(0xAB), "bad")

	resp, err := srv.SubmitDenialBatch(env.sdkCtx(), &types.MsgSubmitDenialBatch{
		Signer: "s", OrgId: "org-1", Epoch: 1,
		Entries: []*types.DenialEntry{unknown},
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
	originatingServe := serveEntry("org-1", 1, h, "sk", "c1", nonce32(0x07))

	_, err := srv.SubmitServeBatch(env.ctx, &types.MsgSubmitServeBatch{
		Signer: "s", OrgId: "org-1", Epoch: 1,
		Serves: []*types.ServeEntry{originatingServe},
	})
	require.NoError(t, err)

	denial := denialEntry("org-1", 1, other, "sk", serveFingerprint(originatingServe, 1), nonce32(0x08), "bad")

	// Deny with a different memory hash than the originating receipt.
	resp, err := srv.SubmitDenialBatch(env.sdkCtx(), &types.MsgSubmitDenialBatch{
		Signer: "s", OrgId: "org-1", Epoch: 1,
		Entries: []*types.DenialEntry{denial},
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
	originatingServe := serveEntry("org-1", 1, h, "sk", "c1", nonce32(0x09))

	_, err := srv.SubmitServeBatch(env.ctx, &types.MsgSubmitServeBatch{
		Signer: "s", OrgId: "org-1", Epoch: 1,
		Serves: []*types.ServeEntry{originatingServe},
	})
	require.NoError(t, err)

	entry := denialEntry("org-1", 1, h, "sk", serveFingerprint(originatingServe, 1), nonce32(0x0A), "bad")
	r1, err := srv.SubmitDenialBatch(env.sdkCtx(), &types.MsgSubmitDenialBatch{
		Signer: "s", OrgId: "org-1", Epoch: 1, Entries: []*types.DenialEntry{entry},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), r1.Accepted)

	// Second denial with same denial fingerprint is rejected as duplicate.
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
		Entries: []*types.DenialEntry{denialEntry("ghost", 1, hash32(1), "dk", hash32(1), nonce32(1), "bad")},
	})
	require.ErrorIs(t, err, types.ErrOrgNotFound)
}

func TestMsgSubmitDenialBatch_ValidateBasic(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	_, err := srv.SubmitDenialBatch(env.sdkCtx(), &types.MsgSubmitDenialBatch{Signer: "", OrgId: "org-1"})
	require.Error(t, err)
}

func TestMsgSubmitDenialBatch_NegAnchorRejectedAtValidateBasic(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	h := hash32(0x0B)
	denial := denialEntry("org-1", 1, h, "sk", hash32(0x0C), nonce32(0x0D), "bad")
	denial.NegAnchor = []byte{1}
	_, err := srv.SubmitDenialBatch(env.sdkCtx(), &types.MsgSubmitDenialBatch{Signer: "s", OrgId: "org-1", Epoch: 1, Entries: []*types.DenialEntry{denial}})
	require.ErrorIs(t, err, types.ErrNegAnchorInert)
}

func TestMsgSubmitEventBatch_GatesAndBatchCap(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	h := hash32(0x61)
	env.mem.approve("org-1", h)
	entry, _ := outcomeEventEntry(t, "org-1", 1, h, 0x61)

	_, err := srv.SubmitEventBatch(env.sdkCtx(), &types.MsgSubmitEventBatch{Signer: "attacker", OrgId: "org-1", Epoch: 1, Events: []*types.EventEntry{entry}})
	require.ErrorIs(t, err, types.ErrUnauthorized)

	parked := *entry
	parked.EventType = types.EventType_EVENT_TYPE_CONTEST
	_, err = srv.SubmitEventBatch(env.sdkCtx(), &types.MsgSubmitEventBatch{Signer: "s", OrgId: "org-1", Epoch: 1, Events: []*types.EventEntry{&parked}})
	require.ErrorIs(t, err, types.ErrEventParked)

	require.NoError(t, env.k.SetParams(env.ctx, types.Params{MaxServesPerBatch: 1, MaxServesPerMemoryPerEpoch: 100}))
	entry2, _ := outcomeEventEntry(t, "org-1", 1, h, 0x62)
	_, err = srv.SubmitEventBatch(env.sdkCtx(), &types.MsgSubmitEventBatch{Signer: "s", OrgId: "org-1", Epoch: 1, Events: []*types.EventEntry{entry, entry2}})
	require.ErrorIs(t, err, types.ErrBatchTooLarge)
}

func TestMsgSubmitEventBatch_HappyPath(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	h := hash32(0x63)
	env.mem.approve("org-1", h)
	entry, _ := outcomeEventEntry(t, "org-1", 2, h, 0x63)
	resp, err := srv.SubmitEventBatch(env.sdkCtx(), &types.MsgSubmitEventBatch{Signer: "s", OrgId: "org-1", Epoch: 2, Events: []*types.EventEntry{entry}})
	require.NoError(t, err)
	require.Equal(t, uint64(1), resp.Accepted)
}

func TestMsgAnchorPolicyVersion_AuthorityGate(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	_, err := srv.AnchorPolicyVersion(env.sdkCtx(), &types.MsgAnchorPolicyVersion{Authority: "intruder", PolicyVersion: "policy-v1", PolicyHash: hash32(0x64)})
	require.ErrorIs(t, err, types.ErrUnauthorized)
	_, err = srv.AnchorPolicyVersion(env.sdkCtx(), &types.MsgAnchorPolicyVersion{Authority: govAuthority, PolicyVersion: "policy-v1", PolicyHash: hash32(0x64)})
	require.NoError(t, err)
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

// ── R-BLAST-RADIUS: only the registered serving key may submit serve/denial ──
// A stolen key that is NOT the org's registered serving key must be rejected.
// (D-S32-CO044-KEY-SEPARATION)

func TestMsgSubmitServeBatch_RejectsNonServingSigner(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	h := hash32(0x31)
	env.mem.approve("org-1", h)

	// "attacker" is a valid, well-formed signer but is NOT the org's serving key.
	_, err := srv.SubmitServeBatch(env.ctx, &types.MsgSubmitServeBatch{
		Signer: "attacker",
		OrgId:  "org-1",
		Epoch:  1,
		Serves: []*types.ServeEntry{serveEntry("org-1", 1, h, "sk", "c1", nonce32(0x31))},
	})
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

func TestMsgSubmitDenialBatch_RejectsNonServingSigner(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)
	h := hash32(0x32)
	env.mem.approve("org-1", h)
	originatingServe := serveEntry("org-1", 1, h, "sk", "c1", nonce32(0x32))

	// Legit serve by the serving key.
	_, err := srv.SubmitServeBatch(env.ctx, &types.MsgSubmitServeBatch{
		Signer: "s", OrgId: "org-1", Epoch: 1,
		Serves: []*types.ServeEntry{originatingServe},
	})
	require.NoError(t, err)

	denial := denialEntry("org-1", 1, h, "sk", serveFingerprint(originatingServe, 1), nonce32(0x33), "bad")

	// Attacker (non-serving key) attempts a denial → rejected before processing.
	_, err = srv.SubmitDenialBatch(env.sdkCtx(), &types.MsgSubmitDenialBatch{
		Signer: "attacker", OrgId: "org-1", Epoch: 1,
		Entries: []*types.DenialEntry{denial},
	})
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

func TestMsgSubmitServeBatch_RejectsWhenNoServingKeyRegistered(t *testing.T) {
	env := setupKeeper(t)
	// org-3 exists but has no registered serving key → no one can serve for it.
	env.org.orgs["org-3"] = true
	srv := keeper.NewMsgServerImpl(env.k)
	h := hash32(0x33)
	env.mem.approve("org-3", h)

	_, err := srv.SubmitServeBatch(env.ctx, &types.MsgSubmitServeBatch{
		Signer: "s", OrgId: "org-3", Epoch: 1,
		Serves: []*types.ServeEntry{serveEntry("org-3", 1, h, "sk", "c1", nonce32(0x33))},
	})
	require.ErrorIs(t, err, types.ErrUnauthorized)
}

func TestSubmitServeBatch_FixedSignedVectorRejectsTampered(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)

	orgID := "org-test"
	epoch := uint64(7)
	hash := make([]byte, 32)
	for i := range hash {
		hash[i] = byte(i + 1)
	}
	env.org.orgs[orgID] = true
	env.org.setServing(orgID, "s")
	env.mem.approve(orgID, hash)

	seed := bytes.Repeat([]byte{0x01}, 32)
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)
	nonce := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	episodeRef := []byte{0xAA, 0xBB, 0xCC, 0xDD}

	body := types.CanonicalServeBody(orgID, hash, epoch, pub, nonce, episodeRef)
	sig := ed25519.Sign(priv, body)
	require.True(t, ed25519.Verify(pub, body, sig))
	fingerprint := types.ComputeServeFingerprint(hash, pub, epoch)
	t.Logf("serve_vector seed=%s pubkey=%s body=%s sig=%s fingerprint=%s",
		hex.EncodeToString(seed),
		hex.EncodeToString(pub),
		strings.ReplaceAll(string(body), "\n", "\\n"),
		hex.EncodeToString(sig),
		hex.EncodeToString(fingerprint),
	)

	validEntry := &types.ServeEntry{
		MemoryContentHash: hash,
		ServeKeyPubkey:    pub,
		ServeSig:          sig,
		ContributorId:     "contrib-1",
		ModelId:           "qwen3:4b",
		TurnCount:         2,
		ContributorWallet: "wallet-1",
		Nonce:             nonce,
		EpisodeRef:        episodeRef,
	}

	resp, err := srv.SubmitServeBatch(env.ctx, &types.MsgSubmitServeBatch{
		Signer: "s", OrgId: orgID, Epoch: epoch, Serves: []*types.ServeEntry{validEntry},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), resp.Accepted)
	require.Equal(t, uint64(0), resp.RejectedDuplicate)
	require.Equal(t, uint64(0), resp.RejectedInvalid)
	require.True(t, env.k.HasServeFingerprint(env.ctx, fingerprint))

	tamperedSig := append([]byte(nil), sig...)
	tamperedSig[0] ^= 0xFF
	require.False(t, ed25519.Verify(pub, body, tamperedSig))
	tamperedEntry := &types.ServeEntry{
		MemoryContentHash: hash,
		ServeKeyPubkey:    pub,
		ServeSig:          tamperedSig,
		ContributorId:     "contrib-1",
		ModelId:           "qwen3:4b",
		TurnCount:         2,
		ContributorWallet: "wallet-1",
		Nonce:             nonce,
		EpisodeRef:        episodeRef,
	}

	tamperedResp, err := srv.SubmitServeBatch(env.ctx, &types.MsgSubmitServeBatch{
		Signer: "s", OrgId: orgID, Epoch: epoch, Serves: []*types.ServeEntry{tamperedEntry},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(0), tamperedResp.Accepted)
	require.Equal(t, uint64(0), tamperedResp.RejectedDuplicate)
	require.Equal(t, uint64(1), tamperedResp.RejectedInvalid)
}

func TestSubmitDenialBatch_FixedSignedVector(t *testing.T) {
	env := setupKeeper(t)
	srv := keeper.NewMsgServerImpl(env.k)

	orgID := "org-test"
	epoch := uint64(7)
	hash := make([]byte, 32)
	for i := range hash {
		hash[i] = byte(i + 1)
	}
	env.org.orgs[orgID] = true
	env.org.setServing(orgID, "s")
	env.mem.approve(orgID, hash)

	seed := bytes.Repeat([]byte{0x01}, 32)
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)
	serveNonce := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	serveEpisodeRef := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	serveBody := types.CanonicalServeBody(orgID, hash, epoch, pub, serveNonce, serveEpisodeRef)
	serveSig := ed25519.Sign(priv, serveBody)
	originatingServe := &types.ServeEntry{
		MemoryContentHash: hash,
		ServeKeyPubkey:    pub,
		ServeSig:          serveSig,
		ContributorId:     "contrib-1",
		ModelId:           "qwen3:4b",
		TurnCount:         2,
		ContributorWallet: "wallet-1",
		Nonce:             serveNonce,
		EpisodeRef:        serveEpisodeRef,
	}
	_, err := srv.SubmitServeBatch(env.ctx, &types.MsgSubmitServeBatch{
		Signer: "s", OrgId: orgID, Epoch: epoch, Serves: []*types.ServeEntry{originatingServe},
	})
	require.NoError(t, err)
	serveFingerprint := types.ComputeServeFingerprint(hash, pub, epoch)

	denialNonce := []byte{0xCA, 0xFE, 0xBA, 0xBE}
	denialBody := types.CanonicalDenialBody(orgID, hash, epoch, pub, serveFingerprint, denialNonce)
	denialSig := ed25519.Sign(priv, denialBody)
	require.True(t, ed25519.Verify(pub, denialBody, denialSig))
	denialFingerprint := types.ComputeDenialFingerprint(orgID, hash, epoch, pub, serveFingerprint)
	t.Logf("denial_vector seed=%s pubkey=%s body=%s sig=%s fingerprint=%s serve_fingerprint=%s",
		hex.EncodeToString(seed),
		hex.EncodeToString(pub),
		strings.ReplaceAll(string(denialBody), "\n", "\\n"),
		hex.EncodeToString(denialSig),
		hex.EncodeToString(denialFingerprint),
		hex.EncodeToString(serveFingerprint),
	)

	entry := &types.DenialEntry{
		MemoryHash:       hash,
		Reason:           "bad",
		ServeKeyPubkey:   pub,
		ServeSig:         denialSig,
		ServeFingerprint: serveFingerprint,
		Nonce:            denialNonce,
	}
	resp, err := srv.SubmitDenialBatch(env.sdkCtx(), &types.MsgSubmitDenialBatch{
		Signer: "s", OrgId: orgID, Epoch: epoch, Entries: []*types.DenialEntry{entry},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), resp.Accepted)
	require.Equal(t, uint64(0), resp.Rejected)

	tamperedSig := append([]byte(nil), denialSig...)
	tamperedSig[0] ^= 0xFF
	require.False(t, ed25519.Verify(pub, denialBody, tamperedSig))
	tampered := &types.DenialEntry{
		MemoryHash:       hash,
		Reason:           "bad",
		ServeKeyPubkey:   pub,
		ServeSig:         tamperedSig,
		ServeFingerprint: serveFingerprint,
		Nonce:            denialNonce,
	}
	tamperedResp, err := srv.SubmitDenialBatch(env.sdkCtx(), &types.MsgSubmitDenialBatch{
		Signer: "s", OrgId: orgID, Epoch: epoch, Entries: []*types.DenialEntry{tampered},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(0), tamperedResp.Accepted)
	require.Equal(t, uint64(1), tamperedResp.Rejected)
}
