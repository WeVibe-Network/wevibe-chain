package keeper_test

import (
	"context"
	"crypto/sha256"
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/wevibe-network/wevibe-chain/testutil/keeper"
	memorytypes "github.com/wevibe-network/wevibe-chain/x/memory/types"
	"github.com/wevibe-network/wevibe-chain/x/reputation/keeper"
	"github.com/wevibe-network/wevibe-chain/x/reputation/types"
)

// ---------------------------------------------------------------------------
// Test fixtures / mocks (unique to this file)
// ---------------------------------------------------------------------------

// queryMockMemoryKeeper is a hand-rolled MemoryKeeper implementation used to
// exercise the upheld-report and verification query paths without a real
// memory module keeper.
type queryMockMemoryKeeper struct {
	reports      []*memorytypes.StoredMemoryReport
	approved     map[string]*memorytypes.MemoryCommitment
	upheldReport *memorytypes.StoredMemoryReport
}

func (m *queryMockMemoryKeeper) IterateUpheldReports(ctx context.Context, cb func(*memorytypes.StoredMemoryReport) bool) error {
	for _, r := range m.reports {
		if !cb(r) {
			break
		}
	}
	return nil
}

func (m *queryMockMemoryKeeper) GetUpheldReport(ctx context.Context, orgID string, memoryHash []byte) (*memorytypes.StoredMemoryReport, error) {
	return m.upheldReport, nil
}

func (m *queryMockMemoryKeeper) GetApprovedMemory(ctx context.Context, orgID string, contentHash []byte) (*memorytypes.MemoryCommitment, error) {
	if m.approved == nil {
		return nil, memorytypes.ErrMemoryNotFound
	}
	c, ok := m.approved[orgID+"/"+string(contentHash)]
	if !ok {
		return nil, memorytypes.ErrMemoryNotFound
	}
	return c, nil
}

// newQueryTestKeeper builds a keeper and a query server. It mirrors the setup
// used by newTestKeeper in keeper_test.go but exposes the keeper so that the
// memory keeper dependency can be wired in for report queries.
func newQueryTestKeeper(t *testing.T) (*keeper.Keeper, context.Context) {
	storeKey := storetypes.NewKVStoreKey("reputation")
	storeService, _ := testkeeper.NewTestStoreService(t, storeKey)
	logger := testkeeper.NewTestLogger()
	k := keeper.NewKeeper(storeService, logger, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry")
	return k, context.Background()
}

// ---------------------------------------------------------------------------
// GetReputation
// ---------------------------------------------------------------------------

func TestQueryGetReputation_Success(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.Activate(ctx)
	mem := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, []string{"golang"}, "commitllm")
	require.NoError(t, k.AddMemory(ctx, []byte("dev1"), mem))

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.GetReputation(ctx, &types.QueryGetReputationRequest{Developer: []byte("dev1")})
	require.NoError(t, err)
	require.Equal(t, "dev1", resp.DeveloperId)
	require.Equal(t, uint64(1), resp.MemoryCount)
	require.Equal(t, uint64(35), resp.Xp)
}

func TestQueryGetReputation_NotFound(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.Activate(ctx)

	q := keeper.NewQueryServerImpl(k)
	_, err := q.GetReputation(ctx, &types.QueryGetReputationRequest{Developer: []byte("unknown")})
	require.ErrorIs(t, err, types.ErrNoStats)
}

// ---------------------------------------------------------------------------
// GetXP
// ---------------------------------------------------------------------------

func TestQueryGetXP_Success(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.Activate(ctx)
	mem := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, nil, "commitllm")
	require.NoError(t, k.AddMemory(ctx, []byte("dev1"), mem))

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.GetXP(ctx, &types.QueryGetXPRequest{Developer: []byte("dev1")})
	require.NoError(t, err)
	require.Equal(t, uint64(35), resp.Xp)
}

func TestQueryGetXP_NotFound(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.Activate(ctx)

	q := keeper.NewQueryServerImpl(k)
	_, err := q.GetXP(ctx, &types.QueryGetXPRequest{Developer: []byte("unknown")})
	require.ErrorIs(t, err, types.ErrNoStats)
}

// ---------------------------------------------------------------------------
// IsActive
// ---------------------------------------------------------------------------

func TestQueryIsActive_Active(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.Activate(ctx)

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.IsActive(ctx, &types.QueryIsActiveRequest{})
	require.NoError(t, err)
	require.True(t, resp.Active)
}

