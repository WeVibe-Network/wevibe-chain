package app_test

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	types "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/wevibe-network/wevibe-chain/app"
)

func TestNewWeVibeApp(t *testing.T) {
	logger := log.NewNopLogger()
	db := dbm.NewMemDB()

	wevibeApp := app.NewWeVibeApp(logger, db, true, simtestutil.EmptyAppOptions{})
	require.NotNil(t, wevibeApp)
	require.NotNil(t, wevibeApp.ModuleManager)

	require.NotNil(t, wevibeApp.AccountKeeper)
	require.NotNil(t, wevibeApp.BankKeeper)
	require.NotNil(t, wevibeApp.StakingKeeper)
	require.NotNil(t, wevibeApp.MintKeeper)
	require.NotNil(t, wevibeApp.DistributionKeeper)
	require.NotNil(t, wevibeApp.SlashingKeeper)
	require.NotNil(t, wevibeApp.FeegrantKeeper)

	require.NotNil(t, wevibeApp.OrgKeeper)
	require.NotNil(t, wevibeApp.EmissionsKeeper)
	require.NotNil(t, wevibeApp.ReputationKeeper)
	require.NotNil(t, wevibeApp.AttestationKeeper)

	require.NotNil(t, wevibeApp.TxConfig().TxDecoder())
	require.NotNil(t, wevibeApp.TxConfig().TxEncoder())

	blocked := app.BlockedAddresses()
	require.Contains(t, blocked, authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String())
	require.Contains(t, blocked, authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName).String())
	require.Contains(t, blocked, authtypes.NewModuleAddress(distrtypes.ModuleName).String())
}

func TestWeVibeAppInitChainAndExport(t *testing.T) {
	logger := log.NewNopLogger()
	db := dbm.NewMemDB()

	wevibeApp := app.NewWeVibeApp(logger, db, true, simtestutil.EmptyAppOptions{})
	require.NotNil(t, wevibeApp)

	genesisState := wevibeApp.BasicManager.DefaultGenesis(wevibeApp.AppCodec())

	// seed staking genesis with a bonded validator so InitGenesis returns updates
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	delAddr := sdk.AccAddress(pubKey.Address())
	valAddr := sdk.ValAddress(pubKey.Address())

	validator := stakingtestutil.NewValidator(t, valAddr, pubKey)
	selfDelegation := sdkmath.NewInt(1_000_000)
	validator.Status = stakingtypes.Bonded
	validator.Tokens = selfDelegation
	validator.DelegatorShares = sdkmath.LegacyNewDecFromInt(selfDelegation)
	validator.MinSelfDelegation = sdkmath.OneInt()

	delegation := stakingtypes.NewDelegation(delAddr.String(), valAddr.String(), sdkmath.LegacyNewDecFromInt(selfDelegation))

	stakingGenesis := stakingtypes.DefaultGenesisState()
	stakingGenesis.Params.BondDenom = app.BondDenom
	stakingGenesis.Validators = []stakingtypes.Validator{validator}
	stakingGenesis.Delegations = []stakingtypes.Delegation{delegation}
	stakingGenesis.LastTotalPower = selfDelegation
	stakingGenesis.LastValidatorPowers = []stakingtypes.LastValidatorPower{{
		Address: valAddr.String(),
		Power:   selfDelegation.Int64(),
	}}

	coins := sdk.NewCoins(sdk.NewCoin(app.BondDenom, selfDelegation))
	bankGenesis := banktypes.DefaultGenesisState()
	bondedPoolAccount := authtypes.NewEmptyModuleAccount(stakingtypes.BondedPoolName, authtypes.Burner, authtypes.Staking)
	notBondedPoolAccount := authtypes.NewEmptyModuleAccount(stakingtypes.NotBondedPoolName, authtypes.Burner, authtypes.Staking)
	feeCollectorAccount := authtypes.NewEmptyModuleAccount(authtypes.FeeCollectorName)
	mintModuleAccount := authtypes.NewEmptyModuleAccount(minttypes.ModuleName, authtypes.Minter)
	distributionModuleAccount := authtypes.NewEmptyModuleAccount(distrtypes.ModuleName)

	bankGenesis.Balances = []banktypes.Balance{
		{
			Address: bondedPoolAccount.GetAddress().String(),
			Coins:   coins,
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
			Address: distributionModuleAccount.GetAddress().String(),
			Coins:   sdk.NewCoins(),
		},
		{
			Address: feeCollectorAccount.GetAddress().String(),
			Coins:   sdk.NewCoins(),
		},
		{
			Address: delAddr.String(),
			Coins:   sdk.NewCoins(),
		},
	}
	bankGenesis.Supply = coins

	baseAccount := authtypes.NewBaseAccountWithAddress(delAddr)
	genAccounts := authtypes.SanitizeGenesisAccounts(authtypes.GenesisAccounts{
		baseAccount,
		bondedPoolAccount,
		notBondedPoolAccount,
		feeCollectorAccount,
		mintModuleAccount,
		distributionModuleAccount,
	})
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccounts)

	genesisState[authtypes.ModuleName] = wevibeApp.AppCodec().MustMarshalJSON(authGenesis)
	genesisState[banktypes.ModuleName] = wevibeApp.AppCodec().MustMarshalJSON(bankGenesis)
	genesisState[stakingtypes.ModuleName] = wevibeApp.AppCodec().MustMarshalJSON(stakingGenesis)
	stateBytes, err := json.Marshal(genesisState)
	require.NoError(t, err)

	ctx := wevibeApp.BaseApp.NewUncachedContext(false, tmproto.Header{})
	_, err = wevibeApp.InitChainer(ctx, &types.RequestInitChain{AppStateBytes: stateBytes})
	require.NoError(t, err)

	blockCtx := ctx.WithBlockHeader(tmproto.Header{Height: 1})
	_, err = wevibeApp.BeginBlocker(blockCtx)
	require.NoError(t, err)

	_, err = wevibeApp.EndBlocker(blockCtx)
	require.NoError(t, err)

	export, err := wevibeApp.ExportAppStateAndValidators(false, nil, nil)
	require.NoError(t, err)
	require.NotEmpty(t, export.AppState)
	require.NotNil(t, export.ConsensusParams)

	stakingParams, err := wevibeApp.StakingKeeper.GetParams(blockCtx)
	require.NoError(t, err)
	require.Equal(t, app.BondDenom, stakingParams.BondDenom)

	mintParams, err := wevibeApp.MintKeeper.Params.Get(blockCtx)
	require.NoError(t, err)
	require.Equal(t, app.BondDenom, mintParams.MintDenom)

	distAddr := wevibeApp.AccountKeeper.GetModuleAddress(distrtypes.ModuleName)
	require.NotNil(t, distAddr)

	slashingParams, err := wevibeApp.SlashingKeeper.GetParams(blockCtx)
	require.NoError(t, err)
	require.NotZero(t, slashingParams.SignedBlocksWindow)

	// CO-040: the emissions module's HasGenesis InitGenesis must seed an
	// emission pool from DefaultParams(). Without this the epoch hook logs
	// "no emission pool found" and never mints.
	emissionPool, err := wevibeApp.EmissionsKeeper.GetEmissionPool(blockCtx)
	require.NoError(t, err)
	require.NotNil(t, emissionPool)
	require.NotZero(t, emissionPool.DailyMint)
	require.Equal(t, uint64(100), emissionPool.OperatorShare+emissionPool.ValidatorShare)

	// CO-040: the reputation module must be active after genesis (GAP-REP-1).
	require.True(t, wevibeApp.ReputationKeeper.IsActive(blockCtx),
		"reputation module must be active after InitChain")
}
