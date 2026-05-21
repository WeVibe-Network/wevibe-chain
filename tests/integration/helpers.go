package integration

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	logv2 "cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	"github.com/wevibe-network/wevibe-chain/app"
	orgtypes "github.com/wevibe-network/wevibe-chain/x/org/types"
)

type TestSuite struct {
	T *testing.T
	*app.WeVibeApp
	Ctx sdk.Context

	ValidatorKey  *secp256k1.PrivKey
	ValidatorAddr sdk.AccAddress
	UserKey       *secp256k1.PrivKey
	UserAddr      sdk.AccAddress
}

var integrationBootstrapOnce sync.Once

func ensureIntegrationBootstrap() {
	integrationBootstrapOnce.Do(func() {
		app.EnsureEncodingRegistered()
	})
}

func NewTestSuite(t *testing.T) *TestSuite {
	t.Helper()
	ensureIntegrationBootstrap()

	db := dbm.NewMemDB()
	wevibeApp := app.NewWeVibeApp(logv2.NewNopLogger(), db, true, simtestutil.EmptyAppOptions{}, baseapp.SetChainID("wevibe-test-1"))

	validatorKey := secp256k1.GenPrivKey()
	validatorAddr := sdk.AccAddress(validatorKey.PubKey().Address())
	userKey := secp256k1.GenPrivKey()
	userAddr := sdk.AccAddress(userKey.PubKey().Address())

	genesisState := wevibeApp.BasicManager.DefaultGenesis(wevibeApp.AppCodec())

	selfDelegation := sdkmath.NewInt(1_000_000)
	valAddr := sdk.ValAddress(validatorKey.PubKey().Address())

	validator := stakingtestutil.NewValidator(t, valAddr, validatorKey.PubKey())
	validator.Status = stakingtypes.Bonded
	validator.Tokens = selfDelegation
	validator.DelegatorShares = sdkmath.LegacyNewDecFromInt(selfDelegation)
	validator.MinSelfDelegation = sdkmath.OneInt()

	delegation := stakingtypes.NewDelegation(validatorAddr.String(), valAddr.String(), sdkmath.LegacyNewDecFromInt(selfDelegation))

	stakingGenesis := stakingtypes.DefaultGenesisState()
	stakingGenesis.Params.BondDenom = "uvibe"
	stakingGenesis.Validators = []stakingtypes.Validator{validator}
	stakingGenesis.Delegations = []stakingtypes.Delegation{delegation}
	stakingGenesis.LastTotalPower = selfDelegation
	stakingGenesis.LastValidatorPowers = []stakingtypes.LastValidatorPower{{
		Address: valAddr.String(),
		Power:   selfDelegation.Int64(),
	}}

	bankGenesis := banktypes.DefaultGenesisState()
	bondedPoolAccount := authtypes.NewEmptyModuleAccount(stakingtypes.BondedPoolName, authtypes.Burner, authtypes.Staking)
	notBondedPoolAccount := authtypes.NewEmptyModuleAccount(stakingtypes.NotBondedPoolName, authtypes.Burner, authtypes.Staking)
	mintModuleAccount := authtypes.NewEmptyModuleAccount("mint", authtypes.Minter)
	distrModuleAccount := authtypes.NewEmptyModuleAccount("distr")
	orgModuleAccount := authtypes.NewEmptyModuleAccount(orgtypes.ModuleName, authtypes.Burner)

	bankGenesis.Balances = []banktypes.Balance{
		{
			Address: validatorAddr.String(),
			Coins:   sdk.NewCoins(sdk.NewCoin("uvibe", selfDelegation)),
		},
		{
			Address: userAddr.String(),
			Coins:   sdk.NewCoins(sdk.NewCoin("uvibe", sdkmath.NewInt(500_000_000))),
		},
		{
			Address: bondedPoolAccount.GetAddress().String(),
			Coins:   sdk.NewCoins(sdk.NewCoin("uvibe", selfDelegation)),
		},
		{
			Address: notBondedPoolAccount.GetAddress().String(),
			Coins:   sdk.NewCoins(),
		},
		{
			Address: mintModuleAccount.GetAddress().String(),
			Coins:   sdk.NewCoins(),
		},
		{
			Address: distrModuleAccount.GetAddress().String(),
			Coins:   sdk.NewCoins(),
		},
		{
			Address: orgModuleAccount.GetAddress().String(),
			Coins:   sdk.NewCoins(),
		},
	}
	bankGenesis.Supply = sdk.NewCoins(sdk.NewCoin("uvibe", selfDelegation.Mul(sdkmath.NewInt(2)).Add(sdkmath.NewInt(500_000_000))))

	genAccounts := authtypes.GenesisAccounts{
		authtypes.NewBaseAccountWithAddress(validatorAddr),
		authtypes.NewBaseAccountWithAddress(userAddr),
		bondedPoolAccount,
		notBondedPoolAccount,
		mintModuleAccount,
		distrModuleAccount,
		orgModuleAccount,
	}
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccounts)

	genesisState[authtypes.ModuleName] = wevibeApp.AppCodec().MustMarshalJSON(authGenesis)
	genesisState[banktypes.ModuleName] = wevibeApp.AppCodec().MustMarshalJSON(bankGenesis)
	genesisState[stakingtypes.ModuleName] = wevibeApp.AppCodec().MustMarshalJSON(stakingGenesis)

	stateBytes, err := json.Marshal(genesisState)
	require.NoError(t, err)

	_, err = wevibeApp.BaseApp.InitChain(&abci.RequestInitChain{
		ChainId:       "wevibe-test-1",
		AppStateBytes: stateBytes,
		Time:          time.Now(),
	})
	require.NoError(t, err)

	_, err = wevibeApp.BaseApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: 1, Time: time.Now()})
	require.NoError(t, err)

	_, err = wevibeApp.BaseApp.Commit()
	require.NoError(t, err)

	return &TestSuite{
		T:             t,
		WeVibeApp:       wevibeApp,
		Ctx:           wevibeApp.BaseApp.NewNextBlockContext(cmtproto.Header{Height: 2}),
		ValidatorKey:  validatorKey,
		ValidatorAddr: validatorAddr,
		UserKey:       userKey,
		UserAddr:      userAddr,
	}
}

func (s *TestSuite) DeliverMsg(msgs ...sdk.Msg) (*sdk.Result, error) {
	handler := s.MsgServiceRouter().Handler(msgs[0])
	if handler == nil {
		return nil, fmt.Errorf("no handler found for %T", msgs[0])
	}

	ctx := sdk.NewContext(s.WeVibeApp.BaseApp.CommitMultiStore(), s.Ctx.BlockHeader(), false, s.Ctx.Logger())
	result, err := handler(ctx, msgs[0])
	if err != nil {
		return nil, err
	}
	s.Ctx = ctx
	return result, nil
}

func (s *TestSuite) QueryOrg(orgID string) (*orgtypes.Org, error) {
	return s.OrgKeeper.GetOrg(s.Ctx, orgID)
}

type MsgDeliverOption func(msg sdk.Msg) sdk.Msg

func SignMsg(msg sdk.Msg, privKey *secp256k1.PrivKey, accNum, accSeq uint64) MsgDeliverOption {
	return func(m sdk.Msg) sdk.Msg {
		return m
	}
}
