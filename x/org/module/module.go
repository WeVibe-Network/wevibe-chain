package org

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/wevibe-network/wevibe-chain/x/org/keeper"
	"github.com/wevibe-network/wevibe-chain/x/org/types"
)

type Module struct {
	keeper *keeper.Keeper
}

var (
	_ depinject.OnePerModuleType = (*Module)(nil)
	_ appmodule.AppModule        = (*Module)(nil)
	_ module.HasServices         = (*Module)(nil)
	_ module.HasGenesis          = (*Module)(nil)
	_ interface {
		RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux)
	} = (*Module)(nil)
)

func NewModule(k *keeper.Keeper) *Module {
	return &Module{keeper: k}
}

func (m *Module) IsOnePerModuleType() {}
func (m *Module) IsAppModule()        {}

func (m *Module) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	state := types.DefaultGenesis()
	bz, err := state.MarshalJSON()
	if err != nil {
		panic(fmt.Errorf("org: marshal default genesis: %w", err))
	}
	return bz
}

func (m *Module) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	if len(bz) == 0 {
		return nil
	}
	var state types.GenesisState
	if err := state.UnmarshalJSON(bz); err != nil {
		return fmt.Errorf("org: unmarshal genesis: %w", err)
	}
	return state.Validate()
}

func (m *Module) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) {
	state := types.DefaultGenesis()
	if len(bz) > 0 {
		if err := state.UnmarshalJSON(bz); err != nil {
			panic(fmt.Errorf("org: unmarshal genesis: %w", err))
		}
	}

	if err := state.Validate(); err != nil {
		panic(fmt.Errorf("org: validate genesis: %w", err))
	}

	if err := m.keeper.InitGenesis(ctx, state); err != nil {
		panic(fmt.Errorf("org: init genesis: %w", err))
	}
}

func (m *Module) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	state, err := m.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(fmt.Errorf("org: export genesis: %w", err))
	}
	bz, err := state.MarshalJSON()
	if err != nil {
		panic(fmt.Errorf("org: marshal genesis: %w", err))
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
