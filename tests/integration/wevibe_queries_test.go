package integration

import (
	"testing"

	"github.com/stretchr/testify/require"

	emissionskeeper "github.com/wevibe-network/wevibe-chain/x/emissions/keeper"
	emissionstypes "github.com/wevibe-network/wevibe-chain/x/emissions/types"
	orgkeeper "github.com/wevibe-network/wevibe-chain/x/org/keeper"
	orgtypes "github.com/wevibe-network/wevibe-chain/x/org/types"
	reputationkeeper "github.com/wevibe-network/wevibe-chain/x/reputation/keeper"
	reputationtypes "github.com/wevibe-network/wevibe-chain/x/reputation/types"
)

func TestQueryOrg_GetOrg(t *testing.T) {
	suite := NewTestSuite(t)
	orgID := orgtypes.DeriveOrgID(suite.UserAddr)

	msg := &orgtypes.MsgRegisterOrg{
		Signer:          suite.UserAddr.String(),
		Leader:          suite.UserAddr.String(),
		StorageQuota:    1000,
		RetrievalBudget: 500,
	}
	_, err := suite.DeliverMsg(msg)
	require.NoError(t, err)

	queryServer := orgkeeper.NewQueryServerImpl(suite.OrgKeeper)
	resp, err := queryServer.GetOrg(suite.Ctx, &orgtypes.QueryGetOrgRequest{
		OrgId: orgID,
	})
	require.NoError(t, err)
	require.Equal(t, orgID, resp.OrgId)
	require.Equal(t, suite.UserAddr.String(), resp.Leader)
	require.Equal(t, uint64(1000), resp.StorageQuota)
	require.Equal(t, uint64(500), resp.RetrievalBudget)
}

func TestQueryOrg_GetMembers(t *testing.T) {
	suite := NewTestSuite(t)
	orgID := orgtypes.DeriveOrgID(suite.UserAddr)

	orgMsg := &orgtypes.MsgRegisterOrg{
		Signer:          suite.UserAddr.String(),
		Leader:          suite.UserAddr.String(),
		StorageQuota:    1000,
		RetrievalBudget: 500,
		LeaderWallet:    suite.UserAddr.String(),
	}
	_, err := suite.DeliverMsg(orgMsg)
	require.NoError(t, err)

	memberMsg := &orgtypes.MsgAddMember{
		Signer: suite.UserAddr.String(),
		OrgId:  orgID,
		Pubkey: "wevibe1member1234567890123456789012345678901234567890",
		Role:   "member",
	}
	_, err = suite.DeliverMsg(memberMsg)
	require.NoError(t, err)
}

func TestQueryOrg_IsMember(t *testing.T) {
	suite := NewTestSuite(t)
	orgID := orgtypes.DeriveOrgID(suite.UserAddr)

	orgMsg := &orgtypes.MsgRegisterOrg{
		Signer:          suite.UserAddr.String(),
		Leader:          suite.UserAddr.String(),
		StorageQuota:    1000,
		RetrievalBudget: 500,
	}
	_, err := suite.DeliverMsg(orgMsg)
	require.NoError(t, err)

	queryServer := orgkeeper.NewQueryServerImpl(suite.OrgKeeper)

	respLeader, err := queryServer.IsMember(suite.Ctx, &orgtypes.QueryIsMemberRequest{
		OrgId:  orgID,
		Pubkey: suite.UserAddr.String(),
	})
	require.NoError(t, err)
	require.True(t, respLeader.IsMember)

	respNotMember, err := queryServer.IsMember(suite.Ctx, &orgtypes.QueryIsMemberRequest{
		OrgId:  orgID,
		Pubkey: "wevibe1notamember123456789012345678901234567890",
	})
	require.NoError(t, err)
	require.False(t, respNotMember.IsMember)
}

func TestQueryEmissions_GetEmissionPool_SeededAtGenesis(t *testing.T) {
	suite := NewTestSuite(t)

	// CO-040: the emissions module seeds an emission pool at genesis (derived
	// from DefaultParams). The pool must therefore be queryable immediately.
	queryServer := emissionskeeper.NewQueryServerImpl(suite.EmissionsKeeper)
	resp, err := queryServer.GetEmissionPool(suite.Ctx, &emissionstypes.QueryGetEmissionPoolRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, emissionstypes.DefaultParams().DailyMintAmount, resp.DailyMint)
	require.Equal(t, uint64(100), resp.OperatorShare+resp.ValidatorShare)
}

func TestQueryReputation_IsActive(t *testing.T) {
	suite := NewTestSuite(t)

	// CO-040 (GAP-REP-1): the reputation module is active at genesis.
	queryServer := reputationkeeper.NewQueryServerImpl(suite.ReputationKeeper)
	resp, err := queryServer.IsActive(suite.Ctx, &reputationtypes.QueryIsActiveRequest{})
	require.NoError(t, err)
	require.True(t, resp.Active)
}

func TestQueryReputation_GetReputation_NotFound(t *testing.T) {
	suite := NewTestSuite(t)

	queryServer := reputationkeeper.NewQueryServerImpl(suite.ReputationKeeper)
	_, err := queryServer.GetReputation(suite.Ctx, &reputationtypes.QueryGetReputationRequest{
		Developer: []byte("nonexistent-developer"),
	})
	require.Error(t, err)
}
