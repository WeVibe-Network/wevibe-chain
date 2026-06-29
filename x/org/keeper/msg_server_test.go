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
	validSigner            = "cosmos1vq0svzat0jyknkc6rfp40l8tr5cz4qxd6m6tyx"
	validLeader            = "cosmos1t9xdz4tvmsm2qj8fxadue6yx5ysp30zv4rnau6"
	validLeaderPubkey      = "leader_pubkey_12345678901234567890123456789012"
	validMember            = "cosmos1gsank9k6ygfnx376cuhw8zp9p8ssnyez44dtmh"
	validAuthority         = "cosmos14taukd54w5eak58yjv4lpzz3a0vr0petthfpc5"
	validHubResponsePubkey = "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"
	validX25519Pubkey      = "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"
)

type mockFeegrantKeeper struct {
	grantCalls    []feegrantGrantCall
	lastGranter   sdk.AccAddress
	lastGrantee   sdk.AccAddress
	lastAllowance feegrant.FeeAllowanceI
	err           error
}

type feegrantGrantCall struct {
	granter   sdk.AccAddress
	grantee   sdk.AccAddress
	allowance feegrant.FeeAllowanceI
}

func newMockFeegrantKeeper() *mockFeegrantKeeper {
	return &mockFeegrantKeeper{}
}

