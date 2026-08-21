package keeper

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/wevibe-network/wevibe-chain/x/org/types"
)

type Keeper struct {
	storeService    store.KVStoreService
	logger          log.Logger
	authority       string
	bankKeeper      types.BankKeeper
	feegrantKeeper  types.FeegrantKeeper
	memoryKeeper    types.MemoryKeeper
	serveKeeper     types.ServeKeeper
	bandwidthKeeper types.BandwidthKeeper
}

func NewKeeper(storeService store.KVStoreService, logger log.Logger, authority string, bankKeeper types.BankKeeper, feegrantKeeper types.FeegrantKeeper) *Keeper {
	return &Keeper{
		storeService:   storeService,
		logger:         logger,
		authority:      authority,
		bankKeeper:     bankKeeper,
		feegrantKeeper: feegrantKeeper,
	}
}

func (k *Keeper) SetMemoryKeeper(mk types.MemoryKeeper)       { k.memoryKeeper = mk }
func (k *Keeper) SetServeKeeper(sk types.ServeKeeper)         { k.serveKeeper = sk }
func (k *Keeper) SetBandwidthKeeper(bk types.BandwidthKeeper) { k.bandwidthKeeper = bk }

func (k *Keeper) getStore(ctx context.Context) store.KVStore {
	return k.storeService.OpenKVStore(ctx)
}

func orgKey(orgID string) []byte {
	return []byte(fmt.Sprintf("org/%s", orgID))
}

func memberKey(orgID, memberPubkey string) []byte {
	return []byte(fmt.Sprintf("member/%s/%s", orgID, memberPubkey))
}

func slotRegistryKey() []byte {
	return []byte("slotreg/")
}

func orgConfigKey(orgID string) []byte {
	return []byte(fmt.Sprintf("orgconfig/%s", orgID))
}

const ParamsKey = "params"

func orgToStored(org *types.Org) *types.StoredOrg {
	return &types.StoredOrg{
		OrgId:               org.OrgID,
		Slot:                org.Slot,
		Leader:              org.Leader,
		Domain:              org.Domain,
		Name:                org.Name,
		Description:         org.Description,
		TechStack:           org.TechStack,
		FocusAreas:          org.FocusAreas,
		CreatedAt:           org.CreatedAt,
		RenewalHeight:       org.RenewalHeight,
		StorageQuota:        org.StorageQuota,
		RetrievalBudget:     org.RetrievalBudget,
		Status:              int32(org.Status),
		HubServingAddress:   org.HubServingAddress,
		HubEndpoints:        org.HubEndpoints,
		HubResponsePubkey:   org.HubResponsePubkey,
		LeaderWalletAddress: org.LeaderWalletAddress,
		AccountAddress:      org.AccountAddress,
	}
}

func storedToOrg(stored types.StoredOrg) types.Org {
	return types.Org{
		OrgID:               stored.OrgId,
		Slot:                stored.Slot,
		Leader:              stored.Leader,
		Domain:              stored.Domain,
		Name:                stored.Name,
		Description:         stored.Description,
		TechStack:           stored.TechStack,
		FocusAreas:          stored.FocusAreas,
		CreatedAt:           stored.CreatedAt,
		RenewalHeight:       stored.RenewalHeight,
		StorageQuota:        stored.StorageQuota,
		RetrievalBudget:     stored.RetrievalBudget,
		Status:              types.OrgStatus(stored.Status),
		HubServingAddress:   stored.HubServingAddress,
		HubEndpoints:        stored.HubEndpoints,
		HubResponsePubkey:   stored.HubResponsePubkey,
		LeaderWalletAddress: stored.LeaderWalletAddress,
		AccountAddress:      stored.AccountAddress,
	}
}

func memberToStored(member *types.MemberRecord) *types.StoredMemberRecord {
	return &types.StoredMemberRecord{
		OrgId:         member.OrgID,
		Pubkey:        member.Pubkey,
		Role:          member.Role,
		X25519Pubkey:  member.X25519Pubkey,
		CanContribute: member.CanContribute,
		CanModerate:   member.CanModerate,
	}
}

