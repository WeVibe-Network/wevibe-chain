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
