package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/wevibe-network/wevibe-chain/x/org/keeper"
	"github.com/wevibe-network/wevibe-chain/x/org/types"
)

// queryLeader is a deterministic leader pubkey reused across query tests.
const queryLeader = "leader_pubkey_12345678901234567890123456789012"

// ---------------------------------------------------------------------------
// GetOrg
// ---------------------------------------------------------------------------

func TestQueryGetOrg_Success(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	org := types.NewOrg("org1", queryLeader, "example.com", 1000000, 5000)
	require.NoError(t, k.RegisterOrg(ctx, org, testCreatorAddr))

	resp, err := qs.GetOrg(ctx, &types.QueryGetOrgRequest{OrgId: "org1"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "org1", resp.OrgId)
	require.Equal(t, queryLeader, resp.Leader)
	require.Equal(t, "example.com", resp.Domain)
	require.Equal(t, uint64(1000000), resp.StorageQuota)
	require.Equal(t, uint64(5000), resp.RetrievalBudget)
	require.Equal(t, int32(types.OrgStatus_ACTIVE), resp.Status)
}

func TestQueryGetOrg_NotFound(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	resp, err := qs.GetOrg(ctx, &types.QueryGetOrgRequest{OrgId: "missing"})
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrOrgNotFound)
	require.Nil(t, resp)
}

// ---------------------------------------------------------------------------
// GetMembers
// ---------------------------------------------------------------------------

