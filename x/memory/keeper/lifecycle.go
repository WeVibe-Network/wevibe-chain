package keeper

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/wevibe-network/wevibe-chain/x/memory/types"
)

const currentEpochKey = "current_epoch"

func (k *Keeper) saveMemoryCommitment(ctx context.Context, memory *types.MemoryCommitment) error {
	store := k.getStore(ctx)
	bz, err := proto.Marshal(memoryToStored(memory))
	if err != nil {
		return fmt.Errorf("marshal memory: %w", err)
	}
	if err := store.Set(approvedKey(memory.OrgID, memory.ContentHash), bz); err != nil {
		return fmt.Errorf("persist memory: %w", err)
	}
	return nil
}

func (k *Keeper) loadMemoryByCID(ctx context.Context, orgID, cid string) (*types.MemoryCommitment, error) {
	contentHash, err := decodeCID(cid)
	if err != nil {
		return nil, err
	}

	store := k.getStore(ctx)
	bz, err := store.Get(approvedKey(orgID, contentHash))
	if err != nil {
		return nil, fmt.Errorf("get memory: %w", err)
	}
	if bz == nil {
		return nil, types.ErrMemoryNotFound
	}

	var stored types.StoredMemoryCommitment
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal memory: %w", err)
	}

	memory := storedToMemory(stored)
	return &memory, nil
}

// GetContributorsWithApprovalsInEpoch returns, network-wide, the number of
// committed (non-archived, non-denied) approved memories per contributor
// pubkey that were approved in the given epoch. It is consumed by the
// emissions keeper to determine the qualifying contributor set for per-epoch
// contributor emissions.
//
// R-CACHEKV-ITER: this is a NEW network-wide iteration. It is read-only and
// MUST NOT use a post-loop iter.Error() block (which returns a false failure
// at normal end-of-iteration under cache-wrapped stores). The Valid() loop
// terminates correctly on its own.
func (k *Keeper) GetContributorsWithApprovalsInEpoch(ctx context.Context, epoch uint64) (map[string]uint64, error) {
	store := k.getStore(ctx)
	prefix := []byte("approved/")
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, fmt.Errorf("iterate approved memories: %w", err)
	}
	defer iter.Close()

	counts := make(map[string]uint64)
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredMemoryCommitment
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		if stored.ApprovedAtEpoch != epoch {
			continue
		}
		if stored.State != types.MemoryState_MEMORY_STATE_COMMITTED {
			continue
		}
		if stored.ContributorPubkey == "" {
			continue
		}
		counts[stored.ContributorPubkey]++
	}

	return counts, nil
}

func (k *Keeper) getCurrentEpoch(ctx context.Context) uint64 {
	store := k.getStore(ctx)
	bz, err := store.Get([]byte(currentEpochKey))
	if err != nil || bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (k *Keeper) setCurrentEpoch(ctx context.Context, epoch uint64) error {
	store := k.getStore(ctx)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, epoch)
	if err := store.Set([]byte(currentEpochKey), buf); err != nil {
		return fmt.Errorf("set current epoch: %w", err)
	}
	return nil
}

func decodeCID(cid string) ([]byte, error) {
	contentHash, err := hex.DecodeString(cid)
	if err != nil {
		return nil, fmt.Errorf("decode cid: %w", err)
	}
	if len(contentHash) != types.ContentHashLen {
		return nil, types.ErrInvalidContentHash
	}
	return contentHash, nil
}
