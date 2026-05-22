package keeper_test

import (
	"context"
	"testing"

	storetypes "cosmossdk.io/store/types"

	testkeeper "github.com/wevibe-network/wevibe-chain/testutil/keeper"
	"github.com/wevibe-network/wevibe-chain/x/reputation/keeper"
	"github.com/wevibe-network/wevibe-chain/x/reputation/types"
)

var (
	reputationStoreKey = storetypes.NewKVStoreKey("reputation")
)

func newTestKeeper(t *testing.T) (*keeper.Keeper, context.Context) {
	storeService, _ := testkeeper.NewTestStoreService(t, reputationStoreKey)
	logger := testkeeper.NewTestLogger()
	k := keeper.NewKeeper(storeService, logger, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry")
	ctx := context.Background()
	return k, ctx
}

func TestInitGenesis_Inactive(t *testing.T) {
	k, ctx := newTestKeeper(t)

	err := k.InitGenesisState(ctx, &types.GenesisState{Active: false})
	if err != nil {
		t.Fatalf("InitGenesisState failed: %v", err)
	}
	if k.IsActive(ctx) {
		t.Fatal("expected inactive")
	}
}

func TestInitGenesis_Active(t *testing.T) {
	k, ctx := newTestKeeper(t)

	err := k.InitGenesisState(ctx, &types.GenesisState{Active: true})
	if err != nil {
		t.Fatalf("InitGenesisState failed: %v", err)
	}
	if !k.IsActive(ctx) {
		t.Fatal("expected active")
	}
}

func TestActivate_Deactivate(t *testing.T) {
	k, ctx := newTestKeeper(t)

	k.Activate(ctx)
	if !k.IsActive(ctx) {
		t.Fatal("expected active after Activate")
	}

	k.Deactivate(ctx)
	if k.IsActive(ctx) {
		t.Fatal("expected inactive after Deactivate")
	}
}

func TestUpdateReputation_NotActive(t *testing.T) {
	k, ctx := newTestKeeper(t)

	memory := types.NewAttestedMemory([]byte("dev1"), "cid123", 5, 7, []string{"golang", "cosmos"}, "commitllm")
	err := k.UpdateReputation(ctx, []byte("dev1"), memory)
	if err != types.ErrReputationNotActive {
		t.Fatalf("expected ErrReputationNotActive, got: %v", err)
	}
}

func TestUpdateReputation_Success(t *testing.T) {
	k, ctx := newTestKeeper(t)

	k.Activate(ctx)

	memory := types.NewAttestedMemory([]byte("dev1"), "cid123", 5, 7, []string{"golang", "cosmos"}, "commitllm")
	err := k.UpdateReputation(ctx, []byte("dev1"), memory)
	if err != nil {
		t.Fatalf("UpdateReputation failed: %v", err)
	}

	stats, err := k.GetReputation(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetReputation failed: %v", err)
	}
	if stats.MemoryCount != 1 {
		t.Fatalf("expected MemoryCount 1, got: %d", stats.MemoryCount)
	}
	if stats.XP != 35 {
		t.Fatalf("expected XP 35, got: %d", stats.XP)
	}
}

func TestUpdateReputation_MultipleMemories(t *testing.T) {
	k, ctx := newTestKeeper(t)

	k.Activate(ctx)

	mem1 := types.NewAttestedMemory([]byte("dev1"), "cid1", 3, 5, []string{"rust"}, "commitllm")
	mem2 := types.NewAttestedMemory([]byte("dev1"), "cid2", 7, 9, []string{"golang"}, "proxy-attested")
	mem3 := types.NewAttestedMemory([]byte("dev1"), "cid3", 5, 8, []string{"golang", "rust"}, "unattested")

	err := k.UpdateReputation(ctx, []byte("dev1"), mem1)
	if err != nil {
		t.Fatalf("UpdateReputation 1 failed: %v", err)
	}
	err = k.UpdateReputation(ctx, []byte("dev1"), mem2)
	if err != nil {
		t.Fatalf("UpdateReputation 2 failed: %v", err)
	}
	err = k.UpdateReputation(ctx, []byte("dev1"), mem3)
	if err != nil {
		t.Fatalf("UpdateReputation 3 failed: %v", err)
	}

	stats, err := k.GetReputation(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetReputation failed: %v", err)
	}
	if stats.MemoryCount != 3 {
		t.Fatalf("expected MemoryCount 3, got: %d", stats.MemoryCount)
	}
	if stats.XP != 15+63+40 {
		t.Fatalf("expected XP %d, got: %d", 15+63+40, stats.XP)
	}
	if stats.DifficultyBucket[3] != 1 {
		t.Fatalf("expected DifficultyBucket[3]=1, got: %d", stats.DifficultyBucket[3])
	}
	if stats.DifficultyBucket[7] != 1 {
		t.Fatalf("expected DifficultyBucket[7]=1, got: %d", stats.DifficultyBucket[7])
	}
	if stats.DifficultyBucket[5] != 1 {
		t.Fatalf("expected DifficultyBucket[5]=1, got: %d", stats.DifficultyBucket[5])
	}
}

func TestGetReputation_NotFound(t *testing.T) {
	k, ctx := newTestKeeper(t)

	_, err := k.GetReputation(ctx, []byte("unknown"))
	if err != types.ErrNoStats {
		t.Fatalf("expected ErrNoStats, got: %v", err)
	}
}

func TestAddMemory(t *testing.T) {
	k, ctx := newTestKeeper(t)

	k.Activate(ctx)

	memory := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, []string{"golang"}, "commitllm")
	err := k.AddMemory(ctx, []byte("dev1"), memory)
	if err != nil {
		t.Fatalf("AddMemory failed: %v", err)
	}

	stats, err := k.GetReputation(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetReputation failed: %v", err)
	}
	if stats.MemoryCount != 1 {
		t.Fatalf("expected MemoryCount 1, got: %d", stats.MemoryCount)
	}
}

