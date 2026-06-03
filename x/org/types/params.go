package types

import "fmt"

func DefaultParams() Params {
	return Params{
		MinRegistrationFee:       10000000,
		AnnualRenewalFee:         5000000,
		DefaultStorageQuota:      1073741824,
		DefaultRetrievalBudget:  10000,
		GracePeriodEpochs:       30,
		BurnPriceDecayEpochs:    10,
		BaseBurnPrice:           10000000,
		BurnPriceIncreasePercent: 20,
		SlotCap:                 32,
	}
}

func (p Params) Validate() error {
	if p.SlotCap == 0 {
		return fmt.Errorf("slot_cap must be greater than 0")
	}
	return nil
}
