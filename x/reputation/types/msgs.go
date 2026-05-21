package types

import "fmt"

func (m *MsgUpdateReputation) ValidateBasic() error {
	if m.Signer == "" {
		return fmt.Errorf("signer cannot be empty")
	}
	if len(m.Developer) == 0 {
		return ErrInvalidDeveloper
	}
	if m.MemoryCid == "" {
		return ErrInvalidMemory
	}
	if m.Difficulty > 10 {
		return ErrInvalidDifficulty
	}
	if m.Quality > 10 {
		return ErrInvalidQuality
	}
	return nil
}

func (m *MsgUpdateParams) ValidateBasic() error {
	if m.Authority == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	return m.Params.Validate()
}

func (m *MsgIncrementContribution) ValidateBasic() error {
	if m.Authority == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	if m.ContributorId == "" {
		return fmt.Errorf("contributor_id cannot be empty")
	}
	return nil
}

func (m *MsgIncrementServe) ValidateBasic() error {
	if m.Authority == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	if m.ContributorId == "" {
		return fmt.Errorf("contributor_id cannot be empty")
	}
	return nil
}

func (m *MsgRecordBan) ValidateBasic() error {
	if m.Authority == "" {
		return fmt.Errorf("authority cannot be empty")
	}
	if m.ContributorId == "" {
		return fmt.Errorf("contributor_id cannot be empty")
	}
	return nil
}