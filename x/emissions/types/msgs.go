package types

import "fmt"

func (m *MsgMintDailyEmission) ValidateBasic() error {
	if m.Authority == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	if m.Epoch == 0 {
		return fmt.Errorf("epoch must be greater than 0")
	}
	return nil
}

func (m *MsgDistributeOperatorRewards) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if len(m.Rewards) == 0 {
		return fmt.Errorf("rewards cannot be empty")
	}
	return nil
}

func (m *MsgUpdateParams) ValidateBasic() error {
	if m.Authority == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	return m.Params.Validate()
}