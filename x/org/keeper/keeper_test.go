package keeper_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/wevibe-network/wevibe-chain/testutil/keeper"
	"github.com/wevibe-network/wevibe-chain/x/org/keeper"
	"github.com/wevibe-network/wevibe-chain/x/org/types"
)

var (
	orgStoreKey      = storetypes.NewKVStoreKey("org")
	testCreatorAddr  = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	testX25519Pubkey = "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"
)

type mockBankKeeper struct {
	balances               map[string]math.Int
	moduleBalances         map[string]math.Int
	alwaysReturnNilForSend bool
}

func newMockBankKeeper() *mockBankKeeper {
	m := &mockBankKeeper{
		balances:               make(map[string]math.Int),
		moduleBalances:         make(map[string]math.Int),
		alwaysReturnNilForSend: true,
	}
	m.balances[testCreatorAddr.String()] = math.NewInt(10000000000)
	m.balances["cosmos1abc"] = math.NewInt(10000000000)
	m.balances["cosmos1abc1234567890123456789012345678901234567"] = math.NewInt(10000000000)
	m.balances["cosmos14taukd54w5eak58yjv4lpzz3a0vr0petthfpc5"] = math.NewInt(10000000000)
	m.balances["cosmos1vq0svzat0jyknkc6rfp40l8tr5cz4qxd6m6tyx"] = math.NewInt(10000000000)
	m.balances["cosmos1t9xdz4tvmsm2qj8fxadue6yx5ysp30zv4rnau6"] = math.NewInt(10000000000)
	m.balances["cosmos1gsank9k6ygfnx376cuhw8zp9p8ssnyez44dtmh"] = math.NewInt(10000000000)
	m.balances["cosmos1wnrlmtryha3hwrep5rn5hlu7jk3g8jka2ff505"] = math.NewInt(10000000000)
	m.moduleBalances["org"] = math.NewInt(10000000000000)
	return m
}

func newMockBankKeeperWithBalance() *mockBankKeeper {
	m := newMockBankKeeper()
	m.balances["cosmos1def"] = math.NewInt(10000000000)
	m.balances["cosmos1def1234567890123456789012345678901234"] = math.NewInt(10000000000)
	m.balances["cosmos1leader123456789012345678901234567890"] = math.NewInt(10000000000)
	m.balances["cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry"] = math.NewInt(10000000000)
	return m
}

func newMockBankKeeperStrict() *mockBankKeeper {
	m := &mockBankKeeper{
		balances:               make(map[string]math.Int),
		moduleBalances:         make(map[string]math.Int),
		alwaysReturnNilForSend: false,
	}
	m.balances[testCreatorAddr.String()] = math.NewInt(10000000000)
	m.balances["cosmos1abc"] = math.NewInt(10000000000)
	m.moduleBalances["org"] = math.NewInt(10000000000000)
	return m
}

func newTestKeeperWithStrictBank(t *testing.T) (*keeper.Keeper, context.Context, *mockBankKeeper) {
	storeService, _ := testkeeper.NewTestStoreService(t, orgStoreKey)
	logger := testkeeper.NewTestLogger()
	bank := newMockBankKeeperStrict()
	k := keeper.NewKeeper(storeService, logger, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", bank, nil)
	ctx := context.Background()
	return k, ctx, bank
}

func (m *mockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	if m.alwaysReturnNilForSend {
		return nil
	}
	addr := senderAddr.String()
	balance, ok := m.balances[addr]
	if !ok {
		balance = math.ZeroInt()
	}
	amtToSub := amt.AmountOf("uvibe")
	if balance.LT(amtToSub) {
		return errors.New("insufficient funds")
	}
	spendable := balance.Sub(amtToSub)
	m.balances[addr] = spendable
	moduleBal, ok := m.moduleBalances[recipientModule]
	if !ok {
		moduleBal = math.ZeroInt()
	}
	m.moduleBalances[recipientModule] = moduleBal.Add(amtToSub)
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	moduleBalance, ok := m.moduleBalances[senderModule]
	if !ok {
		moduleBalance = math.ZeroInt()
	}
	amtToSub := amt.AmountOf("uvibe")
	if moduleBalance.LT(amtToSub) {
		return errors.New("insufficient funds")
	}
	spendable := moduleBalance.Sub(amtToSub)
	m.moduleBalances[senderModule] = spendable
	recipientBal, ok := m.balances[recipientAddr.String()]
	if !ok {
		recipientBal = math.ZeroInt()
	}
	m.balances[recipientAddr.String()] = recipientBal.Add(amtToSub)
	return nil
}

func (m *mockBankKeeper) SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	if m.alwaysReturnNilForSend {
		return nil
	}

	fromBalance, ok := m.balances[fromAddr.String()]
	if !ok {
		fromBalance = math.ZeroInt()
	}
	amtToSub := amt.AmountOf("uvibe")
	if fromBalance.LT(amtToSub) {
		return errors.New("insufficient funds")
	}
	m.balances[fromAddr.String()] = fromBalance.Sub(amtToSub)

	toBalance, ok := m.balances[toAddr.String()]
	if !ok {
		toBalance = math.ZeroInt()
	}
	m.balances[toAddr.String()] = toBalance.Add(amtToSub)
	return nil
}

