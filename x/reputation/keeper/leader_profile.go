package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/gogoproto/proto"
	"github.com/wevibe-network/wevibe-chain/x/reputation/types"
)

const (
	leaderProfilePrefix = "leaderprofile/"
)

func leaderProfileKey(leaderPubkey, orgID string) []byte {
	return []byte(fmt.Sprintf("%s/%s/%s", leaderProfilePrefix, leaderPubkey, orgID))
}

func (k *Keeper) GetLeaderProfile(ctx context.Context, leaderPubkey, orgID string) (*types.StoredLeaderProfile, error) {
	store := k.getStore(ctx)
	key := leaderProfileKey(leaderPubkey, orgID)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return &types.StoredLeaderProfile{
			LeaderPubkey: leaderPubkey,
			OrgId:       orgID,
		}, nil
	}
	var profile types.StoredLeaderProfile
	if err := proto.Unmarshal(bz, &profile); err != nil {
		return nil, fmt.Errorf("unmarshal leader profile: %w", err)
	}
	return &profile, nil
}

func (k *Keeper) SetLeaderProfile(ctx context.Context, profile *types.StoredLeaderProfile) error {
	store := k.getStore(ctx)
	bz, err := proto.Marshal(profile)
	if err != nil {
		return fmt.Errorf("marshal leader profile: %w", err)
	}
	return store.Set(leaderProfileKey(profile.LeaderPubkey, profile.OrgId), bz)
}

func (k *Keeper) IncrementLeaderChainCommit(ctx context.Context, leaderPubkey, orgID string) error {
	profile, err := k.GetLeaderProfile(ctx, leaderPubkey, orgID)
	if err != nil {
		return err
	}

	profile.TotalChainCommitsSigned++

	return k.SetLeaderProfile(ctx, profile)
}

func (k *Keeper) IncrementLeaderUpheldReport(ctx context.Context, leaderPubkey, orgID string) error {
	profile, err := k.GetLeaderProfile(ctx, leaderPubkey, orgID)
	if err != nil {
		return err
	}

	profile.TotalUpheldReportsCommitted++

	return k.SetLeaderProfile(ctx, profile)
}

func (k *Keeper) SetLeaderCurrent(ctx context.Context, leaderPubkey, orgID string, current bool) error {
	profile, err := k.GetLeaderProfile(ctx, leaderPubkey, orgID)
	if err != nil {
		return err
	}

	profile.CurrentLeader = current

	return k.SetLeaderProfile(ctx, profile)
}

func (k *Keeper) ListLeaderProfilesByOrg(ctx context.Context, orgID string) ([]*types.StoredLeaderProfile, error) {
	store := k.getStore(ctx)
	prefix := []byte(fmt.Sprintf("leaderprofile/%%/%s", orgID))
	iter, err := store.Iterator(prefix, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var profiles []*types.StoredLeaderProfile
	for ; iter.Valid(); iter.Next() {
		var profile types.StoredLeaderProfile
		if err := proto.Unmarshal(iter.Value(), &profile); err != nil {
			continue
		}
		profiles = append(profiles, &profile)
	}
	return profiles, nil
}

func (k *Keeper) ListLeaderProfilesByPubkey(ctx context.Context, leaderPubkey string) ([]*types.StoredLeaderProfile, error) {
	store := k.getStore(ctx)
	prefix := []byte(fmt.Sprintf("leaderprofile/%s/%%", leaderPubkey))
	iter, err := store.Iterator(prefix, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var profiles []*types.StoredLeaderProfile
	for ; iter.Valid(); iter.Next() {
		var profile types.StoredLeaderProfile
		if err := proto.Unmarshal(iter.Value(), &profile); err != nil {
			continue
		}
		profiles = append(profiles, &profile)
	}
	return profiles, nil
}