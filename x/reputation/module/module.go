package reputation

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/wevibe-network/wevibe-chain/x/reputation/keeper"
	"github.com/wevibe-network/wevibe-chain/x/reputation/types"
)

type Module struct {
	keeper *keeper.Keeper
}

var (
	_ depinject.OnePerModuleType = (*Module)(nil)
	_ appmodule.AppModule        = (*Module)(nil)
	_ module.HasServices         = (*Module)(nil)
	// HasGenesis (and the embedded HasGenesisBasics) make the SDK ModuleManager
	// run DefaultGenesis/InitGenesis/ExportGenesis for this module. Without it
	// the module's genesis path is silently skipped and the module is never
	// activated at genesis (GAP-REP-1).
	_ module.HasGenesis = (*Module)(nil)
)

func NewModule(k *keeper.Keeper) *Module {
	return &Module{keeper: k}
}

func (m *Module) IsOnePerModuleType() {}
func (m *Module) IsAppModule()        {}

// DefaultGenesis returns the default genesis state as raw JSON. The module is
// active by default, matching DefaultParams().Active == true.
func (m *Module) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	bz, err := types.DefaultGenesis().MarshalJSON()
	if err != nil {
		panic(fmt.Errorf("reputation: marshal default genesis: %w", err))
	}
	return bz
}

// ValidateGenesis validates the reputation genesis state.
func (m *Module) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	if len(bz) == 0 {
		return nil
	}
	var state types.GenesisState
	if err := state.UnmarshalJSON(bz); err != nil {
		return fmt.Errorf("reputation: unmarshal genesis: %w", err)
	}
	return nil
}

// InitGenesis initializes module state from genesis. It persists the active
// flag and the default params so the module is operational the moment the
// chain starts.
func (m *Module) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) {
	state := types.DefaultGenesis()
	if len(bz) > 0 {
		var supplied types.GenesisState
		if err := supplied.UnmarshalJSON(bz); err != nil {
			panic(fmt.Errorf("reputation: unmarshal genesis: %w", err))
		}
		state = &supplied
	}

	if err := m.keeper.SetParams(ctx, types.DefaultParams()); err != nil {
		panic(fmt.Errorf("reputation: set params: %w", err))
	}
	if err := m.keeper.InitGenesisState(ctx, state); err != nil {
		panic(fmt.Errorf("reputation: init genesis: %w", err))
	}
}

// ExportGenesis exports module state as raw JSON.
func (m *Module) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	state, err := m.keeper.ExportGenesisState(ctx)
	if err != nil {
		panic(fmt.Errorf("reputation: export genesis: %w", err))
	}
	bz, err := state.MarshalJSON()
	if err != nil {
		panic(fmt.Errorf("reputation: marshal genesis: %w", err))
	}
	return bz
}

func (m *Module) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(m.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(m.keeper))
}