func (m *mockBankKeeper) MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	moduleBal, ok := m.moduleBalances[moduleName]
	if !ok {
		moduleBal = math.ZeroInt()
	}
	moduleBal = moduleBal.Add(amt.AmountOf("uvibe"))
	m.moduleBalances[moduleName] = moduleBal
	return nil
}

func (m *mockBankKeeper) HasSupply(ctx context.Context, denom string) bool {
	return true
}

func (m *mockBankKeeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	balance, ok := m.balances[addr.String()]
	if !ok {
		balance = math.ZeroInt()
	}
	return sdk.NewCoin(denom, balance)
}

func (m *mockBankKeeper) BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	moduleBalance, ok := m.moduleBalances[moduleName]
	if !ok {
		moduleBalance = math.ZeroInt()
	}
	amtToSub := amt.AmountOf("uvibe")
	if moduleBalance.LT(amtToSub) {
		return errors.New("insufficient funds")
	}
	m.moduleBalances[moduleName] = moduleBalance.Sub(amtToSub)
	return nil
}

func (m *mockBankKeeper) SetBalance(addr string, amount math.Int) {
	m.balances[addr] = amount
}

func (m *mockBankKeeper) SetModuleBalance(module string, amount math.Int) {
	m.moduleBalances[module] = amount
}

