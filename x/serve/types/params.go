package types

func DefaultParams() Params {
	return Params{
		MaxServesPerBatch:             500,
		SelfServeDiscountPercent:      50,
		MaxServesPerMemoryPerEpoch:    100,
		MinOrgAgeEpochs:              1,
		DiminishingReturnsThreshold:   10,
	}
}

func (p *Params) Validate() error {
	return nil
}