package types

import (
	"context"

	memorytypes "github.com/wevibe-network/wevibe-chain/x/memory/types"
)

type ServeKeeper interface {
	GetContributorEpochServesRaw(ctx context.Context, contributorID string, epoch uint64) (serveCount, selfServeCount, totalTurns uint64, orgIDs []string, err error)
}

type MemoryKeeper interface {
	IterateUpheldReports(ctx context.Context, cb func(*memorytypes.StoredMemoryReport) bool) error
	GetUpheldReport(ctx context.Context, orgID string, memoryHash []byte) (*memorytypes.StoredMemoryReport, error)
	GetApprovedMemory(ctx context.Context, orgID string, contentHash []byte) (*memorytypes.MemoryCommitment, error)
}