func storedToMember(stored types.StoredMemberRecord) types.MemberRecord {
	return types.MemberRecord{
		OrgID:         stored.OrgId,
		Pubkey:        stored.Pubkey,
		Role:          stored.Role,
		X25519Pubkey:  stored.X25519Pubkey,
		CanContribute: stored.CanContribute,
		CanModerate:   stored.CanModerate,
	}
}

func (k *Keeper) SetParams(ctx context.Context, params types.Params) error {
	store := k.getStore(ctx)
	bz, err := proto.Marshal(&params)
	if err != nil {
		return fmt.Errorf("marshal params: %w", err)
	}
	return store.Set([]byte(ParamsKey), bz)
}

func (k *Keeper) GetParams(ctx context.Context) (types.Params, error) {
	store := k.getStore(ctx)
	bz, err := store.Get([]byte(ParamsKey))
	if err != nil {
		return types.Params{}, fmt.Errorf("get params: %w", err)
	}
	if bz == nil {
		return types.DefaultParams(), nil
	}
	var params types.Params
	if err := proto.Unmarshal(bz, &params); err != nil {
		return types.Params{}, fmt.Errorf("unmarshal params: %w", err)
	}
	return params, nil
}

func (k *Keeper) GetNextSlot(ctx context.Context) (uint64, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(slotRegistryKey())
	if err != nil {
		return 0, fmt.Errorf("get slot registry: %w", err)
	}
	if bz == nil {
		return 0, nil
	}

	var registry types.StoredSlotRegistry
	if err := proto.Unmarshal(bz, &registry); err != nil {
		return 0, fmt.Errorf("unmarshal slot registry: %w", err)
	}

	return registry.NextSlot, nil
}

func (k *Keeper) SetNextSlot(ctx context.Context, nextSlot uint64) error {
	store := k.getStore(ctx)
	bz, err := proto.Marshal(&types.StoredSlotRegistry{NextSlot: nextSlot})
	if err != nil {
		return fmt.Errorf("marshal slot registry: %w", err)
	}

	return store.Set(slotRegistryKey(), bz)
}

func (k *Keeper) RegisterOrg(ctx context.Context, org *types.Org, creator sdk.AccAddress) error {
	nextSlot, err := k.GetNextSlot(ctx)
	if err != nil {
		return err
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if nextSlot >= params.SlotCap {
		return types.ErrSlotCapReached
	}

	org.Slot = nextSlot
	org.OrgID = types.FormatOrgID(nextSlot)
	org.AccountAddress = types.OrgAccountAddress(org.OrgID).String()

	if err := org.Validate(); err != nil {
		return err
	}

	store := k.getStore(ctx)
	key := orgKey(org.OrgID)

	has, err := store.Has(key)
	if err != nil {
		return err
	}
	if has {
		return types.ErrOrgAlreadyExists
	}

	orgs, err := k.GetAllOrgs(ctx)
	if err != nil {
		return err
	}
	for _, existingOrg := range orgs {
		if existingOrg.Leader == org.Leader {
			return types.ErrLeaderAlreadyOwnsOrg
		}
	}

	price := k.ComputeSlotPrice(ctx, org.Slot)
	if price.IsPositive() {
		burnHalf := price.QuoRaw(2)
		acctHalf := price.Sub(burnHalf)

		if burnHalf.IsPositive() {
			burnCoins := sdk.NewCoins(sdk.NewCoin("uvibe", burnHalf))
			if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creator, types.ModuleName, burnCoins); err != nil {
				return types.ErrInsufficientFund
			}
			if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, burnCoins); err != nil {
				return fmt.Errorf("burn coins: %w", err)
			}
		}

		if acctHalf.IsPositive() {
			acctCoins := sdk.NewCoins(sdk.NewCoin("uvibe", acctHalf))
			if err := k.bankKeeper.SendCoins(ctx, creator, types.OrgAccountAddress(org.OrgID), acctCoins); err != nil {
				return types.ErrInsufficientFund
			}
		}
	}

	storedOrg := orgToStored(org)
	storedOrg.TotalActiveMembers = 1

	bz, err := proto.Marshal(storedOrg)
	if err != nil {
		return fmt.Errorf("marshal org: %w", err)
	}

	store.Set(key, bz)

	if err := k.grantServingFeegrant(ctx, types.OrgAccountAddress(org.OrgID), org.HubServingAddress); err != nil {
		return err
	}
	if org.LeaderWalletAddress != "" {
		if err := k.grantLeaderFeegrant(ctx, types.OrgAccountAddress(org.OrgID), org.LeaderWalletAddress); err != nil {
			return err
		}
	}

	leaderKey := memberKey(org.OrgID, org.Leader)
	leaderBz, err := proto.Marshal(&types.StoredMemberRecord{
		OrgId:  org.OrgID,
		Pubkey: org.Leader,
		Role:   "leader",
	})
	if err != nil {
		return fmt.Errorf("marshal leader: %w", err)
	}
	store.Set(leaderKey, leaderBz)

	if err := k.SetNextSlot(ctx, nextSlot+1); err != nil {
		return err
	}

	k.logger.Info("org registered",
		"org_id", org.OrgID,
		"slot", org.Slot,
		"leader", org.Leader,
	)
	return nil
}

