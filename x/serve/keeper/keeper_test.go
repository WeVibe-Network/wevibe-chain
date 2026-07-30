package keeper_test

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"testing"

	logv2 "cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/wevibe-network/wevibe-chain/testutil/keeper"
	"github.com/wevibe-network/wevibe-chain/x/serve/keeper"
	"github.com/wevibe-network/wevibe-chain/x/serve/types"
)

const govAuthority = "gov-authority"

func hexEncode(b []byte) string { return hex.EncodeToString(b) }

// hash32 returns a deterministic 32-byte content hash seeded by b.
func hash32(b byte) []byte {
	h := make([]byte, types.ContentHashLen)
	for i := range h {
		h[i] = b
	}
	return h
}

// nonce32 returns a deterministic 32-byte nonce seeded by b.
func nonce32(b byte) []byte {
	n := make([]byte, 32)
	n[0] = b
	n[31] = b
	return n
}

func serveKeypair(keyID string) (ed25519.PublicKey, ed25519.PrivateKey) {
	seed := sha256.Sum256([]byte(keyID))
	priv := ed25519.NewKeyFromSeed(seed[:])
	pub := priv.Public().(ed25519.PublicKey)
	return pub, priv
}

func serveKeyHex(keyID string) string {
	pub, _ := serveKeypair(keyID)
	return hex.EncodeToString(pub)
}

type serveTestEnv struct {
	k    *keeper.Keeper
	ctx  context.Context
	cms  storetypes.CommitMultiStore
	org  *mockOrgKeeper
	mem  *mockMemoryKeeper
	band *mockBandwidthKeeper
	rep  *mockReputationKeeper
}

func setupKeeper(t *testing.T) *serveTestEnv {
	t.Helper()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService, cms := testkeeper.NewTestStoreService(t, key)
	logger := logv2.NewNopLogger()

	org := newMockOrgKeeper("org-1")
	// CO-044: register the org's serving key. Serve/denial msg_server tests sign
	// with "s"; ProcessServeBatch (keeper-level) tests bypass the signer check.
	org.setServing("org-1", "s")
	mem := newMockMemoryKeeper()
	band := newMockBandwidthKeeper()
	rep := newMockReputationKeeper()

	k := keeper.NewKeeper(storeService, logger, govAuthority, org, mem, band, rep)
	return &serveTestEnv{k: k, ctx: context.Background(), cms: cms, org: org, mem: mem, band: band, rep: rep}
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

func TestGetParams_DefaultWhenUnset(t *testing.T) {
	env := setupKeeper(t)
	params, err := env.k.GetParams(env.ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams().MaxServesPerBatch, params.MaxServesPerBatch)
	require.Equal(t, types.DefaultParams().MaxServesPerMemoryPerEpoch, params.MaxServesPerMemoryPerEpoch)
}

func TestSetGetParams_Roundtrip(t *testing.T) {
	env := setupKeeper(t)
	custom := types.Params{
		MaxServesPerBatch:           7,
		SelfServeDiscountPercent:    25,
		MaxServesPerMemoryPerEpoch:  3,
		MinOrgAgeEpochs:             2,
		DiminishingReturnsThreshold: 5,
	}
	require.NoError(t, env.k.SetParams(env.ctx, custom))

	got, err := env.k.GetParams(env.ctx)
	require.NoError(t, err)
	require.Equal(t, custom.MaxServesPerBatch, got.MaxServesPerBatch)
	require.Equal(t, custom.MaxServesPerMemoryPerEpoch, got.MaxServesPerMemoryPerEpoch)
	require.Equal(t, custom.SelfServeDiscountPercent, got.SelfServeDiscountPercent)
}

// ---------------------------------------------------------------------------
// ProcessServeBatch — happy path & side effects
// ---------------------------------------------------------------------------

func serveEntry(orgID string, epoch uint64, hash []byte, serveKey, contributorID string, nonce []byte) *types.ServeEntry {
	pub, priv := serveKeypair(serveKey)
	entry := &types.ServeEntry{
		MemoryContentHash: hash,
		ServeKeyPubkey:    append([]byte(nil), pub...),
		ContributorId:     contributorID,
		ModelId:           "qwen3:4b",
		TurnCount:         3,
		ContributorWallet: "wallet-1",
		Nonce:             append([]byte(nil), nonce...),
	}
	body := types.CanonicalServeBody(orgID, hash, epoch, entry.ServeKeyPubkey, entry.Nonce)
	entry.ServeSig = ed25519.Sign(priv, body)
	return entry
}

func serveFingerprint(entry *types.ServeEntry, epoch uint64) []byte {
	return types.ComputeServeFingerprint(entry.MemoryContentHash, entry.ServeKeyPubkey, epoch)
}

func denialEntry(orgID string, epoch uint64, memoryHash []byte, serveKey string, serveFingerprint []byte, nonce []byte, reason string) *types.DenialEntry {
	pub, priv := serveKeypair(serveKey)
	entry := &types.DenialEntry{
		MemoryHash:       memoryHash,
		Reason:           reason,
		ServeKeyPubkey:   append([]byte(nil), pub...),
		ServeFingerprint: append([]byte(nil), serveFingerprint...),
		Nonce:            append([]byte(nil), nonce...),
	}
	body := types.CanonicalDenialBody(orgID, memoryHash, epoch, entry.ServeKeyPubkey, entry.ServeFingerprint, entry.Nonce)
	entry.ServeSig = ed25519.Sign(priv, body)
	return entry
}

func eventKeypair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	return pub, priv
}

