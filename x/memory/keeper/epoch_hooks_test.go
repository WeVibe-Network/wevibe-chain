package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wevibe-network/wevibe-chain/x/memory/types"
)

func TestEpochDurationHookAdvancesCurrentEpoch(t *testing.T) {
	t.Setenv("WEVIBE_EPOCH_DURATION_SECONDS", "2")

	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	require.Equal(t, uint64(0), k.getCurrentEpoch(ctx))
	require.NoError(t, k.AfterEpochEnd(ctx, WeVibeEpochIdentifier, 1))
	require.NoError(t, k.AfterEpochEnd(ctx, WeVibeEpochIdentifier, 2))
	require.Equal(t, uint64(2), k.getCurrentEpoch(ctx))
}

func TestEpochDurationHookIgnoresForeignIdentifier(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	require.NoError(t, k.AfterEpochEnd(ctx, "other_epoch", 99))
	require.Equal(t, uint64(0), k.getCurrentEpoch(ctx))
}

// TestAfterEpochEnd_AppliesAndPersistsDecay is the CO-040 regression test for
// the CO-039 zero-decay bug. ApplyEpochDecay's writes are cached by the epoch
// dispatcher and only committed if the hook returns nil. This test proves that
// running AfterEpochEnd (a) returns nil and (b) actually mutates and persists
// the stored weight — i.e. the decay is not silently rolled back.
func TestAfterEpochEnd_AppliesAndPersistsDecay(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	hash := []byte("eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee")
	storeMemoryWithKeywords(
		t, k, ctx, defaultOrgID, hash,
		types.MemoryState_MEMORY_STATE_COMMITTED, 0,
		withKeywords(&types.KeywordWeight{Keyword: "kw1", Weight: "0.9000"}),
	)

	require.NoError(t, k.AfterEpochEnd(ctx, WeVibeEpochIdentifier, 30))

	memory, err := k.GetApprovedMemory(ctx, defaultOrgID, hash)
	require.NoError(t, err)
	got := weightForKeyword(t, memory, "kw1")
	require.Less(t, got, 0.9000, "idle/untrusted decay must have lowered and persisted the weight through the hook")
}

// TestAfterEpochEnd_ResilientOnEmptyState verifies the hook returns nil (never
// an error) on a fresh keeper with no memories, protecting the dispatcher's
// cached-write batch from rollback (R-EPOCH-HOOK-RESILIENCE).
func TestAfterEpochEnd_ResilientOnEmptyState(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	require.NoError(t, k.AfterEpochEnd(ctx, WeVibeEpochIdentifier, 1))
	require.Equal(t, uint64(1), k.getCurrentEpoch(ctx))
}

// TestAfterEpochEnd_DecaysSettledEpochNotHead is the CO-042 regression test for
// the zero-decay root cause: serve/denial traffic is relayed asynchronously and
// settles 1-3 epochs after the activity occurred. AfterEpochEnd(N) must assess
// epoch (N - IdleDecaySettleEpochs), whose traffic has settled, NOT epoch N. A
// memory created in epoch 0 must remain UNTOUCHED while N is still inside the
// settle window relative to the epoch that exits grace, and must decay once the
// assessed (lagged) epoch clears grace.
func TestAfterEpochEnd_DecaysSettledEpochNotHead(t *testing.T) {
	k, _, _ := makeTestKeeper(t)
	ctx := context.Background()

	hash := []byte("ssssssssssssssssssssssssssssssss")
	storeMemoryWithKeywords(
		t, k, ctx, defaultOrgID, hash,
		types.MemoryState_MEMORY_STATE_COMMITTED, 0,
		withKeywords(&types.KeywordWeight{Keyword: "kw1", Weight: "0.9000"}),
	)

	// Head epoch 24 assesses epoch 24-5=19, still inside the 20-epoch grace
	// window for a memory created at epoch 0 -> no decay.
	require.NoError(t, k.AfterEpochEnd(ctx, WeVibeEpochIdentifier, 24))
	memory, err := k.GetApprovedMemory(ctx, defaultOrgID, hash)
	require.NoError(t, err)
	require.Equal(t, 0.9000, weightForKeyword(t, memory, "kw1"),
		"assessed epoch (19) is still in grace; weight must be untouched")

	// Head epoch 26 assesses epoch 26-5=21, past grace -> decay applies.
	require.NoError(t, k.AfterEpochEnd(ctx, WeVibeEpochIdentifier, 26))
	memory, err = k.GetApprovedMemory(ctx, defaultOrgID, hash)
	require.NoError(t, err)
	require.Less(t, weightForKeyword(t, memory, "kw1"), 0.9000,
		"assessed epoch (21) is past grace; idle decay must lower and persist the weight")
}
