package keeper_test

import (
	"context"
	"errors"
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
	orgStoreKey     = storetypes.NewKVStoreKey("org")
	testCreatorAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
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

func TestRegisterOrg(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)

	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	has, err := k.HasOrg(ctx, "org1")
	if err != nil {
		t.Fatalf("HasOrg failed: %v", err)
	}
	if !has {
		t.Fatal("expected org to exist")
	}
}

func TestRegisterOrg_DuplicateOrg(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)

	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("first RegisterOrg failed: %v", err)
	}

	err = k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != types.ErrOrgAlreadyExists {
		t.Fatalf("expected ErrOrgAlreadyExists, got: %v", err)
	}
}

func TestGetOrg(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)

	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	retrieved, err := k.GetOrg(ctx, "org1")
	if err != nil {
		t.Fatalf("GetOrg failed: %v", err)
	}
	if retrieved.OrgID != "org1" {
		t.Fatalf("expected org1, got: %s", retrieved.OrgID)
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

	member := types.NewMemberRecord("org1", "member_pubkey_123456789012345678901234", "member")
	err = k.AddMember(ctx, member)
	if err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	isMember, err := k.IsMember(ctx, "org1", "member_pubkey_123456789012345678901234")
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

	member := types.NewMemberRecord("org1", "member_pubkey_123456789012345678901234", "member")
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

	member := types.NewMemberRecord("org1", "member_pubkey_123456789012345678901234", "member")
	err = k.AddMember(ctx, member)
	if err != nil {
		t.Fatalf("AddMember failed: %v", err)
	}

	err = k.RemoveMember(ctx, "org1", "member_pubkey_123456789012345678901234")
	if err != nil {
		t.Fatalf("RemoveMember failed: %v", err)
	}

	isMember, err := k.IsMember(ctx, "org1", "member_pubkey_123456789012345678901234")
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

	isLeader, err := k.IsLeader(ctx, "org1", "leader_pubkey_12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("IsLeader failed: %v", err)
	}
	if !isLeader {
		t.Fatal("expected leader to be recognized as leader")
	}

	isLeader, err = k.IsLeader(ctx, "org1", "some_other_pubkey")
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

	err = k.UpdateOrgStatus(ctx, "org1", types.OrgStatus_DORMANT)
	if err != nil {
		t.Fatalf("UpdateOrgStatus failed: %v", err)
	}

	updated, err := k.GetOrg(ctx, "org1")
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

	err = k.UpdateStorageQuota(ctx, "org1", 2000000)
	if err != nil {
		t.Fatalf("UpdateStorageQuota failed: %v", err)
	}

	updated, err := k.GetOrg(ctx, "org1")
	if err != nil {
		t.Fatalf("GetOrg failed: %v", err)
	}
	if updated.StorageQuota != 2000000 {
		t.Fatalf("expected quota 2000000, got: %d", updated.StorageQuota)
	}
}