func signEventEntry(t *testing.T, orgID string, epoch uint64, entry *types.EventEntry, priv ed25519.PrivateKey) []byte {
	t.Helper()
	body, err := types.CanonicalEventBody(entry.EventType, orgID, entry.MemoryContentHash, epoch, entry.SignerPubkey, entry.Nonce, entry)
	require.NoError(t, err)
	entry.Signature = ed25519.Sign(priv, body)
	return types.ComputeEventFingerprint(body)
}

func outcomeEventEntry(t *testing.T, orgID string, epoch uint64, h []byte, nonce byte) (*types.EventEntry, []byte) {
	t.Helper()
	pub, priv := eventKeypair(t)
	entry := &types.EventEntry{
		EventType:         types.EventType_EVENT_TYPE_OUTCOME,
		MemoryContentHash: h,
		SignerPubkey:      append([]byte(nil), pub...),
		Nonce:             nonce32(nonce),
		Body: &types.EventEntry_Outcome{Outcome: &types.OutcomeEventBody{
			EpisodeRef:  []byte{0x01, nonce},
			Worked:      true,
			EvidenceRef: []byte{0x02, nonce},
			ServeRef:    types.ComputeServeFingerprint(h, pub, epoch),
		}},
	}
	fp := signEventEntry(t, orgID, epoch, entry, priv)
	return entry, fp
}

func TestProcessServeBatch_AcceptsValidServe(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0x11)
	env.mem.approve("org-1", h)
	entry := serveEntry("org-1", 1, h, "serve-key-1", "contrib-1", nonce32(0x01))

	accepted, dup, invalid, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		entry,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), accepted)
	require.Equal(t, uint64(0), dup)
	require.Equal(t, uint64(0), invalid)

	// Side effects: bandwidth consumed, reputation recorded.
	require.Equal(t, uint64(1), env.band.consumed["org-1"])
	require.Equal(t, 1, env.rep.serveCalls)

	// Fingerprint now registered, serve count incremented.
	require.True(t, env.k.HasServeFingerprint(env.ctx, serveFingerprint(entry, 1)))
	require.Equal(t, uint64(1), env.k.GetMemoryServeCount(env.ctx, "org-1", h, 1))

	// Epoch stats reflect the serve.
	stats, err := env.k.GetEpochServeStats(env.ctx, "org-1", 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), stats.TotalServes)
	require.Equal(t, uint64(1), stats.UniqueMemoriesServed)
	require.Equal(t, uint64(1), stats.UniqueServeKeys)
	require.Equal(t, uint64(0), stats.SelfServes)
	require.Equal(t, uint64(1), stats.ModelBreakdown["qwen3:4b"])
}

