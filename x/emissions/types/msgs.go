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

func (m *MsgUpdateParams) ValidateBasic() error {
	if m.Authority == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	return m.Params.Validate()
}

func (m *MsgClaimContributorReward) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if m.PasskeyPubkey == "" {
		return fmt.Errorf("passkey_pubkey cannot be empty")
	}
	return nil
}
