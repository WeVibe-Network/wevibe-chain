package types

import (
	"fmt"
)

func DefaultParams() Params {
	return Params{
		Active:              true,
		MaxDifficulty:        10,
		MaxQuality:          10,
		ServeXpPerServe:     5,
		SelfServeXpPerServe: 2,
	}
}

func (p Params) Validate() error {
	if p.MaxDifficulty == 0 {
		return fmt.Errorf("max_difficulty must be non-zero")
	}
	if p.MaxQuality == 0 {
		return fmt.Errorf("max_quality must be non-zero")
	}
	return nil
}