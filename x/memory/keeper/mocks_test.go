package keeper

import (
	"context"
)

type mockReputationKeeper struct {
	incrementContributionCalls       []struct{ contributor, orgID, memoryCID string }
	incrementModeratorApprovalCalls   []struct{ modPubkey, orgID string }
	incrementModeratorUpheldCalls     []struct{ modPubkey, orgID string }
	incrementLeaderChainCommitCalls   []struct{ leader, orgID string }
	incrementLeaderUpheldReportCalls []struct{ leader, orgID string }
}

func (m *mockReputationKeeper) IncrementContribution(ctx context.Context, contributorID, orgID, memoryCID string) error {
	m.incrementContributionCalls = append(m.incrementContributionCalls, struct{ contributor, orgID, memoryCID string }{contributorID, orgID, memoryCID})
	return nil
}

func (m *mockReputationKeeper) IncrementModeratorApproval(ctx context.Context, modPubkey, orgID string, memoryHash []byte, epoch uint64) error {
	m.incrementModeratorApprovalCalls = append(m.incrementModeratorApprovalCalls, struct{ modPubkey, orgID string }{modPubkey, orgID})
	return nil
}

func (m *mockReputationKeeper) IncrementModeratorUpheld(ctx context.Context, modPubkey, orgID string) error {
	m.incrementModeratorUpheldCalls = append(m.incrementModeratorUpheldCalls, struct{ modPubkey, orgID string }{modPubkey, orgID})
	return nil
}

func (m *mockReputationKeeper) IncrementLeaderChainCommit(ctx context.Context, leaderPubkey, orgID string) error {
	m.incrementLeaderChainCommitCalls = append(m.incrementLeaderChainCommitCalls, struct{ leader, orgID string }{leaderPubkey, orgID})
	return nil
}

func (m *mockReputationKeeper) IncrementLeaderUpheldReport(ctx context.Context, leaderPubkey, orgID string) error {
	m.incrementLeaderUpheldReportCalls = append(m.incrementLeaderUpheldReportCalls, struct{ leader, orgID string }{leaderPubkey, orgID})
	return nil
}

func (m *mockOrgKeeper) IncrementOrgCommittedMemories(ctx context.Context, orgID string) error {
	return nil
}

func (m *mockOrgKeeper) IncrementOrgUpheldReports(ctx context.Context, orgID string) error {
	return nil
}

func (m *mockOrgKeeper) SetOrgLastActivityEpoch(ctx context.Context, orgID string, epoch uint64) error {
	return nil
}

func (m *mockOrgKeeperServer) IncrementOrgCommittedMemories(ctx context.Context, orgID string) error {
	return nil
}

func (m *mockOrgKeeperServer) IncrementOrgUpheldReports(ctx context.Context, orgID string) error {
	return nil
}

func (m *mockOrgKeeperServer) SetOrgLastActivityEpoch(ctx context.Context, orgID string, epoch uint64) error {
	return nil
}