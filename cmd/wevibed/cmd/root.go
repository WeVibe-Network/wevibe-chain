package cmd

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"cosmossdk.io/log"
	upgradecli "cosmossdk.io/x/upgrade/client/cli"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/spf13/cobra"

	dbm "github.com/cosmos/cosmos-db"

	cmtcfg "github.com/cometbft/cometbft/config"

	"github.com/cosmos/cosmos-sdk/client"
	clientconfig "github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/pruning"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/server"
	servercmd "github.com/cosmos/cosmos-sdk/server/cmd"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrcli "github.com/cosmos/cosmos-sdk/x/distribution/client/cli"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/wevibe-network/wevibe-chain/app"
)

const epochDurationSecondsEnv = "WEVIBE_EPOCH_DURATION_SECONDS"

// NewRootCmd creates a new root command for the WeVibe daemon.
func NewRootCmd() *cobra.Command {
	encodingConfig := app.MakeEncodingConfig()

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("")

	rootCmd := &cobra.Command{
		Use:          "wevibed",
		Short:        "WeVibe Chain Daemon",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateEpochDurationSecondsEnv(); err != nil {
				return err
			}

			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			clientCtx := initClientCtx.WithCmdContext(cmd.Context())
			var err error
			clientCtx, err = client.ReadPersistentCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			clientCtx, err = clientconfig.ReadFromClientConfig(clientCtx)
			if err != nil {
				return err
			}

			if err := client.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()
			customCMTConfig := initCometBFTConfig()

			return server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, customCMTConfig)
		},
	}

	initRootCmd(rootCmd, encodingConfig)

	return rootCmd
}

func initRootCmd(rootCmd *cobra.Command, encodingConfig app.EncodingConfig) {
	server.AddCommands(rootCmd, app.DefaultNodeHome, newApp, appExport, func(*cobra.Command) {})

	rootCmd.AddCommand(
		genutilcli.InitCmd(app.ModuleBasics, app.DefaultNodeHome),
		snapshot.Cmd(newApp),
		pruning.Cmd(newApp, app.DefaultNodeHome),
		debug.Cmd(),
		server.StatusCommand(),
		keys.Commands(),
		genesisCommand(encodingConfig.TxConfig),
		queryCommand(),
		txCommand(),
	)
}

func genesisCommand(txConfig client.TxConfig) *cobra.Command {
	cmd := genutilcli.Commands(txConfig, app.ModuleBasics, app.DefaultNodeHome)
	return cmd
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.WaitTxCmd(),
		rpc.ValidatorCommand(),
		server.QueryBlockCmd(),
		server.QueryBlocksCmd(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
		server.QueryBlockResultsCmd(),
		queryAuthCommand(),
		queryBankCommand(),
		queryStakingCommand(),
		queryDistributionCommand(),
		querySlashingCommand(),
		queryGovCommand(),
		queryUpgradeCommand(),
	)

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		authcmd.GetSimulateCmd(),
	)

	addressCodec := authcodec.NewBech32Codec(app.Bech32Prefix)
	validatorAddressCodec := authcodec.NewBech32Codec(app.Bech32PrefixVal)

	cmd.AddCommand(
		bankcli.NewTxCmd(addressCodec),
		stakingcli.NewTxCmd(validatorAddressCodec, addressCodec),
		distrcli.NewTxCmd(validatorAddressCodec, addressCodec),
		govcli.NewTxCmd(nil),
		txSlashingCommand(),
		upgradecli.GetTxCmd(addressCodec),
	)

	return cmd
}

func txSlashingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "slashing",
		Short:                      "Slashing transaction subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	unjailCmd := &cobra.Command{
		Use:   "unjail",
		Short: "Unjail a jailed validator",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			validatorAddr, err := authcodec.NewBech32Codec(app.Bech32PrefixVal).BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			msg := slashingtypes.NewMsgUnjail(validatorAddr)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(unjailCmd)

	cmd.AddCommand(unjailCmd)
	return cmd
}

func queryAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "auth",
		Short:                      "Querying auth module",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	accountCmd := &cobra.Command{
		Use:   "account [address]",
		Short: "Query account by address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := authtypes.NewQueryClient(clientCtx).Account(cmd.Context(), &authtypes.QueryAccountRequest{Address: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(accountCmd)

	moduleAccountCmd := &cobra.Command{
		Use:   "module-account [name]",
		Short: "Query module account by module name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := authtypes.NewQueryClient(clientCtx).ModuleAccountByName(cmd.Context(), &authtypes.QueryModuleAccountByNameRequest{Name: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(moduleAccountCmd)

	cmd.AddCommand(accountCmd, moduleAccountCmd)
	return cmd
}

func queryBankCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "bank",
		Short:                      "Querying bank module",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	balanceCmd := &cobra.Command{
		Use:   "balance [address] [denom]",
		Short: "Query coin balance by account and denom",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := banktypes.NewQueryClient(clientCtx).Balance(cmd.Context(), &banktypes.QueryBalanceRequest{Address: args[0], Denom: args[1]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(balanceCmd)

	balancesCmd := &cobra.Command{
		Use:   "balances [address]",
		Short: "Query all balances by account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := banktypes.NewQueryClient(clientCtx).AllBalances(cmd.Context(), &banktypes.QueryAllBalancesRequest{Address: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(balancesCmd)

	cmd.AddCommand(balanceCmd, balancesCmd)
	return cmd
}

func queryStakingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "staking",
		Short:                      "Querying staking module",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	validatorCmd := &cobra.Command{
		Use:   "validator [validator-address]",
		Short: "Query validator by operator address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := stakingtypes.NewQueryClient(clientCtx).Validator(cmd.Context(), &stakingtypes.QueryValidatorRequest{ValidatorAddr: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(validatorCmd)

	validatorsCmd := &cobra.Command{
		Use:   "validators",
		Short: "Query all validators",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := stakingtypes.NewQueryClient(clientCtx).Validators(cmd.Context(), &stakingtypes.QueryValidatorsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(validatorsCmd)

	cmd.AddCommand(validatorCmd, validatorsCmd)
	return cmd
}

func queryDistributionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "distribution",
		Short:                      "Querying distribution module",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	paramsCmd := &cobra.Command{
		Use:   "params",
		Short: "Query distribution params",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := distrtypes.NewQueryClient(clientCtx).Params(cmd.Context(), &distrtypes.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(paramsCmd)

	outstandingRewardsCmd := &cobra.Command{
		Use:   "validator-outstanding-rewards [validator-address]",
		Short: "Query outstanding rewards for a validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := distrtypes.NewQueryClient(clientCtx).ValidatorOutstandingRewards(
				cmd.Context(),
				&distrtypes.QueryValidatorOutstandingRewardsRequest{ValidatorAddress: args[0]},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(outstandingRewardsCmd)

	commissionCmd := &cobra.Command{
		Use:   "commission [validator-address]",
		Short: "Query commission for a validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := distrtypes.NewQueryClient(clientCtx).ValidatorCommission(
				cmd.Context(),
				&distrtypes.QueryValidatorCommissionRequest{ValidatorAddress: args[0]},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(commissionCmd)

	communityPoolCmd := &cobra.Command{
		Use:   "community-pool",
		Short: "Query community pool balance",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := distrtypes.NewQueryClient(clientCtx).CommunityPool(cmd.Context(), &distrtypes.QueryCommunityPoolRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(communityPoolCmd)

	cmd.AddCommand(paramsCmd, outstandingRewardsCmd, commissionCmd, communityPoolCmd)
	return cmd
}

func querySlashingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "slashing",
		Short:                      "Querying slashing module",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	paramsCmd := &cobra.Command{
		Use:   "params",
		Short: "Query slashing params",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := slashingtypes.NewQueryClient(clientCtx).Params(cmd.Context(), &slashingtypes.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(paramsCmd)

	signingInfoCmd := &cobra.Command{
		Use:   "signing-info [cons-address]",
		Short: "Query validator signing info by consensus address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := slashingtypes.NewQueryClient(clientCtx).SigningInfo(
				cmd.Context(),
				&slashingtypes.QuerySigningInfoRequest{ConsAddress: args[0]},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(signingInfoCmd)

	signingInfosCmd := &cobra.Command{
		Use:   "signing-infos",
		Short: "Query signing information for all validators",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := slashingtypes.NewQueryClient(clientCtx).SigningInfos(
				cmd.Context(),
				&slashingtypes.QuerySigningInfosRequest{},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(signingInfosCmd)

	cmd.AddCommand(paramsCmd, signingInfoCmd, signingInfosCmd)
	return cmd
}

func queryGovCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "gov",
		Short:                      "Querying governance module",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	proposalCmd := &cobra.Command{
		Use:   "proposal [proposal-id]",
		Short: "Query governance proposal by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			proposalID, err := parseUint64Arg(args[0], "proposal id")
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := govv1.NewQueryClient(clientCtx).Proposal(cmd.Context(), &govv1.QueryProposalRequest{ProposalId: proposalID})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(proposalCmd)

	proposalsCmd := &cobra.Command{
		Use:   "proposals",
		Short: "Query governance proposals",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := govv1.NewQueryClient(clientCtx).Proposals(cmd.Context(), &govv1.QueryProposalsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(proposalsCmd)

	voteCmd := &cobra.Command{
		Use:   "vote [proposal-id] [voter]",
		Short: "Query a governance vote",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			proposalID, err := parseUint64Arg(args[0], "proposal id")
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := govv1.NewQueryClient(clientCtx).Vote(cmd.Context(), &govv1.QueryVoteRequest{ProposalId: proposalID, Voter: args[1]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(voteCmd)

	votesCmd := &cobra.Command{
		Use:   "votes [proposal-id]",
		Short: "Query governance votes for a proposal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			proposalID, err := parseUint64Arg(args[0], "proposal id")
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := govv1.NewQueryClient(clientCtx).Votes(cmd.Context(), &govv1.QueryVotesRequest{ProposalId: proposalID})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(votesCmd)

	cmd.AddCommand(proposalCmd, proposalsCmd, voteCmd, votesCmd)
	return cmd
}

func queryUpgradeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "upgrade",
		Short:                      "Querying upgrade module",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	planCmd := &cobra.Command{
		Use:   "plan",
		Short: "Query current upgrade plan",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := upgradetypes.NewQueryClient(clientCtx).CurrentPlan(cmd.Context(), &upgradetypes.QueryCurrentPlanRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(planCmd)

	appliedCmd := &cobra.Command{
		Use:   "applied [name]",
		Short: "Query applied upgrade plan by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			res, err := upgradetypes.NewQueryClient(clientCtx).AppliedPlan(cmd.Context(), &upgradetypes.QueryAppliedPlanRequest{Name: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(appliedCmd)

	cmd.AddCommand(planCmd, appliedCmd)
	return cmd
}

func parseUint64Arg(raw, field string) (uint64, error) {
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", field, raw, err)
	}

	return value, nil
}

func validateEpochDurationSecondsEnv() error {
	raw := os.Getenv(epochDurationSecondsEnv)
	if raw == "" {
		return nil
	}

	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds < 1 {
		return fmt.Errorf("invalid %s: %q", epochDurationSecondsEnv, raw)
	}

	return nil
}

func initAppConfig() (string, interface{}) {
	srvCfg := serverconfig.DefaultConfig()
	srvCfg.MinGasPrices = "0.01vibe"

	return serverconfig.DefaultConfigTemplate, srvCfg
}

func initCometBFTConfig() *cmtcfg.Config {
	return cmtcfg.DefaultConfig()
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer, appOpts servertypes.AppOptions) servertypes.Application {
	baseAppOptions := server.DefaultBaseappOptions(appOpts)
	wevibeApp := app.NewWeVibeApp(logger, db, true, appOpts, baseAppOptions...)
	if traceStore != nil {
		wevibeApp.SetCommitMultiStoreTracer(traceStore)
	}
	return wevibeApp
}

func appExport(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	baseAppOptions := server.DefaultBaseappOptions(appOpts)

	var wevibeApp *app.WeVibeApp
	if height != -1 {
		wevibeApp = app.NewWeVibeApp(logger, db, false, appOpts, baseAppOptions...)
		if traceStore != nil {
			wevibeApp.SetCommitMultiStoreTracer(traceStore)
		}
		if err := wevibeApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		wevibeApp = app.NewWeVibeApp(logger, db, true, appOpts, baseAppOptions...)
		if traceStore != nil {
			wevibeApp.SetCommitMultiStoreTracer(traceStore)
		}
	}

	return wevibeApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}

func Execute() error {
	rootCmd := NewRootCmd()
	return servercmd.Execute(rootCmd, "WEVIBED", app.DefaultNodeHome)
}