func newTestKeeper(t *testing.T) (*keeper.Keeper, context.Context, *mockBankKeeper) {
	storeService, _ := testkeeper.NewTestStoreService(t, orgStoreKey)
	logger := testkeeper.NewTestLogger()
	bank := newMockBankKeeper()
	k := keeper.NewKeeper(storeService, logger, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", bank, nil)
	ctx := context.Background()
	return k, ctx, bank
}

func newTestKeeperWithFundedBank(t *testing.T) (*keeper.Keeper, context.Context, *mockBankKeeper) {
	storeService, _ := testkeeper.NewTestStoreService(t, orgStoreKey)
	logger := testkeeper.NewTestLogger()
	bank := newMockBankKeeperWithBalance()
	k := keeper.NewKeeper(storeService, logger, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", bank, nil)
	ctx := context.Background()
	return k, ctx, bank
}

func registerMsgServerOrgWithLeaderWallet(t *testing.T, srv types.MsgServer, ctx context.Context, signer, leader, leaderWallet string) string {
	t.Helper()

	resp, err := srv.RegisterOrg(ctx, &types.MsgRegisterOrg{
		Signer:          signer,
		Leader:          leader,
		StorageQuota:    1000,
		RetrievalBudget: 500,
		LeaderWallet:    leaderWallet,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	return resp.OrgId
}

func TestRegisterOrg(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)

	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	has, err := k.HasOrg(ctx, org.OrgID)
	if err != nil {
		t.Fatalf("HasOrg failed: %v", err)
	}
	if !has {
		t.Fatal("expected org to exist")
	}
	require.Equal(t, types.FormatOrgID(0), org.OrgID)
	require.Equal(t, uint64(0), org.Slot)
	require.Equal(t, types.OrgAccountAddress(org.OrgID).String(), org.AccountAddress)

	retrieved, err := k.GetOrg(ctx, org.OrgID)
	require.NoError(t, err)
	require.Equal(t, org.AccountAddress, retrieved.AccountAddress)
}

func TestRegisterOrg_DuplicateLeader(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)

	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("first RegisterOrg failed: %v", err)
	}

	err = k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != types.ErrLeaderAlreadyOwnsOrg {
		t.Fatalf("expected ErrLeaderAlreadyOwnsOrg, got: %v", err)
	}
}

func TestRegisterOrg_SlotCapReached(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	params, err := k.GetParams(ctx)
	require.NoError(t, err)
	params.SlotCap = 32
	require.NoError(t, k.SetParams(ctx, params))

	for i := 0; i < 32; i++ {
		org := types.NewOrg("", fmt.Sprintf("leader-%d", i), "", 1000000, 5000)
		err := k.RegisterOrg(ctx, org, testCreatorAddr)
		require.NoError(t, err)
		require.Equal(t, types.FormatOrgID(uint64(i)), org.OrgID)
		require.Equal(t, uint64(i), org.Slot)
	}

	err = k.RegisterOrg(ctx, types.NewOrg("", "leader-over-cap", "", 1000000, 5000), testCreatorAddr)
	require.ErrorIs(t, err, types.ErrSlotCapReached)
}

func TestRegisterOrg_SplitsBurnAndOrgAccount(t *testing.T) {
	k, ctx, bank := newTestKeeperWithStrictBank(t)

	price := k.ComputeSlotPrice(ctx, 0)
	acctHalf := price.Sub(price.QuoRaw(2))

	creatorStart := bank.GetBalance(ctx, testCreatorAddr, "uvibe").Amount
	moduleStart := bank.moduleBalances[types.ModuleName]

	org := types.NewOrg("", "leader_pubkey_split_123456789012345678901234", "", 1000000, 5000)
	require.NoError(t, k.RegisterOrg(ctx, org, testCreatorAddr))

	require.Equal(t, creatorStart.Sub(price), bank.GetBalance(ctx, testCreatorAddr, "uvibe").Amount)
	require.Equal(t, moduleStart, bank.moduleBalances[types.ModuleName])

	orgAccountAddr := types.OrgAccountAddress(org.OrgID)
	require.Equal(t, orgAccountAddr.String(), org.AccountAddress)
	require.Equal(t, acctHalf, bank.GetBalance(ctx, orgAccountAddr, "uvibe").Amount)
}

func TestRegisterOrg_InsufficientFunds(t *testing.T) {
	k, ctx, bank := newTestKeeperWithStrictBank(t)

	price := k.ComputeSlotPrice(ctx, 0)
	bank.SetBalance(testCreatorAddr.String(), price.Sub(math.NewInt(1)))

	org := types.NewOrg("", "leader_pubkey_insufficient_1234567890123456", "", 1000000, 5000)
	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	require.ErrorIs(t, err, types.ErrInsufficientFund)
}

func TestGetOrg(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)

	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	retrieved, err := k.GetOrg(ctx, org.OrgID)
	if err != nil {
		t.Fatalf("GetOrg failed: %v", err)
	}
	if retrieved.OrgID != org.OrgID {
		t.Fatalf("expected %s, got: %s", org.OrgID, retrieved.OrgID)
	}
	if retrieved.Leader != "leader_pubkey_12345678901234567890123456789012" {
		t.Fatalf("unexpected leader: %s", retrieved.Leader)
	}
}

func TestGetOrg_NotFound(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	_, err := k.GetOrg(ctx, "nonexistent")
	if err != types.ErrOrgNotFound {
		t.Fatalf("expected ErrOrgNotFound, got: %v", err)
	}
}

func TestAddMember(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)
	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	member := types.NewMemberRecord(org.OrgID, "member_pubkey_123456789012345678901234", "member", testX25519Pubkey)
	err = k.AddMember(ctx, member)
	if err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	isMember, err := k.IsMember(ctx, org.OrgID, "member_pubkey_123456789012345678901234")
	if err != nil {
		t.Fatalf("IsMember failed: %v", err)
	}
	if !isMember {
		t.Fatal("expected member to exist")
	}
}

