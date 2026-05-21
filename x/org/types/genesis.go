package types

type GenesisState struct {
	Orgs         []*Org           `json:"orgs"`
	Members      []*MemberRecord  `json:"members"`
	DynamicPrice *DynamicPrice    `json:"dynamic_price"`
	Treasuries   []*Treasury      `json:"treasuries"`
	RepTiers     []*RepTierConfig `json:"rep_tiers"`
	OrgConfigs   []*OrgConfig     `json:"org_configs"`
}

func NewGenesisState(orgs []*Org, members []*MemberRecord, dynamicPrice *DynamicPrice) *GenesisState {
	return &GenesisState{
		Orgs:         orgs,
		Members:      members,
		DynamicPrice: dynamicPrice,
	}
}