func TestQueryIsActive_Inactive(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	// not activated

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.IsActive(ctx, &types.QueryIsActiveRequest{})
	require.NoError(t, err)
	require.False(t, resp.Active)
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

func TestQueryParams_Default(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp.Params)
	// unset params returns DefaultParams
	require.Equal(t, types.DefaultParams().MaxDifficulty, resp.Params.MaxDifficulty)
	require.Equal(t, types.DefaultParams().MaxQuality, resp.Params.MaxQuality)
}

func TestQueryParams_Stored(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	require.NoError(t, k.SetParams(ctx, types.Params{
		Active:        true,
		MaxDifficulty: 8,
		MaxQuality:    9,
	}))

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, uint32(8), resp.Params.MaxDifficulty)
	require.Equal(t, uint32(9), resp.Params.MaxQuality)
}

// ---------------------------------------------------------------------------
// GetServeStats
// ---------------------------------------------------------------------------

func TestQueryGetServeStats_Success(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.Activate(ctx)
	require.NoError(t, k.RecordServe(ctx, []byte("dev1"), "org1", 5, false))
	require.NoError(t, k.RecordServe(ctx, []byte("dev1"), "org2", 6, true))

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.GetServeStats(ctx, &types.QueryGetServeStatsRequest{Developer: []byte("dev1")})
	require.NoError(t, err)
	require.Equal(t, uint64(2), resp.ServeCount)
	require.Equal(t, uint64(1), resp.SelfServeCount)
	require.Equal(t, uint64(2), resp.OrgBreadth)
	require.Equal(t, uint64(7), resp.ServeXp)
	require.Equal(t, uint64(5), resp.FirstSeenEpoch)
}

func TestQueryGetServeStats_NoServes(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.Activate(ctx)
	// developer has a contribution but no serves; serve fields default to zero.
	mem := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, nil, "commitllm")
	require.NoError(t, k.AddMemory(ctx, []byte("dev1"), mem))

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.GetServeStats(ctx, &types.QueryGetServeStatsRequest{Developer: []byte("dev1")})
	require.NoError(t, err)
	require.Equal(t, uint64(0), resp.ServeCount)
	require.Equal(t, uint64(0), resp.FirstSeenEpoch)
}

func TestQueryGetServeStats_NotFound(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.Activate(ctx)

	q := keeper.NewQueryServerImpl(k)
	_, err := q.GetServeStats(ctx, &types.QueryGetServeStatsRequest{Developer: []byte("unknown")})
	require.ErrorIs(t, err, types.ErrNoStats)
}

// ---------------------------------------------------------------------------
// GetContributorOrgSet
// ---------------------------------------------------------------------------

func TestQueryGetContributorOrgSet_Success(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.Activate(ctx)
	require.NoError(t, k.RecordServe(ctx, []byte("dev1"), "org1", 5, false))
	require.NoError(t, k.RecordServe(ctx, []byte("dev1"), "org2", 6, false))

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.GetContributorOrgSet(ctx, &types.QueryGetContributorOrgSetRequest{Developer: []byte("dev1")})
	require.NoError(t, err)
	require.Len(t, resp.OrgIds, 2)
}

func TestQueryGetContributorOrgSet_Empty(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.Activate(ctx)

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.GetContributorOrgSet(ctx, &types.QueryGetContributorOrgSetRequest{Developer: []byte("unknown")})
	require.NoError(t, err)
	require.Empty(t, resp.OrgIds)
}

// ---------------------------------------------------------------------------
// GetCrossOrgProfile
// ---------------------------------------------------------------------------

func TestQueryGetCrossOrgProfile_Success(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.Activate(ctx)
	mem := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, []string{"golang"}, "commitllm")
	require.NoError(t, k.AddMemory(ctx, []byte("dev1"), mem))
	require.NoError(t, k.RecordServe(ctx, []byte("dev1"), "org1", 5, false))
	require.NoError(t, k.RecordServe(ctx, []byte("dev1"), "org2", 6, false))

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.GetCrossOrgProfile(ctx, &types.QueryGetCrossOrgProfileRequest{Developer: []byte("dev1")})
	require.NoError(t, err)
	require.Equal(t, "dev1", resp.DeveloperId)
	require.Equal(t, uint64(1), resp.MemoryCount)
	require.Equal(t, uint64(35), resp.Xp)
	require.Equal(t, uint64(2), resp.ServeCount)
	require.Equal(t, uint64(2), resp.OrgBreadth)
	require.Len(t, resp.OrgIds, 2)
	require.Equal(t, uint64(1), resp.DomainTags["golang"])
}

