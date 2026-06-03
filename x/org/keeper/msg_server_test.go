package keeper_test

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/wevibe-network/wevibe-chain/testutil/keeper"
	"github.com/wevibe-network/wevibe-chain/x/org/keeper"
	"github.com/wevibe-network/wevibe-chain/x/org/types"
)

var (
	validSigner    = "cosmos1vq0svzat0jyknkc6rfp40l8tr5cz4qxd6m6tyx"
	validLeader    = "cosmos1t9xdz4tvmsm2qj8fxadue6yx5ysp30zv4rnau6"
	validMember    = "cosmos1gsank9k6ygfnx376cuhw8zp9p8ssnyez44dtmh"
	validRecipient = "cosmos1wnrlmtryha3hwrep5rn5hlu7jk3g8jka2ff505"
	validAuthority = "cosmos14taukd54w5eak58yjv4lpzz3a0vr0petthfpc5"
)

type mockFeegrantKeeper struct {
	lastGranter   sdk.AccAddress
	lastGrantee   sdk.AccAddress
	lastAllowance feegrant.FeeAllowanceI
	err           error
}

func newMockFeegrantKeeper() *mockFeegrantKeeper {
	return &mockFeegrantKeeper{}
}

func (m *mockFeegrantKeeper) GrantAllowance(ctx context.Context, granter, grantee sdk.AccAddress, allowance feegrant.FeeAllowanceI) error {
	m.lastGranter = granter
	m.lastGrantee = grantee
	m.lastAllowance = allowance
	return m.err
}

func setupMsgServer(t *testing.T) (types.MsgServer, context.Context, *mockBankKeeper, *mockFeegrantKeeper) {
	storeKey := storetypes.NewKVStoreKey("org")
	storeService, cms := testkeeper.NewTestStoreService(t, storeKey)
	logger := testkeeper.NewTestLogger()
	bank := newMockBankKeeper()
	feegrantKeeper := newMockFeegrantKeeper()
	k := keeper.NewKeeper(storeService, logger, validAuthority, bank, feegrantKeeper)
	sdkCtx := sdk.NewContext(cms, tmproto.Header{Height: 1, Time: time.Now().UTC()}, false, logger)
	return keeper.NewMsgServerImpl(k), sdk.WrapSDKContext(sdkCtx), bank, feegrantKeeper
}

func mustDeriveOrgID(t *testing.T, signer string) string {
	t.Helper()
	addr, err := sdk.AccAddressFromBech32(signer)
	require.NoError(t, err)
	return types.DeriveOrgID(addr)
}

func TestMsgRegisterOrg_ValidateBasic(t *testing.T) {
	msg := &types.MsgRegisterOrg{}
	require.Error(t, msg.ValidateBasic())

	msg.Signer = validSigner
	msg.Leader = validLeader
	require.NoError(t, msg.ValidateBasic())
}

func TestMsgGrantTrialAllowance_ValidateBasic(t *testing.T) {
	msg := &types.MsgGrantTrialAllowance{}
	require.Error(t, msg.ValidateBasic())

	msg.Signer = validLeader
	msg.OrgId = "org1"
	msg.Grantee = validMember
	msg.DailySubmissions = 1
	msg.TrialDays = 1
	require.NoError(t, msg.ValidateBasic())
}

func TestMsgRegisterOrg_Success(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	msg := &types.MsgRegisterOrg{
		Signer:          validSigner,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	}

	resp, err := srv.RegisterOrg(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, mustDeriveOrgID(t, validSigner), resp.OrgId)
}

func TestMsgRegisterOrg_Duplicate(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	msg := &types.MsgRegisterOrg{
		Signer:          validSigner,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	}

	_, err := srv.RegisterOrg(ctx, msg)
	require.NoError(t, err)

	_, err = srv.RegisterOrg(ctx, msg)
	require.Error(t, err)
}

