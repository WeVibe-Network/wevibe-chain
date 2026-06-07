package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	// SDK foundation: Cosmos SDK v0.53.5 + CometBFT v0.38.20 (D-S29-SDK-V053).
	// Do NOT bump to v0.54.x - that line uses store/v2 in baseapp and breaks
	// cosmossdk.io/x/upgrade compatibility. See CO-008 implementation report.
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	feegrant "cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	signingtypes "cosmossdk.io/x/tx/signing"
	upgrade "cosmossdk.io/x/upgrade"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	tx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamTypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/epochs"
	epochskeeper "github.com/cosmos/cosmos-sdk/x/epochs/keeper"
	epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilTypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cast"

	v2 "github.com/wevibe-network/wevibe-chain/app/upgrades/v2"
	attestationkeeper "github.com/wevibe-network/wevibe-chain/x/attestation/keeper"
	attestationmodule "github.com/wevibe-network/wevibe-chain/x/attestation/module"
	attestationTypes "github.com/wevibe-network/wevibe-chain/x/attestation/types"
	bandwidthkeeper "github.com/wevibe-network/wevibe-chain/x/bandwidth/keeper"
	bandwidthmodule "github.com/wevibe-network/wevibe-chain/x/bandwidth/module"
	bandwidthTypes "github.com/wevibe-network/wevibe-chain/x/bandwidth/types"
	emissionskeeper "github.com/wevibe-network/wevibe-chain/x/emissions/keeper"
	emissionsmodule "github.com/wevibe-network/wevibe-chain/x/emissions/module"
	emissionsTypes "github.com/wevibe-network/wevibe-chain/x/emissions/types"
	identitykeeper "github.com/wevibe-network/wevibe-chain/x/identity/keeper"
	identitymodule "github.com/wevibe-network/wevibe-chain/x/identity/module"
	identityTypes "github.com/wevibe-network/wevibe-chain/x/identity/types"
	memorykeeper "github.com/wevibe-network/wevibe-chain/x/memory/keeper"
	memorymodule "github.com/wevibe-network/wevibe-chain/x/memory/module"
	memoryTypes "github.com/wevibe-network/wevibe-chain/x/memory/types"
	orgkeeper "github.com/wevibe-network/wevibe-chain/x/org/keeper"
	orgmodule "github.com/wevibe-network/wevibe-chain/x/org/module"
	orgTypes "github.com/wevibe-network/wevibe-chain/x/org/types"
	reputationkeeper "github.com/wevibe-network/wevibe-chain/x/reputation/keeper"
	reputationmodule "github.com/wevibe-network/wevibe-chain/x/reputation/module"
	reputationTypes "github.com/wevibe-network/wevibe-chain/x/reputation/types"
	servekeeper "github.com/wevibe-network/wevibe-chain/x/serve/keeper"
	servemodule "github.com/wevibe-network/wevibe-chain/x/serve/module"
	serveTypes "github.com/wevibe-network/wevibe-chain/x/serve/types"
)

const (
	AppName          = "wevibed"
	BondDenom        = "uvibe"
	Bech32Prefix     = "wevibe"
	Bech32PrefixVal  = "wevibevaloper"
	Bech32PrefixCons = "wevibevalcons"
)

var (
	DefaultNodeHome string

	maccPerms = map[string][]string{
		authTypes.FeeCollectorName:     nil,
		distrTypes.ModuleName:          nil,
		mintTypes.ModuleName:           {authTypes.Minter},
		stakingTypes.BondedPoolName:    {authTypes.Burner, authTypes.Staking},
		stakingTypes.NotBondedPoolName: {authTypes.Burner, authTypes.Staking},
		govtypes.ModuleName:            {authTypes.Burner},
		"org":                          {authTypes.Burner},
	}
)

var (
	encodingOnce               sync.Once
	basicManagerInterfacesOnce sync.Once
	sharedInterfaceRegistry    codectypes.InterfaceRegistry
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	DefaultNodeHome = filepath.Join(userHomeDir, ".wevibed")

	cfg := sdk.GetConfig()
	sdk.DefaultBondDenom = BondDenom
	cfg.SetBech32PrefixForAccount(Bech32Prefix, Bech32Prefix+"pub")
	cfg.SetBech32PrefixForValidator(Bech32PrefixVal, Bech32PrefixVal+"pub")
	cfg.SetBech32PrefixForConsensusNode(Bech32PrefixCons, Bech32PrefixCons+"pub")
	cfg.Seal()
}

