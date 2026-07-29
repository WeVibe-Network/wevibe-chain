package types

func DefaultParams() Params {
	return Params{
		MaxPendingPerOrg:       1000,
		PendingRetentionEpochs: 7,
		MaxBlobSizeBytes:       1048576,
		MaxKeywordsPerMemory:   20,
		ContestWindowEpochs:    10,
	}
}

func (p Params) Validate() error {
	return nil
}
