package types

type GenesisState struct {
	Orgs       []*Org          `json:"orgs"`
	Members    []*MemberRecord `json:"members"`
	OrgConfigs []*OrgConfig    `json:"org_configs"`
	NextSlot   uint64          `json:"next_slot"`
	Params     Params          `json:"params"`
}

func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Orgs:       []*Org{},
		Members:    []*MemberRecord{},
		OrgConfigs: []*OrgConfig{},
		NextSlot:   0,
		Params:     DefaultParams(),
	}
}

func (g *GenesisState) Validate() error {
	if err := g.Params.Validate(); err != nil {
		return err
	}

	return nil
}

func NewGenesisState(orgs []*Org, members []*MemberRecord) *GenesisState {
	return &GenesisState{
		Orgs:     orgs,
		Members:  members,
		NextSlot: 0,
		Params:   DefaultParams(),
	}
}