func EnsureEncodingRegistered() {
	encodingOnce.Do(func() {
		registry, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
			ProtoFiles: gogoproto.HybridResolver,
			SigningOptions: signingtypes.Options{
				AddressCodec:          authcodec.NewBech32Codec(Bech32Prefix),
				ValidatorAddressCodec: authcodec.NewBech32Codec(Bech32PrefixVal),
			},
		})
		if err != nil {
			panic(err)
		}
		std.RegisterInterfaces(registry)
		ModuleBasics.RegisterInterfaces(registry)
		RegisterInterfaces(registry)
		sharedInterfaceRegistry = registry
	})
}

type WeVibeApp struct {
	*baseapp.BaseApp

	cdc               codec.Codec
	legacyAmino       *codec.LegacyAmino
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry

	keys map[string]*storetypes.KVStoreKey

	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.BaseKeeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	DistributionKeeper    distrkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	EpochsKeeper          epochskeeper.Keeper
	GovKeeper             *govkeeper.Keeper
	FeegrantKeeper        feegrantkeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper

	ModuleManager *module.Manager
	BasicManager  module.BasicManager

	BandwidthKeeper   *bandwidthkeeper.Keeper
	EmissionsKeeper   *emissionskeeper.Keeper
	MemoryKeeper      *memorykeeper.Keeper
	OrgKeeper         *orgkeeper.Keeper
	IdentityKeeper    *identitykeeper.Keeper
	ReputationKeeper  *reputationkeeper.Keeper
	ServeKeeper       *servekeeper.Keeper
	AttestationKeeper *attestationkeeper.Keeper
}

var (
	_ runtime.AppI            = (*WeVibeApp)(nil)
	_ servertypes.Application = (*WeVibeApp)(nil)
)

