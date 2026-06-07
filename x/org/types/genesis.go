package types

import "fmt"

type GenesisState struct {
	Orgs               []*Org                     `json:"orgs"`
	Members            []*MemberRecord            `json:"members"`
	OrgConfigs         []*OrgConfig               `json:"org_configs"`
	ExtractionProfiles []*StoredExtractionProfile `json:"extraction_profiles"`
	NextSlot           uint64                     `json:"next_slot"`
	Params             Params                     `json:"params"`
}

func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Orgs:               []*Org{},
		Members:            []*MemberRecord{},
		OrgConfigs:         []*OrgConfig{},
		ExtractionProfiles: []*StoredExtractionProfile{},
		NextSlot:           0,
		Params:             DefaultParams(),
	}
}

func (g *GenesisState) Validate() error {
	if err := g.Params.Validate(); err != nil {
		return err
	}

	seenProfileOrgIDs := make(map[string]struct{}, len(g.ExtractionProfiles))
	for i, profile := range g.ExtractionProfiles {
		if profile == nil {
			return fmt.Errorf("extraction_profiles[%d] cannot be nil", i)
		}
		if profile.OrgId == "" {
			return ErrInvalidOrgID
		}
		if _, exists := seenProfileOrgIDs[profile.OrgId]; exists {
			return fmt.Errorf("duplicate extraction profile org_id: %s", profile.OrgId)
		}
		seenProfileOrgIDs[profile.OrgId] = struct{}{}
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
