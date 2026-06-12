package types_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wevibe-network/wevibe-chain/x/org/types"
)

const validX25519Pubkey = "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"

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

func TestValidateHubResponsePubkey(t *testing.T) {
	valid := "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"
	require.NoError(t, types.ValidateHubResponsePubkey(""))
	require.NoError(t, types.ValidateHubResponsePubkey(valid))

	tests := []struct {
		name   string
		pubkey string
	}{
		{name: "odd length", pubkey: "abc"},
		{name: "non-hex", pubkey: "zz112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"},
		{name: "31 bytes", pubkey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddee"},
		{name: "33 bytes", pubkey: valid + "00"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateHubResponsePubkey(tc.pubkey)
			require.ErrorIs(t, err, types.ErrInvalidHubResponsePubkey)
		})
	}
}

// ---------------------------------------------------------------------------
// MsgAddMember.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgAddMember_ValidateBasic(t *testing.T) {
	valid := &types.MsgAddMember{
		Signer:       "cosmos1signer",
		OrgId:        "org1",
		Pubkey:       "member_pubkey",
		Role:         "member",
		X25519Pubkey: validX25519Pubkey,
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

	t.Run("empty x25519 pubkey", func(t *testing.T) {
		m := *valid
		m.X25519Pubkey = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("empty role", func(t *testing.T) {
		m := *valid
		m.Role = ""
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidRole)
	})

	t.Run("invalid role", func(t *testing.T) {
		m := *valid
		m.Role = "leader"
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
// MsgSetOrgConfig.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgSetOrgConfig_ValidateBasic(t *testing.T) {
	valid := &types.MsgSetOrgConfig{
		Signer:                   "cosmos1signer",
		OrgId:                    "org1",
		VocabHash:                "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
		EmbeddingModelId:         "text-embedding-3-large",
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

	t.Run("empty vocab hash allowed", func(t *testing.T) {
		m := *valid
		m.VocabHash = ""
		require.NoError(t, m.ValidateBasic())
	})

	t.Run("vocab hash wrong length", func(t *testing.T) {
		m := *valid
		m.VocabHash = "00112233445566778899aabbccddeeff00112233445566778899aabbccddee"
		require.Error(t, m.ValidateBasic())
	})

	t.Run("vocab hash uppercase rejected", func(t *testing.T) {
		m := *valid
		m.VocabHash = "00112233445566778899AABBCCDDEEFF00112233445566778899AABBCCDDEEFF"
		require.Error(t, m.ValidateBasic())
	})

	t.Run("vocab hash non-hex rejected", func(t *testing.T) {
		m := *valid
		m.VocabHash = "g0112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"
		require.Error(t, m.ValidateBasic())
	})

	t.Run("empty embedding model id allowed", func(t *testing.T) {
		m := *valid
		m.EmbeddingModelId = ""
		require.NoError(t, m.ValidateBasic())
	})

	t.Run("embedding model id too long", func(t *testing.T) {
		m := *valid
		m.EmbeddingModelId = strings.Repeat("a", 129)
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
		Signer:          "cosmos1signer",
		OrgId:           "org1",
		NewLeader:       "cosmos1newleader",
		NewLeaderWallet: "cosmos1gsank9k6ygfnx376cuhw8zp9p8ssnyez44dtmh",
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

	t.Run("empty new leader wallet", func(t *testing.T) {
		m := *valid
		m.NewLeaderWallet = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("invalid new leader wallet", func(t *testing.T) {
		m := *valid
		m.NewLeaderWallet = "not-a-bech32-address"
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
	require.Equal(t, uint64(32), p.SlotCap)
}

func TestParamsValidate(t *testing.T) {
	require.NoError(t, types.DefaultParams().Validate())

	zero := types.DefaultParams()
	zero.SlotCap = 0
	require.Error(t, zero.Validate())
}

// ---------------------------------------------------------------------------
// Genesis helpers
// ---------------------------------------------------------------------------

func TestNewGenesisState(t *testing.T) {
	orgs := []*types.Org{types.NewOrg("org1", "leader", "", "", "", "", 1, 1)}
	members := []*types.MemberRecord{types.NewMemberRecord("org1", "pk", "member", validX25519Pubkey)}

	gs := types.NewGenesisState(orgs, members)
	require.NotNil(t, gs)
	require.Len(t, gs.Orgs, 1)
	require.Len(t, gs.Members, 1)
	// fields not passed to the constructor stay empty.
	require.Nil(t, gs.OrgConfigs)
}

func TestNewGenesisState_Empty(t *testing.T) {
	gs := types.NewGenesisState(nil, nil)
	require.NotNil(t, gs)
	require.Empty(t, gs.Orgs)
	require.Empty(t, gs.Members)
}

// ---------------------------------------------------------------------------
// keys.go helpers
// ---------------------------------------------------------------------------

func TestNewOrg(t *testing.T) {
	o := types.NewOrg("org1", "leader", "example.com", "A test org", "Go, TypeScript", "AI, Security", 1024, 512)
	require.Equal(t, "org1", o.OrgID)
	require.Equal(t, "leader", o.Leader)
	require.Equal(t, "example.com", o.Domain)
	require.Equal(t, "A test org", o.Description)
	require.Equal(t, "Go, TypeScript", o.TechStack)
	require.Equal(t, "AI, Security", o.FocusAreas)
	require.Equal(t, uint64(1024), o.StorageQuota)
	require.Equal(t, uint64(512), o.RetrievalBudget)
	require.Equal(t, types.OrgStatus_ACTIVE, o.Status)
	require.True(t, o.IsActive())
}

func TestOrgAccountAddress_Deterministic(t *testing.T) {
	orgID := types.FormatOrgID(7)
	addr1 := types.OrgAccountAddress(orgID)
	addr2 := types.OrgAccountAddress(orgID)

	require.Equal(t, addr1.String(), addr2.String())
	require.NotEmpty(t, addr1.String())
}

func TestOrgValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		o := types.NewOrg("org1", "leader", "example.com", "", "", "", 1, 1)
		require.NoError(t, o.Validate())
	})

	t.Run("valid empty domain", func(t *testing.T) {
		o := types.NewOrg("org1", "leader", "", "", "", "", 1, 1)
		require.NoError(t, o.Validate())
	})

	t.Run("empty org id", func(t *testing.T) {
		o := types.NewOrg("", "leader", "", "", "", "", 1, 1)
		require.ErrorIs(t, o.Validate(), types.ErrInvalidOrgID)
	})

	t.Run("empty leader", func(t *testing.T) {
		o := types.NewOrg("org1", "", "", "", "", "", 1, 1)
		require.ErrorIs(t, o.Validate(), types.ErrInvalidLeader)
	})

	t.Run("domain too long", func(t *testing.T) {
		long := ""
		for i := 0; i < 129; i++ {
			long += "a"
		}
		o := types.NewOrg("org1", "leader", long, "", "", "", 1, 1)
		require.ErrorIs(t, o.Validate(), types.ErrInvalidDomain)
	})

	t.Run("domain with invalid character", func(t *testing.T) {
		o := types.NewOrg("org1", "leader", "bad domain!", "", "", "", 1, 1)
		require.ErrorIs(t, o.Validate(), types.ErrInvalidDomain)
	})

	t.Run("domain with allowed characters", func(t *testing.T) {
		o := types.NewOrg("org1", "leader", "React, Next.js / TypeScript + Node & Go", "", "", "", 1, 1)
		require.NoError(t, o.Validate())
	})

	t.Run("description too long", func(t *testing.T) {
		o := types.NewOrg("org1", "leader", "", strings.Repeat("a", 501), "", "", 1, 1)
		require.ErrorIs(t, o.Validate(), types.ErrInvalidDescription)
	})

	t.Run("description with control character", func(t *testing.T) {
		o := types.NewOrg("org1", "leader", "", "line1\nline2", "", "", 1, 1)
		require.ErrorIs(t, o.Validate(), types.ErrInvalidDescription)
	})

	t.Run("tech stack too long", func(t *testing.T) {
		o := types.NewOrg("org1", "leader", "", "", strings.Repeat("b", 201), "", 1, 1)
		require.ErrorIs(t, o.Validate(), types.ErrInvalidTechStack)
	})

	t.Run("tech stack with control character", func(t *testing.T) {
		o := types.NewOrg("org1", "leader", "", "", "Go\tTypeScript", "", 1, 1)
		require.ErrorIs(t, o.Validate(), types.ErrInvalidTechStack)
	})

	t.Run("focus areas too long", func(t *testing.T) {
		o := types.NewOrg("org1", "leader", "", "", "", strings.Repeat("c", 201), 1, 1)
		require.ErrorIs(t, o.Validate(), types.ErrInvalidFocusAreas)
	})

	t.Run("focus areas with control character", func(t *testing.T) {
		o := types.NewOrg("org1", "leader", "", "", "", "ai\x1fml", 1, 1)
		require.ErrorIs(t, o.Validate(), types.ErrInvalidFocusAreas)
	})
}

func TestOrgIsActive(t *testing.T) {
	o := types.NewOrg("org1", "leader", "", "", "", "", 1, 1)
	require.True(t, o.IsActive())

	o.Status = types.OrgStatus_DORMANT
	require.False(t, o.IsActive())

	o.Status = types.OrgStatus_CLOSED
	require.False(t, o.IsActive())
}

func TestNewMemberRecord(t *testing.T) {
	m := types.NewMemberRecord("org1", "pk", "moderator", validX25519Pubkey)
	require.Equal(t, "org1", m.OrgID)
	require.Equal(t, "pk", m.Pubkey)
	require.Equal(t, "moderator", m.Role)
	require.Equal(t, validX25519Pubkey, m.X25519Pubkey)
}

func TestMemberRecordMemberKey(t *testing.T) {
	m := types.NewMemberRecord("org1", "pk", "member", validX25519Pubkey)
	key := m.MemberKey()
	require.Equal(t, types.MemberKey{OrgID: "org1", Member: "pk"}, key)
}

func TestMemberRecordMemberKey_Empty(t *testing.T) {
	m := types.NewMemberRecord("", "", "", "")
	key := m.MemberKey()
	require.Equal(t, types.MemberKey{}, key)
}
