package types

import "encoding/json"

type GenesisState struct {
	Attestations []*StoredSessionAttestation `json:"attestations"`
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