func NewWeVibeApp(
	logger log.Logger,
	db dbm.DB,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *WeVibeApp {
	EnsureEncodingRegistered()
	interfaceRegistry := sharedInterfaceRegistry
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()
	txConfig := tx.NewTxConfig(appCodec, tx.DefaultSignModes)

	std.RegisterLegacyAminoCodec(legacyAmino)

	bApp := baseapp.NewBaseApp(AppName, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	keys := storetypes.NewKVStoreKeys(
		authTypes.StoreKey,
		bankTypes.StoreKey,
		stakingTypes.StoreKey,
		mintTypes.StoreKey,
		distrTypes.StoreKey,
		slashingTypes.StoreKey,
		consensusparamTypes.StoreKey,
		epochstypes.StoreKey,
		govtypes.StoreKey,
		feegrant.StoreKey,
		authzkeeper.StoreKey,
		upgradetypes.StoreKey,
		"bandwidth",
		"emissions",
		"memory",
		"org",
		"identity",
		"reputation",
		"serve",
		"attestation",
	)

	if err := bApp.RegisterStreamingServices(appOpts, keys); err != nil {
		panic(err)
	}

	wevibeKeys := map[string]*storetypes.KVStoreKey{
		"bandwidth":   keys["bandwidth"],
		"emissions":   keys["emissions"],
		"memory":      keys["memory"],
		"org":         keys["org"],
		"identity":    keys["identity"],
		"reputation":  keys["reputation"],
		"serve":       keys["serve"],
		"attestation": keys["attestation"],
	}

	app := &WeVibeApp{
		BaseApp:           bApp,
		cdc:               appCodec,
		legacyAmino:       legacyAmino,
		txConfig:          txConfig,
		interfaceRegistry: interfaceRegistry,
		keys:              keys,
	}

	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensusparamTypes.StoreKey]),
		authTypes.NewModuleAddress("gov").String(),
		runtime.EventService{},
	)
	bApp.SetParamStore(app.ConsensusParamsKeeper.ParamsStore)

	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authTypes.StoreKey]),
		authTypes.ProtoBaseAccount,
		maccPerms,
		authcodec.NewBech32Codec(Bech32Prefix),
		Bech32Prefix,
		authTypes.NewModuleAddress("gov").String(),
		authkeeper.WithUnorderedTransactions(true),
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[bankTypes.StoreKey]),
		app.AccountKeeper,
		BlockedAddresses(),
		authTypes.NewModuleAddress(authTypes.FeeCollectorName).String(),
		logger,
	)

	app.FeegrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[feegrant.StoreKey]),
		app.AccountKeeper,
	).SetBankKeeper(app.BankKeeper)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[authzkeeper.StoreKey]),
		appCodec,
		bApp.MsgServiceRouter(),
		app.AccountKeeper,
	).SetBankKeeper(app.BankKeeper)

	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[stakingTypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		authTypes.NewModuleAddress("gov").String(),
		authcodec.NewBech32Codec(Bech32PrefixVal),
		authcodec.NewBech32Codec(Bech32PrefixCons),
	)

	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[mintTypes.StoreKey]),
		app.StakingKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		authTypes.FeeCollectorName,
		authTypes.NewModuleAddress("gov").String(),
	)

	app.DistributionKeeper = distrkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[distrTypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		authTypes.FeeCollectorName,
		authTypes.NewModuleAddress("gov").String(),
	)

	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		runtime.NewKVStoreService(keys[slashingTypes.StoreKey]),
		app.StakingKeeper,
		authTypes.NewModuleAddress("gov").String(),
	)

	app.StakingKeeper.SetHooks(
		stakingTypes.NewMultiStakingHooks(
			app.DistributionKeeper.Hooks(),
			app.SlashingKeeper.Hooks(),
		),
	)

	govConfig := govtypes.DefaultConfig()

	app.GovKeeper = govkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[govtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.DistributionKeeper,
		bApp.MsgServiceRouter(),
		govConfig,
		authTypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.EpochsKeeper = epochskeeper.NewKeeper(
		runtime.NewKVStoreService(keys[epochstypes.StoreKey]),
		appCodec,
	)

	govAuthority := authTypes.NewModuleAddress("gov").String()
	skipUpgradeHeights := make(map[int64]bool)
	for _, height := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(height)] = true
	}
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))
	if homePath == "" {
		homePath = DefaultNodeHome
	}

	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		appCodec,
		homePath,
		bApp,
		govAuthority,
	)

	app.OrgKeeper = orgkeeper.NewKeeper(runtime.NewKVStoreService(wevibeKeys["org"]), logger, govAuthority, app.BankKeeper, app.FeegrantKeeper)
	app.BandwidthKeeper = bandwidthkeeper.NewKeeper(runtime.NewKVStoreService(wevibeKeys["bandwidth"]), logger, govAuthority, app.OrgKeeper)
	app.IdentityKeeper = identitykeeper.NewKeeper(runtime.NewKVStoreService(wevibeKeys["identity"]), logger, govAuthority)
	app.ReputationKeeper = reputationkeeper.NewKeeper(runtime.NewKVStoreService(wevibeKeys["reputation"]), logger, govAuthority)
	app.MemoryKeeper = memorykeeper.NewKeeper(runtime.NewKVStoreService(wevibeKeys["memory"]), logger, govAuthority, app.OrgKeeper, app.ReputationKeeper)
	app.ServeKeeper = servekeeper.NewKeeper(runtime.NewKVStoreService(wevibeKeys["serve"]), logger, govAuthority, app.OrgKeeper, app.MemoryKeeper, app.BandwidthKeeper, app.ReputationKeeper)
	app.EmissionsKeeper = emissionskeeper.NewKeeper(runtime.NewKVStoreService(wevibeKeys["emissions"]), logger, govAuthority, app.ServeKeeper, app.MemoryKeeper, app.OrgKeeper, app.ReputationKeeper)
	app.AttestationKeeper = attestationkeeper.NewKeeper(runtime.NewKVStoreService(wevibeKeys["attestation"]), logger, govAuthority, app.OrgKeeper)
	app.MemoryKeeper.SetServeKeeper(app.ServeKeeper)
	app.OrgKeeper.SetMemoryKeeper(app.MemoryKeeper)
	app.OrgKeeper.SetServeKeeper(app.ServeKeeper)
	app.OrgKeeper.SetBandwidthKeeper(app.BandwidthKeeper)
	app.ReputationKeeper.SetServeKeeper(app.ServeKeeper)
	app.ReputationKeeper.SetMemoryKeeper(app.MemoryKeeper)

	app.EpochsKeeper.SetHooks(
		epochstypes.NewMultiEpochHooks(
			app.EmissionsKeeper,
			app.MemoryKeeper,
		),
	)

	authModule := auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, nil)
	stakingModule := staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, nil)
	distrModule := distribution.NewAppModule(appCodec, app.DistributionKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, nil)
	slashingModule := slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, nil, nil)
	mintModule := mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil, nil)
	consensusModule := consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper)
	genutilModule := genutil.NewAppModule(app.AccountKeeper, app.StakingKeeper, app, txConfig)
	govModule := gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper, nil)
	feegrantMod := feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeegrantKeeper, app.interfaceRegistry)
	authzMod := authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry)
	upgradeModule := upgrade.NewAppModule(app.UpgradeKeeper, app.AccountKeeper.AddressCodec())

	emissionsMod := emissionsmodule.NewModule(app.EmissionsKeeper)
	bandwidthMod := bandwidthmodule.NewModule(app.BandwidthKeeper)
	memoryMod := memorymodule.NewModule(app.MemoryKeeper)
	orgMod := orgmodule.NewModule(app.OrgKeeper)
	identityMod := identitymodule.NewModule(app.IdentityKeeper)
	reputationMod := reputationmodule.NewModule(app.ReputationKeeper)
	serveMod := servemodule.NewModule(app.ServeKeeper)
	attestationMod := attestationmodule.NewModule(app.AttestationKeeper)

	app.ModuleManager = module.NewManagerFromMap(map[string]appmodule.AppModule{
		authTypes.ModuleName:           authModule,
		bankTypes.ModuleName:           bankModule,
		stakingTypes.ModuleName:        stakingModule,
		distrTypes.ModuleName:          distrModule,
		slashingTypes.ModuleName:       slashingModule,
		mintTypes.ModuleName:           mintModule,
		consensusparamTypes.ModuleName: consensusModule,
		genutilTypes.ModuleName:        genutilModule,
		epochstypes.ModuleName:         epochs.NewAppModule(app.EpochsKeeper),
		govtypes.ModuleName:            govModule,
		feegrant.ModuleName:            feegrantMod,
		authz.ModuleName:               authzMod,
		upgradetypes.ModuleName:        upgradeModule,

		"bandwidth":   bandwidthMod,
		"emissions":   emissionsMod,
		"memory":      memoryMod,
		"org":         orgMod,
		"identity":    identityMod,
		"reputation":  reputationMod,
		"serve":       serveMod,
		"attestation": attestationMod,
	})

	app.BasicManager = module.NewBasicManagerFromManager(
		app.ModuleManager,
		map[string]module.AppModuleBasic{
			genutilTypes.ModuleName: genutil.NewAppModuleBasic(genutilTypes.DefaultMessageValidator),
		})
	app.BasicManager.RegisterLegacyAminoCodec(legacyAmino)
	basicManagerInterfacesOnce.Do(func() {
		app.BasicManager.RegisterInterfaces(sharedInterfaceRegistry)
	})

	app.ModuleManager.SetOrderInitGenesis(
		authTypes.ModuleName,
		authz.ModuleName,
		bankTypes.ModuleName,
		distrTypes.ModuleName,
		stakingTypes.ModuleName,
		slashingTypes.ModuleName,
		upgradetypes.ModuleName,
		govtypes.ModuleName,
		feegrant.ModuleName,
		mintTypes.ModuleName,
		genutilTypes.ModuleName,
		consensusparamTypes.ModuleName,
		epochstypes.ModuleName,

		"org",
		"bandwidth",
		"memory",
		"serve",
		"attestation",
		"emissions",
		"identity",
		"reputation",
	)

	app.ModuleManager.SetOrderBeginBlockers(
		epochstypes.ModuleName,
		mintTypes.ModuleName,
		distrTypes.ModuleName,
		slashingTypes.ModuleName,
		stakingTypes.ModuleName,
		authTypes.ModuleName,
		authz.ModuleName,
		consensusparamTypes.ModuleName,
	)

	app.ModuleManager.SetOrderEndBlockers(
		stakingTypes.ModuleName,
		bankTypes.ModuleName,
		govtypes.ModuleName,
		feegrant.ModuleName,
	)

	configurator := module.NewConfigurator(appCodec, bApp.MsgServiceRouter(), bApp.GRPCQueryRouter())
	app.UpgradeKeeper.SetUpgradeHandler(
		v2.UpgradeName,
		v2.CreateUpgradeHandler(app.ModuleManager, configurator),
	)
	if err := app.ModuleManager.RegisterServices(configurator); err != nil {
		panic(err)
	}

	anteHandler, err := ante.NewAnteHandler(ante.HandlerOptions{
		AccountKeeper:   app.AccountKeeper,
		BankKeeper:      app.BankKeeper,
		FeegrantKeeper:  app.FeegrantKeeper,
		SignModeHandler: txConfig.SignModeHandler(),
		SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
	})
	if err != nil {
		panic(err)
	}
	app.SetAnteHandler(anteHandler)

	app.SetInitChainer(app.InitChainer)
	app.ModuleManager.SetOrderPreBlockers(
		upgradetypes.ModuleName,
		authTypes.ModuleName,
	)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	app.MountKVStores(keys)

	// CO-005c: Wire UpgradeStoreLoader so baseapp's LoadLatestVersion is upgrade-aware.
	// When the pre-upgrade binary halts at the upgrade height, x/upgrade's PreBlocker
	// writes /root/.wevibed/data/upgrade-info.json. On startup of the post-upgrade binary,
	// we must read that file and tell baseapp to use the upgrade store loader before
	// LoadLatestVersion runs. Without this, baseapp uses DefaultStoreLoader, which fails
	// to load state across the upgrade boundary with "version does not exist."
	// See DECISIONS.md D-S29-UPGRADE-STORE-LOADER.
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk: %s", err))
	}

	if upgradeInfo.Name == v2.UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := storetypes.StoreUpgrades{
			// v2 is a no-op upgrade — no store renames, additions, or deletions.
			// Future upgrades that add/remove module stores populate these slices.
		}
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}

	if loadLatest {
		if err := bApp.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("error loading last version: %w", err))
		}
	}

	return app
}