func TestMsgAddMember_ValidateBasic(t *testing.T) {
	msg := &types.MsgAddMember{}
	require.Error(t, msg.ValidateBasic())

	msg.Signer = validSigner
	msg.OrgId = "org1"
	msg.Pubkey = validMember
	msg.Role = "member"
	require.NoError(t, msg.ValidateBasic())
}

func TestMsgAddMember_Success(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeader, validLeader)

	memberMsg := &types.MsgAddMember{
		Signer: validLeader,
		OrgId:  orgID,
		Pubkey: validMember,
		Role:   "member",
	}

	resp, err := srv.AddMember(ctx, memberMsg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgRemoveMember_ValidateBasic(t *testing.T) {
	msg := &types.MsgRemoveMember{}
	require.Error(t, msg.ValidateBasic())

	msg.Signer = validSigner
	msg.OrgId = "org1"
	msg.Pubkey = validMember
	require.NoError(t, msg.ValidateBasic())
}

func TestMsgRemoveMember_NotFound(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	msg := &types.MsgRemoveMember{
		Signer: validSigner,
		OrgId:  "org1",
		Pubkey: validMember,
	}

	_, err := srv.RemoveMember(ctx, msg)
	require.Error(t, err)
}

func TestMsgFundTreasury_ValidateBasic(t *testing.T) {
	tests := []struct {
		name        string
		msg         *types.MsgFundTreasury
		expectError bool
	}{
		{
			name: "empty signer",
			msg: &types.MsgFundTreasury{
				OrgId:  "org1",
				Amount: "1000",
			},
			expectError: true,
		},
		{
			name: "empty org_id",
			msg: &types.MsgFundTreasury{
				Signer: validSigner,
				Amount: "1000",
			},
			expectError: true,
		},
		{
			name: "empty amount",
			msg: &types.MsgFundTreasury{
				Signer: validSigner,
				OrgId:  "org1",
			},
			expectError: true,
		},
		{
			name: "negative amount",
			msg: &types.MsgFundTreasury{
				Signer: validSigner,
				OrgId:  "org1",
				Amount: "-1000",
			},
			expectError: true,
		},
		{
			name: "valid",
			msg: &types.MsgFundTreasury{
				Signer: validSigner,
				OrgId:  "org1",
				Amount: "1000",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgFundTreasury_Success(t *testing.T) {
	srv, ctx, bank, _ := setupMsgServer(t)

	bank.SetBalance(validSigner, math.NewInt(1000000))

	orgMsg := &types.MsgRegisterOrg{
		Signer:          validSigner,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	}
	_, err := srv.RegisterOrg(ctx, orgMsg)
	require.NoError(t, err)
	orgID := mustDeriveOrgID(t, validSigner)

	treasuryMsg := &types.MsgFundTreasury{
		Signer: validSigner,
		OrgId:  orgID,
		Amount: "500000",
	}

	resp, err := srv.FundTreasury(ctx, treasuryMsg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgWithdrawTreasury_NotLeader(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgMsg := &types.MsgRegisterOrg{
		Signer:          validSigner,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	}
	_, err := srv.RegisterOrg(ctx, orgMsg)
	require.NoError(t, err)
	orgID := mustDeriveOrgID(t, validSigner)

	treasuryMsg := &types.MsgFundTreasury{
		Signer: validSigner,
		OrgId:  orgID,
		Amount: "500000",
	}
	_, err = srv.FundTreasury(ctx, treasuryMsg)
	require.NoError(t, err)

	withdrawMsg := &types.MsgWithdrawTreasury{
		Signer:    validSigner,
		OrgId:     orgID,
		Amount:    "100000",
		Recipient: validRecipient,
	}

	_, err = srv.WithdrawTreasury(ctx, withdrawMsg)
	require.Error(t, err)
	require.Equal(t, types.ErrNotLeader, err)
}

func TestMsgSetRepTiers_Success(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgMsg := &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	}
	_, err := srv.RegisterOrg(ctx, orgMsg)
	require.NoError(t, err)
	orgID := mustDeriveOrgID(t, validLeader)

	setRepTiersMsg := &types.MsgSetRepTiers{
		Signer: validLeader,
		OrgId:  orgID,
		Tiers: []*types.RepTier{
			{
				MinReputation:            0,
				MaxReputation:            50,
				MaxContributionsPerEpoch: 3,
				PayoutPerMemory:          "1",
			},
		},
	}

	resp, err := srv.SetRepTiers(ctx, setRepTiersMsg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgSetRepTiers_NotLeader(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgMsg := &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	}
	_, err := srv.RegisterOrg(ctx, orgMsg)
	require.NoError(t, err)
	orgID := mustDeriveOrgID(t, validLeader)

	setRepTiersMsg := &types.MsgSetRepTiers{
		Signer: validSigner,
		OrgId:  orgID,
		Tiers: []*types.RepTier{
			{
				MinReputation:            0,
				MaxReputation:            50,
				MaxContributionsPerEpoch: 3,
				PayoutPerMemory:          "1",
			},
		},
	}

	_, err = srv.SetRepTiers(ctx, setRepTiersMsg)
	require.Error(t, err)
	require.Equal(t, types.ErrNotLeader, err)
}

func TestMsgSetOrgConfig_Success(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgMsg := &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	}
	_, err := srv.RegisterOrg(ctx, orgMsg)
	require.NoError(t, err)
	orgID := mustDeriveOrgID(t, validLeader)

	setOrgConfigMsg := &types.MsgSetOrgConfig{
		Signer:                   validLeader,
		OrgId:                    orgID,
		ServeAttestationRequired: true,
		MinContributionsPerEpoch: 10,
	}

	resp, err := srv.SetOrgConfig(ctx, setOrgConfigMsg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgGrantTrialAllowance_Success(t *testing.T) {
	srv, ctx, _, feegrantKeeper := setupMsgServer(t)

	_, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validSigner,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	})
	require.NoError(t, err)
	orgID := mustDeriveOrgID(t, validSigner)

	msg := &types.MsgGrantTrialAllowance{
		Signer:           validLeader,
		OrgId:            orgID,
		Grantee:          validMember,
		DailySubmissions: 3,
		TrialDays:        2,
	}

	resp, err := srv.GrantTrialAllowance(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	require.NotNil(t, feegrantKeeper.lastAllowance)
	require.Equal(t, validLeader, feegrantKeeper.lastGranter.String())
	require.Equal(t, validMember, feegrantKeeper.lastGrantee.String())

	periodic, ok := feegrantKeeper.lastAllowance.(*feegrant.PeriodicAllowance)
	require.True(t, ok)
	require.Equal(t, 24*time.Hour, periodic.Period)

	dailyExpected := sdk.NewCoin("uvibe", math.NewInt(2000*3))
	require.Equal(t, sdk.NewCoins(dailyExpected), periodic.PeriodSpendLimit)
	require.Equal(t, sdk.NewCoins(dailyExpected), periodic.PeriodCanSpend)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("uvibe", math.NewInt(2000*3*2))), periodic.Basic.SpendLimit)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	require.NotNil(t, periodic.Basic.Expiration)
	require.True(t, periodic.Basic.Expiration.After(sdkCtx.BlockTime()))
}

func TestMsgGrantTrialAllowance_NotLeader(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	_, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validSigner,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	})
	require.NoError(t, err)
	orgID := mustDeriveOrgID(t, validSigner)

	msg := &types.MsgGrantTrialAllowance{
		Signer:           validSigner,
		OrgId:            orgID,
		Grantee:          validMember,
		DailySubmissions: 1,
		TrialDays:        1,
	}

	_, err = srv.GrantTrialAllowance(ctx, msg)
	require.ErrorIs(t, err, types.ErrNotLeader)
}

func TestMsgUpdateMemberRole_Success(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeader, validLeader)

	_, err := srv.AddMember(ctx, &types.MsgAddMember{
		Signer: validLeader,
		OrgId:  orgID,
		Pubkey: validMember,
		Role:   "member",
	})
	require.NoError(t, err)

	msg := &types.MsgUpdateMemberRole{
		Signer:  validLeader,
		OrgId:   orgID,
		Pubkey:  validMember,
		NewRole: "moderator",
	}

	resp, err := srv.UpdateMemberRole(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgUpdateMemberRole_NotLeader(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeader, validLeader)

	_, err := srv.AddMember(ctx, &types.MsgAddMember{
		Signer: validLeader,
		OrgId:  orgID,
		Pubkey: validMember,
		Role:   "member",
	})
	require.NoError(t, err)

	msg := &types.MsgUpdateMemberRole{
		Signer:  validSigner,
		OrgId:   orgID,
		Pubkey:  validMember,
		NewRole: "moderator",
	}

	_, err = srv.UpdateMemberRole(ctx, msg)
	require.ErrorIs(t, err, types.ErrNotLeader)
}

func TestMsgUpdateMemberRole_MemberNotFound(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	_, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	})
	require.NoError(t, err)
	orgID := mustDeriveOrgID(t, validLeader)

	msg := &types.MsgUpdateMemberRole{
		Signer:  validLeader,
		OrgId:   orgID,
		Pubkey:  validMember,
		NewRole: "moderator",
	}

	_, err = srv.UpdateMemberRole(ctx, msg)
	require.ErrorIs(t, err, types.ErrMemberNotFound)
}

func TestMsgUpdateMemberRole_CannotChangeLeaderRole(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	_, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	})
	require.NoError(t, err)
	orgID := mustDeriveOrgID(t, validLeader)

	msg := &types.MsgUpdateMemberRole{
		Signer:  validLeader,
		OrgId:   orgID,
		Pubkey:  validLeader,
		NewRole: "member",
	}

	_, err = srv.UpdateMemberRole(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot change role of org leader")
}