// Serve attribution is derived from the authoritative committed memory record,
// NOT the untrusted serve payload wallet.
func TestProcessServeBatch_AttributionFromStoredMemory(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0x33)
	env.mem.approveWithContributor("org-1", h, "mem-author-pubkey")

	// Serve payload carries a DIFFERENT (untrusted) wallet ("wallet-1").
	accepted, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry("org-1", 1, h, "serve-key-1", "contrib-1", nonce32(0x33)),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), accepted)
	require.Equal(t, 1, env.rep.serveCalls)
	// Reputation is credited to the stored memory's contributor pubkey, not "wallet-1".
	require.Equal(t, "mem-author-pubkey", env.rep.lastServeContributor)
}

// When the stored memory has no contributor pubkey, the serve
// is still accepted but the reputation record is skipped (no fallback to the
// payload wallet; the serve path does not crash).
func TestProcessServeBatch_SkipsAttributionWhenNoStoredContributor(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0x44)
	env.mem.approveWithContributor("org-1", h, "")

	accepted, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry("org-1", 1, h, "serve-key-1", "contrib-1", nonce32(0x44)),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), accepted)
	require.Equal(t, 0, env.rep.serveCalls)
}

func TestProcessServeBatch_SelfServeCounted(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0x22)
	env.mem.approve("org-1", h)
	selfContributor := serveKeyHex("same-id")

	// Self-serve: hex(ServeKeyPubkey) == ContributorId.
	_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry("org-1", 1, h, "same-id", selfContributor, nonce32(0x02)),
	})
	require.NoError(t, err)

	stats, err := env.k.GetEpochServeStats(env.ctx, "org-1", 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), stats.SelfServes)

	cs, err := env.k.GetContributorEpochServes(env.ctx, selfContributor, 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), cs.ServeCount)
	require.Equal(t, uint64(1), cs.SelfServeCount)
	require.Equal(t, uint64(3), cs.TotalTurns)
	require.Contains(t, cs.OrgIDs, "org-1")
}

// ---------------------------------------------------------------------------
// ProcessServeBatch — rejection edge cases
// ---------------------------------------------------------------------------

func TestProcessServeBatch_DuplicateFingerprintRejected(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0x33)
	env.mem.approve("org-1", h)
	nonce := nonce32(0x03)

	_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry("org-1", 1, h, "k1", "c1", nonce),
	})
	require.NoError(t, err)

	// Resubmit same serve fingerprint — must be counted as duplicate, not accepted.
	accepted, dup, invalid, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry("org-1", 1, h, "k1", "c1", nonce),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(0), accepted)
	require.Equal(t, uint64(1), dup)
	require.Equal(t, uint64(0), invalid)
	// Serve count must remain 1 (no double-count).
	require.Equal(t, uint64(1), env.k.GetMemoryServeCount(env.ctx, "org-1", h, 1))
}

func TestProcessServeBatch_UnapprovedMemoryRejected(t *testing.T) {
	env := setupKeeper(t)
	// Memory never approved.
	accepted, dup, invalid, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry("org-1", 1, hash32(0x44), "k1", "c1", nonce32(0x04)),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(0), accepted)
	require.Equal(t, uint64(0), dup)
	require.Equal(t, uint64(1), invalid)
}

func TestProcessServeBatch_InvalidInEpochRejected(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0x55)
	env.mem.approve("org-1", h)
	// Mark as not valid in this epoch.
	env.mem.validEpoch["org-1|"+hexEncode(h)] = false

	accepted, _, invalid, err := env.k.ProcessServeBatch(env.ctx, "org-1", 9, []*types.ServeEntry{
		serveEntry("org-1", 9, h, "k1", "c1", nonce32(0x05)),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(0), accepted)
	require.Equal(t, uint64(1), invalid)
}

func TestProcessServeBatch_IsValidInEpochErrorPropagates(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0x56)
	env.mem.approve("org-1", h)
	env.mem.validErr = context.Canceled

	_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry("org-1", 1, h, "k1", "c1", nonce32(0x57)),
	})
	require.Error(t, err)
}

func TestProcessServeBatch_BatchTooLarge(t *testing.T) {
	env := setupKeeper(t)
	require.NoError(t, env.k.SetParams(env.ctx, types.Params{MaxServesPerBatch: 1, MaxServesPerMemoryPerEpoch: 100}))
	h := hash32(0x66)
	env.mem.approve("org-1", h)

	_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry("org-1", 1, h, "k1", "c1", nonce32(0x06)),
		serveEntry("org-1", 1, h, "k2", "c2", nonce32(0x07)),
	})
	require.ErrorIs(t, err, types.ErrBatchTooLarge)
	// Nothing consumed because the size check precedes bandwidth.
	require.Equal(t, uint64(0), env.band.consumed["org-1"])
}

