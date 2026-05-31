package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/gogoproto/proto"
	"github.com/wevibe-network/wevibe-chain/x/reputation/types"
)

type Keeper struct {
	storeService store.KVStoreService
	logger       log.Logger
	authority    string
	serveKeeper  types.ServeKeeper
	memoryKeeper types.MemoryKeeper
}

func NewKeeper(storeService store.KVStoreService, logger log.Logger, authority string) *Keeper {
	return &Keeper{
		storeService: storeService,
		logger:       logger,
		authority:    authority,
	}
}

func (k *Keeper) SetServeKeeper(sk types.ServeKeeper)   { k.serveKeeper = sk }
func (k *Keeper) SetMemoryKeeper(mk types.MemoryKeeper) { k.memoryKeeper = mk }

func (k *Keeper) getStore(ctx context.Context) store.KVStore {
	return k.storeService.OpenKVStore(ctx)
}

var (
	prefixActive  = []byte("active/")
	prefixStats   = []byte("stats/")
	prefixMemory  = []byte("memory/")
	prefixOrgSet  = []byte("orgset/")
	prefixProfile = []byte("profile/")
)

func activeKey() []byte {
	return prefixActive
}

func statsKey(developer []byte) []byte {
	return append(prefixStats, developer...)
}

func memoryKey(developer []byte, memoryCID string) []byte {
	return append(append(prefixMemory, developer...), []byte("/"+memoryCID)...)
}

func memoryPrefix(developer []byte) []byte {
	return append(prefixMemory, developer...)
}

func contributorOrgSetKey(developer []byte) []byte {
	return append(prefixOrgSet, developer...)
}

func contributorProfileKey(contributorID, orgID string) []byte {
	return append(prefixProfile, []byte(contributorID+"/"+orgID)...)
}

func reputationStatsToStored(stats *types.ReputationStats) *types.StoredReputationStats {
	difficultyBucket := make([]uint64, len(stats.DifficultyBucket))
	copy(difficultyBucket, stats.DifficultyBucket[:])
	return &types.StoredReputationStats{
		DeveloperId:         stats.DeveloperID,
		MemoryCount:         stats.MemoryCount,
		DifficultyBucket:    difficultyBucket,
		DomainTags:          stats.DomainTags,
		ProvenanceBreakdown: stats.ProvenanceBreakdown,
		Xp:                  stats.XP,
		ServeCount:          stats.ServeCount,
		SelfServeCount:      stats.SelfServeCount,
		OrgBreadth:          stats.OrgBreadth,
		FirstSeenEpoch:      stats.FirstSeenEpoch,
		ServeXp:             stats.ServeXP,
	}
}

func storedToReputationStats(stored *types.StoredReputationStats) *types.ReputationStats {
	var bucket [11]uint64
	copy(bucket[:], stored.DifficultyBucket)
	return &types.ReputationStats{
		DeveloperID:         stored.DeveloperId,
		MemoryCount:         stored.MemoryCount,
		DifficultyBucket:    bucket,
		DomainTags:          stored.DomainTags,
		ProvenanceBreakdown: stored.ProvenanceBreakdown,
		XP:                  stored.Xp,
		ServeCount:          stored.ServeCount,
		SelfServeCount:      stored.SelfServeCount,
		OrgBreadth:          stored.OrgBreadth,
		FirstSeenEpoch:      stored.FirstSeenEpoch,
		ServeXP:             stored.ServeXp,
	}
}

func attestedMemoryToStored(memory *types.AttestedMemory) *types.StoredAttestedMemory {
	return &types.StoredAttestedMemory{
		Developer:  memory.Developer,
		MemoryCid:  memory.MemoryCID,
		Difficulty: uint32(memory.Difficulty),
		Quality:    uint32(memory.Quality),
		DomainTags: memory.DomainTags,
		Provenance: memory.Provenance,
	}
}

func storedToAttestedMemory(stored *types.StoredAttestedMemory) *types.AttestedMemory {
	return &types.AttestedMemory{
		Developer:  stored.Developer,
		MemoryCID:  stored.MemoryCid,
		Difficulty: uint8(stored.Difficulty),
		Quality:    uint8(stored.Quality),
		DomainTags: stored.DomainTags,
		Provenance: stored.Provenance,
	}
}

const ParamsKey = "params"

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