func (k *Keeper) GetOrg(ctx context.Context, orgID string) (*types.Org, error) {
	store := k.getStore(ctx)
	key := orgKey(orgID)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrOrgNotFound
	}

	var stored types.StoredOrg
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal org: %w", err)
	}
	org := storedToOrg(stored)
	return &org, nil
}

func (k *Keeper) HasOrg(ctx context.Context, orgID string) (bool, error) {
	store := k.getStore(ctx)
	has, err := store.Has(orgKey(orgID))
	if err != nil {
		return false, err
	}
	return has, nil
}

func (k *Keeper) AddMember(ctx context.Context, member *types.MemberRecord) error {
	if member.Role != "member" {
		return types.ErrInvalidRole
	}

	store := k.getStore(ctx)
	orgStorageKey := orgKey(member.OrgID)
	orgBz, err := store.Get(orgStorageKey)
	if err != nil {
		return err
	}
	if orgBz == nil {
		return types.ErrOrgNotFound
	}

	var orgStored types.StoredOrg
	if err := proto.Unmarshal(orgBz, &orgStored); err != nil {
		return fmt.Errorf("unmarshal org: %w", err)
	}
	if types.OrgStatus(orgStored.Status) != types.OrgStatus_ACTIVE {
		return types.ErrOrgNotActive
	}

	key := memberKey(member.OrgID, member.Pubkey)

	has, err := store.Has(key)
	if err != nil {
		return err
	}
	if has {
		return types.ErrMemberExists
	}

	bz, err := proto.Marshal(memberToStored(member))
	if err != nil {
		return fmt.Errorf("marshal member: %w", err)
	}

	if err := store.Set(key, bz); err != nil {
		return err
	}

	orgStored.TotalActiveMembers++
	orgBz, err = proto.Marshal(&orgStored)
	if err != nil {
		return fmt.Errorf("marshal org: %w", err)
	}
	if err := store.Set(orgStorageKey, orgBz); err != nil {
		return err
	}

	k.logger.Info("member added",
		"org_id", member.OrgID,
		"member", member.Pubkey,
		"role", member.Role,
	)
	return nil
}

