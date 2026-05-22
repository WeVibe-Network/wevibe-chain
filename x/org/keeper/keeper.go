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

func dynamicPriceKey() []byte {
	return []byte("dynprice/")
}

func treasuryKey(orgID string) []byte {
	return []byte(fmt.Sprintf("treasury/%s", orgID))
}

func repTierKey(orgID string) []byte {
	return []byte(fmt.Sprintf("reptier/%s", orgID))
}

func orgConfigKey(orgID string) []byte {
	return []byte(fmt.Sprintf("orgconfig/%s", orgID))
}

const ParamsKey = "params"

func orgToStored(org *types.Org) *types.StoredOrg {
	return &types.StoredOrg{
		OrgId:           org.OrgID,
		Leader:          org.Leader,
		Domain:          org.Domain,
		CreatedAt:       org.CreatedAt,
		RenewalHeight:   org.RenewalHeight,
		StorageQuota:    org.StorageQuota,
		RetrievalBudget: org.RetrievalBudget,
		Status:          int32(org.Status),
	}
}

func storedToOrg(stored types.StoredOrg) types.Org {
	return types.Org{
		OrgID:           stored.OrgId,
		Leader:          stored.Leader,
		Domain:          stored.Domain,
		CreatedAt:       stored.CreatedAt,
		RenewalHeight:   stored.RenewalHeight,
		StorageQuota:    stored.StorageQuota,
		RetrievalBudget: stored.RetrievalBudget,
		Status:          types.OrgStatus(stored.Status),
	}
}

func memberToStored(member *types.MemberRecord) *types.StoredMemberRecord {
	return &types.StoredMemberRecord{
		OrgId:  member.OrgID,
		Pubkey: member.Pubkey,
		Role:   member.Role,
	}
}

func storedToMember(stored types.StoredMemberRecord) types.MemberRecord {
	return types.MemberRecord{
		OrgID:  stored.OrgId,
		Pubkey: stored.Pubkey,
		Role:   stored.Role,
	}
}

func dynamicPriceToStored(dp *types.DynamicPrice) *types.StoredDynamicPrice {
	return &types.StoredDynamicPrice{
		Price:         dp.Price,
		LastCreation:  dp.LastCreation,
		CreationCount: dp.CreationCount,
	}
}

func storedToDynamicPrice(stored types.StoredDynamicPrice) types.DynamicPrice {
	return types.DynamicPrice{
		Price:         stored.Price,
		LastCreation:  stored.LastCreation,
		CreationCount: stored.CreationCount,
	}
}

func treasuryToStored(treasury *types.Treasury) *types.StoredTreasury {
	return &types.StoredTreasury{
		OrgId:   treasury.OrgID,
		Balance: treasury.Balance,
	}
}