func (k *Keeper) Activate(ctx context.Context) {
	store := k.getStore(ctx)
	store.Set(activeKey(), []byte{0x01})
	k.logger.Info("reputation module activated")
}

func (k *Keeper) Deactivate(ctx context.Context) {
	store := k.getStore(ctx)
	store.Set(activeKey(), []byte{0x00})
	k.logger.Info("reputation module deactivated")
}

func (k *Keeper) IsActive(ctx context.Context) bool {
	store := k.getStore(ctx)
	bz, err := store.Get(activeKey())
	if err != nil {
		return false
	}
	if bz == nil {
		return false
	}
	return len(bz) > 0 && bz[0] == 0x01
}

func (k *Keeper) IsActivated(ctx context.Context) bool {
	return k.IsActive(ctx)
}

func (k *Keeper) UpdateReputation(ctx context.Context, developer []byte, memory *types.AttestedMemory) error {
	if !k.IsActive(ctx) {
		return types.ErrReputationNotActive
	}

	if err := memory.Validate(); err != nil {
		return err
	}

	store := k.getStore(ctx)
	devStr := string(developer)

	statsKey := statsKey(developer)
	bz, err := store.Get(statsKey)
	if err != nil {
		return err
	}
	var stats *types.ReputationStats
	if bz != nil {
		var stored types.StoredReputationStats
		if err := proto.Unmarshal(bz, &stored); err != nil {
			return fmt.Errorf("unmarshal stats: %w", err)
		}
		stats = storedToReputationStats(&stored)
	} else {
		stats = types.NewReputationStats(devStr)
	}

	stats.AddMemory(memory.Difficulty, memory.Quality, memory.DomainTags, memory.Provenance)

	bz, err = proto.Marshal(reputationStatsToStored(stats))
	if err != nil {
		return fmt.Errorf("marshal stats: %w", err)
	}
	if err := store.Set(statsKey, bz); err != nil {
		return err
	}

	memKey := memoryKey(developer, memory.MemoryCID)
	memBz, err := proto.Marshal(attestedMemoryToStored(memory))
	if err != nil {
		return fmt.Errorf("marshal memory: %w", err)
	}
	if err := store.Set(memKey, memBz); err != nil {
		return err
	}

	k.logger.Info("reputation updated",
		"developer", devStr,
		"memory_cid", memory.MemoryCID,
		"difficulty", memory.Difficulty,
		"quality", memory.Quality,
		"provenance", memory.Provenance,
	)

	return nil
}

func (k *Keeper) GetReputation(ctx context.Context, developer []byte) (*types.ReputationStats, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(statsKey(developer))
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrNoStats
	}

	var stored types.StoredReputationStats
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal stats: %w", err)
	}
	return storedToReputationStats(&stored), nil
}

func (k *Keeper) AddMemory(ctx context.Context, developer []byte, memory *types.AttestedMemory) error {
	return k.UpdateReputation(ctx, developer, memory)
}

func (k *Keeper) GetDifficultyHistogram(ctx context.Context, developer []byte) (*types.DifficultyHistogram, error) {
	stats, err := k.GetReputation(ctx, developer)
	if err != nil {
		return nil, err
	}

	return types.NewDifficultyHistogram(developer, stats.DifficultyBucket), nil
}

func (k *Keeper) GetDomainExpertise(ctx context.Context, developer []byte) (*types.DomainExpertise, error) {
	stats, err := k.GetReputation(ctx, developer)
	if err != nil {
		return nil, err
	}

	return types.NewDomainExpertise(developer, stats.DomainTags), nil
}

func (k *Keeper) GetXP(ctx context.Context, developer []byte) (uint64, error) {
	stats, err := k.GetReputation(ctx, developer)
	if err != nil {
		return 0, err
	}
	return stats.XP, nil
}

func (k *Keeper) GetMemoryCount(ctx context.Context, developer []byte) (uint64, error) {
	stats, err := k.GetReputation(ctx, developer)
	if err != nil {
		return 0, err
	}
	return stats.MemoryCount, nil
}

func (k *Keeper) GetProvenanceStats(ctx context.Context, developer []byte) (*types.ProvenanceStats, error) {
	stats, err := k.GetReputation(ctx, developer)
	if err != nil {
		return nil, err
	}

	return types.NewProvenanceStats(developer, stats.ProvenanceBreakdown), nil
}

