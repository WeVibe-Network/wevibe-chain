package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (m *MsgSetBandwidthOverride) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if m.MemoryCap == 0 && m.ServeCap == 0 {
		return ErrInvalidCap
	}
	return nil
}

func (m *MsgUpdateParams) ValidateBasic() error {
	if m.Authority == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	if err := m.Params.Validate(); err != nil {
		return err
	}
	return nil
}

func (p *Params) Validate() error {
	return nil
}

var _ sdk.Msg = &MsgSetBandwidthOverride{}
var _ sdk.Msg = &MsgUpdateParams{}

func (msg *MsgSetBandwidthOverride) Route() string { return RouterKey }
func (msg *MsgSetBandwidthOverride) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Signer)
	return []sdk.AccAddress{addr}
}

func (msg *MsgUpdateParams) Route() string { return RouterKey }
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

const RouterKey = ModuleName