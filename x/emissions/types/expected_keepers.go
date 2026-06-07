package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	GetContributorsWithApprovalsInEpoch(ctx context.Context, epoch uint64) (map[string]uint64, error)
}

type ReputationKeeper interface {
	GetContributorProfile(ctx context.Context, contributorID, orgID string) (*reputationtypes.StoredContributorProfile, error)
}

type OrgKeeper interface {
	GetAllOrgs(ctx context.Context) ([]*orgTypes.Org, error)
	GetOrgConfig(ctx context.Context, orgID string) (*orgTypes.OrgConfig, error)
}

type BankKeeper interface {
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

type IdentityKeeper interface {
	ResolveIdentity(ctx context.Context, passkeyPubkey string) (walletAddress string, isMigrated bool, found bool, err error)
}