func (k *Keeper) GetDeveloperMemories(ctx context.Context, developer []byte) ([]*types.AttestedMemory, error) {
	store := k.getStore(ctx)
	prefix := memoryPrefix(developer)
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var memories []*types.AttestedMemory
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredAttestedMemory
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		memories = append(memories, storedToAttestedMemory(&stored))
	}

	return memories, nil
}

func (k *Keeper) HasDeveloper(ctx context.Context, developer []byte) (bool, error) {
	store := k.getStore(ctx)
	has, err := store.Has(statsKey(developer))
	if err != nil {
		return false, err
	}
	return has, nil
}

func (k *Keeper) GetAllDevelopers(ctx context.Context) ([][]byte, error) {
	store := k.getStore(ctx)
	iter, err := store.Iterator(prefixStats, storetypes.PrefixEndBytes(prefixStats))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var developers [][]byte
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		dev := key[len(prefixStats):]
		developers = append(developers, dev)
	}
	return developers, nil
}

func (k *Keeper) Logger(ctx context.Context) log.Logger {
	return k.logger.With("module", "x/reputation")
}

func (k *Keeper) GetReputationStatsMap(ctx context.Context) (map[string]*types.ReputationStats, error) {
	store := k.getStore(ctx)
	iter, err := store.Iterator(prefixStats, storetypes.PrefixEndBytes(prefixStats))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	result := make(map[string]*types.ReputationStats)
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredReputationStats
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		stats := storedToReputationStats(&stored)
		result[stats.DeveloperID] = stats
	}
	return result, nil
}

func (k *Keeper) InitFromMap(ctx context.Context, statsMap map[string]*types.ReputationStats) error {
	for dev, stats := range statsMap {
		if err := stats.Validate(); err != nil {
			return err
		}
		store := k.getStore(ctx)
		bz, err := proto.Marshal(reputationStatsToStored(stats))
		if err != nil {
			return err
		}
		if err := store.Set(statsKey([]byte(dev)), bz); err != nil {
			return err
		}
	}
	return nil
}

func developerKey(developer []byte) string {
	return string(developer)
}

func (k *Keeper) ValidateAttestedMemory(ctx context.Context, memory *types.AttestedMemory) error {
	if !k.IsActive(ctx) {
		return types.ErrReputationNotActive
	}
	return memory.Validate()
}

func (k *Keeper) GetTopDomains(ctx context.Context, developer []byte, limit int) ([]string, error) {
	stats, err := k.GetReputation(ctx, developer)
	if err != nil {
		return nil, err
	}

	type domainCount struct {
		domain string
		count  uint64
	}

	domains := make([]domainCount, 0, len(stats.DomainTags))
	for domain, count := range stats.DomainTags {
		domains = append(domains, domainCount{domain, count})
	}

	if len(domains) <= limit {
		result := make([]string, len(domains))
		for i, d := range domains {
			result[i] = d.domain
		}
		return result, nil
	}

	for i := 0; i < len(domains)-1; i++ {
		for j := i + 1; j < len(domains); j++ {
			if domains[j].count > domains[i].count {
				domains[i], domains[j] = domains[j], domains[i]
			}
		}
	}

	result := make([]string, limit)
	for i := 0; i < limit; i++ {
		result[i] = domains[i].domain
	}

	return result, nil
}

func (k *Keeper) GetAverageDifficulty(ctx context.Context, developer []byte) (float64, error) {
	stats, err := k.GetReputation(ctx, developer)
	if err != nil {
		return 0, err
	}

	if stats.MemoryCount == 0 {
		return 0, nil
	}

	var weightedSum float64
	for i, count := range stats.DifficultyBucket {
		weightedSum += float64(i) * float64(count)
	}

	return weightedSum / float64(stats.MemoryCount), nil
}

