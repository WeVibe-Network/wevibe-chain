package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (m *MsgSubmitServeBatch) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if len(m.Serves) == 0 {
		return ErrBatchEmpty
	}
	for _, serve := range m.Serves {
		if len(serve.MemoryContentHash) != ContentHashLen {
			return ErrInvalidContentHash
		}
		if serve.ServeKey == "" {
			return ErrInvalidServeKey
		}
		if serve.ContributorId == "" {
			return ErrInvalidContributor
		}
		if len(serve.Nullifier) != NullifierLen {
			return ErrInvalidNullifier
		}
	}
	return nil
}

func (m *MsgSubmitDenialBatch) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if len(m.Entries) == 0 {
		return ErrBatchEmpty
	}
	for _, entry := range m.Entries {
		if len(entry.MemoryHash) != ContentHashLen {
			return ErrInvalidContentHash
		}
		if len(entry.Nullifier) != NullifierLen {
			return ErrInvalidNullifier
		}
		if entry.DenyKey == "" {
			return ErrInvalidServeKey
		}
	}
	return nil
}

func (m *MsgUpdateParams) ValidateBasic() error {
	if m.Authority == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	return nil
}

var _ sdk.Msg = &MsgSubmitServeBatch{}
var _ sdk.Msg = &MsgSubmitDenialBatch{}
var _ sdk.Msg = &MsgUpdateParams{}

func (msg *MsgSubmitServeBatch) Route() string { return RouterKey }
func (msg *MsgSubmitServeBatch) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Signer)
	return []sdk.AccAddress{addr}
}

func (msg *MsgSubmitDenialBatch) Route() string { return RouterKey }
func (msg *MsgSubmitDenialBatch) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Signer)
	return []sdk.AccAddress{addr}
}

func (msg *MsgUpdateParams) Route() string { return RouterKey }
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

const RouterKey = ModuleName