// genesisInitMarkerKey is a sentinel key written to every mounted KV store
// during InitChainer to ensure no module's IAVL tree is empty after init.
//
// Background (CO-005d, D-S29-CHAIN-RESTART-FOUNDATION):
//
// On chain restart, cosmos store/rootmulti.LoadLatestVersion calls
// iavl.LoadStoreWithOpts → MutableTree.LoadVersion(commitID.Version) for
// every mounted store. For trees that received zero user writes across all
// versions, the on-disk root key resolves to ErrVersionDoesNotExist
// (cosmos/iavl@v1.3.4/nodedb.go:891), causing baseapp to panic with
// "failed to load store: version does not exist".
//
// WeVibe modules (bandwidth, emissions, memory, org, serve, attestation,
// reputation) implement only the AppModule marker interface — not
// appmodule.HasGenesis — so the SDK's ModuleManager.InitGenesis silently
// skips their genesis path (no interface match) and their stores receive
// zero writes during InitChain. Several SDK modules without populated
// default state (feegrant, upgrade, consensusparam) similarly land with
// empty trees. This marker writes one byte to every mounted store under a
// 4-byte 0xFF sentinel key — well outside any module's collections.Prefix()
// namespace (collections prefixes are typically single bytes 0x00..0xFE).
var genesisInitMarkerKey = []byte{0xFF, 0xFF, 0xFF, 0xFF}
var genesisInitMarkerValue = []byte{0x01}

