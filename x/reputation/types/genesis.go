package types

import (
	"encoding/json"
)

// DefaultGenesis returns the default reputation genesis state. The module is
// active by default (D-13.5, matching DefaultParams().Active == true). The
// historical DefaultGenesis returned Active: false (GAP-REP-1), which left the
// module inert after genesis even though DefaultParams declared it active.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Active: true,
	}
}

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