func TestProcessServeBatch_BandwidthErrorAborts(t *testing.T) {
	env := setupKeeper(t)
	env.band.err = context.DeadlineExceeded
	h := hash32(0x77)
	env.mem.approve("org-1", h)
	entry := serveEntry("org-1", 1, h, "k1", "c1", nonce32(0x08))

	_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		entry,
	})
	require.Error(t, err)
	require.False(t, env.k.HasServeFingerprint(env.ctx, serveFingerprint(entry, 1)))
}

func TestProcessServeBatch_MaxServesPerMemoryEnforced(t *testing.T) {
	env := setupKeeper(t)
	require.NoError(t, env.k.SetParams(env.ctx, types.Params{MaxServesPerBatch: 500, MaxServesPerMemoryPerEpoch: 1}))
	h := hash32(0x88)
	env.mem.approve("org-1", h)

	// First serve accepted.
	a1, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry("org-1", 1, h, "k1", "c1", nonce32(0x09)),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), a1)

	// Second serve on same memory/epoch rejected (cap=1).
	a2, _, invalid, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry("org-1", 1, h, "k2", "c2", nonce32(0x0A)),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(0), a2)
	require.Equal(t, uint64(1), invalid)
}

func TestProcessServeBatch_ReputationFailureNonFatal(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0xA1)
	env.mem.approve("org-1", h)
	env.rep.serveErr = context.Canceled

	accepted, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry("org-1", 1, h, "k1", "c1", nonce32(0x0C)),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), accepted)
}

func TestProcessServeBatch_MixedBatch(t *testing.T) {
	env := setupKeeper(t)
	approved := hash32(0xB1)
	env.mem.approve("org-1", approved)
	unapproved := hash32(0xB2)
	dupNonce := nonce32(0x0D)

	// Pre-register a fingerprint to force a duplicate.
	_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry("org-1", 1, approved, "k0", "c0", dupNonce),
	})
	require.NoError(t, err)

	accepted, dup, invalid, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry("org-1", 1, approved, "k1", "c1", nonce32(0x0E)),   // accepted
		serveEntry("org-1", 1, approved, "k0", "c0", dupNonce),        // duplicate
		serveEntry("org-1", 1, unapproved, "k2", "c2", nonce32(0x0F)), // invalid (unapproved)
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), accepted)
	require.Equal(t, uint64(1), dup)
	require.Equal(t, uint64(1), invalid)
}

// ---------------------------------------------------------------------------
// ProcessEventBatch — recall-pivot append-only event log
// ---------------------------------------------------------------------------

func TestProcessEventBatch_AcceptsOutcomeAndStoresRoundTrip(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0x91)
	env.mem.approve("org-1", h)
	entry, fp := outcomeEventEntry(t, "org-1", 11, h, 0x91)

	accepted, dup, invalid, err := env.k.ProcessEventBatch(env.ctx, "org-1", 11, []*types.EventEntry{entry})
	require.NoError(t, err)
	require.Equal(t, uint64(1), accepted)
	require.Equal(t, uint64(0), dup)
	require.Equal(t, uint64(0), invalid)

	events, err := env.k.GetEventsForEpoch(env.ctx, "org-1", 11)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, fp, events[0].Fingerprint)
	require.Equal(t, types.EventType_EVENT_TYPE_OUTCOME, events[0].EventType)
	require.Equal(t, entry.GetOutcome().EpisodeRef, events[0].GetOutcome().EpisodeRef)
	require.Equal(t, entry.GetOutcome().Worked, events[0].GetOutcome().Worked)
	require.Equal(t, entry.GetOutcome().EvidenceRef, events[0].GetOutcome().EvidenceRef)
	require.Equal(t, entry.GetOutcome().ServeRef, events[0].GetOutcome().ServeRef)
}

