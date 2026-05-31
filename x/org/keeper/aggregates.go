package keeper

import (
	"context"
	"encoding/binary"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/wevibe-network/wevibe-chain/x/org/types"
)

func (k *Keeper) IncrementOrgCommittedMemories(ctx context.Context, orgID string) error {
	return k.modifyOrgAggregate(ctx, orgID, func(stored *types.StoredOrg, currentEpoch uint64) {
		stored.TotalCommittedMemories++
		stored.LastActivityEpoch = currentEpoch
	})
}

func (k *Keeper) IncrementOrgUpheldReports(ctx context.Context, orgID string) error {
	return k.modifyOrgAggregate(ctx, orgID, func(stored *types.StoredOrg, currentEpoch uint64) {
		stored.TotalUpheldReports++
		stored.LastActivityEpoch = currentEpoch
	})
}

func (k *Keeper) IncrementOrgEpochRotations(ctx context.Context, orgID string) error {
	return k.modifyOrgAggregate(ctx, orgID, func(stored *types.StoredOrg, currentEpoch uint64) {
		stored.TotalEpochRotations++
		stored.LastActivityEpoch = currentEpoch
	})
}

func (k *Keeper) SetOrgLastActivityEpoch(ctx context.Context, orgID string, epoch uint64) error {
	return k.modifyOrgAggregate(ctx, orgID, func(stored *types.StoredOrg, _ uint64) {
		stored.LastActivityEpoch = epoch
	})
}

func (k *Keeper) RecomputeOrgActiveMembers(ctx context.Context, orgID string) error {
	store := k.getStore(ctx)
	prefix := []byte(fmt.Sprintf("member/%s/", orgID))
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return err
	}
	defer iter.Close()

	var count uint64
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredMemberRecord
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		if stored.Role != "inactive" {
			count++
		}
	}

	return k.modifyOrgAggregate(ctx, orgID, func(stored *types.StoredOrg, currentEpoch uint64) {
		stored.TotalActiveMembers = count
		stored.LastActivityEpoch = currentEpoch
	})
}

func (k *Keeper) modifyOrgAggregate(ctx context.Context, orgID string, fn func(*types.StoredOrg, uint64)) error {
	store := k.getStore(ctx)
	key := orgKey(orgID)
	bz, err := store.Get(key)
	if err != nil {
		return err
	}
	if bz == nil {
		return types.ErrOrgNotFound
	}

	var stored types.StoredOrg
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return fmt.Errorf("unmarshal org: %w", err)
	}

	currentEpoch := k.getCurrentEpoch(ctx)
	fn(&stored, currentEpoch)

	bz, err = proto.Marshal(&stored)
	if err != nil {
		return fmt.Errorf("marshal org: %w", err)
	}
	return store.Set(key, bz)
}

func (k *Keeper) getCurrentEpoch(ctx context.Context) uint64 {
	store := k.getStore(ctx)
	bz, err := store.Get([]byte("epoch/current"))
	if err != nil || bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}