func TestGetDifficultyHistogram(t *testing.T) {
	k, ctx := newTestKeeper(t)

	k.Activate(ctx)

	mem1 := types.NewAttestedMemory([]byte("dev1"), "cid1", 3, 5, []string{}, "commitllm")
	mem2 := types.NewAttestedMemory([]byte("dev1"), "cid2", 7, 9, []string{}, "proxy-attested")
	mem3 := types.NewAttestedMemory([]byte("dev1"), "cid3", 3, 8, []string{}, "unattested")

	k.AddMemory(ctx, []byte("dev1"), mem1)
	k.AddMemory(ctx, []byte("dev1"), mem2)
	k.AddMemory(ctx, []byte("dev1"), mem3)

	histogram, err := k.GetDifficultyHistogram(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetDifficultyHistogram failed: %v", err)
	}
	if histogram.Buckets[3] != 2 {
		t.Fatalf("expected Buckets[3]=2, got: %d", histogram.Buckets[3])
	}
	if histogram.Buckets[7] != 1 {
		t.Fatalf("expected Buckets[7]=1, got: %d", histogram.Buckets[7])
	}
	if histogram.TotalCount != 3 {
		t.Fatalf("expected TotalCount=3, got: %d", histogram.TotalCount)
	}
}

func TestGetDomainExpertise(t *testing.T) {
	k, ctx := newTestKeeper(t)

	k.Activate(ctx)

	mem1 := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, []string{"golang", "cosmos"}, "commitllm")
	mem2 := types.NewAttestedMemory([]byte("dev1"), "cid2", 7, 9, []string{"golang", "rust"}, "proxy-attested")

	k.AddMemory(ctx, []byte("dev1"), mem1)
	k.AddMemory(ctx, []byte("dev1"), mem2)

	expertise, err := k.GetDomainExpertise(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetDomainExpertise failed: %v", err)
	}
	if expertise.DomainTags["golang"] != 2 {
		t.Fatalf("expected DomainTags[golang]=2, got: %d", expertise.DomainTags["golang"])
	}
	if expertise.DomainTags["cosmos"] != 1 {
		t.Fatalf("expected DomainTags[cosmos]=1, got: %d", expertise.DomainTags["cosmos"])
	}
	if expertise.DomainTags["rust"] != 1 {
		t.Fatalf("expected DomainTags[rust]=1, got: %d", expertise.DomainTags["rust"])
	}
	if expertise.TotalTags != 4 {
		t.Fatalf("expected TotalTags=4, got: %d", expertise.TotalTags)
	}
}

