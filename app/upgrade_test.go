package app_test

import (
	"testing"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/stretchr/testify/require"

	"github.com/wevibe-network/wevibe-chain/app"
	v2 "github.com/wevibe-network/wevibe-chain/app/upgrades/v2"
)

// TestV2UpgradeHandlerRegistered verifies the v2 governance upgrade handler is
// wired into the UpgradeKeeper so the halt -> swap -> restart flow can fire.
func TestV2UpgradeHandlerRegistered(t *testing.T) {
	wevibeApp := app.NewWeVibeApp(log.NewNopLogger(), dbm.NewMemDB(), true, simtestutil.EmptyAppOptions{})
	require.NotNil(t, wevibeApp)
	require.NotNil(t, wevibeApp.UpgradeKeeper)

	require.True(t, wevibeApp.UpgradeKeeper.HasHandler(v2.UpgradeName),
		"expected upgrade handler %q to be registered", v2.UpgradeName)
	require.False(t, wevibeApp.UpgradeKeeper.HasHandler("nonexistent-upgrade"))
}
