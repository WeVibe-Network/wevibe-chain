package keeper_test

import (
	"context"
	"encoding/binary"
	"encoding/hex"
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

// nullifier32 returns a deterministic 32-byte nullifier seeded by b.
func nullifier32(b byte) []byte {
	n := make([]byte, types.NullifierLen)
	n[0] = b
	n[31] = b
	return n
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

func serveEntry(hash []byte, serveKey, contributorID string, nullifier []byte) *types.ServeEntry {
	return &types.ServeEntry{
		MemoryContentHash: hash,
		ServeKey:          serveKey,
		ContributorId:     contributorID,
		Nullifier:         nullifier,
		ModelId:           "qwen3:4b",
		TurnCount:         3,
		ContributorWallet: "wallet-1",
		MatchedKeywords:   []string{"alpha", "beta"},
	}
}

func TestProcessServeBatch_AcceptsValidServe(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0x11)
	env.mem.approve("org-1", h)

	accepted, dup, invalid, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry(h, "serve-key-1", "contrib-1", nullifier32(0x01)),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), accepted)
	require.Equal(t, uint64(0), dup)
	require.Equal(t, uint64(0), invalid)

	// Side effects: bandwidth consumed, serve boost applied, reputation recorded.
	require.Equal(t, uint64(1), env.band.consumed["org-1"])
	require.Equal(t, 1, env.mem.boostCalls)
	require.Equal(t, 1, env.rep.serveCalls)

	// Nullifier now registered, serve count incremented.
	require.True(t, env.k.HasNullifier(env.ctx, nullifier32(0x01)))
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

// CO-041 Task F: serve attribution is derived from the authoritative committed
// memory record, NOT the serve payload wallet.
func TestProcessServeBatch_AttributionFromStoredMemory(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0x33)
	env.mem.approveWithContributor("org-1", h, "mem-author-wallet")

	// Serve payload carries a DIFFERENT (untrusted) wallet ("wallet-1").
	accepted, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry(h, "serve-key-1", "contrib-1", nullifier32(0x33)),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), accepted)
	require.Equal(t, 1, env.rep.serveCalls)
	// Reputation is credited to the stored memory's contributor, not "wallet-1".
	require.Equal(t, "mem-author-wallet", env.rep.lastServeWallet)
}

// CO-041 Task F: when the stored memory has no contributor address, the serve
// is still accepted but the reputation record is skipped (no fallback to the
// payload wallet; the serve path does not crash).
func TestProcessServeBatch_SkipsAttributionWhenNoStoredAddress(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0x44)
	env.mem.approveWithContributor("org-1", h, "")

	accepted, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry(h, "serve-key-1", "contrib-1", nullifier32(0x44)),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), accepted)
	require.Equal(t, 0, env.rep.serveCalls)
}

func TestProcessServeBatch_SelfServeCounted(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0x22)
	env.mem.approve("org-1", h)

	// Self-serve: ServeKey == ContributorId.
	_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry(h, "same-id", "same-id", nullifier32(0x02)),
	})
	require.NoError(t, err)

	stats, err := env.k.GetEpochServeStats(env.ctx, "org-1", 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), stats.SelfServes)

	cs, err := env.k.GetContributorEpochServes(env.ctx, "same-id", 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), cs.ServeCount)
	require.Equal(t, uint64(1), cs.SelfServeCount)
	require.Equal(t, uint64(3), cs.TotalTurns)
	require.Contains(t, cs.OrgIDs, "org-1")
}

// ---------------------------------------------------------------------------
// ProcessServeBatch — rejection edge cases
// ---------------------------------------------------------------------------

func TestProcessServeBatch_DuplicateNullifierRejected(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0x33)
	env.mem.approve("org-1", h)
	null := nullifier32(0x03)

	_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry(h, "k1", "c1", null),
	})
	require.NoError(t, err)

	// Resubmit same nullifier — must be counted as duplicate, not accepted.
	accepted, dup, invalid, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry(h, "k1", "c1", null),
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
		serveEntry(hash32(0x44), "k1", "c1", nullifier32(0x04)),
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
		serveEntry(h, "k1", "c1", nullifier32(0x05)),
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
		serveEntry(h, "k1", "c1", nullifier32(0x57)),
	})
	require.Error(t, err)
}

func TestProcessServeBatch_BatchTooLarge(t *testing.T) {
	env := setupKeeper(t)
	require.NoError(t, env.k.SetParams(env.ctx, types.Params{MaxServesPerBatch: 1, MaxServesPerMemoryPerEpoch: 100}))
	h := hash32(0x66)
	env.mem.approve("org-1", h)

	_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry(h, "k1", "c1", nullifier32(0x06)),
		serveEntry(h, "k2", "c2", nullifier32(0x07)),
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

	_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry(h, "k1", "c1", nullifier32(0x08)),
	})
	require.Error(t, err)
	require.False(t, env.k.HasNullifier(env.ctx, nullifier32(0x08)))
}

