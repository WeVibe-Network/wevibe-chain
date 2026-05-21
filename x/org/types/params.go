package types

func DefaultParams() Params {
	return Params{
		MinRegistrationFee:       10000000,
		AnnualRenewalFee:       5000000,
		DefaultStorageQuota:    1073741824,
		DefaultRetrievalBudget:  10000,
		GracePeriodEpochs:      30,
		BurnPriceDecayEpochs:   10,
		BaseBurnPrice:          10000000,
		BurnPriceIncreasePercent: 20,
	}
}

func (p Params) Validate() error {
	return nil
}