func TestQueryGetMembers_Success(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	org := types.NewOrg("org1", queryLeader, "", 1000000, 5000)
	require.NoError(t, k.RegisterOrg(ctx, org, testCreatorAddr))

	require.NoError(t, k.AddMember(ctx, types.NewMemberRecord("org1", "member_pubkey_aaaa1234567890123456789012", "member")))
	require.NoError(t, k.AddMember(ctx, types.NewMemberRecord("org1", "member_pubkey_bbbb1234567890123456789012", "moderator")))

	resp, err := qs.GetMembers(ctx, &types.QueryGetMembersRequest{OrgId: "org1"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	// leader is added as a member during RegisterOrg, plus the two added above.
	require.Len(t, resp.Members, 3)
	for _, m := range resp.Members {
		require.Equal(t, "org1", m.OrgId)
		require.NotEmpty(t, m.Pubkey)
	}
}

func TestQueryGetMembers_Empty(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	resp, err := qs.GetMembers(ctx, &types.QueryGetMembersRequest{OrgId: "missing"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.Members)
}

// ---------------------------------------------------------------------------
// IsMember
// ---------------------------------------------------------------------------

func TestQueryIsMember_True(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	org := types.NewOrg("org1", queryLeader, "", 1000000, 5000)
	require.NoError(t, k.RegisterOrg(ctx, org, testCreatorAddr))
	require.NoError(t, k.AddMember(ctx, types.NewMemberRecord("org1", "member_pubkey_cccc1234567890123456789012", "member")))

	resp, err := qs.IsMember(ctx, &types.QueryIsMemberRequest{OrgId: "org1", Pubkey: "member_pubkey_cccc1234567890123456789012"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.IsMember)
}

func TestQueryIsMember_False(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	resp, err := qs.IsMember(ctx, &types.QueryIsMemberRequest{OrgId: "org1", Pubkey: "nobody"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.False(t, resp.IsMember)
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

func TestQueryParams_Default(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	resp, err := qs.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Params)
	require.Equal(t, types.DefaultParams().BaseBurnPrice, resp.Params.BaseBurnPrice)
}

func TestQueryParams_Set(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	params, err := k.GetParams(ctx)
	require.NoError(t, err)
	params.BaseBurnPrice = 42
	k.SetParams(ctx, params)

	resp, err := qs.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(42), resp.Params.BaseBurnPrice)
}

// ---------------------------------------------------------------------------
// GetTreasury
// ---------------------------------------------------------------------------

func TestQueryGetTreasury_Funded(t *testing.T) {
	k, ctx, _ := newTestKeeperWithFundedBank(t)
	qs := keeper.NewQueryServerImpl(k)

	org := types.NewOrg("org1", queryLeader, "", 1000000, 5000)
	require.NoError(t, k.RegisterOrg(ctx, org, testCreatorAddr))

	funder, _ := sdk.AccAddressFromBech32("cosmos1abc")
	require.NoError(t, k.FundTreasury(ctx, "org1", funder, math.NewInt(500000)))

	resp, err := qs.GetTreasury(ctx, &types.QueryGetTreasuryRequest{OrgId: "org1"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "500000", resp.Balance)
}

func TestQueryGetTreasury_NoTreasury(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	resp, err := qs.GetTreasury(ctx, &types.QueryGetTreasuryRequest{OrgId: "missing"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "0", resp.Balance)
}

// ---------------------------------------------------------------------------
// GetRepTiers
// ---------------------------------------------------------------------------

func TestQueryGetRepTiers_Success(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	org := types.NewOrg("org1", queryLeader, "", 1000000, 5000)
	require.NoError(t, k.RegisterOrg(ctx, org, testCreatorAddr))

	tiers := []*types.RepTierRecord{
		{MinReputation: 0, MaxReputation: 50, MaxContributionsPerEpoch: 3, PayoutPerMemory: "1"},
		{MinReputation: 200, MaxReputation: 1000, MaxContributionsPerEpoch: 50, PayoutPerMemory: "5"},
	}
	require.NoError(t, k.SetRepTiers(ctx, "org1", tiers))

	resp, err := qs.GetRepTiers(ctx, &types.QueryGetRepTiersRequest{OrgId: "org1"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Tiers, 2)
	require.Equal(t, "1", resp.Tiers[0].PayoutPerMemory)
}

func TestQueryGetRepTiers_NotFound(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	resp, err := qs.GetRepTiers(ctx, &types.QueryGetRepTiersRequest{OrgId: "missing"})
	require.Error(t, err)
	require.Nil(t, resp)
}

// ---------------------------------------------------------------------------
// GetOrgConfig
// ---------------------------------------------------------------------------

func TestQueryGetOrgConfig_Success(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	org := types.NewOrg("org1", queryLeader, "", 1000000, 5000)
	require.NoError(t, k.RegisterOrg(ctx, org, testCreatorAddr))

	cfg := &types.OrgConfig{
		OrgID:                    "org1",
		ServeAttestationRequired: true,
		MinContributionsPerEpoch: 7,
		ContestStakeVibe:         123,
	}
	require.NoError(t, k.SetOrgConfig(ctx, "org1", cfg))

	resp, err := qs.GetOrgConfig(ctx, &types.QueryGetOrgConfigRequest{OrgId: "org1"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.ServeAttestationRequired)
	require.Equal(t, uint64(7), resp.MinContributionsPerEpoch)
	require.Equal(t, uint64(123), resp.ContestStakeVibe)
}

func TestQueryGetOrgConfig_DefaultWhenUnset(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	// No config stored: keeper returns a zero-valued default config, no error.
	resp, err := qs.GetOrgConfig(ctx, &types.QueryGetOrgConfigRequest{OrgId: "missing"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.False(t, resp.ServeAttestationRequired)
	require.Equal(t, uint64(0), resp.MinContributionsPerEpoch)
	require.Equal(t, uint64(0), resp.ContestStakeVibe)
}

// ---------------------------------------------------------------------------
// GetOrgProfile
// ---------------------------------------------------------------------------

func TestQueryGetOrgProfile_Success(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	org := types.NewOrg("org1", queryLeader, "example.com", 1000000, 5000)
	require.NoError(t, k.RegisterOrg(ctx, org, testCreatorAddr))
	require.NoError(t, k.AddMember(ctx, types.NewMemberRecord("org1", "member_pubkey_dddd1234567890123456789012", "moderator")))

	resp, err := qs.GetOrgProfile(ctx, &types.QueryGetOrgProfileRequest{OrgId: "org1"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "org1", resp.OrgId)
	require.Equal(t, queryLeader, resp.Leader)
	require.Equal(t, "example.com", resp.Domain)
	require.Equal(t, int32(types.OrgStatus_ACTIVE), resp.Status)
	// leader + one moderator member.
	require.Equal(t, uint64(2), resp.MemberCount)
	require.Equal(t, uint64(1), resp.ModeratorCount)
	require.Equal(t, "0", resp.TreasuryBalance)
}

func TestQueryGetOrgProfile_EmptyOrgID(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	resp, err := qs.GetOrgProfile(ctx, &types.QueryGetOrgProfileRequest{OrgId: ""})
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidOrgID)
	require.Nil(t, resp)
}

func TestQueryGetOrgProfile_NotFound(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	resp, err := qs.GetOrgProfile(ctx, &types.QueryGetOrgProfileRequest{OrgId: "missing"})
	require.Error(t, err)
	require.Nil(t, resp)
}