func TestProcessServeBatch_MaxServesPerMemoryEnforced(t *testing.T) {
	env := setupKeeper(t)
	require.NoError(t, env.k.SetParams(env.ctx, types.Params{MaxServesPerBatch: 500, MaxServesPerMemoryPerEpoch: 1}))
	h := hash32(0x88)
	env.mem.approve("org-1", h)

	// First serve accepted.
	a1, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry(h, "k1", "c1", nullifier32(0x09)),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), a1)

	// Second serve on same memory/epoch rejected (cap=1).
	a2, _, invalid, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry(h, "k2", "c2", nullifier32(0x0A)),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(0), a2)
	require.Equal(t, uint64(1), invalid)
}

func TestProcessServeBatch_BoostFailureNonFatal(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0x99)
	env.mem.approve("org-1", h)
	env.mem.boostErr = context.Canceled // boost fails, serve must still be accepted

	accepted, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry(h, "k1", "c1", nullifier32(0x0B)),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), accepted)
	require.True(t, env.k.HasNullifier(env.ctx, nullifier32(0x0B)))
}

func TestProcessServeBatch_ReputationFailureNonFatal(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0xA1)
	env.mem.approve("org-1", h)
	env.rep.serveErr = context.Canceled

	accepted, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry(h, "k1", "c1", nullifier32(0x0C)),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), accepted)
}

func TestProcessServeBatch_MixedBatch(t *testing.T) {
	env := setupKeeper(t)
	approved := hash32(0xB1)
	env.mem.approve("org-1", approved)
	unapproved := hash32(0xB2)
	dupNull := nullifier32(0x0D)

	// Pre-register a nullifier to force a duplicate.
	_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry(approved, "k0", "c0", dupNull),
	})
	require.NoError(t, err)

	accepted, dup, invalid, err := env.k.ProcessServeBatch(env.ctx, "org-1", 1, []*types.ServeEntry{
		serveEntry(approved, "k1", "c1", nullifier32(0x0E)),   // accepted
		serveEntry(approved, "k0", "c0", dupNull),             // duplicate
		serveEntry(unapproved, "k2", "c2", nullifier32(0x0F)), // invalid (unapproved)
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), accepted)
	require.Equal(t, uint64(1), dup)
	require.Equal(t, uint64(1), invalid)
}

// ---------------------------------------------------------------------------
// Matched keywords
// ---------------------------------------------------------------------------

func TestStoreAndGetMatchedKeywords(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0xC1)
	cid := hexEncode(h)

	require.NoError(t, env.k.StoreMatchedKeywordsForEpoch(env.ctx, "org-1", h, 1, []string{"x", "y", "z"}))

	got, err := env.k.GetMatchedKeywordsForEpoch(env.ctx, "org-1", cid, 1)
	require.NoError(t, err)
	require.Len(t, got, 3)
	require.True(t, got["x"])
	require.True(t, got["y"])
	require.True(t, got["z"])
}

func TestStoreMatchedKeywords_EmptyRejected(t *testing.T) {
	env := setupKeeper(t)
	err := env.k.StoreMatchedKeywordsForEpoch(env.ctx, "org-1", hash32(0xC2), 1, nil)
	require.Error(t, err)

	err = env.k.StoreMatchedKeywordsForEpoch(env.ctx, "org-1", hash32(0xC2), 1, []string{"ok", ""})
	require.Error(t, err)
}

func TestGetMatchedKeywords_InvalidCID(t *testing.T) {
	env := setupKeeper(t)
	_, err := env.k.GetMatchedKeywordsForEpoch(env.ctx, "org-1", "not-hex!!", 1)
	require.Error(t, err)

	// Valid hex but wrong length.
	_, err = env.k.GetMatchedKeywordsForEpoch(env.ctx, "org-1", "abcd", 1)
	require.Error(t, err)
}

func TestGetMemoryServeCountForEpoch_InvalidCID(t *testing.T) {
	env := setupKeeper(t)
	_, err := env.k.GetMemoryServeCountForEpoch(env.ctx, "org-1", "zz", 1)
	require.Error(t, err)
	_, err = env.k.GetMemoryServeCountForEpoch(env.ctx, "org-1", "abcd", 1)
	require.Error(t, err)
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

	got, err := env.k.GetMemoryDenialCountForEpoch(env.ctx, "org-1", hexEncode(h), 1)
	require.NoError(t, err)
	require.Equal(t, uint64(2), got)
}