func (m *mockFeegrantKeeper) GrantAllowance(ctx context.Context, granter, grantee sdk.AccAddress, allowance feegrant.FeeAllowanceI) error {
	m.grantCalls = append(m.grantCalls, feegrantGrantCall{granter: granter, grantee: grantee, allowance: allowance})
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

func setupMsgAndQueryServer(t *testing.T) (types.MsgServer, types.QueryServer, context.Context, *mockBankKeeper, *mockFeegrantKeeper) {
	storeKey := storetypes.NewKVStoreKey("org")
	storeService, cms := testkeeper.NewTestStoreService(t, storeKey)
	logger := testkeeper.NewTestLogger()
	bank := newMockBankKeeper()
	feegrantKeeper := newMockFeegrantKeeper()
	k := keeper.NewKeeper(storeService, logger, validAuthority, bank, feegrantKeeper)
	sdkCtx := sdk.NewContext(cms, tmproto.Header{Height: 1, Time: time.Now().UTC()}, false, logger)
	ctx := sdk.WrapSDKContext(sdkCtx)
	return keeper.NewMsgServerImpl(k), keeper.NewQueryServerImpl(k), ctx, bank, feegrantKeeper
}

func assertServingFeegrantAllowance(t *testing.T, allowance feegrant.FeeAllowanceI) {
	t.Helper()

	allowed, ok := allowance.(*feegrant.AllowedMsgAllowance)
	require.True(t, ok)
	require.ElementsMatch(t,
		[]string{"/wevibe.serve.v1.MsgSubmitServeBatch", "/wevibe.serve.v1.MsgSubmitDenialBatch"},
		allowed.AllowedMessages,
	)

	innerAllowance, err := allowed.GetAllowance()
	require.NoError(t, err)
	basic, ok := innerAllowance.(*feegrant.BasicAllowance)
	require.True(t, ok)
	require.Empty(t, basic.SpendLimit)
	require.Nil(t, basic.Expiration)
}

func assertLeaderFeegrantAllowance(t *testing.T, allowance feegrant.FeeAllowanceI) {
	t.Helper()

	allowed, ok := allowance.(*feegrant.AllowedMsgAllowance)
	require.True(t, ok)
	require.ElementsMatch(t,
		[]string{
			"/wevibe.org.v1.MsgAddMember",
			"/wevibe.org.v1.MsgRemoveMember",
			"/wevibe.org.v1.MsgSetMemberCapabilities",
			"/wevibe.org.v1.MsgSetOrgConfig",
			"/wevibe.org.v1.MsgSetServingKey",
			"/wevibe.org.v1.MsgSetServingInfo",
			"/wevibe.org.v1.MsgRotateEpoch",
			"/wevibe.org.v1.MsgTransferLeadership",
			"/wevibe.org.v1.MsgCloseOrg",
			"/wevibe.org.v1.MsgGrantTrialAllowance",
			"/wevibe.memory.v1.MsgSubmitCommitment",
			"/wevibe.memory.v1.MsgApproveMemory",
			"/wevibe.memory.v1.MsgReportMemory",
		},
		allowed.AllowedMessages,
	)

	innerAllowance, err := allowed.GetAllowance()
	require.NoError(t, err)
	basic, ok := innerAllowance.(*feegrant.BasicAllowance)
	require.True(t, ok)
	require.Empty(t, basic.SpendLimit)
	require.Nil(t, basic.Expiration)
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

func TestMsgSetServingInfo_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     types.MsgSetServingInfo
		errIs   error
		wantErr bool
	}{
		{
			name: "valid single endpoint",
			msg: types.MsgSetServingInfo{
				Signer:            validLeader,
				OrgId:             "org1",
				HubEndpoints:      []string{"https://hub-1.example.com"},
				HubResponsePubkey: validHubResponsePubkey,
			},
		},
		{
			name: "valid three endpoints",
			msg: types.MsgSetServingInfo{
				Signer: validLeader,
				OrgId:  "org1",
				HubEndpoints: []string{
					"https://hub-1.example.com",
					"https://hub-2.example.com",
					"http://hub-3.example.com:8080",
				},
			},
		},
		{
			name: "valid with empty response pubkey",
			msg: types.MsgSetServingInfo{
				Signer:            validLeader,
				OrgId:             "org1",
				HubEndpoints:      []string{"https://hub-1.example.com"},
				HubResponsePubkey: "",
			},
		},
		{
			name: "missing signer",
			msg: types.MsgSetServingInfo{
				OrgId:        "org1",
				HubEndpoints: []string{"https://hub-1.example.com"},
			},
			wantErr: true,
		},
		{
			name: "missing org id",
			msg: types.MsgSetServingInfo{
				Signer:       validLeader,
				HubEndpoints: []string{"https://hub-1.example.com"},
			},
			errIs:   types.ErrInvalidOrgID,
			wantErr: true,
		},
		{
			name: "zero endpoints",
			msg: types.MsgSetServingInfo{
				Signer: validLeader,
				OrgId:  "org1",
			},
			errIs:   types.ErrInvalidHubEndpoints,
			wantErr: true,
		},
		{
			name: "four endpoints",
			msg: types.MsgSetServingInfo{
				Signer: validLeader,
				OrgId:  "org1",
				HubEndpoints: []string{
					"https://hub-1.example.com",
					"https://hub-2.example.com",
					"https://hub-3.example.com",
					"https://hub-4.example.com",
				},
			},
			errIs:   types.ErrInvalidHubEndpoints,
			wantErr: true,
		},
		{
			name: "malformed plain text",
			msg: types.MsgSetServingInfo{
				Signer:       validLeader,
				OrgId:        "org1",
				HubEndpoints: []string{"not a url"},
			},
			errIs:   types.ErrInvalidHubEndpoints,
			wantErr: true,
		},
		{
			name: "invalid scheme",
			msg: types.MsgSetServingInfo{
				Signer:       validLeader,
				OrgId:        "org1",
				HubEndpoints: []string{"ftp://x"},
			},
			errIs:   types.ErrInvalidHubEndpoints,
			wantErr: true,
		},
		{
			name: "missing host",
			msg: types.MsgSetServingInfo{
				Signer:       validLeader,
				OrgId:        "org1",
				HubEndpoints: []string{"https://"},
			},
			errIs:   types.ErrInvalidHubEndpoints,
			wantErr: true,
		},
		{
			name: "response pubkey odd-length hex",
			msg: types.MsgSetServingInfo{
				Signer:            validLeader,
				OrgId:             "org1",
				HubEndpoints:      []string{"https://hub-1.example.com"},
				HubResponsePubkey: "abc",
			},
			errIs:   types.ErrInvalidHubResponsePubkey,
			wantErr: true,
		},
		{
			name: "response pubkey non-hex",
			msg: types.MsgSetServingInfo{
				Signer:            validLeader,
				OrgId:             "org1",
				HubEndpoints:      []string{"https://hub-1.example.com"},
				HubResponsePubkey: "gg112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
			},
			errIs:   types.ErrInvalidHubResponsePubkey,
			wantErr: true,
		},
		{
			name: "response pubkey decodes to 31 bytes",
			msg: types.MsgSetServingInfo{
				Signer:            validLeader,
				OrgId:             "org1",
				HubEndpoints:      []string{"https://hub-1.example.com"},
				HubResponsePubkey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddee",
			},
			errIs:   types.ErrInvalidHubResponsePubkey,
			wantErr: true,
		},
		{
			name: "response pubkey decodes to 33 bytes",
			msg: types.MsgSetServingInfo{
				Signer:            validLeader,
				OrgId:             "org1",
				HubEndpoints:      []string{"https://hub-1.example.com"},
				HubResponsePubkey: validHubResponsePubkey + "00",
			},
			errIs:   types.ErrInvalidHubResponsePubkey,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if !tc.wantErr {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			if tc.errIs != nil {
				require.ErrorIs(t, err, tc.errIs)
			}
		})
	}
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
	require.Equal(t, types.FormatOrgID(0), resp.OrgId)
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
	msg.X25519Pubkey = validX25519Pubkey
	require.NoError(t, msg.ValidateBasic())
}