func TestMsgRotateEpoch_Success(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeader, validLeader)

	msg := &types.MsgRotateEpoch{
		Signer: validLeader,
		OrgId:  orgID,
	}

	resp, err := srv.RotateEpoch(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(1), resp.NewEpoch)
}

func TestMsgRotateEpoch_NotLeader(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	_, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	})
	require.NoError(t, err)
	orgID := mustDeriveOrgID(t, validLeader)

	msg := &types.MsgRotateEpoch{
		Signer: validSigner,
		OrgId:  orgID,
	}

	_, err = srv.RotateEpoch(ctx, msg)
	require.ErrorIs(t, err, types.ErrNotLeader)
}

func TestMsgRotateEpoch_ClosedOrg(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeader, validLeader)

	_, err := srv.CloseOrg(ctx, &types.MsgCloseOrg{
		Signer: validLeader,
		OrgId:  orgID,
	})
	require.NoError(t, err)

	msg := &types.MsgRotateEpoch{
		Signer: validLeader,
		OrgId:  orgID,
	}

	_, err = srv.RotateEpoch(ctx, msg)
	require.ErrorIs(t, err, types.ErrOrgNotActive)
}

func TestMsgTransferLeadership_Success(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeader, validLeader)

	_, err := srv.AddMember(ctx, &types.MsgAddMember{
		Signer: validLeader,
		OrgId:  orgID,
		Pubkey: validMember,
		Role:   "member",
	})
	require.NoError(t, err)

	msg := &types.MsgTransferLeadership{
		Signer:    validLeader,
		OrgId:     orgID,
		NewLeader: validMember,
	}

	resp, err := srv.TransferLeadership(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgTransferLeadership_NotLeader(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	_, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	})
	require.NoError(t, err)
	orgID := mustDeriveOrgID(t, validLeader)

	msg := &types.MsgTransferLeadership{
		Signer:    validSigner,
		OrgId:     orgID,
		NewLeader: validMember,
	}

	_, err = srv.TransferLeadership(ctx, msg)
	require.ErrorIs(t, err, types.ErrNotLeader)
}