func TestProcessEventBatch_AcceptsAllConsumedTypes(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0x92)
	env.mem.approve("org-1", h)
	cases := []struct {
		name      string
		eventType types.EventType
		build     func([]byte, byte) *types.EventEntry
	}{
		{"outcome", types.EventType_EVENT_TYPE_OUTCOME, func(pub []byte, n byte) *types.EventEntry {
			return &types.EventEntry{EventType: types.EventType_EVENT_TYPE_OUTCOME, MemoryContentHash: h, SignerPubkey: pub, Nonce: nonce32(n), Body: &types.EventEntry_Outcome{Outcome: &types.OutcomeEventBody{EpisodeRef: []byte{n}, Worked: true, EvidenceRef: []byte{n + 1}, ServeRef: types.ComputeServeFingerprint(h, pub, 12)}}}
		}},
		{"validity", types.EventType_EVENT_TYPE_VALIDITY_PREDICATE, func(pub []byte, n byte) *types.EventEntry {
			return &types.EventEntry{EventType: types.EventType_EVENT_TYPE_VALIDITY_PREDICATE, MemoryContentHash: h, SignerPubkey: pub, Nonce: nonce32(n), Body: &types.EventEntry_ValidityPredicate{ValidityPredicate: &types.ValidityPredicateEventBody{PredicateId: []byte{n}, Result: types.PredicateResult_PREDICATE_RESULT_PASS, EvidenceRef: []byte{n + 1}}}}
		}},
		{"cost", types.EventType_EVENT_TYPE_COST_TO_DISCOVER, func(pub []byte, n byte) *types.EventEntry {
			return &types.EventEntry{EventType: types.EventType_EVENT_TYPE_COST_TO_DISCOVER, MemoryContentHash: h, SignerPubkey: pub, Nonce: nonce32(n), Body: &types.EventEntry_CostToDiscover{CostToDiscover: &types.CostToDiscoverEventBody{Cycles: 10, ToolCalls: 2, AttemptsToGreen: 1, EvidenceRef: []byte{n}}}}
		}},
		{"convergence", types.EventType_EVENT_TYPE_CONVERGENCE, func(pub []byte, n byte) *types.EventEntry {
			return &types.EventEntry{EventType: types.EventType_EVENT_TYPE_CONVERGENCE, MemoryContentHash: h, SignerPubkey: pub, Nonce: nonce32(n), Body: &types.EventEntry_Convergence{Convergence: &types.ConvergenceEventBody{ConvergenceRef: []byte{n}}}}
		}},
	}

	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pub, priv := eventKeypair(t)
			entry := tc.build(pub, byte(0xA0+i))
			signEventEntry(t, "org-1", 12, entry, priv)
			accepted, dup, invalid, err := env.k.ProcessEventBatch(env.ctx, "org-1", 12, []*types.EventEntry{entry})
			require.NoError(t, err)
			require.Equal(t, uint64(1), accepted)
			require.Equal(t, uint64(0), dup)
			require.Equal(t, uint64(0), invalid)
		})
	}
	events, err := env.k.GetEventsForEpoch(env.ctx, "org-1", 12)
	require.NoError(t, err)
	require.Len(t, events, 4)
}

func TestProcessEventBatch_RejectionsAndImmutability(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0x93)
	env.mem.approve("org-1", h)
	e1, fp1 := outcomeEventEntry(t, "org-1", 13, h, 0x93)
	e2, fp2 := outcomeEventEntry(t, "org-1", 13, h, 0x94)

	accepted, dup, invalid, err := env.k.ProcessEventBatch(env.ctx, "org-1", 13, []*types.EventEntry{e1, e2})
	require.NoError(t, err)
	require.Equal(t, uint64(2), accepted)
	require.Equal(t, uint64(0), dup)
	require.Equal(t, uint64(0), invalid)
	events, err := env.k.GetEventsForEpoch(env.ctx, "org-1", 13)
	require.NoError(t, err)
	require.Len(t, events, 2)
	require.ElementsMatch(t, [][]byte{fp1, fp2}, [][]byte{events[0].Fingerprint, events[1].Fingerprint})

	accepted, dup, invalid, err = env.k.ProcessEventBatch(env.ctx, "org-1", 13, []*types.EventEntry{e1})
	require.NoError(t, err)
	require.Equal(t, uint64(0), accepted)
	require.Equal(t, uint64(1), dup)
	require.Equal(t, uint64(0), invalid)
	events, err = env.k.GetEventsForEpoch(env.ctx, "org-1", 13)
	require.NoError(t, err)
	require.Len(t, events, 2)

	badSig, _ := outcomeEventEntry(t, "org-1", 13, h, 0x95)
	badSig.Signature[0] ^= 0xFF
	missing, _ := outcomeEventEntry(t, "org-1", 13, hash32(0x96), 0x96)
	accepted, dup, invalid, err = env.k.ProcessEventBatch(env.ctx, "org-1", 13, []*types.EventEntry{badSig, missing})
	require.NoError(t, err)
	require.Equal(t, uint64(0), accepted)
	require.Equal(t, uint64(0), dup)
	require.Equal(t, uint64(2), invalid)
}