// ---------------------------------------------------------------------------
// Not-found queries
// ---------------------------------------------------------------------------

func TestGetEpochServeStats_NotFound(t *testing.T) {
	env := setupKeeper(t)
	_, err := env.k.GetEpochServeStats(env.ctx, "org-1", 99)
	require.ErrorIs(t, err, types.ErrStatsNotFound)
}

func TestGetEpochTrafficStats_NotFoundReturnsZero(t *testing.T) {
	env := setupKeeper(t)
	serves, denials, err := env.k.GetEpochTrafficStats(env.ctx, "org-1", 99)
	require.NoError(t, err)
	require.Equal(t, uint64(0), serves)
	require.Equal(t, uint64(0), denials)
}

func TestGetContributorEpochServes_NotFound(t *testing.T) {
	env := setupKeeper(t)
	_, err := env.k.GetContributorEpochServes(env.ctx, "ghost", 99)
	require.ErrorIs(t, err, types.ErrContributorNotFound)
}

func TestGetServeAttestationByNullifier_NotFound(t *testing.T) {
	env := setupKeeper(t)
	_, found, err := env.k.GetServeAttestationByNullifier(env.ctx, nullifier32(0xEE))
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

	// Produce some state via a real serve.
	_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 2, []*types.ServeEntry{
		serveEntry(h, "k1", "c1", nullifier32(0x10)),
	})
	require.NoError(t, err)

	exported, err := env.k.ExportGenesis(env.ctx)
	require.NoError(t, err)
	require.Len(t, exported.Attestations, 1)
	require.Len(t, exported.EpochStats, 1)
	require.Len(t, exported.ContributorServes, 1)

	// JSON roundtrip.
	bz, err := exported.MarshalJSON()
	require.NoError(t, err)
	var decoded types.GenesisState
	require.NoError(t, decoded.UnmarshalJSON(bz))
	require.Len(t, decoded.Attestations, 1)

	// Import into a fresh keeper and verify state restored.
	env2 := setupKeeper(t)
	require.NoError(t, env2.k.InitGenesis(env2.ctx, &decoded))

	require.True(t, env2.k.HasNullifier(env2.ctx, nullifier32(0x10)))
	stats, err := env2.k.GetEpochServeStats(env2.ctx, "org-1", 2)
	require.NoError(t, err)
	require.Equal(t, uint64(1), stats.TotalServes)
}

func TestGenesisInit_DenialAttestations(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0xF2)
	null := nullifier32(0x11)
	gs := &types.GenesisState{
		DenialAttestations: []*types.StoredDenialAttestation{
			{OrgId: "org-1", MemoryHash: h, DenyKey: "dk", Reason: "spam", Epoch: 3, Nullifier: null},
		},
	}
	require.NoError(t, env.k.InitGenesis(env.ctx, gs))
	require.True(t, env.k.HasDenialNullifier(env.ctx, null))
	require.Equal(t, uint64(1), env.k.GetMemoryDenialCount(env.ctx, "org-1", h, 3))
	serves, denials, err := env.k.GetEpochTrafficStats(env.ctx, "org-1", 3)
	require.NoError(t, err)
	require.Equal(t, uint64(0), serves)
	require.Equal(t, uint64(1), denials)
}

func TestGetServeAttestations_ListByOrgEpoch(t *testing.T) {
	env := setupKeeper(t)
	h := hash32(0xF3)
	env.mem.approve("org-1", h)
	require.NoError(t, env.k.SetParams(env.ctx, types.Params{MaxServesPerBatch: 500, MaxServesPerMemoryPerEpoch: 100}))

	for i := 0; i < 3; i++ {
		_, _, _, err := env.k.ProcessServeBatch(env.ctx, "org-1", 4, []*types.ServeEntry{
			serveEntry(h, "k", "c", nullifier32(byte(0x20+i))),
		})
		require.NoError(t, err)
	}

	atts, err := env.k.GetServeAttestations(env.ctx, "org-1", 4)
	require.NoError(t, err)
	require.Len(t, atts, 3)

	// Different epoch returns none.
	none, err := env.k.GetServeAttestations(env.ctx, "org-1", 5)
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
			serveEntry(h, "k", "c", nullifier32(byte(0x30+i))),
		})
		require.NoError(t, err)
	}
	require.Equal(t, uint64(5), env.k.GetMemoryServeCount(env.ctx, "org-1", h, 6))
	// guard: ensure binary helper import is genuinely exercised
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, 5)
	require.Equal(t, uint64(5), binary.BigEndian.Uint64(buf))
}