func storedToTreasury(stored types.StoredTreasury) types.Treasury {
	return types.Treasury{
		OrgID:   stored.OrgId,
		Balance: stored.Balance,
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

func (k *Keeper) RegisterOrg(ctx context.Context, org *types.Org, creator sdk.AccAddress) error {
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

	burnPrice := k.ComputeBurnPrice(ctx)
	if burnPrice.IsPositive() {
		burnCoins := sdk.NewCoins(sdk.NewCoin("uvibe", burnPrice))
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creator, types.ModuleName, burnCoins); err != nil {
			return types.ErrInsufficientFund
		}
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, burnCoins); err != nil {
			return fmt.Errorf("burn coins: %w", err)
		}
	}

	bz, err := proto.Marshal(orgToStored(org))
	if err != nil {
		return fmt.Errorf("marshal org: %w", err)
	}

	store.Set(key, bz)

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

	k.updateDynamicPriceOnCreation(ctx)

	k.logger.Info("org registered",
		"org_id", org.OrgID,
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
	store := k.getStore(ctx)
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

	has, err := store.Has(key)
	if err != nil {
		return err
	}
	if !has {
		return types.ErrMemberNotFound
	}

	if err := store.Delete(key); err != nil {
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
	if err := iter.Error(); err != nil {
		return nil, err
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
	return member.Role == "moderator", nil
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

func (k *Keeper) GetDynamicPrice(ctx context.Context) (*types.DynamicPrice, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(dynamicPriceKey())
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, errors.New("dynamic price not set")
	}

	var stored types.StoredDynamicPrice
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal dynamic price: %w", err)
	}
	dp := storedToDynamicPrice(stored)
	return &dp, nil
}

func (k *Keeper) SetDynamicPrice(ctx context.Context, dp *types.DynamicPrice) error {
	store := k.getStore(ctx)
	bz, err := proto.Marshal(dynamicPriceToStored(dp))
	if err != nil {
		return fmt.Errorf("marshal dynamic price: %w", err)
	}
	store.Set(dynamicPriceKey(), bz)
	return nil
}

func (k *Keeper) updateDynamicPriceOnCreation(ctx context.Context) {
	store := k.getStore(ctx)
	bz, err := store.Get(dynamicPriceKey())
	if err != nil {
		return
	}

	var dp types.DynamicPrice
	if bz != nil {
		var stored types.StoredDynamicPrice
		if err := proto.Unmarshal(bz, &stored); err != nil {
			return
		}
		dp = storedToDynamicPrice(stored)
	}

	dp.CreationCount++
	dp.Price = k.ComputeDynamicPriceValue(ctx, dp.CreationCount)

	bz, err = proto.Marshal(dynamicPriceToStored(&dp))
	if err != nil {
		return
	}
	store.Set(dynamicPriceKey(), bz)
}

func (k *Keeper) ComputeDynamicPriceValue(ctx context.Context, creationCount uint64) uint64 {
	params, err := k.GetParams(ctx)
	if err != nil {
		return params.BaseBurnPrice
	}

	basePrice := params.BaseBurnPrice
	increasePercent := params.BurnPriceIncreasePercent

	multiplier := big.NewRat(100, 100)
	inc := big.NewRat(int64(increasePercent), 100)
	for i := uint64(0); i < creationCount; i++ {
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

func (k *Keeper) ComputeBurnPrice(ctx context.Context) math.Int {
	store := k.getStore(ctx)
	bz, err := store.Get(dynamicPriceKey())
	if err != nil {
		return math.ZeroInt()
	}

	var dp types.DynamicPrice
	if bz != nil {
		var stored types.StoredDynamicPrice
		if err := proto.Unmarshal(bz, &stored); err != nil {
			return math.ZeroInt()
		}
		dp = storedToDynamicPrice(stored)
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return math.ZeroInt()
	}

	if dp.CreationCount > params.BurnPriceDecayEpochs {
		dp.CreationCount = dp.CreationCount - params.BurnPriceDecayEpochs
	} else {
		dp.CreationCount = 0
	}

	price := k.ComputeDynamicPriceValue(ctx, dp.CreationCount)
	return math.NewInt(int64(price))
}

func (k *Keeper) GetTreasuryBalance(ctx context.Context, orgID string) (string, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(treasuryKey(orgID))
	if err != nil {
		return "0", err
	}
	if bz == nil {
		return "0", nil
	}

	var stored types.StoredTreasury
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return "0", fmt.Errorf("unmarshal treasury: %w", err)
	}
	return stored.Balance, nil
}

func (k *Keeper) GetTreasuryBalanceInt(ctx context.Context, orgID string) (math.Int, error) {
	balance, err := k.GetTreasuryBalance(ctx, orgID)
	if err != nil {
		return math.ZeroInt(), err
	}
	if balance == "" {
		return math.ZeroInt(), nil
	}
	val, ok := math.NewIntFromString(balance)
	if !ok {
		return math.ZeroInt(), fmt.Errorf("invalid treasury balance: %s", balance)
	}
	return val, nil
}

func (k *Keeper) FundTreasury(ctx context.Context, orgID string, funder sdk.AccAddress, amount math.Int) error {
	store := k.getStore(ctx)

	balance, err := k.GetTreasuryBalanceInt(ctx, orgID)
	if err != nil {
		return err
	}

	coins := sdk.NewCoins(sdk.NewCoin("uvibe", amount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, funder, types.ModuleName, coins); err != nil {
		return err
	}

	newBalance := balance.Add(amount)
	bz, err := proto.Marshal(&types.StoredTreasury{
		OrgId:   orgID,
		Balance: newBalance.String(),
	})
	if err != nil {
		return fmt.Errorf("marshal treasury: %w", err)
	}
	return store.Set(treasuryKey(orgID), bz)
}

func (k *Keeper) WithdrawTreasury(ctx context.Context, orgID string, recipient sdk.AccAddress, amount math.Int) error {
	store := k.getStore(ctx)

	balance, err := k.GetTreasuryBalanceInt(ctx, orgID)
	if err != nil {
		return err
	}

	if balance.LT(amount) {
		return types.ErrInsufficientTreasury
	}

	coins := sdk.NewCoins(sdk.NewCoin("uvibe", amount))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, coins); err != nil {
		return err
	}

	newBalance := balance.Sub(amount)
	bz, err := proto.Marshal(&types.StoredTreasury{
		OrgId:   orgID,
		Balance: newBalance.String(),
	})
	if err != nil {
		return fmt.Errorf("marshal treasury: %w", err)
	}
	return store.Set(treasuryKey(orgID), bz)
}

func (k *Keeper) DebitTreasury(ctx context.Context, orgID string, amount math.Int) error {
	store := k.getStore(ctx)

	balance, err := k.GetTreasuryBalanceInt(ctx, orgID)
	if err != nil {
		return err
	}

	if balance.LT(amount) {
		return types.ErrInsufficientTreasury
	}

	newBalance := balance.Sub(amount)
	bz, err := proto.Marshal(&types.StoredTreasury{
		OrgId:   orgID,
		Balance: newBalance.String(),
	})
	if err != nil {
		return fmt.Errorf("marshal treasury: %w", err)
	}
	return store.Set(treasuryKey(orgID), bz)
}

func (k *Keeper) GetRepTiers(ctx context.Context, orgID string) (*types.RepTierConfig, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(repTierKey(orgID))
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrTreasuryNotFound
	}

	var stored types.StoredRepTierConfig
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal rep tier config: %w", err)
	}
	tiers := make([]*types.RepTierRecord, len(stored.Tiers))
	for i, t := range stored.Tiers {
		tiers[i] = &types.RepTierRecord{
			MinReputation:            t.MinReputation,
			MaxReputation:            t.MaxReputation,
			MaxContributionsPerEpoch: t.MaxContributionsPerEpoch,
			PayoutPerMemory:          t.PayoutPerMemory,
		}
	}
	return &types.RepTierConfig{OrgID: stored.OrgId, Tiers: tiers}, nil
}

func (k *Keeper) SetRepTiers(ctx context.Context, orgID string, tiers []*types.RepTierRecord) error {
	store := k.getStore(ctx)
	storedTiers := make([]*types.RepTier, len(tiers))
	for i, t := range tiers {
		storedTiers[i] = &types.RepTier{
			MinReputation:            t.MinReputation,
			MaxReputation:            t.MaxReputation,
			MaxContributionsPerEpoch: t.MaxContributionsPerEpoch,
			PayoutPerMemory:          t.PayoutPerMemory,
		}
	}
	bz, err := proto.Marshal(&types.StoredRepTierConfig{OrgId: orgID, Tiers: storedTiers})
	if err != nil {
		return fmt.Errorf("marshal rep tier config: %w", err)
	}
	return store.Set(repTierKey(orgID), bz)
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
			ServeAttestationRequired: false,
			ContestStakeVibe:         0,
			MinContributionsPerEpoch: 0,
		}, nil
	}

	var stored types.StoredOrgConfig
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal org config: %w", err)
	}
	return &types.OrgConfig{
		OrgID:                    stored.OrgId,
		ServeAttestationRequired: stored.ServeAttestationRequired,
		ContestStakeVibe:         stored.ContestStakeVibe,
		MinContributionsPerEpoch: stored.MinContributionsPerEpoch,
	}, nil
}