func TestAddMember_DuplicateMember(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)
	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	member := types.NewMemberRecord(org.OrgID, "member_pubkey_123456789012345678901234", "member", testX25519Pubkey)
	err = k.AddMember(ctx, member)
	if err != nil {
		t.Fatalf("first AddMember failed: %v", err)
	}

	err = k.AddMember(ctx, member)
	if err != types.ErrMemberExists {
		t.Fatalf("expected ErrMemberExists, got: %v", err)
	}
}

func TestRemoveMember(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)
	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	member := types.NewMemberRecord(org.OrgID, "member_pubkey_123456789012345678901234", "member", testX25519Pubkey)
	err = k.AddMember(ctx, member)
	if err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	err = k.RemoveMember(ctx, org.OrgID, "member_pubkey_123456789012345678901234")
	if err != nil {
		t.Fatalf("RemoveMember failed: %v", err)
	}

	isMember, err := k.IsMember(ctx, org.OrgID, "member_pubkey_123456789012345678901234")
	if err != nil {
		t.Fatalf("IsMember failed: %v", err)
	}
	if isMember {
		t.Fatal("expected member to not exist after removal")
	}
}

func TestRemoveMember_NotFound(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	err := k.RemoveMember(ctx, "org1", "nonexistent_member")
	if err != types.ErrMemberNotFound {
		t.Fatalf("expected ErrMemberNotFound, got: %v", err)
	}
}

func TestIsLeader(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)
	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	isLeader, err := k.IsLeader(ctx, org.OrgID, "leader_pubkey_12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("IsLeader failed: %v", err)
	}
	if !isLeader {
		t.Fatal("expected leader to be recognized as leader")
	}

	isLeader, err = k.IsLeader(ctx, org.OrgID, "some_other_pubkey")
	if err != nil {
		t.Fatalf("IsLeader failed: %v", err)
	}
	if isLeader {
		t.Fatal("expected non-leader to not be recognized as leader")
	}
}

func TestUpdateOrgStatus(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)
	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	err = k.UpdateOrgStatus(ctx, org.OrgID, types.OrgStatus_DORMANT)
	if err != nil {
		t.Fatalf("UpdateOrgStatus failed: %v", err)
	}

	updated, err := k.GetOrg(ctx, org.OrgID)
	if err != nil {
		t.Fatalf("GetOrg failed: %v", err)
	}
	if updated.Status != types.OrgStatus_DORMANT {
		t.Fatalf("expected status DORMANT, got: %v", updated.Status)
	}
}

func TestUpdateStorageQuota(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)
	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	err = k.UpdateStorageQuota(ctx, org.OrgID, 2000000)
	if err != nil {
		t.Fatalf("UpdateStorageQuota failed: %v", err)
	}

	updated, err := k.GetOrg(ctx, org.OrgID)
	if err != nil {
		t.Fatalf("GetOrg failed: %v", err)
	}
	if updated.StorageQuota != 2000000 {
		t.Fatalf("expected quota 2000000, got: %d", updated.StorageQuota)
	}
}

func TestGetAllOrgs(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	orgs := []*types.Org{
		types.NewOrg("org1", "leader1_pubkey_1234567890123456789012345", "", 1000000, 5000),
		types.NewOrg("org2", "leader2_pubkey_1234567890123456789012345", "", 1000000, 5000),
	}

	for _, org := range orgs {
		err := k.RegisterOrg(ctx, org, testCreatorAddr)
		if err != nil {
			t.Fatalf("RegisterOrg failed: %v", err)
		}
	}

	allOrgs, err := k.GetAllOrgs(ctx)
	if err != nil {
		t.Fatalf("GetAllOrgs failed: %v", err)
	}
	if len(allOrgs) != 2 {
		t.Fatalf("expected 2 orgs, got: %d", len(allOrgs))
	}
}

