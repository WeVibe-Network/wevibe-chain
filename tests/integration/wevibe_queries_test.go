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

	msg := &orgtypes.MsgRegisterOrg{
		Signer:          suite.UserAddr.String(),
		OrgId:           "test-org-1",
		Leader:          suite.UserAddr.String(),
		StorageQuota:    1000,
		RetrievalBudget: 500,
	}
	_, err := suite.DeliverMsg(msg)
	require.NoError(t, err)

	queryServer := orgkeeper.NewQueryServerImpl(suite.OrgKeeper)
	resp, err := queryServer.GetOrg(suite.Ctx, &orgtypes.QueryGetOrgRequest{
		OrgId: "test-org-1",
	})
	require.NoError(t, err)
	require.Equal(t, "test-org-1", resp.OrgId)
	require.Equal(t, suite.UserAddr.String(), resp.Leader)
	require.Equal(t, uint64(1000), resp.StorageQuota)
	require.Equal(t, uint64(500), resp.RetrievalBudget)
}

func TestQueryOrg_GetMembers(t *testing.T) {
	suite := NewTestSuite(t)

	orgMsg := &orgtypes.MsgRegisterOrg{
		Signer:          suite.UserAddr.String(),
		OrgId:           "test-org-1",
		Leader:          suite.UserAddr.String(),
		StorageQuota:    1000,
		RetrievalBudget: 500,
	}
	_, err := suite.DeliverMsg(orgMsg)
	require.NoError(t, err)

	memberMsg := &orgtypes.MsgAddMember{
		Signer: suite.UserAddr.String(),
		OrgId:  "test-org-1",
		Pubkey: "wevibe1member1234567890123456789012345678901234567890",
		Role:   "member",
	}
	_, err = suite.DeliverMsg(memberMsg)
	require.NoError(t, err)
}

func TestQueryOrg_IsMember(t *testing.T) {
	suite := NewTestSuite(t)

	orgMsg := &orgtypes.MsgRegisterOrg{
		Signer:          suite.UserAddr.String(),
		OrgId:           "test-org-1",
		Leader:          suite.UserAddr.String(),
		StorageQuota:    1000,
		RetrievalBudget: 500,
	}
	_, err := suite.DeliverMsg(orgMsg)
	require.NoError(t, err)

	queryServer := orgkeeper.NewQueryServerImpl(suite.OrgKeeper)

	respLeader, err := queryServer.IsMember(suite.Ctx, &orgtypes.QueryIsMemberRequest{
		OrgId:  "test-org-1",
		Pubkey: suite.UserAddr.String(),
	})
	require.NoError(t, err)
	require.True(t, respLeader.IsMember)

	respNotMember, err := queryServer.IsMember(suite.Ctx, &orgtypes.QueryIsMemberRequest{
		OrgId:  "test-org-1",
		Pubkey: "wevibe1notamember123456789012345678901234567890",
	})
	require.NoError(t, err)
	require.False(t, respNotMember.IsMember)
}

func TestQueryEmissions_GetEmissionPool_NotFound(t *testing.T) {
	suite := NewTestSuite(t)

	queryServer := emissionskeeper.NewQueryServerImpl(suite.EmissionsKeeper)
	_, err := queryServer.GetEmissionPool(suite.Ctx, &emissionstypes.QueryGetEmissionPoolRequest{})
	require.Error(t, err)
}

func TestQueryReputation_IsActive(t *testing.T) {
	suite := NewTestSuite(t)

	queryServer := reputationkeeper.NewQueryServerImpl(suite.ReputationKeeper)
	resp, err := queryServer.IsActive(suite.Ctx, &reputationtypes.QueryIsActiveRequest{})
	require.NoError(t, err)
	require.False(t, resp.Active)
}

func TestQueryReputation_GetReputation_NotFound(t *testing.T) {
	suite := NewTestSuite(t)

	queryServer := reputationkeeper.NewQueryServerImpl(suite.ReputationKeeper)
	_, err := queryServer.GetReputation(suite.Ctx, &reputationtypes.QueryGetReputationRequest{
		Developer: []byte("nonexistent-developer"),
	})
	require.Error(t, err)
}