func (k *Keeper) RemoveMember(ctx context.Context, orgID, memberPubkey string) error {
	store := k.getStore(ctx)
	key := memberKey(orgID, memberPubkey)

	memberBz, err := store.Get(key)
	if err != nil {
		return err
	}
	if memberBz == nil {
		return types.ErrMemberNotFound
	}

	var memberStored types.StoredMemberRecord
	if err := proto.Unmarshal(memberBz, &memberStored); err != nil {
		return fmt.Errorf("unmarshal member: %w", err)
	}
	if memberStored.Role == "leader" {
		return types.ErrCannotRemoveLeader
	}

	orgStorageKey := orgKey(orgID)
	orgBz, err := store.Get(orgStorageKey)
	if err != nil {
		return err
	}
	if orgBz == nil {
		return types.ErrOrgNotFound
	}

	var orgStored types.StoredOrg
	if err := proto.Unmarshal(orgBz, &orgStored); err != nil {
		return fmt.Errorf("unmarshal org: %w", err)
	}
	if types.OrgStatus(orgStored.Status) != types.OrgStatus_ACTIVE {
		return types.ErrOrgNotActive
	}

	if err := store.Delete(key); err != nil {
		return err
	}

	if orgStored.TotalActiveMembers > 0 {
		orgStored.TotalActiveMembers--
	}
	orgBz, err = proto.Marshal(&orgStored)
	if err != nil {
		return fmt.Errorf("marshal org: %w", err)
	}
	if err := store.Set(orgStorageKey, orgBz); err != nil {
		return err
	}

	k.logger.Info("member removed",
		"org_id", orgID,
		"member", memberPubkey,
	)
	return nil
}

func (k *Keeper) GetMember(ctx context.Context, orgID, memberPubkey string) (*types.MemberRecord, error) {
	store := k.getStore(ctx)
	key := memberKey(orgID, memberPubkey)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrMemberNotFound
	}

	var stored types.StoredMemberRecord
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal member: %w", err)
	}
	member := storedToMember(stored)
	return &member, nil
}

func (k *Keeper) GetAllMembers(ctx context.Context, orgID string) ([]*types.MemberRecord, error) {
	store := k.getStore(ctx)
	prefix := []byte(fmt.Sprintf("member/%s/", orgID))
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var members []*types.MemberRecord
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredMemberRecord
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		member := storedToMember(stored)
		members = append(members, &member)
	}
	return members, nil
}

func (k *Keeper) IsMember(ctx context.Context, orgID, memberPubkey string) (bool, error) {
	store := k.getStore(ctx)
	has, err := store.Has(memberKey(orgID, memberPubkey))
	if err != nil {
		return false, err
	}
	return has, nil
}

func (k *Keeper) IsLeader(ctx context.Context, orgID, memberPubkey string) (bool, error) {
	member, err := k.GetMember(ctx, orgID, memberPubkey)
	if err != nil {
		if errors.Is(err, types.ErrMemberNotFound) {
			return false, nil
		}
		return false, err
	}
	return member.Role == "leader", nil
}

func (k *Keeper) IsModerator(ctx context.Context, orgID, memberPubkey string) (bool, error) {
	member, err := k.GetMember(ctx, orgID, memberPubkey)
	if err != nil {
		if errors.Is(err, types.ErrMemberNotFound) {
			return false, nil
		}
		return false, err
	}
	return member.CanModerate, nil
}

// GetServingAddress returns the org's currently-registered hub serving key
// chain address — the only signer authorized to submit serve/denial batches
// (D-S32-CO044-KEY-SEPARATION). Returns ErrOrgNotFound if the org does not
// exist; the serving address may be empty if the org was never provisioned
// with one (in which case no serve/denial batch can ever be accepted).
func (k *Keeper) GetServingAddress(ctx context.Context, orgID string) (string, error) {
	org, err := k.GetOrg(ctx, orgID)
	if err != nil {
		return "", err
	}
	return org.HubServingAddress, nil
}

// GetLeaderWallet returns the org's registered leader chain wallet address —
// the authenticated tx signer authorized to commit org decisions
// (D-S32-CO044-KEY-SEPARATION). May be empty if the org was never provisioned
// with one (in which case no org-decision message can be authorized).
func (k *Keeper) GetLeaderWallet(ctx context.Context, orgID string) (string, error) {
	org, err := k.GetOrg(ctx, orgID)
	if err != nil {
		return "", err
	}
	return org.LeaderWalletAddress, nil
}

