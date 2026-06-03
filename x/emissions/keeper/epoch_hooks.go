package keeper

import (
	"context"
)

const WeVibeEpochIdentifier = "wevibe_epoch"

func (k *Keeper) AfterEpochEnd(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	if epochIdentifier != WeVibeEpochIdentifier {
		return nil
	}

	epoch := uint64(epochNumber)

	emission, err := k.MintDailyEmission(ctx, epoch)
	if err != nil {
		k.logger.Info("failed to mint daily emission", "epoch", epoch, "error", err)
		return nil
	}

	k.logger.Info("epoch emission processed",
		"epoch", epoch,
		"total_emitted", emission.TotalEmitted,
	)

	return nil
}

func (k *Keeper) BeforeEpochStart(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}
