package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
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
