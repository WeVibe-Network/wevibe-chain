package types

import "fmt"

const (
	DefaultServeDBps             = uint64(220)
	DefaultDenialDBps            = uint64(900)
	DefaultIdleDBps              = uint64(600)
	DefaultServeFloorBps         = uint64(4000)
	DefaultDenialFloorBps        = uint64(3000)
	DefaultIdleProtectBps        = uint64(500)
	DefaultIdleUntrustedBps      = uint64(10000)
	DefaultIdleTrafficRefBps     = uint64(2200)
	DefaultIdleTrafficFloorBps   = uint64(10000)
	DefaultTrustMinServes        = uint64(1)
	DefaultTrustMaxRateBps       = uint64(3000)
	DefaultGraceEpochs           = uint64(20)
	DefaultRetrievalThresholdBps = uint64(1500)
)

func DefaultParams() Params {
	return Params{
		MaxPendingPerOrg:        1000,
		PendingRetentionEpochs:  7,
		MaxBlobSizeBytes:        1048576,
		MaxKeywordsPerMemory:    20,
		RetrievalThresholdBps:   DefaultRetrievalThresholdBps,
		InitialConfidenceBps:    0,
		ContestWindowEpochs:     10,
		GraceEpochs:             DefaultGraceEpochs,
		ServeDBps:               DefaultServeDBps,
		DenialDBps:              DefaultDenialDBps,
		IdleDBps:                DefaultIdleDBps,
		ServeFloorBps:           DefaultServeFloorBps,
		DenialFloorBps:          DefaultDenialFloorBps,
		IdleProtectBps:          DefaultIdleProtectBps,
		IdleUntrustedBps:        DefaultIdleUntrustedBps,
		IdleTrafficRefBpsPerMem: DefaultIdleTrafficRefBps,
		IdleTrafficFloorBps:     DefaultIdleTrafficFloorBps,
		TrustMinServes:          DefaultTrustMinServes,
		TrustMaxRateBps:         DefaultTrustMaxRateBps,
	}
}

func (p Params) Validate() error {
	if p.GraceEpochs < 1 {
		return fmt.Errorf("grace epochs must be at least 1")
	}
	if p.TrustMinServes < 1 {
		return fmt.Errorf("trust min serves must be at least 1")
	}
	if p.IdleTrafficRefBpsPerMem < 1 {
		return fmt.Errorf("idle_traffic_ref_bps_per_mem must be at least 1")
	}

	bpsFields := map[string]uint64{
		"retrieval_threshold_bps":      p.RetrievalThresholdBps,
		"initial_confidence_bps":       p.InitialConfidenceBps,
		"serve_d_bps":                  p.ServeDBps,
		"denial_d_bps":                 p.DenialDBps,
		"idle_d_bps":                   p.IdleDBps,
		"serve_floor_bps":              p.ServeFloorBps,
		"denial_floor_bps":             p.DenialFloorBps,
		"idle_protect_bps":             p.IdleProtectBps,
		"idle_untrusted_bps":           p.IdleUntrustedBps,
		"idle_traffic_ref_bps_per_mem": p.IdleTrafficRefBpsPerMem,
		"idle_traffic_floor_bps":       p.IdleTrafficFloorBps,
		"trust_max_rate_bps":           p.TrustMaxRateBps,
	}
	for name, value := range bpsFields {
		if value > 10000 {
			return fmt.Errorf("%s must be <= 10000", name)
		}
	}

	return nil
}