func (k *Keeper) RecordServe(ctx context.Context, contributorWallet []byte, orgID string, epoch uint64, isSelfServe bool) error {
	if !k.IsActive(ctx) {
		return types.ErrReputationNotActive
	}

	store := k.getStore(ctx)
	devStr := string(contributorWallet)

	statsKey := statsKey(contributorWallet)
	bz, err := store.Get(statsKey)
	if err != nil {
		return err
	}
	var stats *types.ReputationStats
	if bz != nil {
		var stored types.StoredReputationStats
		if err := proto.Unmarshal(bz, &stored); err != nil {
			return fmt.Errorf("unmarshal stats: %w", err)
		}
		stats = storedToReputationStats(&stored)
	} else {
		stats = types.NewReputationStats(devStr)
	}

	stats.ServeCount++
	if isSelfServe {
		stats.SelfServeCount++
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if isSelfServe {
		stats.ServeXP += params.SelfServeXpPerServe
	} else {
		stats.ServeXP += params.ServeXpPerServe
	}

	if stats.FirstSeenEpoch == 0 {
		stats.FirstSeenEpoch = epoch
	}

	orgSetKey := contributorOrgSetKey(contributorWallet)
	orgSetBz, err := store.Get(orgSetKey)
	var orgSet *types.StoredContributorOrgSet
	if orgSetBz != nil {
		var stored types.StoredContributorOrgSet
		if err := proto.Unmarshal(orgSetBz, &stored); err != nil {
			return fmt.Errorf("unmarshal org set: %w", err)
		}
		orgSet = &stored
	} else {
		orgSet = &types.StoredContributorOrgSet{
			ContributorId: devStr,
			OrgIds:        make([]string, 0),
		}
	}

	orgFound := false
	for _, existing := range orgSet.OrgIds {
		if existing == orgID {
			orgFound = true
			break
		}
	}
	if !orgFound {
		orgSet.OrgIds = append(orgSet.OrgIds, orgID)
		stats.OrgBreadth++
	}

	orgSetBz, err = proto.Marshal(orgSet)
	if err != nil {
		return fmt.Errorf("marshal org set: %w", err)
	}
	if err := store.Set(orgSetKey, orgSetBz); err != nil {
		return err
	}

	bz, err = proto.Marshal(reputationStatsToStored(stats))
	if err != nil {
		return fmt.Errorf("marshal stats: %w", err)
	}
	if err := store.Set(statsKey, bz); err != nil {
		return err
	}

	k.logger.Info("serve recorded",
		"developer", devStr,
		"org_id", orgID,
		"epoch", epoch,
		"is_self_serve", isSelfServe,
		"serve_count", stats.ServeCount,
		"org_breadth", stats.OrgBreadth,
	)

	return nil
}

func (k *Keeper) GetServeStats(ctx context.Context, developer []byte) (*types.ReputationStats, error) {
	stats, err := k.GetReputation(ctx, developer)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (k *Keeper) GetContributorOrgSet(ctx context.Context, developer []byte) (*types.StoredContributorOrgSet, error) {
	store := k.getStore(ctx)
	orgSetKey := contributorOrgSetKey(developer)
	bz, err := store.Get(orgSetKey)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return &types.StoredContributorOrgSet{
			ContributorId: string(developer),
			OrgIds:        make([]string, 0),
		}, nil
	}
	var orgSet types.StoredContributorOrgSet
	if err := proto.Unmarshal(bz, &orgSet); err != nil {
		return nil, fmt.Errorf("unmarshal org set: %w", err)
	}
	return &orgSet, nil
}

func (k *Keeper) GetCrossOrgProfile(ctx context.Context, developer []byte) (*types.StoredContributorOrgSet, error) {
	return k.GetContributorOrgSet(ctx, developer)
}

func (k *Keeper) InitGenesisState(ctx context.Context, state *types.GenesisState) error {
	store := k.getStore(ctx)
	if state.Active {
		store.Set(activeKey(), []byte{0x01})
		k.Activate(ctx)
	}

	for _, stats := range state.Stats {
		bz, err := proto.Marshal(reputationStatsToStored(stats))
		if err != nil {
			return err
		}
		if err := store.Set(statsKey([]byte(stats.DeveloperID)), bz); err != nil {
			return err
		}
	}

	for _, orgSet := range state.ContributorOrgSets {
		bz, err := proto.Marshal(orgSet)
		if err != nil {
			return err
		}
		if err := store.Set(contributorOrgSetKey([]byte(orgSet.ContributorId)), bz); err != nil {
			return err
		}
	}

	return nil
}

func (k *Keeper) ExportGenesisState(ctx context.Context) (*types.GenesisState, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(activeKey())
	if err != nil {
		return nil, err
	}
	active := bz != nil && len(bz) > 0 && bz[0] == 0x01

	statsMap := make(map[string]*types.ReputationStats)
	iter, err := store.Iterator(prefixStats, storetypes.PrefixEndBytes(prefixStats))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var stored types.StoredReputationStats
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		stats := storedToReputationStats(&stored)
		statsMap[stats.DeveloperID] = stats
	}

	var statsList []*types.ReputationStats
	for _, stats := range statsMap {
		statsList = append(statsList, stats)
	}

	var orgSets []*types.StoredContributorOrgSet
	iter2, err := store.Iterator(prefixOrgSet, storetypes.PrefixEndBytes(prefixOrgSet))
	if err != nil {
		return nil, err
	}
	defer iter2.Close()

	for ; iter2.Valid(); iter2.Next() {
		var orgSet types.StoredContributorOrgSet
		if err := proto.Unmarshal(iter2.Value(), &orgSet); err != nil {
			continue
		}
		orgSets = append(orgSets, &orgSet)
	}

	return &types.GenesisState{
		Active:             active,
		Stats:              statsList,
		ContributorOrgSets: orgSets,
	}, nil
}