func TestGetXP(t *testing.T) {
	k, ctx := newTestKeeper(t)

	k.Activate(ctx)

	mem1 := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, []string{}, "commitllm")
	mem2 := types.NewAttestedMemory([]byte("dev1"), "cid2", 3, 10, []string{}, "proxy-attested")

	k.AddMemory(ctx, []byte("dev1"), mem1)
	k.AddMemory(ctx, []byte("dev1"), mem2)

	xp, err := k.GetXP(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetXP failed: %v", err)
	}
	if xp != 35+30 {
		t.Fatalf("expected XP %d, got: %d", 35+30, xp)
	}
}

func TestGetMemoryCount(t *testing.T) {
	k, ctx := newTestKeeper(t)

	k.Activate(ctx)

	mem1 := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, []string{}, "commitllm")
	mem2 := types.NewAttestedMemory([]byte("dev1"), "cid2", 3, 10, []string{}, "proxy-attested")

	k.AddMemory(ctx, []byte("dev1"), mem1)
	k.AddMemory(ctx, []byte("dev1"), mem2)

	count, err := k.GetMemoryCount(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetMemoryCount failed: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected MemoryCount 2, got: %d", count)
	}
}

func TestGetProvenanceStats(t *testing.T) {
	k, ctx := newTestKeeper(t)

	k.Activate(ctx)

	mem1 := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, []string{}, "commitllm")
	mem2 := types.NewAttestedMemory([]byte("dev1"), "cid2", 3, 10, []string{}, "proxy-attested")
	mem3 := types.NewAttestedMemory([]byte("dev1"), "cid3", 4, 8, []string{}, "unattested")

	k.AddMemory(ctx, []byte("dev1"), mem1)
	k.AddMemory(ctx, []byte("dev1"), mem2)
	k.AddMemory(ctx, []byte("dev1"), mem3)

	stats, err := k.GetProvenanceStats(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetProvenanceStats failed: %v", err)
	}
	if stats.Tier1Count != 1 {
		t.Fatalf("expected Tier1Count=1, got: %d", stats.Tier1Count)
	}
	if stats.Tier2Count != 1 {
		t.Fatalf("expected Tier2Count=1, got: %d", stats.Tier2Count)
	}
	if stats.UnattestedCount != 1 {
		t.Fatalf("expected UnattestedCount=1, got: %d", stats.UnattestedCount)
	}
	if stats.TotalCount != 3 {
		t.Fatalf("expected TotalCount=3, got: %d", stats.TotalCount)
	}
}

func TestGetDeveloperMemories(t *testing.T) {
	k, ctx := newTestKeeper(t)

	k.Activate(ctx)

	mem1 := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, []string{}, "commitllm")
	mem2 := types.NewAttestedMemory([]byte("dev1"), "cid2", 3, 10, []string{}, "proxy-attested")

	k.AddMemory(ctx, []byte("dev1"), mem1)
	k.AddMemory(ctx, []byte("dev1"), mem2)

	memories, err := k.GetDeveloperMemories(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetDeveloperMemories failed: %v", err)
	}
	if len(memories) != 2 {
		t.Fatalf("expected 2 memories, got: %d", len(memories))
	}
}

func TestHasDeveloper(t *testing.T) {
	k, ctx := newTestKeeper(t)

	k.Activate(ctx)

	mem := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, []string{}, "commitllm")
	k.AddMemory(ctx, []byte("dev1"), mem)

	has, err := k.HasDeveloper(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("HasDeveloper failed: %v", err)
	}
	if !has {
		t.Fatal("expected has developer")
	}

	has, err = k.HasDeveloper(ctx, []byte("unknown"))
	if err != nil {
		t.Fatalf("HasDeveloper for unknown failed: %v", err)
	}
	if has {
		t.Fatal("expected no developer for unknown")
	}
}

func TestGetAllDevelopers(t *testing.T) {
	k, ctx := newTestKeeper(t)

	k.Activate(ctx)

	mem1 := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, []string{}, "commitllm")
	mem2 := types.NewAttestedMemory([]byte("dev2"), "cid2", 3, 10, []string{}, "proxy-attested")

	k.AddMemory(ctx, []byte("dev1"), mem1)
	k.AddMemory(ctx, []byte("dev2"), mem2)

	developers, err := k.GetAllDevelopers(ctx)
	if err != nil {
		t.Fatalf("GetAllDevelopers failed: %v", err)
	}
	if len(developers) != 2 {
		t.Fatalf("expected 2 developers, got: %d", len(developers))
	}
}