func TestDynamicPrice(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	dp := &types.DynamicPrice{
		Price:         1000,
		LastCreation:  100,
		CreationCount: 5,
	}

	err := k.SetDynamicPrice(ctx, dp)
	if err != nil {
		t.Fatalf("SetDynamicPrice failed: %v", err)
	}

	retrieved, err := k.GetDynamicPrice(ctx)
	if err != nil {
		t.Fatalf("GetDynamicPrice failed: %v", err)
	}
	if retrieved.Price != 1000 {
		t.Fatalf("expected price 1000, got: %d", retrieved.Price)
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
		types.NewMemberRecord("org1", "member1_pubkey_123456789012345678901", "member"),
		types.NewMemberRecord("org1", "member2_pubkey_123456789012345678901", "member"),
	}

	for _, m := range members {
		err := k.AddMember(ctx, m)
		if err != nil {
			t.Fatalf("AddMember failed: %v", err)
		}
	}

	allMembers, err := k.GetAllMembers(ctx, "org1")
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
			types.NewMemberRecord("org1", "member_pubkey_123456789012345678901234", "member"),
		},
		DynamicPrice: &types.DynamicPrice{
			Price:         1000,
			LastCreation:  100,
			CreationCount: 5,
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
	if exported.DynamicPrice.Price != state.DynamicPrice.Price {
		t.Fatalf("expected price %d, got: %d", state.DynamicPrice.Price, exported.DynamicPrice.Price)
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

	has, err := k1.HasOrg(ctx1, "org1")
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

	has, err = k2.HasOrg(ctx2, "org1")
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

	has, err := k2.HasOrg(ctx2, "org1")
	if err != nil {
		t.Fatalf("HasOrg failed: %v", err)
	}
	if !has {
		t.Fatal("expected org to persist after commit")
	}

	retrieved, err := k2.GetOrg(ctx2, "org1")
	if err != nil {
		t.Fatalf("GetOrg failed: %v", err)
	}
	if retrieved.OrgID != "org1" {
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
			types.NewMemberRecord("org1", "member1_pubkey_123456789012345678901", "member"),
		},
		DynamicPrice: &types.DynamicPrice{
			Price:         1000,
			LastCreation:  100,
			CreationCount: 5,
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

	members := []*types.MemberRecord{
		types.NewMemberRecord("org1", "member1_pubkey_123456789012345678901", "member"),
		types.NewMemberRecord("org1", "member2_pubkey_123456789012345678901", "member"),
		types.NewMemberRecord("org2", "member3_pubkey_123456789012345678901", "member"),
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

	for _, orgID := range []string{"org1", "org2"} {
		has, err := k.HasOrg(ctx, orgID)
		if err != nil {
			t.Fatalf("HasOrg failed: %v", err)
		}
		if !has {
			t.Errorf("expected org %s to exist", orgID)
		}
	}

	org1Members, err := k.GetAllMembers(ctx, "org1")
	if err != nil {
		t.Fatalf("GetAllMembers failed: %v", err)
	}
	if len(org1Members) != 3 {
		t.Fatalf("expected 3 members for org1, got: %d", len(org1Members))
	}
}

func TestFundTreasury(t *testing.T) {
	k, ctx, _ := newTestKeeperWithFundedBank(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)
	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	funder, _ := sdk.AccAddressFromBech32("cosmos1abc")
	err = k.FundTreasury(ctx, "org1", funder, math.NewInt(500000))
	if err != nil {
		t.Fatalf("FundTreasury failed: %v", err)
	}

	balance, err := k.GetTreasuryBalance(ctx, "org1")
	if err != nil {
		t.Fatalf("GetTreasuryBalance failed: %v", err)
	}
	require.Equal(t, "500000", balance)
}

func TestFundTreasury_InsufficientBalance(t *testing.T) {
	k, ctx, _ := newTestKeeperWithStrictBank(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)
	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	funder, _ := sdk.AccAddressFromBech32("cosmos1abc")
	err = k.FundTreasury(ctx, "org1", funder, math.NewInt(500000))
	if err == nil {
		t.Fatal("expected FundTreasury to fail with insufficient funds")
	}
}

func TestWithdrawTreasury(t *testing.T) {
	k, ctx, _ := newTestKeeperWithFundedBank(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)
	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	funder, _ := sdk.AccAddressFromBech32("cosmos1abc")
	err = k.FundTreasury(ctx, "org1", funder, math.NewInt(500000))
	if err != nil {
		t.Fatalf("FundTreasury failed: %v", err)
	}

	recipient, _ := sdk.AccAddressFromBech32("cosmos1def")
	err = k.WithdrawTreasury(ctx, "org1", recipient, math.NewInt(200000))
	if err != nil {
		t.Fatalf("WithdrawTreasury failed: %v", err)
	}

	balance, err := k.GetTreasuryBalance(ctx, "org1")
	if err != nil {
		t.Fatalf("GetTreasuryBalance failed: %v", err)
	}
	require.Equal(t, "300000", balance)
}

func TestWithdrawTreasury_InsufficientTreasury(t *testing.T) {
	k, ctx, _ := newTestKeeperWithFundedBank(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)
	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	funder, _ := sdk.AccAddressFromBech32("cosmos1abc")
	err = k.FundTreasury(ctx, "org1", funder, math.NewInt(100000))
	if err != nil {
		t.Fatalf("FundTreasury failed: %v", err)
	}

	recipient, _ := sdk.AccAddressFromBech32("cosmos1def")
	err = k.WithdrawTreasury(ctx, "org1", recipient, math.NewInt(200000))
	if err != types.ErrInsufficientTreasury {
		t.Fatalf("expected ErrInsufficientTreasury, got: %v", err)
	}
}

func TestDebitTreasury(t *testing.T) {
	k, ctx, _ := newTestKeeperWithFundedBank(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)
	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	balance, err := k.GetTreasuryBalanceInt(ctx, "org1")
	if err != nil {
		t.Fatalf("GetTreasuryBalanceInt failed: %v", err)
	}
	require.True(t, balance.IsZero())
}

func TestGetTreasuryBalance_NoTreasury(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)
	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	balance, err := k.GetTreasuryBalance(ctx, "org1")
	if err != nil {
		t.Fatalf("GetTreasuryBalance failed: %v", err)
	}
	require.Equal(t, "0", balance)
}

func TestSetRepTiers(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)
	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	tiers := []*types.RepTierRecord{
		{
			MinReputation:            0,
			MaxReputation:            50,
			MaxContributionsPerEpoch: 3,
			PayoutPerMemory:          "1",
		},
		{
			MinReputation:            200,
			MaxReputation:            1000,
			MaxContributionsPerEpoch: 50,
			PayoutPerMemory:          "5",
		},
	}

	err = k.SetRepTiers(ctx, "org1", tiers)
	if err != nil {
		t.Fatalf("SetRepTiers failed: %v", err)
	}

	cfg, err := k.GetRepTiers(ctx, "org1")
	if err != nil {
		t.Fatalf("GetRepTiers failed: %v", err)
	}
	require.Len(t, cfg.Tiers, 2)
}

func TestSetOrgConfig(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	org := types.NewOrg("org1", "leader_pubkey_12345678901234567890123456789012", "", 1000000, 5000)
	err := k.RegisterOrg(ctx, org, testCreatorAddr)
	if err != nil {
		t.Fatalf("RegisterOrg failed: %v", err)
	}

	cfg := &types.OrgConfig{
		OrgID:                    "org1",
		ServeAttestationRequired: true,
	}

	err = k.SetOrgConfig(ctx, "org1", cfg)
	if err != nil {
		t.Fatalf("SetOrgConfig failed: %v", err)
	}

	retrieved, err := k.GetOrgConfig(ctx, "org1")
	if err != nil {
		t.Fatalf("GetOrgConfig failed: %v", err)
	}
	require.True(t, retrieved.ServeAttestationRequired)
}

func TestComputeBurnPrice_Base(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	params, _ := k.GetParams(ctx)
	params.BaseBurnPrice = 10000000
	params.BurnPriceIncreasePercent = 20
	params.BurnPriceDecayEpochs = 10
	k.SetParams(ctx, params)

	price := k.ComputeBurnPrice(ctx)
	require.Equal(t, math.NewInt(10000000), price)
}

func TestComputeBurnPrice_Increase(t *testing.T) {
	k, ctx, _ := newTestKeeper(t)

	params, _ := k.GetParams(ctx)
	params.BaseBurnPrice = 10000000
	params.BurnPriceIncreasePercent = 20
	params.BurnPriceDecayEpochs = 10
	k.SetParams(ctx, params)

	dp := &types.DynamicPrice{
		Price:         10000000,
		LastCreation:  0,
		CreationCount: 15,
	}
	k.SetDynamicPrice(ctx, dp)

	price := k.ComputeBurnPrice(ctx)
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
			types.NewMemberRecord("org1", "member1_pubkey_123456789012345678901", "member"),
		},
		DynamicPrice: &types.DynamicPrice{
			Price:         1000,
			LastCreation:  100,
			CreationCount: 5,
		},
		Treasuries: []*types.Treasury{
			{OrgID: "org1", Balance: "500000"},
		},
		RepTiers: []*types.RepTierConfig{
			{
				OrgID: "org1",
				Tiers: []*types.RepTierRecord{
					{
						MinReputation:            0,
						MaxReputation:            50,
						MaxContributionsPerEpoch: 3,
						PayoutPerMemory:          "1",
					},
				},
			},
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

	if len(exportedState.Treasuries) != 1 {
		t.Fatalf("expected 1 treasury, got: %d", len(exportedState.Treasuries))
	}
	if len(exportedState.RepTiers) != 1 {
		t.Fatalf("expected 1 rep tier config, got: %d", len(exportedState.RepTiers))
	}
	if len(exportedState.OrgConfigs) != 1 {
		t.Fatalf("expected 1 org config, got: %d", len(exportedState.OrgConfigs))
	}
}
