package types

import (
	"crypto/sha256"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func DeriveOrgID(addr sdk.AccAddress) string {
	h := sha256.Sum256(addr.Bytes())
	orgID, err := bech32.ConvertAndEncode("weorg", h[:20])
	if err != nil {
		panic(err)
	}
	return orgID
}