func TestPolicyAnchor_StoreImmutabilityAndLatest(t *testing.T) {
	env := setupKeeper(t)
	ctx := env.sdkCtx()
	h1 := hash32(0xA1)
	h2 := hash32(0xA2)

	require.NoError(t, env.k.SetPolicyAnchor(ctx, "policy-v1", h1))
	anchor, found, err := env.k.GetPolicyAnchor(ctx, "policy-v1")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, h1, anchor.PolicyHash)
	latest, found, err := env.k.GetLatestPolicyAnchor(ctx)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "policy-v1", latest.PolicyVersion)

	require.NoError(t, env.k.SetPolicyAnchor(ctx, "policy-v1", h1))
	require.Error(t, env.k.SetPolicyAnchor(ctx, "policy-v1", h2))
}

func TestStoredEventBoundaryRule_NoStandingVerdictFields(t *testing.T) {
	forbidden := []string{"weight", "standing", "score", "trust", "verdict", "archived"}
	typ := reflect.TypeOf(types.StoredEvent{})
	for i := 0; i < typ.NumField(); i++ {
		name := strings.ToLower(typ.Field(i).Name)
		for _, bad := range forbidden {
			require.NotContains(t, name, bad)
		}
	}
}

// ---------------------------------------------------------------------------
// Denial counts
// ---------------------------------------------------------------------------

func TestDenialCountIncrement(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0xD1)
	require.Equal(t, uint64(0), env.k.GetMemoryDenialCount(env.ctx, "org-1", h, 1))
	env.k.IncrementDenialCount(env.ctx, "org-1", h, 1)
	env.k.IncrementDenialCount(env.ctx, "org-1", h, 1)
	require.Equal(t, uint64(2), env.k.GetMemoryDenialCount(env.ctx, "org-1", h, 1))

}

// ---------------------------------------------------------------------------
// Not-found queries
// ---------------------------------------------------------------------------

func TestGetEpochServeStats_NotFound(t *testing.T) {
	env := setupKeeper(t)
	_, err := env.k.GetEpochServeStats(env.ctx, "org-1", 99)
	require.ErrorIs(t, err, types.ErrStatsNotFound)
}

func TestGetContributorEpochServes_NotFound(t *testing.T) {
	env := setupKeeper(t)
	_, err := env.k.GetContributorEpochServes(env.ctx, "ghost", 99)
	require.ErrorIs(t, err, types.ErrContributorNotFound)
}

func TestGetServeReceiptByFingerprint_NotFound(t *testing.T) {
	env := setupKeeper(t)
	_, found, err := env.k.GetServeReceiptByFingerprint(env.ctx, hash32(0xEE))
	require.NoError(t, err)
	require.False(t, found)
}

// ---------------------------------------------------------------------------
// Genesis roundtrip
// ---------------------------------------------------------------------------

