package keeper_test

import (
	"context"

	reputationtypes "github.com/wevibe-network/wevibe-chain/x/reputation/types"
)

type mockReputationKeeper struct {
	profiles map[string]*reputationtypes.StoredContributorProfile
}

func newMockReputationKeeper() *mockReputationKeeper {
	return &mockReputationKeeper{
		profiles: make(map[string]*reputationtypes.StoredContributorProfile),
	}
}

func (m *mockReputationKeeper) GetContributorProfile(ctx context.Context, contributorID, orgID string) (*reputationtypes.StoredContributorProfile, error) {
	key := contributorID + "_" + orgID
	if profile, ok := m.profiles[key]; ok {
		return profile, nil
	}
	return &reputationtypes.StoredContributorProfile{
		ContributorId:             contributorID,
		TotalApprovedMemories:     0,
		TotalServesReceived:       0,
		TotalDenialsReceived:      0,
		TotalReportsFiledAgainst:  0,
		TotalReportsUpheldAgainst:  0,
	}, nil
}