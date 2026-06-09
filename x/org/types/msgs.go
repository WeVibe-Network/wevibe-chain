package types

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (m *MsgRegisterOrg) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.Leader == "" {
		return ErrInvalidLeader
	}
	return nil
}

func (m *MsgSetServingKey) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if m.NewServingKey == "" {
		return fmt.Errorf("new_serving_key cannot be empty")
	}
	return nil
}

func (m *MsgSetServingInfo) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if err := ValidateHubEndpoints(m.HubEndpoints); err != nil {
		return err
	}
	if err := ValidateHubResponsePubkey(m.HubResponsePubkey); err != nil {
		return err
	}
	return nil
}

func ValidateHubEndpoints(endpoints []string) error {
	if len(endpoints) < 1 || len(endpoints) > 3 {
		return fmt.Errorf("%w: hub_endpoints must contain between 1 and 3 endpoints", ErrInvalidHubEndpoints)
	}
	for i, endpoint := range endpoints {
		u, err := url.Parse(endpoint)
		if err != nil {
			return fmt.Errorf("%w: hub_endpoints[%d] parse error: %v", ErrInvalidHubEndpoints, i, err)
		}
		scheme := strings.ToLower(u.Scheme)
		if (scheme != "http" && scheme != "https") || u.Host == "" {
			return fmt.Errorf("%w: hub_endpoints[%d] must be an http(s) URL with host", ErrInvalidHubEndpoints, i)
		}
	}
	return nil
}

func ValidateHubResponsePubkey(pubkey string) error {
	if pubkey == "" {
		return nil
	}
	if pubkey != strings.ToLower(pubkey) {
		return fmt.Errorf("%w: hub_response_pubkey must be lowercase hex", ErrInvalidHubResponsePubkey)
	}
	bz, err := hex.DecodeString(pubkey)
	if err != nil {
		return fmt.Errorf("%w: hub_response_pubkey must be valid hex: %v", ErrInvalidHubResponsePubkey, err)
	}
	if len(bz) != 32 {
		return fmt.Errorf("%w: hub_response_pubkey must decode to exactly 32 bytes", ErrInvalidHubResponsePubkey)
	}
	return nil
}

func (m *MsgAddMember) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if m.Pubkey == "" {
		return fmt.Errorf("pubkey cannot be empty")
	}
	if m.Role != "member" && m.Role != "moderator" && m.Role != "contributor" {
		return ErrInvalidRole
	}
	return nil
}

func (m *MsgRemoveMember) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if m.Pubkey == "" {
		return fmt.Errorf("pubkey cannot be empty")
	}
	return nil
}

func (m *MsgUpdateParams) ValidateBasic() error {
	if m.Authority == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	return m.Params.Validate()
}

func (m *MsgSetOrgConfig) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if m.MinContributionsPerEpoch > 100 {
		return fmt.Errorf("min_contributions_per_epoch must be <= 100")
	}
	return nil
}

func (m *MsgGrantTrialAllowance) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if m.Grantee == "" {
		return fmt.Errorf("grantee cannot be empty")
	}
	if m.DailySubmissions == 0 {
		return fmt.Errorf("daily_submissions must be positive")
	}
	if m.TrialDays == 0 {
		return fmt.Errorf("trial_days must be positive")
	}
	return nil
}

func (m *MsgUpdateMemberRole) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if m.Pubkey == "" {
		return fmt.Errorf("pubkey cannot be empty")
	}
	if m.NewRole != "member" && m.NewRole != "moderator" && m.NewRole != "contributor" {
		return fmt.Errorf("new_role must be 'member', 'moderator', or 'contributor'")
	}
	return nil
}

func (m *MsgRotateEpoch) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	return nil
}

func (m *MsgTransferLeadership) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if m.NewLeader == "" {
		return fmt.Errorf("new_leader cannot be empty")
	}
	if m.NewLeaderWallet == "" {
		return fmt.Errorf("new_leader_wallet cannot be empty")
	}
	if _, err := sdk.AccAddressFromBech32(m.NewLeaderWallet); err != nil {
		return fmt.Errorf("invalid new_leader_wallet: %w", err)
	}
	if m.Signer == m.NewLeader {
		return fmt.Errorf("signer cannot transfer leadership to self")
	}
	return nil
}

func (m *MsgCloseOrg) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	return nil
}
