package types

type GenesisState struct {
	EmissionPool     *EmissionPool        `json:"emission_pool"`
	DailyEmissions   []*DailyEmission    `json:"daily_emissions"`
	OperatorRewards  []*OperatorReward   `json:"operator_rewards"`
	ValidatorRewards []*ValidatorReward  `json:"validator_rewards"`
	BootstrapCredits []*BootstrapCredit  `json:"bootstrap_credits"`
	WorkScores       []*WorkScore        `json:"work_scores"`
	AsymmetricGates  []*AsymmetricGate   `json:"asymmetric_gates"`
	BootstrapExpiry  uint64              `json:"bootstrap_expiry"`
}

func NewGenesisState(
	emissionPool *EmissionPool,
	dailyEmissions []*DailyEmission,
	operatorRewards []*OperatorReward,
	validatorRewards []*ValidatorReward,
	bootstrapCredits []*BootstrapCredit,
	workScores []*WorkScore,
	asymmetricGates []*AsymmetricGate,
	bootstrapExpiry uint64,
) *GenesisState {
	return &GenesisState{
		EmissionPool:     emissionPool,
		DailyEmissions:   dailyEmissions,
		OperatorRewards:  operatorRewards,
		ValidatorRewards: validatorRewards,
		BootstrapCredits: bootstrapCredits,
		WorkScores:       workScores,
		AsymmetricGates:  asymmetricGates,
		BootstrapExpiry:  bootstrapExpiry,
	}
}

// DefaultEmissionPool returns the emission pool seeded at genesis. It derives
// its values from DefaultParams() so that DefaultParams remains the single
// source of truth for the emission schedule.
//
// The pool seeds the full locked 32-year emission schedule:
//   - ValidatorPoolRemainingUvibe is the entire validator emission pool
//     (ValidatorEmissionPoolUvibe = 570,000,000,000,000 uvibe).
//   - ContributorPoolRemainingUvibe is the contributor annual cap multiplied
//     by the 32 years of the schedule (ContributorAnnualCapUvibe * years,
//     where years = ScheduleDurationDays / EpochsPerYear = 11680/365 = 32,
//     yielding 320,000,000,000,000 uvibe).
//
// Counters (rollover, start epoch, total epochs elapsed) begin at zero. Once
// seeded, the existing MintDailyEmission logic operates against this pool.
func DefaultEmissionPool() *EmissionPool {
	p := DefaultParams()
	years := p.ScheduleDurationDays / EpochsPerYear // 11680/365 = 32
	return &EmissionPool{
		TotalSupply:                   0,
		DailyMint:                     p.DailyMintAmount,
		OperatorShare:                 p.OperatorSharePercent,
		ValidatorShare:                p.ValidatorSharePercent,
		Epoch:                         0,
		ValidatorPoolRemainingUvibe:   p.ValidatorEmissionPoolUvibe,        // 570_000_000_000_000
		ContributorPoolRemainingUvibe: p.ContributorAnnualCapUvibe * years, // 10_000_000_000_000 * 32 = 320_000_000_000_000
		ContributorRolloverUvibe:      0,
		StartEpoch:                    0,
		TotalEpochsElapsed:            0,
	}
}

// DefaultGenesis returns the default genesis state for the emissions module:
// an initialized emission pool and otherwise empty collections.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		EmissionPool: DefaultEmissionPool(),
	}
}

// Validate performs stateless validation of the emissions genesis state.
func (g *GenesisState) Validate() error {
	if g.EmissionPool != nil {
		if err := g.EmissionPool.Validate(); err != nil {
			return err
		}
	}
	return nil
}