func TestQueryGetCrossOrgProfile_NotFound(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.Activate(ctx)

	q := keeper.NewQueryServerImpl(k)
	_, err := q.GetCrossOrgProfile(ctx, &types.QueryGetCrossOrgProfileRequest{Developer: []byte("unknown")})
	require.ErrorIs(t, err, types.ErrNoStats)
}

// ---------------------------------------------------------------------------
// GetContributorProfile
// ---------------------------------------------------------------------------

func TestQueryGetContributorProfile_Success(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.Activate(ctx)
	mem := types.NewAttestedMemory([]byte("contrib1"), "cid1", 5, 7, []string{"golang"}, "commitllm")
	require.NoError(t, k.AddMemory(ctx, []byte("contrib1"), mem))

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.GetContributorProfile(ctx, &types.QueryGetContributorProfileRequest{ContributorId: "contrib1"})
	require.NoError(t, err)
	require.Equal(t, "contrib1", resp.ContributorId)
	require.Equal(t, uint64(35), resp.Xp)
	require.Equal(t, uint64(1), resp.MemoryCount)
	require.Len(t, resp.DifficultyHistogram, 11)
	require.Equal(t, uint64(1), resp.ProvenanceBreakdown["commitllm"])
}

func TestQueryGetContributorProfile_EmptyContributorId(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.Activate(ctx)

	q := keeper.NewQueryServerImpl(k)
	_, err := q.GetContributorProfile(ctx, &types.QueryGetContributorProfileRequest{ContributorId: ""})
	require.ErrorIs(t, err, types.ErrInvalidDeveloper)
}

func TestQueryGetContributorProfile_NotFound(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.Activate(ctx)

	q := keeper.NewQueryServerImpl(k)
	_, err := q.GetContributorProfile(ctx, &types.QueryGetContributorProfileRequest{ContributorId: "missing"})
	require.ErrorIs(t, err, types.ErrNoStats)
}

// ---------------------------------------------------------------------------
// LeaderProfile
// ---------------------------------------------------------------------------

func TestQueryLeaderProfile_Success(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)

	q := keeper.NewQueryServerImpl(k)
	// GetLeaderProfile returns an empty (non-nil) profile when none stored.
	resp, err := q.LeaderProfile(ctx, &types.QueryLeaderProfileRequest{
		LeaderPubkey: "leader1",
		OrgId:        "org1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Profile)
	require.Equal(t, "leader1", resp.Profile.LeaderPubkey)
	require.Equal(t, "org1", resp.Profile.OrgId)
}

func TestQueryLeaderProfile_MissingArgs(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)

	q := keeper.NewQueryServerImpl(k)
	_, err := q.LeaderProfile(ctx, &types.QueryLeaderProfileRequest{LeaderPubkey: "", OrgId: "org1"})
	require.ErrorIs(t, err, types.ErrInvalidDeveloper)

	_, err = q.LeaderProfile(ctx, &types.QueryLeaderProfileRequest{LeaderPubkey: "leader1", OrgId: ""})
	require.ErrorIs(t, err, types.ErrInvalidDeveloper)
}

// ---------------------------------------------------------------------------
// ModeratorProfile
// ---------------------------------------------------------------------------