func TestMsgAddMember_Success(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeader, validLeader)

	memberMsg := &types.MsgAddMember{
		Signer:       validLeader,
		OrgId:        orgID,
		Pubkey:       validMember,
		Role:         "member",
		X25519Pubkey: validX25519Pubkey,
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

func TestMsgSetOrgConfig_Success(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeaderPubkey, validLeader)

	setOrgConfigMsg := &types.MsgSetOrgConfig{
		Signer:                   validLeader,
		OrgId:                    orgID,
		ServeReceiptRequired:     true,
		VocabHash:                "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
		EmbeddingModelId:         "text-embedding-3-large",
		MinContributionsPerEpoch: 10,
	}

	resp, err := srv.SetOrgConfig(ctx, setOrgConfigMsg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgGrantTrialAllowance_Success(t *testing.T) {
	srv, ctx, _, feegrantKeeper := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeaderPubkey, validLeader)

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

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeaderPubkey, validLeader)

	msg := &types.MsgGrantTrialAllowance{
		Signer:           validSigner,
		OrgId:            orgID,
		Grantee:          validMember,
		DailySubmissions: 1,
		TrialDays:        1,
	}

	_, err := srv.GrantTrialAllowance(ctx, msg)
	require.ErrorIs(t, err, types.ErrNotLeader)
}

