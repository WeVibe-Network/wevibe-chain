package types

import (
	"fmt"
	"strings"

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
	if m.ProducerModelId != "" {
		if len(m.ProducerModelId) > MaxProducerModelIdLen {
			return fmt.Errorf("producer_model_id must be <= %d characters", MaxProducerModelIdLen)
		}
		if strings.TrimSpace(m.ProducerModelId) == "" {
			return fmt.Errorf("producer_model_id cannot be whitespace")
		}
	}
	if len(m.AttestationSessionHash) > 0 {
		if len(m.AttestationSessionHash) != 32 {
			return fmt.Errorf("attestation_session_hash must be exactly 32 bytes when set")
		}
		if m.ProducerModelId == "" {
			return fmt.Errorf("producer_model_id is required when attestation_session_hash is set")
		}
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
