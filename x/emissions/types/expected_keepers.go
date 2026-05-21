package types

import (
	"context"

	"cosmossdk.io/math"

	orgTypes "github.com/wevibe-network/wevibe-chain/x/org/types"
	reputationtypes "github.com/wevibe-network/wevibe-chain/x/reputation/types"
	serveTypes "github.com/wevibe-network/wevibe-chain/x/serve/types"
)

type ServeKeeper interface {
	GetEpochServeStats(ctx context.Context, orgID string, epoch uint64) (*serveTypes.EpochServeStats, error)
	GetServeAttestations(ctx context.Context, orgID string, epoch uint64) ([]*serveTypes.ServeAttestation, error)
}

type MemoryKeeper interface {
	GetApprovedCountByContributor(ctx context.Context, orgID string, epoch uint64) (map[string]uint64, error)
}

type ReputationKeeper interface {
	GetContributorProfile(ctx context.Context, contributorID, orgID string) (*reputationtypes.StoredContributorProfile, error)
}

type OrgKeeper interface {
	GetAllOrgs(ctx context.Context) ([]*orgTypes.Org, error)
	GetTreasuryBalanceInt(ctx context.Context, orgID string) (math.Int, error)
	DebitTreasury(ctx context.Context, orgID string, amount math.Int) error
	GetRepTiers(ctx context.Context, orgID string) (*orgTypes.RepTierConfig, error)
	GetOrgConfig(ctx context.Context, orgID string) (*orgTypes.OrgConfig, error)
}