func TestMsgSetMemberCapabilities_Success(t *testing.T) {
	srv, qs, ctx, _, _ := setupMsgAndQueryServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeaderPubkey, validLeader)

	_, err := srv.AddMember(ctx, &types.MsgAddMember{
		Signer:       validLeader,
		OrgId:        orgID,
		Pubkey:       validMember,
		Role:         "member",
		X25519Pubkey: validX25519Pubkey,
	})
	require.NoError(t, err)

	resp, err := srv.SetMemberCapabilities(ctx, &types.MsgSetMemberCapabilities{
		Signer:        validLeader,
		OrgId:         orgID,
		Pubkey:        validMember,
		CanContribute: true,
		CanModerate:   true,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	membersResp, err := qs.GetMembers(ctx, &types.QueryGetMembersRequest{OrgId: orgID})
	require.NoError(t, err)

	for _, m := range membersResp.Members {
		if m.Pubkey != validMember {
			continue
		}
		require.True(t, m.CanContribute)
		require.True(t, m.CanModerate)
		return
	}

	t.Fatalf("member %s not found in org %s", validMember, orgID)
}

func TestMsgSetMemberCapabilities_NotLeader(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeaderPubkey, validLeader)

	_, err := srv.AddMember(ctx, &types.MsgAddMember{
		Signer:       validLeader,
		OrgId:        orgID,
		Pubkey:       validMember,
		Role:         "member",
		X25519Pubkey: validX25519Pubkey,
	})
	require.NoError(t, err)

	_, err = srv.SetMemberCapabilities(ctx, &types.MsgSetMemberCapabilities{
		Signer:        validSigner,
		OrgId:         orgID,
		Pubkey:        validMember,
		CanContribute: true,
		CanModerate:   true,
	})
	require.ErrorIs(t, err, types.ErrNotLeader)
}

func TestMsgSetMemberCapabilities_MemberNotFound(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeaderPubkey, validLeader)

	_, err := srv.SetMemberCapabilities(ctx, &types.MsgSetMemberCapabilities{
		Signer:        validLeader,
		OrgId:         orgID,
		Pubkey:        validMember,
		CanContribute: true,
		CanModerate:   true,
	})
	require.ErrorIs(t, err, types.ErrMemberNotFound)
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

	orgResp, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	})
	require.NoError(t, err)
	orgID := orgResp.OrgId

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

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeaderPubkey, validLeader)

	_, err := srv.AddMember(ctx, &types.MsgAddMember{
		Signer:       validLeader,
		OrgId:        orgID,
		Pubkey:       validMember,
		Role:         "member",
		X25519Pubkey: validX25519Pubkey,
	})
	require.NoError(t, err)

	msg := &types.MsgTransferLeadership{
		Signer:          validLeader,
		OrgId:           orgID,
		NewLeader:       validMember,
		NewLeaderWallet: validMember,
	}

	resp, err := srv.TransferLeadership(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestMsgTransferLeadership_NotLeader(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeaderPubkey, validLeader)

	msg := &types.MsgTransferLeadership{
		Signer:          validSigner,
		OrgId:           orgID,
		NewLeader:       validMember,
		NewLeaderWallet: validMember,
	}

	_, err := srv.TransferLeadership(ctx, msg)
	require.ErrorIs(t, err, types.ErrNotLeader)
}

func TestMsgTransferLeadership_NewLeaderNotMember(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeaderPubkey, validLeader)

	msg := &types.MsgTransferLeadership{
		Signer:          validLeader,
		OrgId:           orgID,
		NewLeader:       validMember,
		NewLeaderWallet: validMember,
	}

	_, err := srv.TransferLeadership(ctx, msg)
	require.Error(t, err)
	require.Contains(t, err.Error(), "new_leader must be a member of the org")
}

func TestMsgTransferLeadership_SelfTransfer(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgResp, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
	})
	require.NoError(t, err)
	orgID := orgResp.OrgId

	msg := &types.MsgTransferLeadership{
		Signer:          validLeader,
		OrgId:           orgID,
		NewLeader:       validLeader,
		NewLeaderWallet: validLeader,
	}

	err = msg.ValidateBasic()
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot transfer leadership to self")
}

func TestMsgCloseOrg_Success(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeaderPubkey, validLeader)

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

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeaderPubkey, validLeader)

	msg := &types.MsgCloseOrg{
		Signer: validSigner,
		OrgId:  orgID,
	}

	_, err := srv.CloseOrg(ctx, msg)
	require.ErrorIs(t, err, types.ErrNotLeader)
}

