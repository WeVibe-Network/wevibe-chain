package keeper_test

import (
	"testing"

	"github.com/wevibe-network/wevibe-chain/x/emissions/keeper"
)

// TestAfterEpochEnd_NoPool_ReturnsNil verifies R-EPOCH-HOOK-RESILIENCE: when no
// emission pool exists, AfterEpochEnd logs and returns nil rather than
// propagating an error. Returning an error here would cause the Cosmos SDK
// epoch dispatcher to discard ALL cached writes for the epoch-end hook batch
// (e.g. x/memory's epoch bookkeeping) — the CO-039 lost-writes bug class.
func TestAfterEpochEnd_NoPool_ReturnsNil(t *testing.T) {
	k, ctx := newTestKeeper(t)

	if err := k.AfterEpochEnd(ctx, keeper.WeVibeEpochIdentifier, 1); err != nil {
		t.Fatalf("AfterEpochEnd must return nil when no pool is seeded, got: %v", err)
	}
}

// TestAfterEpochEnd_ForeignIdentifier_ReturnsNil verifies the hook ignores
// other epoch identifiers without error.
func TestAfterEpochEnd_ForeignIdentifier_ReturnsNil(t *testing.T) {
	k, ctx := newTestKeeper(t)

	if err := k.AfterEpochEnd(ctx, "other_epoch", 7); err != nil {
		t.Fatalf("AfterEpochEnd must return nil for foreign identifier, got: %v", err)
	}
}
