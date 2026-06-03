package types

type GenesisState struct {
	Orgs       []*Org          `json:"orgs"`
	Members    []*MemberRecord `json:"members"`
	OrgConfigs []*OrgConfig    `json:"org_configs"`
	NextSlot   uint64          `json:"next_slot"`
	Params     Params          `json:"params"`
}

func NewGenesisState(orgs []*Org, members []*MemberRecord) *GenesisState {
	return &GenesisState{
		Orgs:     orgs,
		Members:  members,
		NextSlot: 0,
		Params:   DefaultParams(),
	}
}
