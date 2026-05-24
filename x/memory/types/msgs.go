package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (m *MsgSubmitCommitment) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if len(m.ContentHash) != ContentHashLen {
		return ErrInvalidContentHash
	}
	if m.ContributorId == "" {
		return ErrInvalidContributor
	}
	return nil
}

func (m *MsgApproveMemory) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if len(m.ContentHash) != ContentHashLen {
		return ErrInvalidContentHash
	}
	if len(m.EncryptedBlob) == 0 {
		return ErrInvalidBlob
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

var _ sdk.Msg = &MsgSubmitCommitment{}
var _ sdk.Msg = &MsgApproveMemory{}
var _ sdk.Msg = &MsgUpdateParams{}
var _ sdk.Msg = &MsgReportMemory{}

func (msg *MsgSubmitCommitment) Route() string { return RouterKey }
func (msg *MsgSubmitCommitment) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Signer)
	return []sdk.AccAddress{addr}
}

func (msg *MsgApproveMemory) Route() string { return RouterKey }
func (msg *MsgApproveMemory) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Signer)
	return []sdk.AccAddress{addr}
}

func (msg *MsgUpdateParams) Route() string { return RouterKey }
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

const RouterKey = ModuleName

func (m *MsgReportMemory) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if len(m.ContentHash) != ContentHashLen {
		return ErrInvalidContentHash
	}
	if m.ReporterPubkey == "" {
		return ErrInvalidReporter
	}
	if m.Reason == "" {
		return ErrInvalidReportReason
	}
	return nil
}

func (m *MsgReportMemory) Route() string { return RouterKey }
func (m *MsgReportMemory) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Signer)
	return []sdk.AccAddress{addr}
}
