package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func FormatOrgID(slot uint64) string {
	return fmt.Sprintf("weorg-%d", slot)
}

func OrgAccountAddress(orgID string) sdk.AccAddress {
	return authtypes.NewModuleAddress("orgacct/" + orgID)
}
