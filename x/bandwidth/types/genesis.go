package types

import "encoding/json"

type GenesisState struct {
	BandwidthStates    []*BandwidthState    `json:"bandwidth_states"`
	BandwidthOverrides []*BandwidthOverride `json:"bandwidth_overrides"`
}

func NewGenesisState(states []*BandwidthState, overrides []*BandwidthOverride) *GenesisState {
	return &GenesisState{
		BandwidthStates:    states,
		BandwidthOverrides: overrides,
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