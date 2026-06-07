package emissions

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/wevibe-network/wevibe-chain/x/emissions/keeper"
	"github.com/wevibe-network/wevibe-chain/x/emissions/types"
)

type Module struct {
	keeper *keeper.Keeper
}

var (
	_ depinject.OnePerModuleType = (*Module)(nil)
	_ appmodule.AppModule        = (*Module)(nil)
	_ module.HasServices         = (*Module)(nil)
	_ interface {
		RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux)
	} = (*Module)(nil)
	// HasGenesis (and the embedded HasGenesisBasics) make the SDK ModuleManager
	// run DefaultGenesis/InitGenesis/ExportGenesis for this module. Without it
	// the module's genesis path is silently skipped and no emission pool is
	// seeded — the root cause of the "no emission pool found" epoch-hook log.
	_ module.HasGenesis = (*Module)(nil)
)

func NewModule(k *keeper.Keeper) *Module {
	return &Module{keeper: k}
}

func (m *Module) IsOnePerModuleType() {}
func (m *Module) IsAppModule()        {}

// DefaultGenesis returns the default genesis state as raw JSON. The default
// state includes an initialized emission pool derived from DefaultParams().
func (m *Module) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	bz, err := json.Marshal(types.DefaultGenesis())
	if err != nil {
		panic(fmt.Errorf("emissions: marshal default genesis: %w", err))
	}
	return bz
}

// ValidateGenesis validates the emissions genesis state.
func (m *Module) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	if len(bz) == 0 {
		return nil
	}
	var state types.GenesisState
	if err := json.Unmarshal(bz, &state); err != nil {
		return fmt.Errorf("emissions: unmarshal genesis: %w", err)
	}
	return state.Validate()
}

// InitGenesis initializes module state from genesis. It always persists the
// emission pool (falling back to the DefaultParams-derived pool when genesis
// supplies none) and the default params, so the chain always starts with a
// usable emission schedule.
func (m *Module) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) {
	state := types.DefaultGenesis()
	if len(bz) > 0 {
		var supplied types.GenesisState
		if err := json.Unmarshal(bz, &supplied); err != nil {
			panic(fmt.Errorf("emissions: unmarshal genesis: %w", err))
		}
		state = &supplied
	}
	if state.EmissionPool == nil {
		state.EmissionPool = types.DefaultEmissionPool()
	}

	if err := m.keeper.SetParams(ctx, types.DefaultParams()); err != nil {
		panic(fmt.Errorf("emissions: set params: %w", err))
	}
	if err := m.keeper.InitGenesis(ctx, state); err != nil {
		panic(fmt.Errorf("emissions: init genesis: %w", err))
	}
}

// ExportGenesis exports module state as raw JSON.
func (m *Module) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	state, err := m.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(fmt.Errorf("emissions: export genesis: %w", err))
	}
	bz, err := json.Marshal(state)
	if err != nil {
		panic(fmt.Errorf("emissions: marshal genesis: %w", err))
	}
	return bz
}

func (m *Module) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(m.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(m.keeper))
}

func (m *Module) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}