func TestMsgTransferLeadership_NewLeaderNotMember(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	_, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	})
	require.NoError(t, err)
	orgID := mustDeriveOrgID(t, validLeader)

	msg := &types.MsgTransferLeadership{
		Signer:    validLeader,
		OrgId:     orgID,
		NewLeader: validMember,
	}

	_, err = srv.TransferLeadership(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "new_leader must be a member of the org")
}

func TestMsgTransferLeadership_SelfTransfer(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	_, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	})
	require.NoError(t, err)
	orgID := mustDeriveOrgID(t, validLeader)

	msg := &types.MsgTransferLeadership{
		Signer:    validLeader,
		OrgId:     orgID,
		NewLeader: validLeader,
	}

	err = msg.ValidateBasic()
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot transfer leadership to self")
}

func TestMsgCloseOrg_Success(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	_, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	})
	require.NoError(t, err)
	orgID := mustDeriveOrgID(t, validLeader)

	msg := &types.MsgCloseOrg{
		Signer: validLeader,
		OrgId:  orgID,
	}

	resp, err := srv.CloseOrg(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgCloseOrg_NotLeader(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	_, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	})
	require.NoError(t, err)
	orgID := mustDeriveOrgID(t, validLeader)

	msg := &types.MsgCloseOrg{
		Signer: validSigner,
		OrgId:  orgID,
	}

	_, err = srv.CloseOrg(ctx, msg)
	require.ErrorIs(t, err, types.ErrNotLeader)
}