func TestAttestedMemory_Validate(t *testing.T) {
	validMemory := types.NewAttestedMemory([]byte("dev1"), "cid123", 5, 7, []string{"golang"}, "commitllm")
	err := validMemory.Validate()
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	emptyDevMemory := types.NewAttestedMemory([]byte(""), "cid123", 5, 7, []string{}, "")
	err = emptyDevMemory.Validate()
	if err == nil {
		t.Fatal("expected error for empty developer")
	}

	emptyCIDMemory := types.NewAttestedMemory([]byte("dev1"), "", 5, 7, []string{}, "")
	err = emptyCIDMemory.Validate()
	if err == nil {
		t.Fatal("expected error for empty CID")
	}

	invalidDiffMemory := types.NewAttestedMemory([]byte("dev1"), "cid123", 11, 7, []string{}, "")
	err = invalidDiffMemory.Validate()
	if err == nil {
		t.Fatal("expected error for invalid difficulty")
	}

	invalidQualMemory := types.NewAttestedMemory([]byte("dev1"), "cid123", 5, 11, []string{}, "")
	err = invalidQualMemory.Validate()
	if err == nil {
		t.Fatal("expected error for invalid quality")
	}
}

func TestGetTopDomains(t *testing.T) {
	k, ctx := newTestKeeper(t)

	k.Activate(ctx)

	mem1 := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, []string{"golang", "cosmos", "sdk"}, "commitllm")
	mem2 := types.NewAttestedMemory([]byte("dev1"), "cid2", 3, 10, []string{"golang", "rust"}, "proxy-attested")

	k.AddMemory(ctx, []byte("dev1"), mem1)
	k.AddMemory(ctx, []byte("dev1"), mem2)

	top2, err := k.GetTopDomains(ctx, []byte("dev1"), 2)
	if err != nil {
		t.Fatalf("GetTopDomains failed: %v", err)
	}
	if len(top2) != 2 {
		t.Fatalf("expected 2 domains, got: %d", len(top2))
	}
}

func TestGetAverageDifficulty(t *testing.T) {
	k, ctx := newTestKeeper(t)

	k.Activate(ctx)

	mem1 := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, []string{}, "commitllm")
	mem2 := types.NewAttestedMemory([]byte("dev1"), "cid2", 7, 10, []string{}, "proxy-attested")

	k.AddMemory(ctx, []byte("dev1"), mem1)
	k.AddMemory(ctx, []byte("dev1"), mem2)

	avgDiff, err := k.GetAverageDifficulty(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetAverageDifficulty failed: %v", err)
	}
	expectedAvg := (5.0 + 7.0) / 2.0
	if avgDiff != expectedAvg {
		t.Fatalf("expected avg difficulty %f, got: %f", expectedAvg, avgDiff)
	}
}

func TestGetAverageDifficulty_NoMemories(t *testing.T) {
	k, ctx := newTestKeeper(t)

	k.Activate(ctx)

	_, err := k.GetAverageDifficulty(ctx, []byte("dev1"))
	if err == nil {
		t.Fatal("expected error for no memories")
	}
}

func TestExportGenesis_ImportGenesis(t *testing.T) {
	k1, ctx1 := newTestKeeper(t)

	k1.Activate(ctx1)

	mem1 := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, []string{"golang"}, "commitllm")
	k1.AddMemory(ctx1, []byte("dev1"), mem1)

	state, err := k1.ExportGenesisState(ctx1)
	if err != nil {
		t.Fatalf("ExportGenesisState failed: %v", err)
	}
	if !state.Active {
		t.Fatal("expected active=true from export")
	}

	k2, ctx2 := newTestKeeper(t)
	err = k2.InitGenesisState(ctx2, &types.GenesisState{Active: false})
	if err != nil {
		t.Fatalf("InitGenesisState failed: %v", err)
	}

	_, err = k2.GetReputation(ctx2, []byte("dev1"))
	if err == nil {
		t.Fatal("expected error for no stats after reset")
	}
}

