package types

import (
	"context"

	"github.com/wevibe-network/wevibe-chain/x/memory/types"
)

type OrgKeeper interface {
	HasOrg(ctx context.Context, orgID string) (bool, error)
	// GetServingAddress returns the org's registered hub serving key chain
	// address — the only signer permitted to submit serve/denial batches
	// (D-S32-CO044-KEY-SEPARATION).
	GetServingAddress(ctx context.Context, orgID string) (string, error)
}

type MemoryKeeper interface {
	GetApprovedMemory(ctx context.Context, orgID string, contentHash []byte) (*types.MemoryCommitment, error)
	IsValidInEpoch(ctx context.Context, orgID string, cid string, epoch uint64) (bool, error)
}

type BandwidthKeeper interface {
	ConsumeServeBandwidth(ctx context.Context, orgID string, epoch uint64, count uint64) error
}

type ReputationKeeper interface {
	RecordServe(ctx context.Context, contributorWallet []byte, orgID string, epoch uint64, isSelfServe bool) error
	IncrementContribution(ctx context.Context, contributorWallet, orgID, memoryCID string) error
	IncrementServe(ctx context.Context, contributorWallet, orgID, memoryCID string, count uint64) error
	RecordBan(ctx context.Context, contributorWallet, orgID, memoryCID string) error
}
