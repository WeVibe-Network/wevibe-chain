package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wevibe-network/wevibe-chain/x/org/types"
)

// ---------------------------------------------------------------------------
// MsgRegisterOrg.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgRegisterOrg_ValidateBasic(t *testing.T) {
	valid := &types.MsgRegisterOrg{
		Signer: "cosmos1signer",
		Leader: "leader_pubkey",
	}
	require.NoError(t, valid.ValidateBasic())

	t.Run("empty signer", func(t *testing.T) {
		m := *valid
		m.Signer = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("empty leader", func(t *testing.T) {
		m := *valid
		m.Leader = ""
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidLeader)
	})
}

// ---------------------------------------------------------------------------
// MsgAddMember.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgAddMember_ValidateBasic(t *testing.T) {
	valid := &types.MsgAddMember{
		Signer: "cosmos1signer",
		OrgId:  "org1",
		Pubkey: "member_pubkey",
		Role:   "member",
	}
	require.NoError(t, valid.ValidateBasic())

	t.Run("empty signer", func(t *testing.T) {
		m := *valid
		m.Signer = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("empty org id", func(t *testing.T) {
		m := *valid
		m.OrgId = ""
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidOrgID)
	})

	t.Run("empty pubkey", func(t *testing.T) {
		m := *valid
		m.Pubkey = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("empty role", func(t *testing.T) {
		m := *valid
		m.Role = ""
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidRole)
	})
}

// ---------------------------------------------------------------------------
// MsgRemoveMember.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgRemoveMember_ValidateBasic(t *testing.T) {
	valid := &types.MsgRemoveMember{
		Signer: "cosmos1signer",
		OrgId:  "org1",
		Pubkey: "member_pubkey",
	}
	require.NoError(t, valid.ValidateBasic())

	t.Run("empty signer", func(t *testing.T) {
		m := *valid
		m.Signer = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("empty org id", func(t *testing.T) {
		m := *valid
		m.OrgId = ""
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidOrgID)
	})

	t.Run("empty pubkey", func(t *testing.T) {
		m := *valid
		m.Pubkey = ""
		require.Error(t, m.ValidateBasic())
	})
}

// ---------------------------------------------------------------------------
// MsgUpdateParams.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgUpdateParams_ValidateBasic(t *testing.T) {
	params := types.DefaultParams()
	valid := &types.MsgUpdateParams{
		Authority: "cosmos1authority",
		Params:    &params,
	}
	require.NoError(t, valid.ValidateBasic())

	t.Run("empty authority", func(t *testing.T) {
		m := *valid
		m.Authority = ""
		require.Error(t, m.ValidateBasic())
	})
}

// ---------------------------------------------------------------------------
// MsgFundTreasury.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgFundTreasury_ValidateBasic(t *testing.T) {
	valid := &types.MsgFundTreasury{
		Signer: "cosmos1signer",
		OrgId:  "org1",
		Amount: "1000",
	}
	require.NoError(t, valid.ValidateBasic())

	t.Run("empty signer", func(t *testing.T) {
		m := *valid
		m.Signer = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("empty org id", func(t *testing.T) {
		m := *valid
		m.OrgId = ""
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidOrgID)
	})

	t.Run("empty amount", func(t *testing.T) {
		m := *valid
		m.Amount = ""
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidAmount)
	})

	t.Run("non-numeric amount", func(t *testing.T) {
		m := *valid
		m.Amount = "notanumber"
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidAmount)
	})

	t.Run("zero amount", func(t *testing.T) {
		m := *valid
		m.Amount = "0"
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidAmount)
	})

	t.Run("negative amount", func(t *testing.T) {
		m := *valid
		m.Amount = "-100"
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidAmount)
	})
}

// ---------------------------------------------------------------------------
// MsgWithdrawTreasury.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgWithdrawTreasury_ValidateBasic(t *testing.T) {
	valid := &types.MsgWithdrawTreasury{
		Signer:    "cosmos1signer",
		OrgId:     "org1",
		Amount:    "1000",
		Recipient: "cosmos1recipient",
	}
	require.NoError(t, valid.ValidateBasic())

	t.Run("empty signer", func(t *testing.T) {
		m := *valid
		m.Signer = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("empty org id", func(t *testing.T) {
		m := *valid
		m.OrgId = ""
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidOrgID)
	})

	t.Run("empty amount", func(t *testing.T) {
		m := *valid
		m.Amount = ""
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidAmount)
	})

	t.Run("non-numeric amount", func(t *testing.T) {
		m := *valid
		m.Amount = "xyz"
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidAmount)
	})

	t.Run("zero amount", func(t *testing.T) {
		m := *valid
		m.Amount = "0"
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidAmount)
	})

	t.Run("negative amount", func(t *testing.T) {
		m := *valid
		m.Amount = "-1"
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidAmount)
	})

	t.Run("empty recipient", func(t *testing.T) {
		m := *valid
		m.Recipient = ""
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidRecipient)
	})
}