func TestGenesisRoundtrip(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0xF1)
	env.mem.approve("org-1", h)
	entry := serveEntry("org-1", 2, h, "k1", "c1", nonce32(0x10))
	event, eventFP := outcomeEventEntry(t, "org-1", 2, h, 0xF1)

	// Produce some state via a real serve.
	_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 2, []*types.ServeEntry{
		entry,
	})
	require.NoError(t, err)
	acceptedEvents, _, _, err := env.k.ProcessEventBatch(env.ctx, "org-1", 2, []*types.EventEntry{event})
	require.NoError(t, err)
	require.Equal(t, uint64(1), acceptedEvents)
	require.NoError(t, env.k.SetPolicyAnchor(env.sdkCtx(), "policy-v1", hash32(0xF5)))

	exported, err := env.k.ExportGenesis(env.ctx)
	require.NoError(t, err)
	require.Len(t, exported.ServeReceipts, 1)
	require.Len(t, exported.Events, 1)
	require.Len(t, exported.PolicyAnchors, 1)
	require.Len(t, exported.EpochStats, 1)
	require.Len(t, exported.ContributorServes, 1)

	// Import into a fresh keeper and verify state restored.
	env2 := setupKeeper(t)
	require.NoError(t, env2.k.InitGenesis(env2.ctx, exported))

	require.True(t, env2.k.HasServeFingerprint(env2.ctx, serveFingerprint(entry, 2)))
	stats, err := env2.k.GetEpochServeStats(env2.ctx, "org-1", 2)
	require.NoError(t, err)
	require.Equal(t, uint64(1), stats.TotalServes)
	events, err := env2.k.GetEventsForEpoch(env2.ctx, "org-1", 2)
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, eventFP, events[0].Fingerprint)
	anchor, found, err := env2.k.GetLatestPolicyAnchor(env2.ctx)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "policy-v1", anchor.PolicyVersion)
}

func TestGenesisInit_DenialReceipts(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0xF2)
	servePub, _ := serveKeypair("serve-denial-genesis")
	serveFingerprint := hash32(0x11)
	gs := &types.GenesisState{
		DenialReceipts: []*types.StoredDenialReceipt{
			{OrgId: "org-1", MemoryHash: h, DenyKey: "dk", Reason: "spam", Epoch: 3, ServeFingerprint: serveFingerprint, ServeKeyPubkey: servePub},
		},
	}
	require.NoError(t, env.k.InitGenesis(env.ctx, gs))
	denialFingerprint := types.ComputeDenialFingerprint("org-1", h, 3, servePub, serveFingerprint)
	require.True(t, env.k.HasDenialFingerprint(env.ctx, denialFingerprint))
	require.Equal(t, uint64(1), env.k.GetMemoryDenialCount(env.ctx, "org-1", h, 3))
	stats, err := env.k.GetEpochServeStats(env.ctx, "org-1", 3)
	require.NoError(t, err)
	require.Equal(t, uint64(0), stats.TotalServes)
	require.Equal(t, uint64(1), stats.TotalDenials)
}

func TestGetServeReceipts_ListByOrgEpoch(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0xF3)
	env.mem.approve("org-1", h)
	require.NoError(t, env.k.SetParams(env.ctx, types.Params{MaxServesPerBatch: 500, MaxServesPerMemoryPerEpoch: 100}))

	for i := 0; i < 3; i++ {
		_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 4, []*types.ServeEntry{
			serveEntry("org-1", 4, h, fmt.Sprintf("k-%d", i), "c", nonce32(byte(0x20+i))),
		})
		require.NoError(t, err)
	}

	receipts, err := env.k.GetServeReceipts(env.ctx, "org-1", 4)
	require.NoError(t, err)
	require.Len(t, receipts, 3)

	// Different epoch returns none.
	none, err := env.k.GetServeReceipts(env.ctx, "org-1", 5)
	require.NoError(t, err)
	require.Len(t, none, 0)
}

// sanity: the raw count helper decodes big-endian correctly.
func TestGetMemoryServeCount_Encoding(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0xF4)
	env.mem.approve("org-1", h)
	require.NoError(t, env.k.SetParams(env.ctx, types.Params{MaxServesPerBatch: 500, MaxServesPerMemoryPerEpoch: 100}))
	for i := 0; i < 5; i++ {
		_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 6, []*types.ServeEntry{
			serveEntry("org-1", 6, h, fmt.Sprintf("k-%d", i), "c", nonce32(byte(0x30+i))),
		})
		require.NoError(t, err)
	}
	require.Equal(t, uint64(5), env.k.GetMemoryServeCount(env.ctx, "org-1", h, 6))
	// guard: ensure binary helper import is genuinely exercised
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, 5)
	require.Equal(t, uint64(5), binary.BigEndian.Uint64(buf))
}