func TestMsgCloseOrg_AlreadyClosed(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validSigner, validLeaderPubkey, validLeader)

	_, err := srv.CloseOrg(ctx, &types.MsgCloseOrg{
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
	srv, ctx, _, feegrantKeeper := setupMsgServer(t)

	resp, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
		HubServingKey:   validMember,
		LeaderWallet:    validLeader,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, types.FormatOrgID(0), resp.OrgId)
	require.Len(t, feegrantKeeper.grantCalls, 2)

	servingGrantCall := feegrantKeeper.grantCalls[0]
	require.Equal(t, types.OrgAccountAddress(resp.OrgId).String(), servingGrantCall.granter.String())
	require.Equal(t, validMember, servingGrantCall.grantee.String())
	assertServingFeegrantAllowance(t, servingGrantCall.allowance)

	leaderGrantCall := feegrantKeeper.grantCalls[1]
	require.Equal(t, types.OrgAccountAddress(resp.OrgId).String(), leaderGrantCall.granter.String())
	require.Equal(t, validLeader, leaderGrantCall.grantee.String())
	assertLeaderFeegrantAllowance(t, leaderGrantCall.allowance)
}

func TestMsgSetServingKey_RotatesWhenLeaderWalletSigns(t *testing.T) {
	srv, ctx, _, feegrantKeeper := setupMsgServer(t)

	orgResp, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
		HubServingKey:   validMember,
		LeaderWallet:    validLeader,
	})
	require.NoError(t, err)
	orgID := orgResp.OrgId

	_, err = srv.SetServingKey(ctx, &types.MsgSetServingKey{
		Signer:        validLeader,
		OrgId:         orgID,
		NewServingKey: validSigner,
	})
	require.NoError(t, err)
	require.Len(t, feegrantKeeper.grantCalls, 3)

	reGrantCall := feegrantKeeper.grantCalls[2]
	require.Equal(t, types.OrgAccountAddress(orgID).String(), reGrantCall.granter.String())
	require.Equal(t, validSigner, reGrantCall.grantee.String())
	assertServingFeegrantAllowance(t, reGrantCall.allowance)
}

// R-BLAST-RADIUS: a non-leader-wallet signer cannot rotate the serving key.
func TestMsgSetServingKey_RejectsNonLeaderWallet(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgResp, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          validLeader,
		Leader:          validLeader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
		HubServingKey:   validMember,
		LeaderWallet:    validLeader,
	})
	require.NoError(t, err)
	orgID := orgResp.OrgId

	_, err = srv.SetServingKey(ctx, &types.MsgSetServingKey{
		Signer:        validSigner, // not the leader wallet
		OrgId:         orgID,
		NewServingKey: validAuthority,
	})
	require.ErrorIs(t, err, types.ErrNotLeader)
}

func TestMsgSetServingInfo_SetsHubEndpointsInOrder(t *testing.T) {
	srv, qs, ctx, _, _ := setupMsgAndQueryServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validLeader, validLeader, validLeader)

	tests := []struct {
		name      string
		endpoints []string
	}{
		{
			name:      "single endpoint",
			endpoints: []string{"https://hub-1.example.com"},
		},
		{
			name: "two endpoints",
			endpoints: []string{
				"https://hub-1.example.com",
				"http://hub-2.example.com:8080",
			},
		},
		{
			name: "three endpoints",
			endpoints: []string{
				"https://hub-1.example.com",
				"https://hub-2.example.com",
				"https://hub-3.example.com/rpc",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := srv.SetServingInfo(ctx, &types.MsgSetServingInfo{
				Signer:       validLeader,
				OrgId:        orgID,
				HubEndpoints: tc.endpoints,
			})
			require.NoError(t, err)

			resp, err := qs.GetOrg(ctx, &types.QueryGetOrgRequest{OrgId: orgID})
			require.NoError(t, err)
			require.Equal(t, tc.endpoints, resp.HubEndpoints)
		})
	}
}

