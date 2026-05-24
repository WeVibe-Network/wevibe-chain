package keeper

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/wevibe-network/wevibe-chain/x/memory/types"
)

func relationshipKey(orgID, sourceCID, targetCID string) []byte {
	suffix := fmt.Sprintf("%s:%s:%s", orgID, sourceCID, targetCID)
	key := make([]byte, len(types.RelationshipKeyPrefix)+len(suffix))
	copy(key, types.RelationshipKeyPrefix)
	copy(key[len(types.RelationshipKeyPrefix):], suffix)
	return key
}

func relationshipPrefix(orgID string) []byte {
	suffix := fmt.Sprintf("%s:", orgID)
	prefix := make([]byte, len(types.RelationshipKeyPrefix)+len(suffix))
	copy(prefix, types.RelationshipKeyPrefix)
	copy(prefix[len(types.RelationshipKeyPrefix):], suffix)
	return prefix
}

func (k *Keeper) GetRelationship(ctx context.Context, orgID, sourceCID, targetCID string) (*types.MemoryRelationship, error) {
	relationship, err := k.loadRelationship(ctx, orgID, sourceCID, targetCID)
	if err != nil {
		return nil, err
	}
	return relationship, nil
}

func (k *Keeper) ListRelationshipsForMemory(ctx context.Context, orgID, cid string) ([]types.MemoryRelationship, error) {
	store := k.getStore(ctx)
	prefix := relationshipPrefix(orgID)
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, fmt.Errorf("iterate relationships: %w", err)
	}
	defer iter.Close()

	var relationships []types.MemoryRelationship
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredMemoryRelationship
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			return nil, fmt.Errorf("unmarshal relationship: %w", err)
		}
		rel := storedToRelationship(stored)
		if rel.SourceCID == cid || rel.TargetCID == cid {
			relationships = append(relationships, rel)
		}
	}

	return relationships, nil
}

func (k *Keeper) DeleteRelationship(ctx context.Context, orgID, sourceCID, targetCID string) error {
	store := k.getStore(ctx)
	if err := store.Delete(relationshipKey(orgID, sourceCID, targetCID)); err != nil {
		return fmt.Errorf("delete relationship: %w", err)
	}
	return nil
}

func (k *Keeper) loadRelationship(ctx context.Context, orgID, sourceCID, targetCID string) (*types.MemoryRelationship, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(relationshipKey(orgID, sourceCID, targetCID))
	if err != nil {
		return nil, fmt.Errorf("get relationship: %w", err)
	}
	if bz == nil {
		return nil, types.ErrRelationshipNotFound
	}

	var stored types.StoredMemoryRelationship
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal relationship: %w", err)
	}
	rel := storedToRelationship(stored)
	return &rel, nil
}

func (k *Keeper) saveRelationship(ctx context.Context, orgID string, relationship *types.MemoryRelationship) error {
	stored := relationshipToStored(*relationship)
	stored.OrgId = orgID
	bz, err := proto.Marshal(&stored)
	if err != nil {
		return fmt.Errorf("marshal relationship: %w", err)
	}
	if err := k.getStore(ctx).Set(relationshipKey(orgID, relationship.SourceCID, relationship.TargetCID), bz); err != nil {
		return fmt.Errorf("store relationship: %w", err)
	}
	return nil
}

func (k *Keeper) applyConfidencePenalty(ctx context.Context, orgID, cid string, penalty uint64) error {
	return nil
}

func relationshipToStored(rel types.MemoryRelationship) types.StoredMemoryRelationship {
	return types.StoredMemoryRelationship{
		SourceCid:    rel.SourceCID,
		TargetCid:    rel.TargetCID,
		RelationType: rel.RelationType,
		Proposer:     rel.Proposer,
		Approved:     rel.Approved,
		Epoch:        rel.Epoch,
	}
}

func storedToRelationship(stored types.StoredMemoryRelationship) types.MemoryRelationship {
	return types.MemoryRelationship{
		SourceCID:    stored.SourceCid,
		TargetCID:    stored.TargetCid,
		RelationType: stored.RelationType,
		OrgID:        stored.OrgId,
		Proposer:     stored.Proposer,
		Approved:     stored.Approved,
		Epoch:        stored.Epoch,
	}
}

func isRelationshipState(state types.MemoryState) bool {
	return state == types.MemoryState_MEMORY_STATE_COMMITTED
}
