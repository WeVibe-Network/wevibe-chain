package types

import (
	"encoding/json"
)

func (g *GenesisState) MarshalJSON() ([]byte, error) {
	type GenesisStateAlias GenesisState
	alias := struct {
		GenesisStateAlias
	}{
		GenesisStateAlias: GenesisStateAlias{
			Active:              g.Active,
			Stats:               g.Stats,
			ContributorOrgSets: g.ContributorOrgSets,
		},
	}
	return json.Marshal(alias)
}

func (g *GenesisState) UnmarshalJSON(bz []byte) error {
	type GenesisStateAlias GenesisState
	var alias struct {
		*GenesisStateAlias
	}
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}
	g.Active = alias.Active
	g.Stats = alias.Stats
	g.ContributorOrgSets = alias.ContributorOrgSets
	return nil
}
