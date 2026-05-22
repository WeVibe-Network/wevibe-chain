package app

import (
	"encoding/json"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func (app *WeVibeApp) ExportAppStateAndValidators(forZeroHeight bool, jailAllowedAddrs []string, modulesToExport []string) (servertypes.ExportedApp, error) {
	ctx := app.NewUncachedContext(false, cmtproto.Header{Height: app.LastBlockHeight()})

	genState, err := app.ModuleManager.ExportGenesis(ctx, app.cdc)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	appState, err := json.MarshalIndent(genState, "", "  ")
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	validators, err := staking.WriteValidators(ctx, app.StakingKeeper)
	if err != nil {
		return servertypes.ExportedApp{}, err
	}

	return servertypes.ExportedApp{
		AppState:        appState,
		Validators:      validators,
		Height:          app.LastBlockHeight(),
		ConsensusParams: app.GetConsensusParams(ctx),
	}, nil
}