func (k *Keeper) GetContributorProfile(ctx context.Context, contributorID, orgID string) (*types.StoredContributorProfile, error) {
	store := k.getStore(ctx)
	key := contributorProfileKey(contributorID, orgID)
	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return &types.StoredContributorProfile{
			ContributorId: contributorID,
		}, nil
	}
	var profile types.StoredContributorProfile
	if err := proto.Unmarshal(bz, &profile); err != nil {
		return nil, fmt.Errorf("unmarshal profile: %w", err)
	}
	return &profile, nil
}

func (k *Keeper) IncrementContribution(ctx context.Context, contributorID, orgID, memoryCID string) error {
	store := k.getStore(ctx)
	key := contributorProfileKey(contributorID, orgID)
	bz, err := store.Get(key)
	var profile types.StoredContributorProfile
	if err != nil {
		return err
	}
	if bz != nil {
		if err := proto.Unmarshal(bz, &profile); err != nil {
			return fmt.Errorf("unmarshal profile: %w", err)
		}
	}
	if profile.ContributorId == "" {
		profile.ContributorId = contributorID
	}
	profile.TotalApprovedMemories++
	bz, err = proto.Marshal(&profile)
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}
	return store.Set(key, bz)
}

func (k *Keeper) IncrementServe(ctx context.Context, contributorID, orgID, memoryCID string, count uint64) error {
	store := k.getStore(ctx)
	key := contributorProfileKey(contributorID, orgID)
	bz, err := store.Get(key)
	var profile types.StoredContributorProfile
	if err != nil {
		return err
	}
	if bz != nil {
		if err := proto.Unmarshal(bz, &profile); err != nil {
			return fmt.Errorf("unmarshal profile: %w", err)
		}
	}
	if profile.ContributorId == "" {
		profile.ContributorId = contributorID
	}
	profile.TotalServesReceived += count
	bz, err = proto.Marshal(&profile)
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}
	return store.Set(key, bz)
}

func (k *Keeper) RecordBan(ctx context.Context, contributorID, orgID, memoryCID string) error {
	store := k.getStore(ctx)
	key := contributorProfileKey(contributorID, orgID)
	bz, err := store.Get(key)
	var profile types.StoredContributorProfile
	if err != nil {
		return err
	}
	if bz != nil {
		if err := proto.Unmarshal(bz, &profile); err != nil {
			return fmt.Errorf("unmarshal profile: %w", err)
		}
	}
	if profile.ContributorId == "" {
		profile.ContributorId = contributorID
	}
	profile.TotalReportsUpheldAgainst++
	bz, err = proto.Marshal(&profile)
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}
	return store.Set(key, bz)
}

type ContributorDelta struct {
	ApprovedMemories     int64
	ServesReceived       int64
	DenialsReceived      int64
	ReportsFiledAgainst  int64
	ReportsUpheldAgainst int64
}

func (k *Keeper) UpdateContributorAggregate(ctx context.Context, contributorID string, delta ContributorDelta) error {
	store := k.getStore(ctx)
	key := contributorProfileKey(contributorID, "")
	bz, err := store.Get(key)
	var profile types.StoredContributorProfile
	if err != nil {
		return err
	}
	if bz != nil {
		if err := proto.Unmarshal(bz, &profile); err != nil {
			return fmt.Errorf("unmarshal profile: %w", err)
		}
	}
	if profile.ContributorId == "" {
		profile.ContributorId = contributorID
	}
	profile.TotalApprovedMemories = uint64(int64(profile.TotalApprovedMemories) + delta.ApprovedMemories)
	profile.TotalServesReceived = uint64(int64(profile.TotalServesReceived) + delta.ServesReceived)
	profile.TotalDenialsReceived = uint64(int64(profile.TotalDenialsReceived) + delta.DenialsReceived)
	profile.TotalReportsFiledAgainst = uint64(int64(profile.TotalReportsFiledAgainst) + delta.ReportsFiledAgainst)
	profile.TotalReportsUpheldAgainst = uint64(int64(profile.TotalReportsUpheldAgainst) + delta.ReportsUpheldAgainst)
	bz, err = proto.Marshal(&profile)
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}
	return store.Set(key, bz)
}