func TestGetAllMembers(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)
	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	members := []*types.MemberRecord{
		types.NewMemberRecord(org.OrgID, "member1_pubkey_123456789012345678901", "member", testX25519Pubkey),
		types.NewMemberRecord(org.OrgID, "member2_pubkey_123456789012345678901", "member", testX25519Pubkey),
	}

	for _, m := range members {
		err := k.AddMember(ctx, m)
		if err != nil {
			t.Fatalf("AddMember failed: %v", err)
		}
	}

	allMembers, err := k.GetAllMembers(ctx, org.OrgID)
	if err != nil {
		t.Fatalf("GetAllMembers failed: %v", err)
	}
	if len(allMembers) != 3 {
		t.Fatalf("expected 3 members, got: %d", len(allMembers))
	}
}

func TestInitGenesisAndExportGenesis(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	state := &types.GenesisState{
		Orgs: []*types.Org{
			types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000),
		},
		Members: []*types.MemberRecord{
			types.NewMemberRecord("org1", "member_pubkey_123456789012345678901234", "member", testX25519Pubkey),
		},
	}

	err := k.InitGenesis(ctx, state)
	if err != nil {
		t.Fatalf("InitGenesis failed: %v", err)
	}

	exported, err := k.ExportGenesis(ctx)
	if err != nil {
		t.Fatalf("ExportGenesis failed: %v", err)
	}

	if len(exported.Orgs) != len(state.Orgs) {
		t.Fatalf("expected %d orgs, got: %d", len(state.Orgs), len(exported.Orgs))
	}
	if len(exported.Members) != len(state.Members) {
		t.Fatalf("expected %d members, got: %d", len(state.Members), len(exported.Members))
	}
}