func (app *WeVibeApp) InitChainer(ctx sdk.Context, req *types.RequestInitChain) (*types.ResponseInitChain, error) {
	var genesisState map[string]json.RawMessage
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		return nil, err
	}

	resp, err := app.ModuleManager.InitGenesis(ctx, app.cdc, genesisState)
	if err != nil {
		return nil, err
	}

	// CO-005e: Persist the module version map so future ApplyUpgrade reads a
	// non-empty fromVM and RunMigrations runs per-module migrations rather
	// than re-running InitGenesis on already-initialized state.
	//
	// For app-wired (depinject) chains, x/upgrade's PopulateVersionMap is
	// invoked at construction time and the upgrade module's AppModule.InitGenesis
	// (cosmossdk.io/x/upgrade@v0.2.0/module.go:127-148) persists the map as a
	// side effect of ModuleManager.InitGenesis above. wevibe-chain is manually
	// wired, so GetInitVersionMap() returns nil and that side effect does not
	// fire. The official guidance for this case is the comment at
	// upgrade@v0.2.0/module.go:130-131:
	//   "chains can still use a custom init chainer for setting the version map"
	//
	// See DECISIONS.md D-S29-INITCHAINER-VERSION-MAP.
	if err := app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap()); err != nil {
		return nil, err
	}

	// CO-005d: Write a genesis init marker to every mounted KV store so no
	// module's IAVL tree is empty after InitChain. See D-S29-CHAIN-RESTART-FOUNDATION.
	for _, key := range app.keys {
		ctx.KVStore(key).Set(genesisInitMarkerKey, genesisInitMarkerValue)
	}

	return resp, nil
}