func TestRecordServe_HappyPath(t *testing.T) {
	k, ctx := newTestKeeper(t)
	k.Activate(ctx)

	err := k.RecordServe(ctx, []byte("dev1"), "org1", 5, false)
	if err != nil {
		t.Fatalf("RecordServe failed: %v", err)
	}

	stats, err := k.GetReputation(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetReputation failed: %v", err)
	}
	if stats.ServeCount != 1 {
		t.Fatalf("expected ServeCount 1, got: %d", stats.ServeCount)
	}
	if stats.SelfServeCount != 0 {
		t.Fatalf("expected SelfServeCount 0, got: %d", stats.SelfServeCount)
	}
	if stats.OrgBreadth != 1 {
		t.Fatalf("expected OrgBreadth 1, got: %d", stats.OrgBreadth)
	}
	if stats.ServeXP != 5 {
		t.Fatalf("expected ServeXP 5, got: %d", stats.ServeXP)
	}
	if stats.FirstSeenEpoch != 5 {
		t.Fatalf("expected FirstSeenEpoch 5, got: %d", stats.FirstSeenEpoch)
	}
}

func TestRecordServe_MultipleServes(t *testing.T) {
	k, ctx := newTestKeeper(t)
	k.Activate(ctx)

	err := k.RecordServe(ctx, []byte("dev1"), "org1", 5, false)
	if err != nil {
		t.Fatalf("RecordServe 1 failed: %v", err)
	}
	err = k.RecordServe(ctx, []byte("dev1"), "org2", 6, false)
	if err != nil {
		t.Fatalf("RecordServe 2 failed: %v", err)
	}
	err = k.RecordServe(ctx, []byte("dev1"), "org3", 7, false)
	if err != nil {
		t.Fatalf("RecordServe 3 failed: %v", err)
	}

	stats, err := k.GetReputation(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetReputation failed: %v", err)
	}
	if stats.ServeCount != 3 {
		t.Fatalf("expected ServeCount 3, got: %d", stats.ServeCount)
	}
}

func TestRecordServe_SelfServe(t *testing.T) {
	k, ctx := newTestKeeper(t)
	k.Activate(ctx)

	err := k.RecordServe(ctx, []byte("dev1"), "org1", 5, true)
	if err != nil {
		t.Fatalf("RecordServe failed: %v", err)
	}

	stats, err := k.GetReputation(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetReputation failed: %v", err)
	}
	if stats.ServeCount != 1 {
		t.Fatalf("expected ServeCount 1, got: %d", stats.ServeCount)
	}
	if stats.SelfServeCount != 1 {
		t.Fatalf("expected SelfServeCount 1, got: %d", stats.SelfServeCount)
	}
	if stats.ServeXP != 2 {
		t.Fatalf("expected ServeXP 2 (self-serve discount), got: %d", stats.ServeXP)
	}
}

func TestRecordServe_MultiOrg(t *testing.T) {
	k, ctx := newTestKeeper(t)
	k.Activate(ctx)

	err := k.RecordServe(ctx, []byte("dev1"), "org1", 5, false)
	if err != nil {
		t.Fatalf("RecordServe 1 failed: %v", err)
	}
	err = k.RecordServe(ctx, []byte("dev1"), "org2", 6, false)
	if err != nil {
		t.Fatalf("RecordServe 2 failed: %v", err)
	}
	err = k.RecordServe(ctx, []byte("dev1"), "org3", 7, false)
	if err != nil {
		t.Fatalf("RecordServe 3 failed: %v", err)
	}

	stats, err := k.GetReputation(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetReputation failed: %v", err)
	}
	if stats.OrgBreadth != 3 {
		t.Fatalf("expected OrgBreadth 3, got: %d", stats.OrgBreadth)
	}

	orgSet, err := k.GetContributorOrgSet(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetContributorOrgSet failed: %v", err)
	}
	if len(orgSet.OrgIds) != 3 {
		t.Fatalf("expected 3 orgs in set, got: %d", len(orgSet.OrgIds))
	}
}

func TestRecordServe_SameOrgTwice(t *testing.T) {
	k, ctx := newTestKeeper(t)
	k.Activate(ctx)

	err := k.RecordServe(ctx, []byte("dev1"), "org1", 5, false)
	if err != nil {
		t.Fatalf("RecordServe 1 failed: %v", err)
	}
	err = k.RecordServe(ctx, []byte("dev1"), "org1", 6, false)
	if err != nil {
		t.Fatalf("RecordServe 2 failed: %v", err)
	}

	stats, err := k.GetReputation(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetReputation failed: %v", err)
	}
	if stats.OrgBreadth != 1 {
		t.Fatalf("expected OrgBreadth 1 (not double-counted), got: %d", stats.OrgBreadth)
	}
}