// ---------------------------------------------------------------------------
// MsgSetRepTiers.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgSetRepTiers_ValidateBasic(t *testing.T) {
	valid := &types.MsgSetRepTiers{
		Signer: "cosmos1signer",
		OrgId:  "org1",
		Tiers: []*types.RepTier{
			{MinReputation: 0, MaxReputation: 50, MaxContributionsPerEpoch: 3, PayoutPerMemory: "1"},
			{MinReputation: 200, MaxReputation: 1000, MaxContributionsPerEpoch: 50, PayoutPerMemory: "5"},
		},
	}
	require.NoError(t, valid.ValidateBasic())

	t.Run("empty signer", func(t *testing.T) {
		m := *valid
		m.Signer = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("empty org id", func(t *testing.T) {
		m := *valid
		m.OrgId = ""
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidOrgID)
	})

	t.Run("no tiers", func(t *testing.T) {
		m := *valid
		m.Tiers = nil
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidRepTier)
	})

	t.Run("min greater than max", func(t *testing.T) {
		m := *valid
		m.Tiers = []*types.RepTier{
			{MinReputation: 100, MaxReputation: 50, PayoutPerMemory: "1"},
		}
		require.Error(t, m.ValidateBasic())
	})

	t.Run("empty payout", func(t *testing.T) {
		m := *valid
		m.Tiers = []*types.RepTier{
			{MinReputation: 0, MaxReputation: 50, PayoutPerMemory: ""},
		}
		require.Error(t, m.ValidateBasic())
	})

	t.Run("invalid payout", func(t *testing.T) {
		m := *valid
		m.Tiers = []*types.RepTier{
			{MinReputation: 0, MaxReputation: 50, PayoutPerMemory: "abc"},
		}
		require.Error(t, m.ValidateBasic())
	})

	t.Run("negative payout", func(t *testing.T) {
		m := *valid
		m.Tiers = []*types.RepTier{
			{MinReputation: 0, MaxReputation: 50, PayoutPerMemory: "-1"},
		}
		require.Error(t, m.ValidateBasic())
	})

	t.Run("overlapping tiers", func(t *testing.T) {
		m := *valid
		m.Tiers = []*types.RepTier{
			{MinReputation: 0, MaxReputation: 100, PayoutPerMemory: "1"},
			{MinReputation: 50, MaxReputation: 200, PayoutPerMemory: "2"},
		}
		require.ErrorIs(t, m.ValidateBasic(), types.ErrRepTierOverlap)
	})

	t.Run("zero payout is allowed", func(t *testing.T) {
		m := *valid
		m.Tiers = []*types.RepTier{
			{MinReputation: 0, MaxReputation: 50, PayoutPerMemory: "0"},
		}
		require.NoError(t, m.ValidateBasic())
	})
}

// ---------------------------------------------------------------------------
// MsgSetOrgConfig.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgSetOrgConfig_ValidateBasic(t *testing.T) {
	valid := &types.MsgSetOrgConfig{
		Signer:                   "cosmos1signer",
		OrgId:                    "org1",
		MinContributionsPerEpoch: 10,
	}
	require.NoError(t, valid.ValidateBasic())

	t.Run("empty signer", func(t *testing.T) {
		m := *valid
		m.Signer = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("empty org id", func(t *testing.T) {
		m := *valid
		m.OrgId = ""
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidOrgID)
	})

	t.Run("boundary 100 allowed", func(t *testing.T) {
		m := *valid
		m.MinContributionsPerEpoch = 100
		require.NoError(t, m.ValidateBasic())
	})

	t.Run("min contributions above limit", func(t *testing.T) {
		m := *valid
		m.MinContributionsPerEpoch = 101
		require.Error(t, m.ValidateBasic())
	})
}