func TestSDKTransactionIsolation(t *testing.T) {
	storeService1, cms1 := testkeeper.NewTestStoreService(t, orgStoreKey)
	logger1 := testkeeper.NewTestLogger()
	bank1 := newMockBankKeeper()
	k1 := keeper.NewKeeper(storeService1, logger1, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", bank1, nil)
	ctx1 := context.Background()

	org1 := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)
	err := k1.RegisterOrg(ctx1, org1, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg in tx1 failed: %v", err)
	}

	has, err := k1.HasOrg(ctx1, org1.OrgID)
	if err != nil {
		t.Fatalf("HasOrg in tx1 failed: %v", err)
	}
	if !has {
		t.Fatal("tx1 should see its own org")
	}

	cms1.Commit()

	storeService2 := testkeeper.NewTestStoreServiceWithCMS(t, orgStoreKey, cms1)
	logger2 := testkeeper.NewTestLogger()
	bank2 := newMockBankKeeper()
	k2 := keeper.NewKeeper(storeService2, logger2, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", bank2, nil)
	ctx2 := context.Background()

	has, err = k2.HasOrg(ctx2, org1.OrgID)
	if err != nil {
		t.Fatalf("HasOrg in tx2 failed: %v", err)
	}
	if !has {
		t.Fatal("tx2 should see tx1's org after commit")
	}
}

func TestSDKStoreCommitPersist(t *testing.T) {
	storeService1, cms1 := testkeeper.NewTestStoreService(t, orgStoreKey)
	logger1 := testkeeper.NewTestLogger()
	bank1 := newMockBankKeeper()
	k1 := keeper.NewKeeper(storeService1, logger1, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", bank1, nil)
	ctx1 := context.Background()

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)

	err := k1.RegisterOrg(ctx1, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	cms1.Commit()

	storeService2 := testkeeper.NewTestStoreServiceWithCMS(t, orgStoreKey, cms1)
	logger2 := testkeeper.NewTestLogger()
	bank2 := newMockBankKeeper()
	k2 := keeper.NewKeeper(storeService2, logger2, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", bank2, nil)
	ctx2 := context.Background()

	has, err := k2.HasOrg(ctx2, org.OrgID)
	if err != nil {
		t.Fatalf("HasOrg failed: %v", err)
	}
	if !has {
		t.Fatal("expected org to persist after commit")
	}

	retrieved, err := k2.GetOrg(ctx2, org.OrgID)
	if err != nil {
		t.Fatalf("GetOrg failed: %v", err)
	}
	if retrieved.OrgID != org.OrgID {
		t.Fatal("retrieved org doesn't match stored org")
	}
}

func TestSDKGenesisRoundTrip(t *testing.T) {
	storeService1, _ := testkeeper.NewTestStoreService(t, orgStoreKey)
	logger1 := testkeeper.NewTestLogger()
	bank1 := newMockBankKeeper()
	k1 := keeper.NewKeeper(storeService1, logger1, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", bank1, nil)
	ctx1 := context.Background()

	genesisState := &types.GenesisState{
		Orgs: []*types.Org{
			types.NewOrg("org1", "leader1_pubkey_1234567890123456789012345", "", 1000000, 5000),
			types.NewOrg("org2", "leader2_pubkey_1234567890123456789012345", "", 1000000, 5000),
		},
		Members: []*types.MemberRecord{
			types.NewMemberRecord("org1", "member1_pubkey_123456789012345678901", "member", testX25519Pubkey),
		},
	}

	err := k1.InitGenesis(ctx1, genesisState)
	if err != nil {
		t.Fatalf("InitGenesis failed: %v", err)
	}

	exportedState, err := k1.ExportGenesis(ctx1)
	if err != nil {
		t.Fatalf("ExportGenesis failed: %v", err)
	}

	if len(exportedState.Orgs) != len(genesisState.Orgs) {
		t.Fatalf("expected %d orgs, got: %d", len(genesisState.Orgs), len(exportedState.Orgs))
	}

	storeService2, _ := testkeeper.NewTestStoreService(t, orgStoreKey)
	logger2 := testkeeper.NewTestLogger()
	bank2 := newMockBankKeeper()
	k2 := keeper.NewKeeper(storeService2, logger2, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", bank2, nil)
	ctx2 := context.Background()

	err = k2.InitGenesis(ctx2, exportedState)
	if err != nil {
		t.Fatalf("InitGenesis from exported failed: %v", err)
	}

	reExportedState, err := k2.ExportGenesis(ctx2)
	if err != nil {
		t.Fatalf("re-ExportGenesis failed: %v", err)
	}

	if len(reExportedState.Orgs) != len(exportedState.Orgs) {
		t.Fatalf("expected %d orgs after re-export, got: %d", len(exportedState.Orgs), len(reExportedState.Orgs))
	}
}

func TestSDKMultipleOrgsAndMembers(t *testing.T) {
	storeService, _ := testkeeper.NewTestStoreService(t, orgStoreKey)
	logger := testkeeper.NewTestLogger()
	bank := newMockBankKeeper()
	k := keeper.NewKeeper(storeService, logger, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", bank, nil)
	ctx := context.Background()

	orgs := []*types.Org{
		types.NewOrg("org1", "leader1_pubkey_1234567890123456789012345", "", 1000000, 5000),
		types.NewOrg("org2", "leader2_pubkey_1234567890123456789012345", "", 1000000, 5000),
	}

	for _, org := range orgs {
		err := k.RegisterOrg(ctx, org, testCreatorAddr)
		if err != nil {
			t.Fatalf("RegisterOrg failed: %v", err)
		}
	}
	org0ID := orgs[0].OrgID
	org1ID := orgs[1].OrgID

	members := []*types.MemberRecord{
		types.NewMemberRecord(org0ID, "member1_pubkey_123456789012345678901", "member", testX25519Pubkey),
		types.NewMemberRecord(org0ID, "member2_pubkey_123456789012345678901", "member", testX25519Pubkey),
		types.NewMemberRecord(org1ID, "member3_pubkey_123456789012345678901", "member", testX25519Pubkey),
	}

	for _, m := range members {
		err := k.AddMember(ctx, m)
		if err != nil {
			t.Fatalf("AddMember failed: %v", err)
		}
	}

	allOrgs, err := k.GetAllOrgs(ctx)
	if err != nil {
		t.Fatalf("GetAllOrgs failed: %v", err)
	}
	if len(allOrgs) != 2 {
		t.Fatalf("expected 2 orgs, got: %d", len(allOrgs))
	}

	for _, orgID := range []string{org0ID, org1ID} {
		has, err := k.HasOrg(ctx, orgID)
		if err != nil {
			t.Fatalf("HasOrg failed: %v", err)
		}
		if !has {
			t.Errorf("expected org %s to exist", orgID)
		}
	}

	org1Members, err := k.GetAllMembers(ctx, org0ID)
	if err != nil {
		t.Fatalf("GetAllMembers failed: %v", err)
	}
	if len(org1Members) != 3 {
		t.Fatalf("expected 3 members for org1, got: %d", len(org1Members))
	}
}

func TestSetOrgConfig(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)
	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	cfg := &types.OrgConfig{
		OrgID:                    org.OrgID,
		ServeAttestationRequired: true,
	}

	err = k.SetOrgConfig(ctx, org.OrgID, cfg)
	if err != nil {
		t.Fatalf("SetOrgConfig failed: %v", err)
	}

	retrieved, err := k.GetOrgConfig(ctx, org.OrgID)
	if err != nil {
		t.Fatalf("GetOrgConfig failed: %v", err)
	}
	require.True(t, retrieved.ServeAttestationRequired)
}

func TestComputeSlotPrice_Base(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	params, _ := k.GetParams(ctx)
	params.BaseBurnPrice = 10000000
	params.BurnPriceIncreasePercent = 20
	params.BurnPriceDecayEpochs = 10
	k.SetParams(ctx, params)

	price := k.ComputeSlotPrice(ctx, 0)
	require.Equal(t, math.NewInt(10000000), price)
}

func TestComputeSlotPrice_IncreaseWithSlot(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	params, _ := k.GetParams(ctx)
	params.BaseBurnPrice = 10000000
	params.BurnPriceIncreasePercent = 20
	params.BurnPriceDecayEpochs = 10
	k.SetParams(ctx, params)

	price := k.ComputeSlotPrice(ctx, 15)
	require.True(t, price.GT(math.NewInt(10000000)))
}

func TestGenesisRoundTrip_Extended(t *testing.T) {
	storeService1, _ := testkeeper.NewTestStoreService(t, orgStoreKey)
	logger1 := testkeeper.NewTestLogger()
	bank1 := newMockBankKeeper()
	k1 := keeper.NewKeeper(storeService1, logger1, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", bank1, nil)
	ctx1 := context.Background()

	genesisState := &types.GenesisState{
		Orgs: []*types.Org{
			types.NewOrg("org1", "leader1_pubkey_1234567890123456789012345", "", 1000000, 5000),
		},
		Members: []*types.MemberRecord{
			types.NewMemberRecord("org1", "member1_pubkey_123456789012345678901", "member", testX25519Pubkey),
		},
		OrgConfigs: []*types.OrgConfig{
			{
				OrgID:                    "org1",
				ServeAttestationRequired: true,
			},
		},
	}

	err := k1.InitGenesis(ctx1, genesisState)
	if err != nil {
		t.Fatalf("InitGenesis failed: %v", err)
	}

	exportedState, err := k1.ExportGenesis(ctx1)
	if err != nil {
		t.Fatalf("ExportGenesis failed: %v", err)
	}
	if len(exportedState.OrgConfigs) != 1 {
		t.Fatalf("expected 1 org config, got: %d", len(exportedState.OrgConfigs))
	}
}

func TestMsgAddMember_RejectsNonLeaderWalletSigner(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(
		t,
		srv,
		ctx,
		"cosmos1vq0svzat0jyknkc6rfp40l8tr5cz4qxd6m6tyx",
		"leader_pubkey_12345678901234567890123456789012",
		"cosmos1t9xdz4tvmsm2qj8fxadue6yx5ysp30zv4rnau6",
	)

	_, err := srv.AddMember(ctx, &types.MsgAddMember{
		Signer:       "cosmos1vq0svzat0jyknkc6rfp40l8tr5cz4qxd6m6tyx",
		OrgId:        orgID,
		Pubkey:       "member_pubkey_12345678901234567890123456789012",
		Role:         "contributor",
		X25519Pubkey: testX25519Pubkey,
	})
	require.ErrorIs(t, err, types.ErrNotLeader)
}

func TestMsgRemoveMember_RejectsNonLeaderWalletSigner(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(
		t,
		srv,
		ctx,
		"cosmos1vq0svzat0jyknkc6rfp40l8tr5cz4qxd6m6tyx",
		"leader_pubkey_12345678901234567890123456789012",
		"cosmos1t9xdz4tvmsm2qj8fxadue6yx5ysp30zv4rnau6",
	)

	_, err := srv.AddMember(ctx, &types.MsgAddMember{
		Signer:       "cosmos1t9xdz4tvmsm2qj8fxadue6yx5ysp30zv4rnau6",
		OrgId:        orgID,
		Pubkey:       "member_pubkey_remove_1234567890123456789012",
		Role:         "member",
		X25519Pubkey: testX25519Pubkey,
	})
	require.NoError(t, err)

	_, err = srv.RemoveMember(ctx, &types.MsgRemoveMember{
		Signer: "cosmos1vq0svzat0jyknkc6rfp40l8tr5cz4qxd6m6tyx",
		OrgId:  orgID,
		Pubkey: "member_pubkey_remove_1234567890123456789012",
	})
	require.ErrorIs(t, err, types.ErrNotLeader)
}

func TestMsgRotateEpoch_RejectsNonLeaderWalletSigner(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(
		t,
		srv,
		ctx,
		"cosmos1vq0svzat0jyknkc6rfp40l8tr5cz4qxd6m6tyx",
		"leader_pubkey_12345678901234567890123456789012",
		"cosmos1t9xdz4tvmsm2qj8fxadue6yx5ysp30zv4rnau6",
	)

	_, err := srv.RotateEpoch(ctx, &types.MsgRotateEpoch{
		Signer: "cosmos1vq0svzat0jyknkc6rfp40l8tr5cz4qxd6m6tyx",
		OrgId:  orgID,
	})
	require.ErrorIs(t, err, types.ErrNotLeader)
}

func TestMsgAddMember_RejectsLeaderRole(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(
		t,
		srv,
		ctx,
		"cosmos1vq0svzat0jyknkc6rfp40l8tr5cz4qxd6m6tyx",
		"leader_pubkey_12345678901234567890123456789012",
		"cosmos1t9xdz4tvmsm2qj8fxadue6yx5ysp30zv4rnau6",
	)

	_, err := srv.AddMember(ctx, &types.MsgAddMember{
		Signer:       "cosmos1t9xdz4tvmsm2qj8fxadue6yx5ysp30zv4rnau6",
		OrgId:        orgID,
		Pubkey:       "member_pubkey_leader_1234567890123456789012",
		Role:         "leader",
		X25519Pubkey: testX25519Pubkey,
	})
	require.ErrorIs(t, err, types.ErrInvalidRole)
}

func TestOrgDecisionMsgs_AcceptsLeaderWalletSigner(t *testing.T) {
	srv, ctx, _, _ := setupMsgServer(t)

	orgID := registerMsgServerOrgWithLeaderWallet(
		t,
		srv,
		ctx,
		"cosmos1vq0svzat0jyknkc6rfp40l8tr5cz4qxd6m6tyx",
		"leader_pubkey_12345678901234567890123456789012",
		"cosmos1t9xdz4tvmsm2qj8fxadue6yx5ysp30zv4rnau6",
	)

	addResp, err := srv.AddMember(ctx, &types.MsgAddMember{
		Signer:       "cosmos1t9xdz4tvmsm2qj8fxadue6yx5ysp30zv4rnau6",
		OrgId:        orgID,
		Pubkey:       "member_pubkey_success_1234567890123456789012",
		Role:         "contributor",
		X25519Pubkey: testX25519Pubkey,
	})
	require.NoError(t, err)
	require.NotNil(t, addResp)

	rotateResp, err := srv.RotateEpoch(ctx, &types.MsgRotateEpoch{
		Signer: "cosmos1t9xdz4tvmsm2qj8fxadue6yx5ysp30zv4rnau6",
		OrgId:  orgID,
	})
	require.NoError(t, err)
	require.NotNil(t, rotateResp)
	require.Equal(t, uint64(1), rotateResp.NewEpoch)

	removeResp, err := srv.RemoveMember(ctx, &types.MsgRemoveMember{
		Signer: "cosmos1t9xdz4tvmsm2qj8fxadue6yx5ysp30zv4rnau6",
		OrgId:  orgID,
		Pubkey: "member_pubkey_success_1234567890123456789012",
	})
	require.NoError(t, err)
	require.NotNil(t, removeResp)
}