func TestRecordServe_NotActive(t *testing.T) {
	k, ctx := newTestKeeper(t)

	err := k.RecordServe(ctx, []byte("dev1"), "org1", 5, false)
	if err != types.ErrReputationNotActive {
		t.Fatalf("expected ErrReputationNotActive, got: %v", err)
	}
}

func TestRecordServe_FirstSeenEpochOnlySetOnce(t *testing.T) {
	k, ctx := newTestKeeper(t)
	k.Activate(ctx)

	err := k.RecordServe(ctx, []byte("dev1"), "org1", 5, false)
	if err != nil {
		t.Fatalf("RecordServe 1 failed: %v", err)
	}
	err = k.RecordServe(ctx, []byte("dev1"), "org2", 10, false)
	if err != nil {
		t.Fatalf("RecordServe 2 failed: %v", err)
	}

	stats, err := k.GetReputation(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetReputation failed: %v", err)
	}
	if stats.FirstSeenEpoch != 5 {
		t.Fatalf("expected FirstSeenEpoch 5 (first serve), got: %d", stats.FirstSeenEpoch)
	}
}

func TestGetServeStats(t *testing.T) {
	k, ctx := newTestKeeper(t)
	k.Activate(ctx)

	err := k.RecordServe(ctx, []byte("dev1"), "org1", 5, false)
	if err != nil {
		t.Fatalf("RecordServe failed: %v", err)
	}
	err = k.RecordServe(ctx, []byte("dev1"), "org2", 6, true)
	if err != nil {
		t.Fatalf("RecordServe 2 failed: %v", err)
	}

	stats, err := k.GetServeStats(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetServeStats failed: %v", err)
	}
	if stats.ServeCount != 2 {
		t.Fatalf("expected ServeCount 2, got: %d", stats.ServeCount)
	}
	if stats.SelfServeCount != 1 {
		t.Fatalf("expected SelfServeCount 1, got: %d", stats.SelfServeCount)
	}
	if stats.OrgBreadth != 2 {
		t.Fatalf("expected OrgBreadth 2, got: %d", stats.OrgBreadth)
	}
	if stats.ServeXP != 7 {
		t.Fatalf("expected ServeXP 7 (5+2), got: %d", stats.ServeXP)
	}
}

func TestGetServeStats_NoServes(t *testing.T) {
	k, ctx := newTestKeeper(t)
	k.Activate(ctx)

	memory := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, []string{"golang"}, "commitllm")
	k.AddMemory(ctx, []byte("dev1"), memory)

	stats, err := k.GetServeStats(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetServeStats failed: %v", err)
	}
	if stats.ServeCount != 0 {
		t.Fatalf("expected ServeCount 0, got: %d", stats.ServeCount)
	}
	if stats.FirstSeenEpoch != 0 {
		t.Fatalf("expected FirstSeenEpoch 0 (no serves), got: %d", stats.FirstSeenEpoch)
	}
}

func TestGetContributorOrgSet(t *testing.T) {
	k, ctx := newTestKeeper(t)
	k.Activate(ctx)

	err := k.RecordServe(ctx, []byte("dev1"), "org1", 5, false)
	if err != nil {
		t.Fatalf("RecordServe 1 failed: %v", err)
	}
	err = k.RecordServe(ctx, []byte("dev1"), "org2", 6, false)
	if err != nil {
		t.Fatalf("RecordServe 2 failed: %v", err)
	}

	orgSet, err := k.GetContributorOrgSet(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetContributorOrgSet failed: %v", err)
	}
	if len(orgSet.OrgIds) != 2 {
		t.Fatalf("expected 2 orgs, got: %d", len(orgSet.OrgIds))
	}
}

func TestGetContributorOrgSet_Empty(t *testing.T) {
	k, ctx := newTestKeeper(t)
	k.Activate(ctx)

	orgSet, err := k.GetContributorOrgSet(ctx, []byte("unknown"))
	if err != nil {
		t.Fatalf("GetContributorOrgSet failed: %v", err)
	}
	if len(orgSet.OrgIds) != 0 {
		t.Fatalf("expected empty org set, got: %d", len(orgSet.OrgIds))
	}
}

