package keeper

import (
	"context"
	"sort"
	"strings"

	storetypes "cosmossdk.io/store/types"
)

const WeVibeEpochIdentifier = "wevibe_epoch"

// IdleDecaySettleEpochs is how many epochs the traffic-adaptive decay assessment
// lags behind the chain head. Serve/denial traffic is relayed to the chain
// asynchronously by the hub (decoupled from the request path) and is included
// on-chain a variable number of blocks after the activity occurred. Under fast
// epochs this settlement delay routinely spans 1-3 epochs. ApplyEpochDecay's
// Goldilocks metric and zero-signal guard aggregate the per-epoch serve/denial
// stats for a SPECIFIC epoch; if it assessed epoch N at the instant epoch N ends
// (AfterEpochEnd(N)), epoch N's traffic would not have landed yet, the guard
// would fire on every epoch, and no memory would ever idle-decay (the CO-042
// zero-decay root cause). By assessing epoch (N - IdleDecaySettleEpochs) we
// guarantee that epoch's relayed traffic has fully settled before it is scored,
// while preserving the exact per-epoch metric the decay model is calibrated to.
const IdleDecaySettleEpochs = uint64(5)

func (k *Keeper) AfterEpochEnd(ctx context.Context, epochIdentifier string, epochNumber int64) error {
	if epochIdentifier != WeVibeEpochIdentifier {
		return nil
	}

	epoch := uint64(epochNumber)

	// R-EPOCH-HOOK-RESILIENCE: this hook runs inside the Cosmos SDK epoch
	// dispatcher's cached-write batch. If it returns a non-nil error, the
	// dispatcher discards ALL cached writes for the batch — silently rolling
	// back ApplyEpochDecay's weight changes (the CO-039 zero-decay bug). Every
	// recoverable failure below therefore logs a warning and continues; the
	// function returns nil unconditionally. Each step is independent, so a
	// failure in one must not skip the others — most importantly, a failure in
	// setCurrentEpoch or CheckEpochExpiry must not prevent ApplyEpochDecay from
	// running and persisting.
	if err := k.setCurrentEpoch(ctx, epoch); err != nil {
		k.logger.Warn("epoch hook: set current epoch failed (continuing)", "epoch", epoch, "error", err)
	}

	if err := k.CheckEpochExpiry(ctx, epoch); err != nil {
		k.logger.Warn("epoch hook: check epoch expiry failed (continuing)", "epoch", epoch, "error", err)
	}

	// Traffic-adaptive decay assesses a SETTLED epoch, not the epoch that just
	// ended. Serve/denial batches for epoch (epoch-IdleDecaySettleEpochs) have all
	// been relayed and included on-chain by now, so their per-epoch stats are
	// complete. Assessing the just-ended epoch instead would read zero traffic
	// (async relay has not caught up), firing the zero-signal guard every epoch
	// and suppressing all idle decay (CO-042 zero-decay root cause).
	if epoch >= IdleDecaySettleEpochs {
		decayEpoch := epoch - IdleDecaySettleEpochs
		if err := k.ApplyEpochDecay(ctx, decayEpoch); err != nil {
			k.logger.Warn("epoch hook: apply epoch decay failed (continuing)", "epoch", decayEpoch, "head_epoch", epoch, "error", err)
		}
	} else {
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
