package keeper

import (
	"context"
	"sort"
	"strings"

	storetypes "cosmossdk.io/store/types"
)

const WeVibeEpochIdentifier = "wevibe_epoch"

func (k *Keeper) AfterEpochEnd(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	if epochIdentifier != WeVibeEpochIdentifier {
		return nil
	}

	epoch := uint64(epochNumber)

	// R-EPOCH-HOOK-RESILIENCE: this hook runs inside the Cosmos SDK epoch
	// dispatcher's cached-write batch. If it returns a non-nil error, the
	// dispatcher discards ALL cached writes for the batch. Every
	// recoverable failure below therefore logs a warning and continues; the
	// function returns nil unconditionally. Each step is independent, so a
	// failure in one must not skip the others — most importantly, a failure in
	// setCurrentEpoch or CheckEpochExpiry must not prevent merkle roots from
	// being computed and persisted.
	if err := k.setCurrentEpoch(ctx, epoch); err != nil {
		k.logger.Warn("epoch hook: set current epoch failed (continuing)", "epoch", epoch, "error", err)
	}

	if err := k.CheckEpochExpiry(ctx, epoch); err != nil {
		k.logger.Warn("epoch hook: check epoch expiry failed (continuing)", "epoch", epoch, "error", err)
	}

	orgs, err := k.getAllOrgsWithMemories(ctx)
	if err != nil {
		k.logger.Warn("epoch hook: list orgs with memories failed (skipping merkle roots)", "epoch", epoch, "error", err)
		return nil
	}

	for _, orgID := range orgs {
		if err := k.ComputeAndStoreEpochMerkleRoot(ctx, orgID, epoch); err != nil {
			k.logger.Warn("epoch hook: compute merkle root failed (continuing)", "epoch", epoch, "org", orgID, "error", err)
			continue
		}
	}

	k.logger.Info("epoch merkle roots computed", "epoch", epoch, "orgs", len(orgs))
	return nil
}

func (k *Keeper) BeforeEpochStart(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	return nil
}

func (k *Keeper) getAllOrgsWithMemories(ctx context.Context) ([]string, error) {
	store := k.getStore(ctx)
	prefix := []byte("approved/")
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	orgSet := make(map[string]bool)
	for ; iter.Valid(); iter.Next() {
		key := string(iter.Key())
		remainder := key[len("approved/"):]
		parts := strings.SplitN(remainder, "/", 2)
		if len(parts) > 0 && parts[0] != "" {
			orgSet[parts[0]] = true
		}
	}

	result := make([]string, 0, len(orgSet))
	for orgID := range orgSet {
		result = append(result, orgID)
	}
	sort.Strings(result)
	return result, nil
}