// SetServingKey rotates/revokes the org's hub serving key. Authorized ONLY by
// the org's registered leader chain wallet (signer must equal
// leader_wallet_address). D-S32-CO044-SERVING-KEY-REVOCATION.
func (k *Keeper) SetServingKey(ctx context.Context, orgID, newServingKey, signer string) error {
	org, err := k.GetOrg(ctx, orgID)
	if err != nil {
		return err
	}
	if org.LeaderWalletAddress == "" || signer != org.LeaderWalletAddress {
		return types.ErrNotLeader
	}

	store := k.getStore(ctx)
	key := orgKey(orgID)
	bz, err := store.Get(key)
	if err != nil {
		return err
	}

	var stored types.StoredOrg
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return fmt.Errorf("unmarshal org: %w", err)
	}
	stored.HubServingAddress = newServingKey
	bz, err = proto.Marshal(&stored)
	if err != nil {
		return fmt.Errorf("marshal org: %w", err)
	}
	if err := store.Set(key, bz); err != nil {
		return err
	}

	// Feegrant keeper in this SDK version does not expose a keeper-level revoke
	// API. old grant is harmless — x/serve auth requires
	// signer==HubServingAddress, so the de-whitelisted old key cannot submit
	// serve/deny regardless of a lingering feegrant.
	if err := k.grantServingFeegrant(ctx, types.OrgAccountAddress(orgID), newServingKey); err != nil {
		return err
	}

	k.logger.Info("serving key rotated",
		"org_id", orgID,
		"new_serving_key", newServingKey,
	)
	return nil
}

// SetServingInfo updates the ordered hub transport endpoint list and the
// optional hub response signing pubkey for an org. Authorized ONLY by the
// org's registered leader chain wallet (signer must equal
// leader_wallet_address).
func (k *Keeper) SetServingInfo(ctx context.Context, orgID string, endpoints []string, responsePubkey string, signer string) error {
	org, err := k.GetOrg(ctx, orgID)
	if err != nil {
		return err
	}
	if org.LeaderWalletAddress == "" || signer != org.LeaderWalletAddress {
		return types.ErrNotLeader
	}

	if err := types.ValidateHubEndpoints(endpoints); err != nil {
		return err
	}

	store := k.getStore(ctx)
	key := orgKey(orgID)
	bz, err := store.Get(key)
	if err != nil {
		return err
	}

	var stored types.StoredOrg
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return fmt.Errorf("unmarshal org: %w", err)
	}
	stored.HubEndpoints = append([]string(nil), endpoints...)
	stored.HubResponsePubkey = responsePubkey
	bz, err = proto.Marshal(&stored)
	if err != nil {
		return fmt.Errorf("marshal org: %w", err)
	}
	if err := store.Set(key, bz); err != nil {
		return err
	}

	return nil
}

func (k *Keeper) SetMemberCapabilities(ctx context.Context, orgID, pubkey string, canContribute, canModerate bool, signer string) error {
	org, err := k.GetOrg(ctx, orgID)
	if err != nil {
		return err
	}
	if org.LeaderWalletAddress == "" || signer != org.LeaderWalletAddress {
		return types.ErrNotLeader
	}
	if org.Status != types.OrgStatus_ACTIVE {
		return types.ErrOrgNotActive
	}

	member, err := k.GetMember(ctx, orgID, pubkey)
	if err != nil {
		return err
	}

	member.CanContribute = canContribute
	member.CanModerate = canModerate
	bz, err := proto.Marshal(memberToStored(member))
	if err != nil {
		return fmt.Errorf("marshal member: %w", err)
	}

	store := k.getStore(ctx)
	if err := store.Set(memberKey(orgID, pubkey), bz); err != nil {
		return err
	}

	k.logger.Info("member capabilities set",
		"org_id", orgID,
		"member", pubkey,
		"can_contribute", canContribute,
		"can_moderate", canModerate,
	)
	return nil
}

