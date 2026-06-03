package types

import (
	"context"

	"cosmossdk.io/x/feegrant"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

type FeegrantKeeper interface {
	GrantAllowance(ctx context.Context, granter, grantee sdk.AccAddress, feeAllowance feegrant.FeeAllowanceI) error
}

type MemoryKeeper interface {
	GetApprovedCount(ctx context.Context, orgID string) (uint64, error)
}

type ServeKeeper interface {
	GetEpochServeStatsRaw(ctx context.Context, orgID string, epoch uint64) (totalServes, uniqueMemories, selfServes uint64, modelBreakdown map[string]uint64, err error)
}

type BandwidthKeeper interface {
	GetOrInitBandwidthStateRaw(ctx context.Context, orgID string, epoch uint64) (memoryUsed, memoryCap, serveUsed, serveCap uint64, err error)
}
