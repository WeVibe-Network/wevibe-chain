package types

import (
	"crypto/sha256"
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
		if len(serve.ServeKeyPubkey) != ServePubKeyLen {
			return ErrInvalidServeKeyPubkey
		}
		if len(serve.ServeSig) != ServeSigLen {
			return ErrInvalidServeSignature
		}
		if serve.ContributorId == "" {
			return ErrInvalidContributor
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
		if len(entry.ServeKeyPubkey) != ServePubKeyLen {
			return ErrInvalidServeKeyPubkey
		}
		if len(entry.ServeSig) != ServeSigLen {
			return ErrInvalidServeSignature
		}
		if len(entry.ServeFingerprint) != FingerprintLen {
			return ErrInvalidServeFingerprint
		}
		if len(entry.NegAnchor) != 0 {
			return ErrNegAnchorInert
		}
	}
	return nil
}

func (m *MsgSubmitEventBatch) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if len(m.Events) == 0 {
		return ErrBatchEmpty
	}
	for _, entry := range m.Events {
		if err := validateEventEntry(entry); err != nil {
			return err
		}
	}
	return nil
}

func validateEventEntry(entry *EventEntry) error {
	if entry == nil {
		return fmt.Errorf("event entry cannot be nil")
	}
	if err := validateEventType(entry.EventType); err != nil {
		return err
	}
	if len(entry.MemoryContentHash) != ContentHashLen {
		return ErrInvalidContentHash
	}
	if len(entry.SignerPubkey) != ServePubKeyLen {
		return ErrInvalidServeKeyPubkey
	}
	if len(entry.Signature) != ServeSigLen {
		return ErrInvalidServeSignature
	}
	if len(entry.Nonce) == 0 {
		return fmt.Errorf("event nonce cannot be empty")
	}
	if len(entry.Nonce) > 64 {
		return fmt.Errorf("event nonce must be at most 64 bytes")
	}
	return validateEventBody(entry)
}

func validateEventType(eventType EventType) error {
	switch eventType {
	case EventType_EVENT_TYPE_OUTCOME, EventType_EVENT_TYPE_VALIDITY_PREDICATE, EventType_EVENT_TYPE_COST_TO_DISCOVER, EventType_EVENT_TYPE_CONVERGENCE:
		return nil
	case EventType_EVENT_TYPE_CONTEST, EventType_EVENT_TYPE_SPONSORSHIP:
		return ErrEventParked
	case EventType_EVENT_TYPE_SERVE, EventType_EVENT_TYPE_BLOCK:
		return fmt.Errorf("event type recorded via serve/denial receipts")
	default:
		return ErrInvalidEventType
	}
}

func validateEventBody(entry *EventEntry) error {
	switch entry.EventType {
	case EventType_EVENT_TYPE_OUTCOME:
		body := entry.GetOutcome()
		if body == nil {
			return fmt.Errorf("outcome event body is required")
		}
		if len(body.ServeRef) != sha256.Size {
			return fmt.Errorf("outcome serve_ref must be exactly 32 bytes")
		}
		return validateRefCaps(body.EpisodeRef, body.EvidenceRef)
	case EventType_EVENT_TYPE_VALIDITY_PREDICATE:
		body := entry.GetValidityPredicate()
		if body == nil {
			return fmt.Errorf("validity_predicate event body is required")
		}
		if _, err := predicateResultToken(body.Result); err != nil {
			return err
		}
		return validateRefCaps(body.PredicateId, body.EvidenceRef)
	case EventType_EVENT_TYPE_COST_TO_DISCOVER:
		body := entry.GetCostToDiscover()
		if body == nil {
			return fmt.Errorf("cost_to_discover event body is required")
		}
		return validateRefCaps(body.EvidenceRef)
	case EventType_EVENT_TYPE_CONVERGENCE:
		body := entry.GetConvergence()
		if body == nil {
			return fmt.Errorf("convergence event body is required")
		}
		return validateRefCaps(body.ConvergenceRef)
	default:
		return ErrInvalidEventType
	}
}

func validateRefCaps(refs ...[]byte) error {
	for _, ref := range refs {
		if len(ref) > 64 {
			return fmt.Errorf("event refs must be at most 64 bytes — fingerprints/refs only (RECALL-PIVOT-SPEC §3.2 boundary)")
		}
	}
	return nil
}

func (m *MsgAnchorPolicyVersion) ValidateBasic() error {
	if m.Authority == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	if m.PolicyVersion == "" {
		return fmt.Errorf("policy_version cannot be empty")
	}
	if len(m.PolicyVersion) > 128 {
		return fmt.Errorf("policy_version must be at most 128 chars")
	}
	// PolicyHash is the sha256 of the published edge-policy artifact — standing is NEVER written on-chain; it is recomputed at the edge as f(events, policy_version).
	if len(m.PolicyHash) != 32 {
		return fmt.Errorf("policy_hash must be 32 bytes")
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
var _ sdk.Msg = &MsgSubmitEventBatch{}
var _ sdk.Msg = &MsgAnchorPolicyVersion{}
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

func (msg *MsgSubmitEventBatch) Route() string { return RouterKey }
func (msg *MsgSubmitEventBatch) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Signer)
	return []sdk.AccAddress{addr}
}

func (msg *MsgAnchorPolicyVersion) Route() string { return RouterKey }
func (msg *MsgAnchorPolicyVersion) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

func (msg *MsgUpdateParams) Route() string { return RouterKey }
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

const RouterKey = ModuleName