func TestGetCrossOrgProfile(t *testing.T) {
	k, ctx := newTestKeeper(t)
	k.Activate(ctx)

	memory := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, []string{"golang"}, "commitllm")
	k.AddMemory(ctx, []byte("dev1"), memory)

	err := k.RecordServe(ctx, []byte("dev1"), "org1", 5, false)
	if err != nil {
		t.Fatalf("RecordServe failed: %v", err)
	}
	err = k.RecordServe(ctx, []byte("dev1"), "org2", 6, false)
	if err != nil {
		t.Fatalf("RecordServe 2 failed: %v", err)
	}

	stats, err := k.GetReputation(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetReputation failed: %v", err)
	}

	orgSet, err := k.GetContributorOrgSet(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetContributorOrgSet failed: %v", err)
	}

	if stats.MemoryCount != 1 {
		t.Fatalf("expected MemoryCount 1, got: %d", stats.MemoryCount)
	}
	if stats.XP != 35 {
		t.Fatalf("expected XP 35, got: %d", stats.XP)
	}
	if stats.ServeCount != 2 {
		t.Fatalf("expected ServeCount 2, got: %d", stats.ServeCount)
	}
	if stats.OrgBreadth != 2 {
		t.Fatalf("expected OrgBreadth 2, got: %d", stats.OrgBreadth)
	}
	if len(orgSet.OrgIds) != 2 {
		t.Fatalf("expected 2 orgs, got: %d", len(orgSet.OrgIds))
	}
}

func TestContributionAndServeCoexist(t *testing.T) {
	k, ctx := newTestKeeper(t)
	k.Activate(ctx)

	memory := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, []string{"golang"}, "commitllm")
	k.AddMemory(ctx, []byte("dev1"), memory)

	err := k.RecordServe(ctx, []byte("dev1"), "org1", 5, false)
	if err != nil {
		t.Fatalf("RecordServe failed: %v", err)
	}

	stats, err := k.GetReputation(ctx, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetReputation failed: %v", err)
	}

	if stats.XP != 35 {
		t.Fatalf("expected contribution XP 35, got: %d", stats.XP)
	}
	if stats.ServeXP != 5 {
		t.Fatalf("expected serve XP 5, got: %d", stats.ServeXP)
	}
	if stats.MemoryCount != 1 {
		t.Fatalf("expected MemoryCount 1, got: %d", stats.MemoryCount)
	}
	if stats.ServeCount != 1 {
		t.Fatalf("expected ServeCount 1, got: %d", stats.ServeCount)
	}
}

func TestGenesisRoundTrip_Extended(t *testing.T) {
	k1, ctx1 := newTestKeeper(t)
	k1.Activate(ctx1)

	memory := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, []string{"golang"}, "commitllm")
	k1.AddMemory(ctx1, []byte("dev1"), memory)

	err := k1.RecordServe(ctx1, []byte("dev1"), "org1", 5, false)
	if err != nil {
		t.Fatalf("RecordServe failed: %v", err)
	}

	state, err := k1.ExportGenesisState(ctx1)
	if err != nil {
		t.Fatalf("ExportGenesisState failed: %v", err)
	}
	if !state.Active {
		t.Fatal("expected active=true from export")
	}
	if len(state.Stats) != 1 {
		t.Fatalf("expected 1 stat entry, got: %d", len(state.Stats))
	}
	if len(state.ContributorOrgSets) != 1 {
		t.Fatalf("expected 1 org set entry, got: %d", len(state.ContributorOrgSets))
	}

	k2, ctx2 := newTestKeeper(t)
	err = k2.InitGenesisState(ctx2, state)
	if err != nil {
		t.Fatalf("InitGenesisState failed: %v", err)
	}

	stats, err := k2.GetReputation(ctx2, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetReputation after import failed: %v", err)
	}
	if stats.MemoryCount != 1 {
		t.Fatalf("expected MemoryCount 1 after import, got: %d", stats.MemoryCount)
	}
	if stats.ServeCount != 1 {
		t.Fatalf("expected ServeCount 1 after import, got: %d", stats.ServeCount)
	}

	orgSet, err := k2.GetContributorOrgSet(ctx2, []byte("dev1"))
	if err != nil {
		t.Fatalf("GetContributorOrgSet after import failed: %v", err)
	}
	if len(orgSet.OrgIds) != 1 {
		t.Fatalf("expected 1 org after import, got: %d", len(orgSet.OrgIds))
	}
}
