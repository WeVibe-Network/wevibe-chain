package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/gogoproto/proto"
	"github.com/wevibe-network/wevibe-chain/x/reputation/types"
)

const (
	modProfilePrefix = "modprofile/"
	MaxApprovedHashIndex = 1000
)

func moderatorProfileKey(modPubkey, orgID string) []byte {
	return []byte(fmt.Sprintf("%s/%s/%s", modProfilePrefix, modPubkey, orgID))
}

func (k *Keeper) GetModeratorProfile(ctx context.Context, modPubkey, orgID string) (*types.StoredModeratorProfile, error) {
	store := k.getStore(ctx)
	key := moderatorProfileKey(modPubkey, orgID)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return &types.StoredModeratorProfile{
			ModeratorPubkey: modPubkey,
			OrgId:           orgID,
		}, nil
	}
	var profile types.StoredModeratorProfile
	if err := proto.Unmarshal(bz, &profile); err != nil {
		return nil, fmt.Errorf("unmarshal moderator profile: %w", err)
	}
	return &profile, nil
}

func (k *Keeper) SetModeratorProfile(ctx context.Context, profile *types.StoredModeratorProfile) error {
	store := k.getStore(ctx)
	bz, err := proto.Marshal(profile)
	if err != nil {
		return fmt.Errorf("marshal moderator profile: %w", err)
	}
	return store.Set(moderatorProfileKey(profile.ModeratorPubkey, profile.OrgId), bz)
}

func (k *Keeper) IncrementModeratorApproval(ctx context.Context, modPubkey, orgID string, memoryHash []byte, epoch uint64) error {
	profile, err := k.GetModeratorProfile(ctx, modPubkey, orgID)
	if err != nil {
		return err
	}

	profile.TotalApprovals++

	if len(profile.ApprovedMemoryHashes) >= MaxApprovedHashIndex {
		profile.ApprovedMemoryHashes = profile.ApprovedMemoryHashes[1:]
	}
	profile.ApprovedMemoryHashes = append(profile.ApprovedMemoryHashes, memoryHash)

	if profile.FirstApprovalEpoch == 0 {
		profile.FirstApprovalEpoch = epoch
	}
	profile.LastApprovalEpoch = epoch

	return k.SetModeratorProfile(ctx, profile)
}

func (k *Keeper) IncrementModeratorUpheld(ctx context.Context, modPubkey, orgID string) error {
	profile, err := k.GetModeratorProfile(ctx, modPubkey, orgID)
	if err != nil {
		return err
	}

	profile.ApprovalsLaterUpheldCount++

	return k.SetModeratorProfile(ctx, profile)
}

func (k *Keeper) ListModeratorProfilesByOrg(ctx context.Context, orgID string) ([]*types.StoredModeratorProfile, error) {
	store := k.getStore(ctx)
	prefix := []byte(fmt.Sprintf("modprofile/%%/%s", orgID))
	iter, err := store.Iterator(prefix, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var profiles []*types.StoredModeratorProfile
	for ; iter.Valid(); iter.Next() {
		var profile types.StoredModeratorProfile
		if err := proto.Unmarshal(iter.Value(), &profile); err != nil {
			continue
		}
		profiles = append(profiles, &profile)
	}
	return profiles, nil
}

func (k *Keeper) ListModeratorProfilesByPubkey(ctx context.Context, modPubkey string) ([]*types.StoredModeratorProfile, error) {
	store := k.getStore(ctx)
	prefix := []byte(fmt.Sprintf("modprofile/%s/%%", modPubkey))
	iter, err := store.Iterator(prefix, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var profiles []*types.StoredModeratorProfile
	for ; iter.Valid(); iter.Next() {
		var profile types.StoredModeratorProfile
		if err := proto.Unmarshal(iter.Value(), &profile); err != nil {
			continue
		}
		profiles = append(profiles, &profile)
	}
	return profiles, nil
}