package integration

import (
	"testing"

	"github.com/stretchr/testify/require"

	orgtypes "github.com/wevibe-network/wevibe-chain/x/org/types"
)

func TestMsgRegisterOrg_Integration(t *testing.T) {
	suite := NewTestSuite(t)

	msg := &orgtypes.MsgRegisterOrg{
		Signer:          suite.UserAddr.String(),
		Leader:          suite.UserAddr.String(),
		StorageQuota:    1000,
		RetrievalBudget: 500,
	}

	result, err := suite.DeliverMsg(msg)
	require.NoError(t, err)
	require.NotNil(t, result)

	orgID := orgtypes.FormatOrgID(0)

	org, err := suite.QueryOrg(orgID)
	require.NoError(t, err)
	require.Equal(t, orgID, org.OrgID)
	require.Equal(t, suite.UserAddr.String(), org.Leader)
	require.Equal(t, uint64(1000), org.StorageQuota)
	require.Equal(t, uint64(500), org.RetrievalBudget)
}

func TestMsgRegisterOrg_Duplicate(t *testing.T) {
	suite := NewTestSuite(t)

	msg := &orgtypes.MsgRegisterOrg{
		Signer:          suite.UserAddr.String(),
		Leader:          suite.UserAddr.String(),
		StorageQuota:    1000,
		RetrievalBudget: 500,
	}

	_, err := suite.DeliverMsg(msg)
	require.NoError(t, err)

	_, err = suite.DeliverMsg(msg)
	require.Error(t, err)
	require.ErrorIs(t, err, orgtypes.ErrLeaderAlreadyOwnsOrg)
}

func TestMsgAddMember_Integration(t *testing.T) {
	suite := NewTestSuite(t)

	orgMsg := &orgtypes.MsgRegisterOrg{
		Signer:          suite.UserAddr.String(),
		Leader:          suite.UserAddr.String(),
		StorageQuota:    1000,
		RetrievalBudget: 500,
		LeaderWallet:    suite.UserAddr.String(),
	}
	_, err := suite.DeliverMsg(orgMsg)
	require.NoError(t, err)
	orgID := orgtypes.FormatOrgID(0)

	memberMsg := &orgtypes.MsgAddMember{
		Signer: suite.UserAddr.String(),
		OrgId:  orgID,
		Pubkey: "wevibe1member1234567890123456789012345678901234567890",
		Role:   "member",
	}

	result, err := suite.DeliverMsg(memberMsg)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestMsgRemoveMember_Integration(t *testing.T) {
	suite := NewTestSuite(t)

	orgMsg := &orgtypes.MsgRegisterOrg{
		Signer:          suite.UserAddr.String(),
		Leader:          suite.UserAddr.String(),
		StorageQuota:    1000,
		RetrievalBudget: 500,
		LeaderWallet:    suite.UserAddr.String(),
	}
	_, err := suite.DeliverMsg(orgMsg)
	require.NoError(t, err)
	orgID := orgtypes.FormatOrgID(0)

	memberMsg := &orgtypes.MsgAddMember{
		Signer: suite.UserAddr.String(),
		OrgId:  orgID,
		Pubkey: "wevibe1member1234567890123456789012345678901234567890",
		Role:   "member",
	}
	_, err = suite.DeliverMsg(memberMsg)
	require.NoError(t, err)

	removeMsg := &orgtypes.MsgRemoveMember{
		Signer: suite.UserAddr.String(),
		OrgId:  orgID,
		Pubkey: "wevibe1member1234567890123456789012345678901234567890",
	}

	result, err := suite.DeliverMsg(removeMsg)
	require.NoError(t, err)
	require.NotNil(t, result)
}