func (k *Keeper) TransferLeadership(ctx context.Context, orgID, newLeader, newLeaderWallet, signer string) error {
	if newLeaderWallet == "" {
		return fmt.Errorf("new_leader_wallet cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(newLeaderWallet); err != nil {
		return fmt.Errorf("invalid new_leader_wallet: %w", err)
	}

	org, err := k.GetOrg(ctx, orgID)
	if err != nil {
		return err
	}

	if org.LeaderWalletAddress == "" || signer != org.LeaderWalletAddress {
		return types.ErrNotLeader
	}

	_, err = k.GetMember(ctx, orgID, newLeader)
	if err != nil {
		return fmt.Errorf("new_leader must be a member of the org")
	}

	store := k.getStore(ctx)

	oldLeaderKey := memberKey(orgID, org.Leader)
	oldLeaderBz, err := store.Get(oldLeaderKey)
	if err != nil {
		return err
	}
	var oldLeaderStored types.StoredMemberRecord
	if err := proto.Unmarshal(oldLeaderBz, &oldLeaderStored); err != nil {
		return fmt.Errorf("unmarshal old leader: %w", err)
	}
	oldLeaderStored.Role = "member"
	oldLeaderBz, err = proto.Marshal(&oldLeaderStored)
	if err != nil {
		return fmt.Errorf("marshal old leader: %w", err)
	}
	if err := store.Set(oldLeaderKey, oldLeaderBz); err != nil {
		return err
	}

	newLeaderKey := memberKey(orgID, newLeader)
	newLeaderBz, err := store.Get(newLeaderKey)
	if err != nil {
		return err
	}
	var newLeaderStored types.StoredMemberRecord
	if err := proto.Unmarshal(newLeaderBz, &newLeaderStored); err != nil {
		return fmt.Errorf("unmarshal new leader: %w", err)
	}
	newLeaderStored.Role = "leader"
	newLeaderBz, err = proto.Marshal(&newLeaderStored)
	if err != nil {
		return fmt.Errorf("marshal new leader: %w", err)
	}
	if err := store.Set(newLeaderKey, newLeaderBz); err != nil {
		return err
	}

	orgKey := orgKey(orgID)
	orgBz, err := store.Get(orgKey)
	if err != nil {
		return err
	}
	var orgStored types.StoredOrg
	if err := proto.Unmarshal(orgBz, &orgStored); err != nil {
		return fmt.Errorf("unmarshal org: %w", err)
	}
	orgStored.Leader = newLeader
	orgStored.LeaderWalletAddress = newLeaderWallet
	orgBz, err = proto.Marshal(&orgStored)
	if err != nil {
		return fmt.Errorf("marshal org: %w", err)
	}
	if err := store.Set(orgKey, orgBz); err != nil {
		return err
	}

	orgAccountAddr := types.OrgAccountAddress(orgID)
	// This SDK feegrant keeper does not expose a public keeper-level revoke
	// method, so old leader grants cannot be proactively removed here. The old
	// leader allowance is inert: org-purpose messages gate on
	// signer == current LeaderWalletAddress, and this feegrant is scoped only to
	// org-purpose message type URLs.

	if err := k.grantLeaderFeegrant(ctx, orgAccountAddr, newLeaderWallet); err != nil {
		return err
	}

	k.logger.Info("leadership transferred",
		"org_id", orgID,
		"old_leader", org.Leader,
		"new_leader", newLeader,
	)
	return nil
}

func (k *Keeper) CloseOrg(ctx context.Context, orgID, signer string) error {
	org, err := k.GetOrg(ctx, orgID)
	if err != nil {
		return err
	}

	if org.LeaderWalletAddress == "" || signer != org.LeaderWalletAddress {
		return types.ErrNotLeader
	}

	if org.Status != types.OrgStatus_ACTIVE {
		return types.ErrOrgNotActive
	}

	store := k.getStore(ctx)
	key := orgKey(orgID)
	bz, err := store.Get(key)
	if err != nil {
		return err
	}

	var stored types.StoredOrg
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return fmt.Errorf("unmarshal org: %w", err)
	}

	stored.Status = int32(types.OrgStatus_CLOSED)

	bz, err = proto.Marshal(&stored)
	if err != nil {
		return fmt.Errorf("marshal org: %w", err)
	}
	if err := store.Set(key, bz); err != nil {
		return err
	}

	k.logger.Info("org closed",
		"org_id", orgID,
	)
	return nil
}

