package types

import (
	"fmt"
)

func DefaultParams() Params {
	return Params{
		DailyMintAmount:        1000000000,
		OperatorSharePercent:   80,
		ValidatorSharePercent:  20,
		StorageWeightPercent:   30,
		RetrievalWeightPercent: 70,
		RarityMultiplierCap:    "3.0",
		BootstrapDurationEpochs: 365,
		TotalSupplyUvibe:            1_000_000_000_000_000,
		ValidatorEmissionPoolUvibe:  570_000_000_000_000,
		ContributorAnnualCapUvibe:   10_000_000_000_000,
		ScheduleDurationDays:        11_680,
		ContributorQualifyThreshold: 1,
	}
}

func (p Params) Validate() error {
	if p.OperatorSharePercent+p.ValidatorSharePercent != 100 {
		return fmt.Errorf("operator_share_percent (%d) + validator_share_percent (%d) must equal 100", p.OperatorSharePercent, p.ValidatorSharePercent)
	}
	if p.StorageWeightPercent+p.RetrievalWeightPercent != 100 {
		return fmt.Errorf("storage_weight_percent (%d) + retrieval_weight_percent (%d) must equal 100", p.StorageWeightPercent, p.RetrievalWeightPercent)
	}
	return nil
}