func (app *WeVibeApp) PreBlocker(ctx sdk.Context, req *types.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.ModuleManager.PreBlock(ctx)
}

func (app *WeVibeApp) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.ModuleManager.BeginBlock(ctx)
}

func (app *WeVibeApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.ModuleManager.EndBlock(ctx)
}

func (app *WeVibeApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	tx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	app.BasicManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	if err := server.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

func (app *WeVibeApp) RegisterTxService(clientCtx client.Context) {
	tx.RegisterTxService(app.GRPCQueryRouter(), clientCtx, app.Simulate, app.interfaceRegistry)
}

func (app *WeVibeApp) RegisterTendermintService(clientCtx client.Context) {
	cmtApp := server.NewCometABCIWrapper(app)
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.GRPCQueryRouter(),
		app.interfaceRegistry,
		cmtApp.Query,
	)
}

func (app *WeVibeApp) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

func (app *WeVibeApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

func (app *WeVibeApp) AppCodec() codec.Codec {
	return app.cdc
}

func (app *WeVibeApp) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.interfaceRegistry
}

func (app *WeVibeApp) TxConfig() client.TxConfig {
	return app.txConfig
}

func (app *WeVibeApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

func (app *WeVibeApp) SimulationManager() *module.SimulationManager {
	return nil
}

func (app *WeVibeApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

func (app *WeVibeApp) GetStoreKeys() []storetypes.StoreKey {
	keys := make([]storetypes.StoreKey, 0, len(app.keys))
	for _, key := range app.keys {
		keys = append(keys, key)
	}
	return keys
}

func BlockedAddresses() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authTypes.NewModuleAddress(acc).String()] = true
	}
	return modAccAddrs
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	authTypes.RegisterInterfaces(registry)
	bankTypes.RegisterInterfaces(registry)
	stakingTypes.RegisterInterfaces(registry)
	distrTypes.RegisterInterfaces(registry)
	slashingTypes.RegisterInterfaces(registry)
	mintTypes.RegisterInterfaces(registry)
	consensusparamTypes.RegisterInterfaces(registry)
	feegrant.RegisterInterfaces(registry)
	authz.RegisterInterfaces(registry)
	upgradetypes.RegisterInterfaces(registry)

	emissionsTypes.RegisterInterfaces(registry)
	bandwidthTypes.RegisterInterfaces(registry)
	memoryTypes.RegisterInterfaces(registry)
	orgTypes.RegisterInterfaces(registry)
	identityTypes.RegisterInterfaces(registry)
	reputationTypes.RegisterInterfaces(registry)
	serveTypes.RegisterInterfaces(registry)
	attestationTypes.RegisterInterfaces(registry)
}