func (k *Keeper) UpdateOrgStatus(ctx context.Context, orgID string, status types.OrgStatus) error {
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

	stored.Status = int32(status)

	bz, err = proto.Marshal(&stored)
	if err != nil {
		return fmt.Errorf("marshal org: %w", err)
	}
	if err := store.Set(key, bz); err != nil {
		return err
	}
	return nil
}

func (k *Keeper) UpdateStorageQuota(ctx context.Context, orgID string, quota uint64) error {
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

	stored.StorageQuota = quota

	bz, err = proto.Marshal(&stored)
	if err != nil {
		return fmt.Errorf("marshal org: %w", err)
	}
	if err := store.Set(key, bz); err != nil {
		return err
	}
	return nil
}

func (k *Keeper) computeAscendingPrice(ctx context.Context, slot uint64) uint64 {
	params, err := k.GetParams(ctx)
	if err != nil {
		return params.BaseBurnPrice
	}

	basePrice := params.BaseBurnPrice
	increasePercent := params.BurnPriceIncreasePercent

	multiplier := big.NewRat(100, 100)
	inc := big.NewRat(int64(increasePercent), 100)
	for i := uint64(0); i < slot; i++ {
		multiplier = multiplier.Add(multiplier, inc)
	}

	result := new(big.Int).Mul(big.NewInt(int64(basePrice)), multiplier.Num())
	result = result.Div(result, multiplier.Denom())

	if result.Sign() <= 0 {
		return params.BaseBurnPrice
	}

	if result.BitLen() > 64 {
		return params.BaseBurnPrice
	}
	return result.Uint64()
}

func (k *Keeper) ComputeSlotPrice(ctx context.Context, slot uint64) math.Int {
	price := k.computeAscendingPrice(ctx, slot)
	return math.NewIntFromUint64(price)
}

func (k *Keeper) GetOrgConfig(ctx context.Context, orgID string) (*types.OrgConfig, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(orgConfigKey(orgID))
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return &types.OrgConfig{
			OrgID:                    orgID,
			ServeReceiptRequired:     false,
			ContestStakeVibe:         0,
			VocabHash:                "",
			EmbeddingModelID:         "",
			MinContributionsPerEpoch: 0,
		}, nil
	}

	var stored types.StoredOrgConfig
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal org config: %w", err)
	}
	return &types.OrgConfig{
		OrgID:                    stored.OrgId,
		ServeReceiptRequired:     stored.ServeReceiptRequired,
		ContestStakeVibe:         stored.ContestStakeVibe,
		VocabHash:                stored.VocabHash,
		EmbeddingModelID:         stored.EmbeddingModelId,
		MinContributionsPerEpoch: stored.MinContributionsPerEpoch,
	}, nil
}

func (k *Keeper) SetOrgConfig(ctx context.Context, orgID string, cfg *types.OrgConfig) error {
	store := k.getStore(ctx)
	bz, err := proto.Marshal(&types.StoredOrgConfig{
		OrgId:                    cfg.OrgID,
		ServeReceiptRequired:     cfg.ServeReceiptRequired,
		ContestStakeVibe:         cfg.ContestStakeVibe,
		VocabHash:                cfg.VocabHash,
		EmbeddingModelId:         cfg.EmbeddingModelID,
		MinContributionsPerEpoch: cfg.MinContributionsPerEpoch,
	})
	if err != nil {
		return fmt.Errorf("marshal org config: %w", err)
	}
	return store.Set(orgConfigKey(orgID), bz)
}