func TestQueryModeratorProfile_Success(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.ModeratorProfile(ctx, &types.QueryModeratorProfileRequest{
		ModeratorPubkey: "mod1",
		OrgId:           "org1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Profile)
	require.Equal(t, "mod1", resp.Profile.ModeratorPubkey)
	require.Equal(t, "org1", resp.Profile.OrgId)
}

func TestQueryModeratorProfile_MissingArgs(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)

	q := keeper.NewQueryServerImpl(k)
	_, err := q.ModeratorProfile(ctx, &types.QueryModeratorProfileRequest{ModeratorPubkey: "", OrgId: "org1"})
	require.ErrorIs(t, err, types.ErrInvalidDeveloper)

	_, err = q.ModeratorProfile(ctx, &types.QueryModeratorProfileRequest{ModeratorPubkey: "mod1", OrgId: ""})
	require.ErrorIs(t, err, types.ErrInvalidDeveloper)
}

// ---------------------------------------------------------------------------
// UpheldReportsByContributor
// ---------------------------------------------------------------------------

func TestQueryUpheldReportsByContributor_Success(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	mk := &queryMockMemoryKeeper{
		reports: []*memorytypes.StoredMemoryReport{
			{OrgId: "org1", ContributorPubkey: "contrib1", Reason: "spam", UpheldAtEpoch: 3},
			{OrgId: "org2", ContributorPubkey: "other", Reason: "noise"},
		},
	}
	k.SetMemoryKeeper(mk)

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.UpheldReportsByContributor(ctx, &types.QueryUpheldReportsByContributorRequest{
		ContributorId: "contrib1",
	})
	require.NoError(t, err)
	require.Len(t, resp.Reports, 1)
	require.Equal(t, "org1", resp.Reports[0].OrgId)
	require.Equal(t, uint64(1), resp.Pagination.Total)
}

func TestQueryUpheldReportsByContributor_EmptyContributorId(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.SetMemoryKeeper(&queryMockMemoryKeeper{})

	q := keeper.NewQueryServerImpl(k)
	_, err := q.UpheldReportsByContributor(ctx, &types.QueryUpheldReportsByContributorRequest{ContributorId: ""})
	require.Error(t, err)
}

func TestQueryUpheldReportsByContributor_NoMatches(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.SetMemoryKeeper(&queryMockMemoryKeeper{
		reports: []*memorytypes.StoredMemoryReport{
			{OrgId: "org1", ContributorPubkey: "someone-else"},
		},
	})

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.UpheldReportsByContributor(ctx, &types.QueryUpheldReportsByContributorRequest{
		ContributorId: "contrib1",
	})
	require.NoError(t, err)
	require.Empty(t, resp.Reports)
	require.Equal(t, uint64(0), resp.Pagination.Total)
}

// ---------------------------------------------------------------------------
// UpheldReportsByModerator
// ---------------------------------------------------------------------------

func TestQueryUpheldReportsByModerator_Success(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.SetMemoryKeeper(&queryMockMemoryKeeper{
		reports: []*memorytypes.StoredMemoryReport{
			{OrgId: "org1", ApprovingModerators: []string{"mod1", "mod2"}},
			{OrgId: "org2", ApprovingModerators: []string{"mod3"}},
		},
	})

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.UpheldReportsByModerator(ctx, &types.QueryUpheldReportsByModeratorRequest{
		ModeratorPubkey: "mod1",
	})
	require.NoError(t, err)
	require.Len(t, resp.Reports, 1)
	require.Equal(t, "org1", resp.Reports[0].OrgId)
}

func TestQueryUpheldReportsByModerator_EmptyPubkey(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.SetMemoryKeeper(&queryMockMemoryKeeper{})

	q := keeper.NewQueryServerImpl(k)
	_, err := q.UpheldReportsByModerator(ctx, &types.QueryUpheldReportsByModeratorRequest{ModeratorPubkey: ""})
	require.Error(t, err)
}

func TestQueryUpheldReportsByModerator_NoMatches(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.SetMemoryKeeper(&queryMockMemoryKeeper{
		reports: []*memorytypes.StoredMemoryReport{
			{OrgId: "org1", ApprovingModerators: []string{"other"}},
		},
	})

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.UpheldReportsByModerator(ctx, &types.QueryUpheldReportsByModeratorRequest{
		ModeratorPubkey: "mod1",
	})
	require.NoError(t, err)
	require.Empty(t, resp.Reports)
}

// ---------------------------------------------------------------------------
// UpheldReportsByLeader
// ---------------------------------------------------------------------------

func TestQueryUpheldReportsByLeader_Success(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.SetMemoryKeeper(&queryMockMemoryKeeper{
		reports: []*memorytypes.StoredMemoryReport{
			{OrgId: "org1", CommittingLeaderPubkey: "leader1"},
			{OrgId: "org2", CommittingLeaderPubkey: "leader2"},
		},
	})

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.UpheldReportsByLeader(ctx, &types.QueryUpheldReportsByLeaderRequest{
		LeaderPubkey: "leader1",
	})
	require.NoError(t, err)
	require.Len(t, resp.Reports, 1)
	require.Equal(t, "org1", resp.Reports[0].OrgId)
}

func TestQueryUpheldReportsByLeader_EmptyPubkey(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.SetMemoryKeeper(&queryMockMemoryKeeper{})

	q := keeper.NewQueryServerImpl(k)
	_, err := q.UpheldReportsByLeader(ctx, &types.QueryUpheldReportsByLeaderRequest{LeaderPubkey: ""})
	require.Error(t, err)
}

func TestQueryUpheldReportsByLeader_NoMatches(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.SetMemoryKeeper(&queryMockMemoryKeeper{
		reports: []*memorytypes.StoredMemoryReport{
			{OrgId: "org1", CommittingLeaderPubkey: "other"},
		},
	})

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.UpheldReportsByLeader(ctx, &types.QueryUpheldReportsByLeaderRequest{
		LeaderPubkey: "leader1",
	})
	require.NoError(t, err)
	require.Empty(t, resp.Reports)
}

// ---------------------------------------------------------------------------
// VerifyUpheldReport
// ---------------------------------------------------------------------------

func newVerifiableMemory(orgID string) *memorytypes.MemoryCommitment {
	encryptedBlob := []byte("encrypted-blob")
	wrappedDek := []byte("wrapped-dek")

	hasher := sha256.New()
	_, _ = hasher.Write(encryptedBlob)
	_, _ = hasher.Write(wrappedDek)
	contentHash := hasher.Sum(nil)

	wrappedDekHash := sha256.Sum256(wrappedDek)

	return &memorytypes.MemoryCommitment{
		OrgID:          orgID,
		ContentHash:    contentHash,
		EncryptedBlob:  encryptedBlob,
		WrappedDekEnc:  wrappedDek,
		WrappedDekHash: wrappedDekHash[:],
		PlaintextHash:  []byte("plaintext-hash"),
		Salt:           []byte("salt"),
		CiphertextHash: []byte("ciphertext-hash"),
		ContributorSig: []byte("contributor-sig"),
		Contributor:    "contrib1",
		Epoch:          7,
		MemoryType:     memorytypes.MemoryType_MEMORY_TYPE_MEMORY,
	}
}

func TestQueryVerifyUpheldReport_Success(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	mem := newVerifiableMemory("org1")
	mk := &queryMockMemoryKeeper{
		approved: map[string]*memorytypes.MemoryCommitment{
			"org1/" + string(mem.ContentHash): mem,
		},
		upheldReport: &memorytypes.StoredMemoryReport{
			OrgId:               "org1",
			ApprovingModerators: []string{"mod1"},
			UpholdingModerators: []string{"mod1"},
			UpheldAtEpoch:       9,
			Plaintext:           []byte("pt"),
		},
	}
	k.SetMemoryKeeper(mk)

	q := keeper.NewQueryServerImpl(k)
	resp, err := q.VerifyUpheldReport(ctx, &types.QueryVerifyUpheldReportRequest{
		OrgId:       "org1",
		ContentHash: mem.ContentHash,
	})
	require.NoError(t, err)
	require.Equal(t, "org1", resp.OrgId)
	require.Equal(t, uint64(7), resp.Epoch)
	require.Equal(t, "memory", resp.MemoryType)
	require.Equal(t, uint64(9), resp.UpheldAtEpoch)
	require.NotEmpty(t, resp.CanonicalBody)
}

func TestQueryVerifyUpheldReport_NotFound(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.SetMemoryKeeper(&queryMockMemoryKeeper{approved: map[string]*memorytypes.MemoryCommitment{}})

	q := keeper.NewQueryServerImpl(k)
	_, err := q.VerifyUpheldReport(ctx, &types.QueryVerifyUpheldReportRequest{
		OrgId:       "org1",
		ContentHash: []byte("does-not-exist"),
	})
	require.Error(t, err)
}

func TestQueryVerifyUpheldReport_MissingArgs(t *testing.T) {
	k, ctx := newQueryTestKeeper(t)
	k.SetMemoryKeeper(&queryMockMemoryKeeper{})

	q := keeper.NewQueryServerImpl(k)

	_, err := q.VerifyUpheldReport(ctx, &types.QueryVerifyUpheldReportRequest{OrgId: "", ContentHash: []byte("h")})
	require.Error(t, err)

	_, err = q.VerifyUpheldReport(ctx, &types.QueryVerifyUpheldReportRequest{OrgId: "org1", ContentHash: nil})
	require.Error(t, err)
}
