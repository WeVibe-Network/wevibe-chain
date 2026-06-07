package keeper_test

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
		TotalReportsUpheldAgainst: 0,
	}, nil
}

type mintCoinsCall struct {
	moduleName string
	amt        sdk.Coins
}

type sendCoinsFromModuleCall struct {
	senderModule  string
	recipientAddr sdk.AccAddress
	amt           sdk.Coins
}

type mockBankKeeper struct {
	mintCalls []mintCoinsCall
	sendCalls []sendCoinsFromModuleCall
	mintErr   error
	sendErr   error
}

func newMockBankKeeper() *mockBankKeeper {
	return &mockBankKeeper{}
}

func (m *mockBankKeeper) MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	m.mintCalls = append(m.mintCalls, mintCoinsCall{moduleName: moduleName, amt: amt})
	return m.mintErr
}

func (m *mockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	m.sendCalls = append(m.sendCalls, sendCoinsFromModuleCall{senderModule: senderModule, recipientAddr: recipientAddr, amt: amt})
	return m.sendErr
}

type mockIdentityKeeper struct {
	walletAddress string
	isMigrated    bool
	found         bool
	err           error
	calls         []string
}

func newMockIdentityKeeper() *mockIdentityKeeper {
	return &mockIdentityKeeper{}
}

func (m *mockIdentityKeeper) ResolveIdentity(ctx context.Context, passkeyPubkey string) (walletAddress string, isMigrated bool, found bool, err error) {
	m.calls = append(m.calls, passkeyPubkey)
	return m.walletAddress, m.isMigrated, m.found, m.err
}