func (k *Keeper) GetAllOrgs(ctx context.Context) ([]*types.Org, error) {
	store := k.getStore(ctx)
	prefix := []byte("org/")
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var orgs []*types.Org
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredOrg
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		org := storedToOrg(stored)
		orgs = append(orgs, &org)
	}
	return orgs, nil
}

func (k *Keeper) InitGenesis(ctx context.Context, state *types.GenesisState) error {
	store := k.getStore(ctx)
	params := state.Params
	if params == (types.Params{}) {
		params = types.DefaultParams()
	}
	if err := k.SetParams(ctx, params); err != nil {
		return err
	}

	if err := k.SetNextSlot(ctx, state.NextSlot); err != nil {
		return err
	}

	for _, org := range state.Orgs {
		if err := org.Validate(); err != nil {
			return err
		}
		bz, err := proto.Marshal(orgToStored(org))
		if err != nil {
			return err
		}
		if err := store.Set(orgKey(org.OrgID), bz); err != nil {
			return err
		}
	}

	for _, member := range state.Members {
		bz, err := proto.Marshal(memberToStored(member))
		if err != nil {
			return err
		}
		if err := store.Set(memberKey(member.OrgID, member.Pubkey), bz); err != nil {
			return err
		}
	}

	for _, orgConfig := range state.OrgConfigs {
		bz, err := proto.Marshal(&types.StoredOrgConfig{
			OrgId:                orgConfig.OrgID,
			ServeReceiptRequired: orgConfig.ServeReceiptRequired,
		})
		if err != nil {
			return err
		}
		if err := store.Set(orgConfigKey(orgConfig.OrgID), bz); err != nil {
			return err
		}
	}

	return nil
}

func (k *Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	store := k.getStore(ctx)
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	nextSlot, err := k.GetNextSlot(ctx)
	if err != nil {
		return nil, err
	}

	orgPrefix := []byte("org/")
	orgIter, err := store.Iterator(orgPrefix, storetypes.PrefixEndBytes(orgPrefix))
	if err != nil {
		return nil, err
	}
	defer orgIter.Close()

	var orgs []*types.Org
	for ; orgIter.Valid(); orgIter.Next() {
		var stored types.StoredOrg
		if err := proto.Unmarshal(orgIter.Value(), &stored); err != nil {
			continue
		}
		org := storedToOrg(stored)
		orgs = append(orgs, &org)
	}

	memberPrefix := []byte("member/")
	memberIter, err := store.Iterator(memberPrefix, storetypes.PrefixEndBytes(memberPrefix))
	if err != nil {
		return nil, err
	}
	defer memberIter.Close()

	var members []*types.MemberRecord
	for ; memberIter.Valid(); memberIter.Next() {
		var stored types.StoredMemberRecord
		if err := proto.Unmarshal(memberIter.Value(), &stored); err != nil {
			continue
		}
		member := storedToMember(stored)
		members = append(members, &member)
	}

	orgconfigPrefix := []byte("orgconfig/")
	orgconfigIter, err := store.Iterator(orgconfigPrefix, storetypes.PrefixEndBytes(orgconfigPrefix))
	if err != nil {
		return nil, err
	}
	defer orgconfigIter.Close()

	var orgConfigs []*types.OrgConfig
	for ; orgconfigIter.Valid(); orgconfigIter.Next() {
		var stored types.StoredOrgConfig
		if err := proto.Unmarshal(orgconfigIter.Value(), &stored); err != nil {
			continue
		}
		orgConfigs = append(orgConfigs, &types.OrgConfig{
			OrgID:                    stored.OrgId,
			ServeReceiptRequired:     stored.ServeReceiptRequired,
			ContestStakeVibe:         stored.ContestStakeVibe,
			MinContributionsPerEpoch: stored.MinContributionsPerEpoch,
		})
	}

	return &types.GenesisState{
		Orgs:       orgs,
		Members:    members,
		OrgConfigs: orgConfigs,
		NextSlot:   nextSlot,
		Params:     params,
	}, nil
}

func (k *Keeper) Logger(ctx context.Context) log.Logger {
	return k.logger.With("module", "x/org")
}
