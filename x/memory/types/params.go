package types

import "fmt"

const (
	DefaultIdleDecayRateBps      = uint64(50)
	DefaultDenialDecayBps        = uint64(500)
	DefaultServeBoostBps         = uint64(100)
	DefaultMaxServeBoostPerEpoch = uint64(5)
	DefaultBootstrapGraceEpochs  = uint64(14)
)

func DefaultParams() Params {
	return Params{
		MaxPendingPerOrg:       1000,
		PendingRetentionEpochs: 7,
		MaxBlobSizeBytes:       1048576,
		MaxKeywordsPerMemory:   20,
		MinRetrievalDecayBps:   DefaultIdleDecayRateBps,
		ContestWindowEpochs:    10,
		IdleDecayRateBps:      DefaultIdleDecayRateBps,
		DenialDecayBps:        DefaultDenialDecayBps,
		ServeBoostBps:         DefaultServeBoostBps,
		MaxServeBoostPerEpoch: DefaultMaxServeBoostPerEpoch,
		BootstrapGraceEpochs:  DefaultBootstrapGraceEpochs,
	}
}

func (p Params) Validate() error {
	if p.MinRetrievalDecayBps == 0 {
		return fmt.Errorf("min retrieval decay must be positive")
	}
	if p.IdleDecayRateBps == 0 {
		return fmt.Errorf("idle decay rate must be positive")
	}
	if p.DenialDecayBps == 0 {
		return fmt.Errorf("denial decay must be positive")
	}
	if p.ServeBoostBps == 0 {
		return fmt.Errorf("serve boost must be positive")
	}
	if p.MaxServeBoostPerEpoch == 0 {
		return fmt.Errorf("max serve boost per epoch must be positive")
	}
	return nil
}