func TestMsgSetServingInfo_SetsHubEndpointsAndResponsePubkey(t *testing.T) {
	srv, qs, ctx, _, _ := setupMsgAndQueryServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validLeader, validLeader, validLeader)
	endpoints := []string{
		"https://hub-1.example.com",
		"https://hub-2.example.com",
	}

	_, err := srv.SetServingInfo(ctx, &types.MsgSetServingInfo{
		Signer:            validLeader,
		OrgId:             orgID,
		HubEndpoints:      endpoints,
		HubResponsePubkey: validHubResponsePubkey,
	})
	require.NoError(t, err)

	resp, err := qs.GetOrg(ctx, &types.QueryGetOrgRequest{OrgId: orgID})
	require.NoError(t, err)
	require.Equal(t, endpoints, resp.HubEndpoints)
	require.Equal(t, validHubResponsePubkey, resp.HubResponsePubkey)
}

func TestMsgSetServingInfo_AllowsEmptyResponsePubkey(t *testing.T) {
	srv, qs, ctx, _, _ := setupMsgAndQueryServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validLeader, validLeader, validLeader)

	_, err := srv.SetServingInfo(ctx, &types.MsgSetServingInfo{
		Signer:            validLeader,
		OrgId:             orgID,
		HubEndpoints:      []string{"https://hub-1.example.com"},
		HubResponsePubkey: "",
	})
	require.NoError(t, err)

	resp, err := qs.GetOrg(ctx, &types.QueryGetOrgRequest{OrgId: orgID})
	require.NoError(t, err)
	require.Equal(t, []string{"https://hub-1.example.com"}, resp.HubEndpoints)
	require.Empty(t, resp.HubResponsePubkey)
}

func TestMsgSetServingInfo_RejectsNonLeaderWallet(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validLeader, validLeader, validLeader)

	_, err := srv.SetServingInfo(ctx, &types.MsgSetServingInfo{
		Signer:       validSigner,
		OrgId:        orgID,
		HubEndpoints: []string{"https://hub-1.example.com"},
	})
	require.ErrorIs(t, err, types.ErrNotLeader)
}

func TestMsgSetServingInfo_RejectsInvalidEndpoints(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validLeader, validLeader, validLeader)

	tests := []struct {
		name      string
		endpoints []string
	}{
		{
			name:      "zero endpoints",
			endpoints: []string{},
		},
		{
			name: "four endpoints",
			endpoints: []string{
				"https://hub-1.example.com",
				"https://hub-2.example.com",
				"https://hub-3.example.com",
				"https://hub-4.example.com",
			},
		},
		{
			name:      "malformed url",
			endpoints: []string{"not a url"},
		},
		{
			name:      "invalid scheme",
			endpoints: []string{"ftp://x"},
		},
		{
			name:      "missing host",
			endpoints: []string{"https://"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := srv.SetServingInfo(ctx, &types.MsgSetServingInfo{
				Signer:       validLeader,
				OrgId:        orgID,
				HubEndpoints: tc.endpoints,
			})
			require.ErrorIs(t, err, types.ErrInvalidHubEndpoints)
		})
	}
}

func TestMsgSetServingInfo_RejectsInvalidResponsePubkey(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(t, srv, ctx, validLeader, validLeader, validLeader)

	tests := []struct {
		name   string
		pubkey string
	}{
		{
			name:   "odd-length hex",
			pubkey: "abc",
		},
		{
			name:   "non-hex characters",
			pubkey: "zz112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
		},
		{
			name:   "31-byte key",
			pubkey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddee",
		},
		{
			name:   "33-byte key",
			pubkey: validHubResponsePubkey + "00",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := srv.SetServingInfo(ctx, &types.MsgSetServingInfo{
				Signer:            validLeader,
				OrgId:             orgID,
				HubEndpoints:      []string{"https://hub-1.example.com"},
				HubResponsePubkey: tc.pubkey,
			})
			require.ErrorIs(t, err, types.ErrInvalidHubResponsePubkey)
		})
	}
}
