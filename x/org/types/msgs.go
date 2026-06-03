package types

import (
	"fmt"
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
	if m.Role == "" {
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
