package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/wevibe-network/wevibe-chain/x/memory/types"
)

func validityKey(orgID, cid string) []byte {
	suffix := fmt.Sprintf("%s:%s", orgID, cid)
	key := make([]byte, len(types.ValidityKeyPrefix)+len(suffix))
	copy(key, types.ValidityKeyPrefix)
	copy(key[len(types.ValidityKeyPrefix):], suffix)
	return key
}

func (k *Keeper) GetValidityMetadata(ctx context.Context, orgID, cid string) (types.ValidityMetadata, bool, error) {
	bz, err := k.getStore(ctx).Get(validityKey(orgID, cid))
	if err != nil {
		return types.ValidityMetadata{}, false, fmt.Errorf("get validity metadata: %w", err)
	}
	if bz == nil {
		return types.ValidityMetadata{}, false, nil
	}

	var stored types.StoredValidityMetadata
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return types.ValidityMetadata{}, false, fmt.Errorf("unmarshal validity metadata: %w", err)
	}

	metadata, err := decodeValidityMetadata(stored)
	if err != nil {
		return types.ValidityMetadata{}, false, err
	}

	return metadata, true, nil
}

func (k *Keeper) CheckEpochExpiry(ctx context.Context, epoch uint64) error {
	store := k.getStore(ctx)
	iter, err := store.Iterator(types.ValidityKeyPrefix, storetypes.PrefixEndBytes(types.ValidityKeyPrefix))
	if err != nil {
		return fmt.Errorf("iterate validity metadata: %w", err)
	}

	// R-CACHEKV-ITER: never delete/mutate the store while iterating it under a
	// cache-wrapped context. Collect the expiring tuples first, close the
	// iterator, then archive the memory and delete the validity key.
	type expiringEntry struct {
		key       []byte
		orgID     string
		memoryCid string
	}
	var expired []expiringEntry
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredValidityMetadata
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			iter.Close()
			return fmt.Errorf("unmarshal validity metadata: %w", err)
		}

		if stored.ValidUntilEpoch == 0 || epoch <= stored.ValidUntilEpoch {
			continue
		}

		key := append([]byte(nil), iter.Key()...)
		expired = append(expired, expiringEntry{key: key, orgID: stored.OrgId, memoryCid: stored.MemoryCid})
	}
	iter.Close()

	for _, e := range expired {
		memory, err := k.loadMemoryByCID(ctx, e.orgID, e.memoryCid)
		if err != nil {
			return err
		}
		if memory.State != types.MemoryState_MEMORY_STATE_ARCHIVED && memory.State != types.MemoryState_MEMORY_STATE_DENIED {
			memory.State = types.MemoryState_MEMORY_STATE_ARCHIVED
			if err := k.saveMemoryCommitment(ctx, memory); err != nil {
				return err
			}
		}

		if err := store.Delete(e.key); err != nil {
			return fmt.Errorf("delete expired validity metadata: %w", err)
		}
	}

	return nil
}

func (k *Keeper) IsValidInEpoch(ctx context.Context, orgID, cid string, epoch uint64) (bool, error) {
	metadata, found, err := k.GetValidityMetadata(ctx, orgID, cid)
	if err != nil {
		return false, err
	}
	if !found {
		return true, nil
	}
	if metadata.ValidAfterEpoch != 0 && epoch < metadata.ValidAfterEpoch {
		return false, nil
	}
	if metadata.ValidUntilEpoch != 0 && epoch > metadata.ValidUntilEpoch {
		return false, nil
	}
	return true, nil
}

func decodeValidityMetadata(stored types.StoredValidityMetadata) (types.ValidityMetadata, error) {
	scope := map[string]string{}
	if len(stored.ScopeTagsBz) > 0 {
		if err := json.Unmarshal(stored.ScopeTagsBz, &scope); err != nil {
			return types.ValidityMetadata{}, fmt.Errorf("unmarshal scope tags: %w", err)
		}
	}
	return types.ValidityMetadata{
		ValidAfterEpoch: stored.ValidAfterEpoch,
		ValidUntilEpoch: stored.ValidUntilEpoch,
		ScopeTags:       scope,
	}, nil
}

func isValidityEligibleState(state types.MemoryState) bool {
	return state == types.MemoryState_MEMORY_STATE_COMMITTED
}
