package types

import (
	"fmt"
	"math/big"
)

func (m *MsgRegisterOrg) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if m.Leader == "" {
		return ErrInvalidLeader
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

func (m *MsgFundTreasury) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if m.Amount == "" {
		return ErrInvalidAmount
	}
	amount := new(big.Int)
	_, ok := amount.SetString(m.Amount, 10)
	if !ok {
		return ErrInvalidAmount
	}
	if amount.Sign() <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

func (m *MsgWithdrawTreasury) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if m.Amount == "" {
		return ErrInvalidAmount
	}
	amount := new(big.Int)
	_, ok := amount.SetString(m.Amount, 10)
	if !ok {
		return ErrInvalidAmount
	}
	if amount.Sign() <= 0 {
		return ErrInvalidAmount
	}
	if m.Recipient == "" {
		return ErrInvalidRecipient
	}
	return nil
}

func (m *MsgSetRepTiers) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.OrgId == "" {
		return ErrInvalidOrgID
	}
	if len(m.Tiers) == 0 {
		return ErrInvalidRepTier
	}
	for i, tier := range m.Tiers {
		if tier.MinReputation > tier.MaxReputation {
			return fmt.Errorf("tier %d: min_reputation > max_reputation", i)
		}
		if tier.PayoutPerMemory == "" {
			return fmt.Errorf("tier %d: payout_per_memory cannot be empty", i)
		}
		payout := new(big.Int)
		_, ok := payout.SetString(tier.PayoutPerMemory, 10)
		if !ok {
			return fmt.Errorf("tier %d: invalid payout_per_memory", i)
		}
		if payout.Sign() < 0 {
			return fmt.Errorf("tier %d: payout_per_memory cannot be negative", i)
		}
	}
	for i := 0; i < len(m.Tiers); i++ {
		for j := i + 1; j < len(m.Tiers); j++ {
			if m.Tiers[i].MinReputation <= m.Tiers[j].MaxReputation && m.Tiers[j].MinReputation <= m.Tiers[i].MaxReputation {
				return ErrRepTierOverlap
			}
		}
	}
	return nil
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
