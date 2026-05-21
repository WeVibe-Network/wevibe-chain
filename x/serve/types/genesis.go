package types

import "encoding/json"

type GenesisState struct {
	Attestations       []*ServeAttestation        `json:"attestations"`
	DenialAttestations []*StoredDenialAttestation `json:"denial_attestations"`
	EpochStats         []*EpochServeStats         `json:"epoch_stats"`
	ContributorServes  []*ContributorEpochServes  `json:"contributor_serves"`
}

func NewGenesisState(attestations []*ServeAttestation, denialAttestations []*StoredDenialAttestation, epochStats []*EpochServeStats, contributorServes []*ContributorEpochServes) *GenesisState {
	return &GenesisState{
		Attestations:       attestations,
		DenialAttestations: denialAttestations,
		EpochStats:         epochStats,
		ContributorServes:  contributorServes,
	}
}

func (gs *GenesisState) MarshalJSON() ([]byte, error) {
	type Alias GenesisState
	return json.Marshal(&struct{ *Alias }{Alias: (*Alias)(gs)})
}

func (gs *GenesisState) UnmarshalJSON(b []byte) error {
	type Alias GenesisState
	aux := &struct{ *Alias }{Alias: (*Alias)(gs)}
	return json.Unmarshal(b, &aux)
}
