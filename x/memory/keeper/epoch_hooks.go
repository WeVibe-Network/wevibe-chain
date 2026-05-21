package keeper

import (
	"context"
	"fmt"
	"sort"
	"strings"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
)

const WeVibeEpochIdentifier = "wevibe_epoch"

func (k *Keeper) AfterEpochEnd(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	if epochIdentifier != WeVibeEpochIdentifier {
		return nil
	}

	epoch := uint64(epochNumber)

	if err := k.setCurrentEpoch(ctx, epoch); err != nil {
		return fmt.Errorf("set current epoch: %w", err)
	}

	if err := k.CheckEpochExpiry(ctx, epoch); err != nil {
		return fmt.Errorf("check epoch expiry: %w", err)
	}

	if err := k.ApplyIdleDecay(ctx, "", epoch); err != nil {
		return fmt.Errorf("apply idle decay: %w", err)
	}

	orgs, err := k.getAllOrgsWithMemories(ctx)
	if err != nil {
		return fmt.Errorf("list orgs with memories: %w", err)
	}

	for _, orgID := range orgs {
		if err := k.ComputeAndStoreEpochMerkleRoot(ctx, orgID, epoch); err != nil {
			return fmt.Errorf("compute merkle root for org %s: %w", orgID, err)
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
	if err := iter.Error(); err != nil {
		return nil, err
	}

	result := make([]string, 0, len(orgSet))
	for orgID := range orgSet {
		result = append(result, orgID)
	}
	sort.Strings(result)
	return result, nil
}