func (k *Keeper) AddContributorOrgMembership(ctx context.Context, contributorID, orgID, role string, joinedEpoch uint64) error {
	store := k.getStore(ctx)
	key := contributorProfileKey(contributorID, "")
	bz, err := store.Get(key)
	var profile types.StoredContributorProfile
	if err != nil {
		return err
	}
	if bz != nil {
		if err := proto.Unmarshal(bz, &profile); err != nil {
			return fmt.Errorf("unmarshal profile: %w", err)
		}
	}
	if profile.ContributorId == "" {
		profile.ContributorId = contributorID
	}
	for _, m := range profile.Memberships {
		if m.OrgId == orgID {
			m.Role = role
			m.Active = true
			bz, err = proto.Marshal(&profile)
			if err != nil {
				return fmt.Errorf("marshal profile: %w", err)
			}
			return store.Set(key, bz)
		}
	}
	profile.Memberships = append(profile.Memberships, &types.OrgMembership{
		OrgId:       orgID,
		Role:        role,
		JoinedEpoch: joinedEpoch,
		Active:      true,
	})
	bz, err = proto.Marshal(&profile)
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}
	return store.Set(key, bz)
}

func (k *Keeper) UpdateContributorOrgRole(ctx context.Context, contributorID, orgID, newRole string) error {
	store := k.getStore(ctx)
	key := contributorProfileKey(contributorID, "")
	bz, err := store.Get(key)
	var profile types.StoredContributorProfile
	if err != nil {
		return err
	}
	if bz == nil {
		return fmt.Errorf("contributor profile not found")
	}
	if err := proto.Unmarshal(bz, &profile); err != nil {
		return fmt.Errorf("unmarshal profile: %w", err)
	}
	for _, m := range profile.Memberships {
		if m.OrgId == orgID {
			m.Role = newRole
			bz, err = proto.Marshal(&profile)
			if err != nil {
				return fmt.Errorf("marshal profile: %w", err)
			}
			return store.Set(key, bz)
		}
	}
	return fmt.Errorf("membership not found for org %s", orgID)
}

func (k *Keeper) MarkContributorOrgInactive(ctx context.Context, contributorID, orgID string) error {
	store := k.getStore(ctx)
	key := contributorProfileKey(contributorID, "")
	bz, err := store.Get(key)
	var profile types.StoredContributorProfile
	if err != nil {
		return err
	}
	if bz == nil {
		return fmt.Errorf("contributor profile not found")
	}
	if err := proto.Unmarshal(bz, &profile); err != nil {
		return fmt.Errorf("unmarshal profile: %w", err)
	}
	for _, m := range profile.Memberships {
		if m.OrgId == orgID {
			m.Active = false
			bz, err = proto.Marshal(&profile)
			if err != nil {
				return fmt.Errorf("marshal profile: %w", err)
			}
			return store.Set(key, bz)
		}
	}
	return fmt.Errorf("membership not found for org %s", orgID)
}

func (k *Keeper) RecordContributorActivity(ctx context.Context, contributorID string, epoch uint64) error {
	store := k.getStore(ctx)
	key := contributorProfileKey(contributorID, "")
	bz, err := store.Get(key)
	var profile types.StoredContributorProfile
	if err != nil {
		return err
	}
	if bz != nil {
		if err := proto.Unmarshal(bz, &profile); err != nil {
			return fmt.Errorf("unmarshal profile: %w", err)
		}
	}
	if profile.ContributorId == "" {
		profile.ContributorId = contributorID
	}
	if profile.FirstContributionEpoch == 0 {
		profile.FirstContributionEpoch = epoch
	}
	profile.LastContributionEpoch = epoch
	bz, err = proto.Marshal(&profile)
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}
	return store.Set(key, bz)
}