func TestMsgCloseOrg_AlreadyClosed(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	_, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	})
	require.NoError(t, err)
	orgID := mustDeriveOrgID(t, validLeader)

	_, err = srv.CloseOrg(ctx, &types.MsgCloseOrg{
		Signer: validLeader,
		OrgId:  orgID,
	})
	require.NoError(t, err)

	msg := &types.MsgCloseOrg{
		Signer: validLeader,
		OrgId:  orgID,
	}

	_, err = srv.CloseOrg(ctx, msg)
	require.ErrorIs(t, err, types.ErrOrgNotActive)
}

// ── CO-044: serving-key registration + rotation (D-S32-CO044-*) ──

func TestMsgRegisterOrg_WithServingKeyAndLeaderWallet(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	resp, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
		HubServingKey:   "wevibe1serving000000",
		LeaderWallet:    validLeader,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, mustDeriveOrgID(t, validLeader), resp.OrgId)
}

func TestMsgSetServingKey_RotatesWhenLeaderWalletSigns(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	_, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
		HubServingKey:   "wevibe1serving000000",
		LeaderWallet:    validLeader,
	})
	require.NoError(t, err)
	orgID := mustDeriveOrgID(t, validLeader)

	_, err = srv.SetServingKey(ctx, &types.MsgSetServingKey{
		Signer:        validLeader,
		OrgId:         orgID,
		NewServingKey: "wevibe1rotated000000",
	})
	require.NoError(t, err)
}

// R-BLAST-RADIUS: a non-leader-wallet signer cannot rotate the serving key.
func TestMsgSetServingKey_RejectsNonLeaderWallet(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	_, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
		HubServingKey:   "wevibe1serving000000",
		LeaderWallet:    validLeader,
	})
	require.NoError(t, err)
	orgID := mustDeriveOrgID(t, validLeader)

	_, err = srv.SetServingKey(ctx, &types.MsgSetServingKey{
		Signer:        validSigner, // not the leader wallet
		OrgId:         orgID,
		NewServingKey: "wevibe1evil000000000",
	})
	require.ErrorIs(t, err, types.ErrNotLeader)
}
