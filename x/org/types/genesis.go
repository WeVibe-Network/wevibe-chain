package types

type GenesisState struct {
	Orgs         []*Org          `json:"orgs"`
	Members      []*MemberRecord `json:"members"`
	DynamicPrice *DynamicPrice   `json:"dynamic_price"`
	OrgConfigs   []*OrgConfig    `json:"org_configs"`
	NextSlot     uint64          `json:"next_slot"`
	Params       Params          `json:"params"`
}

func NewGenesisState(orgs []*Org, members []*MemberRecord, dynamicPrice *DynamicPrice) *GenesisState {
	return &GenesisState{
		Orgs:         orgs,
		Members:      members,
		DynamicPrice: dynamicPrice,
		NextSlot:     0,
		Params:       DefaultParams(),
	}
}
