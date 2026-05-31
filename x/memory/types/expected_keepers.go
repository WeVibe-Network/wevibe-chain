package types

import (
	"context"

	orgtypes "github.com/wevibe-network/wevibe-chain/x/org/types"
)

type OrgKeeper interface {
	HasOrg(ctx context.Context, orgID string) (bool, error)
	IsLeader(ctx context.Context, orgID string, memberPubkey string) (bool, error)
	IsModerator(ctx context.Context, orgID string, memberPubkey string) (bool, error)
	GetOrgConfig(ctx context.Context, orgID string) (*orgtypes.OrgConfig, error)
	IncrementOrgCommittedMemories(ctx context.Context, orgID string) error
	IncrementOrgUpheldReports(ctx context.Context, orgID string) error
	SetOrgLastActivityEpoch(ctx context.Context, orgID string, epoch uint64) error
	// GetLeaderWallet returns the org's registered leader chain wallet address —
	// the authenticated tx signer authorized to commit org decisions
	// (approvals/commits/reports). D-S32-CO044-KEY-SEPARATION.
	GetLeaderWallet(ctx context.Context, orgID string) (string, error)
}

type ServeKeeper interface {
	GetMemoryServeCountForEpoch(ctx context.Context, orgID string, memoryCID string, epoch uint64) (uint64, error)
	GetMemoryDenialCountForEpoch(ctx context.Context, orgID string, memoryCID string, epoch uint64) (uint64, error)
	GetEpochTrafficStats(ctx context.Context, orgID string, epoch uint64) (serves uint64, denials uint64, err error)
	// Returns the union of keywords matched
	// by any serve to (orgID, memoryCID) during the given epoch. Empty
	// non-nil map indicates no serves this epoch.
	// Per D-4.2 Implementation Clarifications (DMO-007).
	GetMatchedKeywordsForEpoch(ctx context.Context, orgID string, memoryCID string, epoch uint64) (map[string]bool, error)
}

type ReputationKeeper interface {
	IncrementContribution(ctx context.Context, contributorID, orgID, memoryCID string) error
	IncrementModeratorApproval(ctx context.Context, modPubkey, orgID string, memoryHash []byte, epoch uint64) error
	IncrementModeratorUpheld(ctx context.Context, modPubkey, orgID string) error
	IncrementLeaderChainCommit(ctx context.Context, leaderPubkey, orgID string) error
	IncrementLeaderUpheldReport(ctx context.Context, leaderPubkey, orgID string) error
}