func (k *Keeper) SetOrgConfig(ctx context.Context, orgID string, cfg *types.OrgConfig) error {
	store := k.getStore(ctx)
	bz, err := proto.Marshal(&types.StoredOrgConfig{
		OrgId:                    cfg.OrgID,
		ServeAttestationRequired: cfg.ServeAttestationRequired,
		ContestStakeVibe:         cfg.ContestStakeVibe,
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
	if err := iter.Error(); err != nil {
		return nil, err
	}
	return orgs, nil
}

func (k *Keeper) InitGenesis(ctx context.Context, state *types.GenesisState) error {
	store := k.getStore(ctx)

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

	if state.DynamicPrice != nil {
		bz, err := proto.Marshal(dynamicPriceToStored(state.DynamicPrice))
		if err != nil {
			return err
		}
		if err := store.Set(dynamicPriceKey(), bz); err != nil {
			return err
		}
	}

	for _, treasury := range state.Treasuries {
		bz, err := proto.Marshal(treasuryToStored(treasury))
		if err != nil {
			return err
		}
		if err := store.Set(treasuryKey(treasury.OrgID), bz); err != nil {
			return err
		}
	}

	for _, repTier := range state.RepTiers {
		storedTiers := make([]*types.RepTier, len(repTier.Tiers))
		for i, t := range repTier.Tiers {
			storedTiers[i] = &types.RepTier{
				MinReputation:            t.MinReputation,
				MaxReputation:            t.MaxReputation,
				MaxContributionsPerEpoch: t.MaxContributionsPerEpoch,
				PayoutPerMemory:          t.PayoutPerMemory,
			}
		}
		bz, err := proto.Marshal(&types.StoredRepTierConfig{OrgId: repTier.OrgID, Tiers: storedTiers})
		if err != nil {
			return err
		}
		if err := store.Set(repTierKey(repTier.OrgID), bz); err != nil {
			return err
		}
	}

	for _, orgConfig := range state.OrgConfigs {
		bz, err := proto.Marshal(&types.StoredOrgConfig{
			OrgId:                    orgConfig.OrgID,
			ServeAttestationRequired: orgConfig.ServeAttestationRequired,
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
	if err := orgIter.Error(); err != nil {
		return nil, err
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
	if err := memberIter.Error(); err != nil {
		return nil, err
	}

	var dynamicPrice *types.DynamicPrice
	dpBz, err := store.Get(dynamicPriceKey())
	if err != nil {
		return nil, err
	}
	if dpBz != nil {
		var stored types.StoredDynamicPrice
		if err := proto.Unmarshal(dpBz, &stored); err == nil {
			dp := storedToDynamicPrice(stored)
			dynamicPrice = &dp
		}
	}

	treasuryPrefix := []byte("treasury/")
	treasuryIter, err := store.Iterator(treasuryPrefix, storetypes.PrefixEndBytes(treasuryPrefix))
	if err != nil {
		return nil, err
	}
	defer treasuryIter.Close()

	var treasuries []*types.Treasury
	for ; treasuryIter.Valid(); treasuryIter.Next() {
		var stored types.StoredTreasury
		if err := proto.Unmarshal(treasuryIter.Value(), &stored); err != nil {
			continue
		}
		treasury := storedToTreasury(stored)
		treasuries = append(treasuries, &treasury)
	}

	reptierPrefix := []byte("reptier/")
	reptierIter, err := store.Iterator(reptierPrefix, storetypes.PrefixEndBytes(reptierPrefix))
	if err != nil {
		return nil, err
	}
	defer reptierIter.Close()

	var repTiers []*types.RepTierConfig
	for ; reptierIter.Valid(); reptierIter.Next() {
		var stored types.StoredRepTierConfig
		if err := proto.Unmarshal(reptierIter.Value(), &stored); err != nil {
			continue
		}
		tiers := make([]*types.RepTierRecord, len(stored.Tiers))
		for i, t := range stored.Tiers {
			tiers[i] = &types.RepTierRecord{
				MinReputation:            t.MinReputation,
				MaxReputation:            t.MaxReputation,
				MaxContributionsPerEpoch: t.MaxContributionsPerEpoch,
				PayoutPerMemory:          t.PayoutPerMemory,
			}
		}
		repTiers = append(repTiers, &types.RepTierConfig{OrgID: stored.OrgId, Tiers: tiers})
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
			ServeAttestationRequired: stored.ServeAttestationRequired,
			ContestStakeVibe:         stored.ContestStakeVibe,
			MinContributionsPerEpoch: stored.MinContributionsPerEpoch,
		})
	}

	return &types.GenesisState{
		Orgs:         orgs,
		Members:      members,
		DynamicPrice: dynamicPrice,
		Treasuries:   treasuries,
		RepTiers:     repTiers,
		OrgConfigs:   orgConfigs,
	}, nil
}

func (k *Keeper) Logger(ctx context.Context) log.Logger {
	return k.logger.With("module", "x/org")
}
