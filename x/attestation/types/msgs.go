package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgSubmitSessionAttestation{}
var _ sdk.Msg = &MsgUpdateParams{}

func (m *MsgSubmitSessionAttestation) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if len(m.SessionHash) != SessionHashLen {
		return ErrInvalidSessionHash
	}
	if m.ModelId == "" {
		return ErrInvalidModelID
	}
	if m.ContributorId == "" {
		return ErrInvalidContributor
	}
	if m.ProviderType == ProviderType_PROVIDER_TYPE_UNSPECIFIED {
		return ErrInvalidProviderType
	}
	return nil
}

func (msg *MsgSubmitSessionAttestation) Route() string { return RouterKey }

func (msg *MsgSubmitSessionAttestation) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Signer)
	return []sdk.AccAddress{addr}
}

func (m *MsgUpdateParams) ValidateBasic() error {
	if m.Authority == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	return nil
}

func (msg *MsgUpdateParams) Route() string { return RouterKey }

func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}