// ---------------------------------------------------------------------------
// MsgGrantTrialAllowance.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgGrantTrialAllowance_ValidateBasic(t *testing.T) {
	valid := &types.MsgGrantTrialAllowance{
		Signer:           "cosmos1signer",
		OrgId:            "org1",
		Grantee:          "cosmos1grantee",
		DailySubmissions: 5,
		TrialDays:        7,
	}
	require.NoError(t, valid.ValidateBasic())

	t.Run("empty signer", func(t *testing.T) {
		m := *valid
		m.Signer = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("empty org id", func(t *testing.T) {
		m := *valid
		m.OrgId = ""
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidOrgID)
	})

	t.Run("empty grantee", func(t *testing.T) {
		m := *valid
		m.Grantee = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("zero daily submissions", func(t *testing.T) {
		m := *valid
		m.DailySubmissions = 0
		require.Error(t, m.ValidateBasic())
	})

	t.Run("zero trial days", func(t *testing.T) {
		m := *valid
		m.TrialDays = 0
		require.Error(t, m.ValidateBasic())
	})
}

// ---------------------------------------------------------------------------
// MsgUpdateMemberRole.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgUpdateMemberRole_ValidateBasic(t *testing.T) {
	valid := &types.MsgUpdateMemberRole{
		Signer:  "cosmos1signer",
		OrgId:   "org1",
		Pubkey:  "member_pubkey",
		NewRole: "moderator",
	}
	require.NoError(t, valid.ValidateBasic())

	t.Run("member role allowed", func(t *testing.T) {
		m := *valid
		m.NewRole = "member"
		require.NoError(t, m.ValidateBasic())
	})

	t.Run("empty signer", func(t *testing.T) {
		m := *valid
		m.Signer = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("empty org id", func(t *testing.T) {
		m := *valid
		m.OrgId = ""
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidOrgID)
	})

	t.Run("empty pubkey", func(t *testing.T) {
		m := *valid
		m.Pubkey = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("invalid role", func(t *testing.T) {
		m := *valid
		m.NewRole = "leader"
		require.Error(t, m.ValidateBasic())
	})

	t.Run("empty role", func(t *testing.T) {
		m := *valid
		m.NewRole = ""
		require.Error(t, m.ValidateBasic())
	})
}

// ---------------------------------------------------------------------------
// MsgRotateEpoch.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgRotateEpoch_ValidateBasic(t *testing.T) {
	valid := &types.MsgRotateEpoch{
		Signer: "cosmos1signer",
		OrgId:  "org1",
	}
	require.NoError(t, valid.ValidateBasic())

	t.Run("empty signer", func(t *testing.T) {
		m := *valid
		m.Signer = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("empty org id", func(t *testing.T) {
		m := *valid
		m.OrgId = ""
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidOrgID)
	})
}

// ---------------------------------------------------------------------------
// MsgTransferLeadership.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgTransferLeadership_ValidateBasic(t *testing.T) {
	valid := &types.MsgTransferLeadership{
		Signer:    "cosmos1signer",
		OrgId:     "org1",
		NewLeader: "cosmos1newleader",
	}
	require.NoError(t, valid.ValidateBasic())

	t.Run("empty signer", func(t *testing.T) {
		m := *valid
		m.Signer = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("empty org id", func(t *testing.T) {
		m := *valid
		m.OrgId = ""
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidOrgID)
	})

	t.Run("empty new leader", func(t *testing.T) {
		m := *valid
		m.NewLeader = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("transfer to self", func(t *testing.T) {
		m := *valid
		m.NewLeader = m.Signer
		require.Error(t, m.ValidateBasic())
	})
}

// ---------------------------------------------------------------------------
// MsgCloseOrg.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgCloseOrg_ValidateBasic(t *testing.T) {
	valid := &types.MsgCloseOrg{
		Signer: "cosmos1signer",
		OrgId:  "org1",
	}
	require.NoError(t, valid.ValidateBasic())

	t.Run("empty signer", func(t *testing.T) {
		m := *valid
		m.Signer = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("empty org id", func(t *testing.T) {
		m := *valid
		m.OrgId = ""
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidOrgID)
	})
}

// ---------------------------------------------------------------------------
// Params
// ---------------------------------------------------------------------------

func TestDefaultParams(t *testing.T) {
	p := types.DefaultParams()
	require.Equal(t, uint64(10000000), p.MinRegistrationFee)
	require.Equal(t, uint64(5000000), p.AnnualRenewalFee)
	require.Equal(t, uint64(1073741824), p.DefaultStorageQuota)
	require.Equal(t, uint64(10000), p.DefaultRetrievalBudget)
	require.Equal(t, uint64(30), p.GracePeriodEpochs)
	require.Equal(t, uint64(10), p.BurnPriceDecayEpochs)
	require.Equal(t, uint64(10000000), p.BaseBurnPrice)
	require.Equal(t, uint64(20), p.BurnPriceIncreasePercent)
}

func TestParamsValidate(t *testing.T) {
	require.NoError(t, types.DefaultParams().Validate())

	// zero-value params are also valid (Validate is currently a no-op).
	require.NoError(t, (types.Params{}).Validate())
}

// ---------------------------------------------------------------------------
// Genesis helpers
// ---------------------------------------------------------------------------

func TestNewGenesisState(t *testing.T) {
	orgs := []*types.Org{types.NewOrg("org1", "leader", "", 1, 1)}
	members := []*types.MemberRecord{types.NewMemberRecord("org1", "pk", "member")}
	dp := &types.DynamicPrice{Price: 1000}

	gs := types.NewGenesisState(orgs, members, dp)
	require.NotNil(t, gs)
	require.Len(t, gs.Orgs, 1)
	require.Len(t, gs.Members, 1)
	require.Equal(t, dp, gs.DynamicPrice)
	// fields not passed to the constructor stay empty.
	require.Nil(t, gs.Treasuries)
	require.Nil(t, gs.RepTiers)
	require.Nil(t, gs.OrgConfigs)
}

func TestNewGenesisState_Empty(t *testing.T) {
	gs := types.NewGenesisState(nil, nil, nil)
	require.NotNil(t, gs)
	require.Empty(t, gs.Orgs)
	require.Empty(t, gs.Members)
	require.Nil(t, gs.DynamicPrice)
}

// ---------------------------------------------------------------------------
// keys.go helpers
// ---------------------------------------------------------------------------

func TestNewOrg(t *testing.T) {
	o := types.NewOrg("org1", "leader", "example.com", 1024, 512)
	require.Equal(t, "org1", o.OrgID)
	require.Equal(t, "leader", o.Leader)
	require.Equal(t, "example.com", o.Domain)
	require.Equal(t, uint64(1024), o.StorageQuota)
	require.Equal(t, uint64(512), o.RetrievalBudget)
	require.Equal(t, types.OrgStatus_ACTIVE, o.Status)
	require.True(t, o.IsActive())
}

func TestOrgValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		o := types.NewOrg("org1", "leader", "example.com", 1, 1)
		require.NoError(t, o.Validate())
	})

	t.Run("valid empty domain", func(t *testing.T) {
		o := types.NewOrg("org1", "leader", "", 1, 1)
		require.NoError(t, o.Validate())
	})

	t.Run("empty org id", func(t *testing.T) {
		o := types.NewOrg("", "leader", "", 1, 1)
		require.ErrorIs(t, o.Validate(), types.ErrInvalidOrgID)
	})

	t.Run("empty leader", func(t *testing.T) {
		o := types.NewOrg("org1", "", "", 1, 1)
		require.ErrorIs(t, o.Validate(), types.ErrInvalidLeader)
	})

	t.Run("domain too long", func(t *testing.T) {
		long := ""
		for i := 0; i < 129; i++ {
			long += "a"
		}
		o := types.NewOrg("org1", "leader", long, 1, 1)
		require.ErrorIs(t, o.Validate(), types.ErrInvalidDomain)
	})

	t.Run("domain with invalid character", func(t *testing.T) {
		o := types.NewOrg("org1", "leader", "bad domain!", 1, 1)
		require.ErrorIs(t, o.Validate(), types.ErrInvalidDomain)
	})

	t.Run("domain with allowed characters", func(t *testing.T) {
		o := types.NewOrg("org1", "leader", "React, Next.js / TypeScript + Node & Go", 1, 1)
		require.NoError(t, o.Validate())
	})
}

func TestOrgIsActive(t *testing.T) {
	o := types.NewOrg("org1", "leader", "", 1, 1)
	require.True(t, o.IsActive())

	o.Status = types.OrgStatus_DORMANT
	require.False(t, o.IsActive())

	o.Status = types.OrgStatus_CLOSED
	require.False(t, o.IsActive())
}

func TestNewMemberRecord(t *testing.T) {
	m := types.NewMemberRecord("org1", "pk", "moderator")
	require.Equal(t, "org1", m.OrgID)
	require.Equal(t, "pk", m.Pubkey)
	require.Equal(t, "moderator", m.Role)
}

func TestMemberRecordMemberKey(t *testing.T) {
	m := types.NewMemberRecord("org1", "pk", "member")
	key := m.MemberKey()
	require.Equal(t, types.MemberKey{OrgID: "org1", Member: "pk"}, key)
}

func TestMemberRecordMemberKey_Empty(t *testing.T) {
	m := types.NewMemberRecord("", "", "")
	key := m.MemberKey()
	require.Equal(t, types.MemberKey{}, key)
}
