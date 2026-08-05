package module

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/wevibe-network/wevibe-chain/x/serve/keeper"
	"github.com/wevibe-network/wevibe-chain/x/serve/types"
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
	// the manager silently skips the module's genesis path — which is exactly
	// what happened to the previous core-shaped (but non-matching) methods.
	_ module.HasGenesis = (*Module)(nil)
	_ interface {
		RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux)
	} = (*Module)(nil)
)

func NewModule(k *keeper.Keeper) *Module {
	return &Module{keeper: k}
}

func (m *Module) IsOnePerModuleType() {}
func (m *Module) IsAppModule()        {}

// DefaultGenesis returns the default (empty) genesis state as raw JSON.
func (m *Module) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	bz, err := json.Marshal(&types.GenesisState{})
	if err != nil {
		panic(fmt.Errorf("serve: marshal default genesis: %w", err))
	}
	return bz
}

// ValidateGenesis checks that the serve genesis state at least decodes.
func (m *Module) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	if len(bz) == 0 {
		return nil
	}
	var state types.GenesisState
	if err := json.Unmarshal(bz, &state); err != nil {
		return fmt.Errorf("serve: unmarshal genesis: %w", err)
	}
	return nil
}

// InitGenesis initializes module state from genesis, persisting any seeded
// policy anchors (and the latest-anchor pointer) plus receipts/events/stats.
func (m *Module) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) {
	if len(bz) == 0 {
		return
	}
	var state types.GenesisState
	if err := json.Unmarshal(bz, &state); err != nil {
		panic(fmt.Errorf("serve: unmarshal genesis: %w", err))
	}
	if err := m.keeper.InitGenesis(ctx, &state); err != nil {
		panic(fmt.Errorf("serve: init genesis: %w", err))
	}
}

// ExportGenesis exports module state as raw JSON.
func (m *Module) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	state, err := m.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(fmt.Errorf("serve: export genesis: %w", err))
	}
	bz, err := json.Marshal(state)
	if err != nil {
		panic(fmt.Errorf("serve: marshal genesis: %w", err))
	}
	return bz
}

func (m *Module) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(m.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(m.keeper))
}

// RegisterGRPCGatewayRoutes registers the module's REST query routes on the
// gRPC-gateway mux so the chain serves them over the REST API (port 1317).
func (m *Module) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}
