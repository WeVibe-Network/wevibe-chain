package v2_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/stretchr/testify/require"

	v2 "github.com/wevibe-network/wevibe-chain/app/upgrades/v2"
)

func TestUpgradeName(t *testing.T) {
	require.Equal(t, "v2", v2.UpgradeName)
}

func TestCreateUpgradeHandler_ReturnsNonNilHandler(t *testing.T) {
	mm := module.NewManager()
	handler := v2.CreateUpgradeHandler(mm, nil)
	require.NotNil(t, handler)
}
