package identity

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/wevibe-network/wevibe-chain/x/identity/keeper"
	"github.com/wevibe-network/wevibe-chain/x/identity/types"
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

func (m *Module) Name() string {
	return types.ModuleName
}

func (m *Module) ConsensusVersion() uint64 {
	return 1
}

func (m *Module) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

func (m *Module) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	bz, err := types.DefaultGenesis().MarshalJSON()
	if err != nil {
		panic(fmt.Errorf("identity: marshal default genesis: %w", err))
	}
	return bz
}

func (m *Module) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	if len(bz) == 0 {
		return nil
	}

	var state types.GenesisState
	if err := state.UnmarshalJSON(bz); err != nil {
		return fmt.Errorf("identity: unmarshal genesis: %w", err)
	}

	return state.Validate()
}

func (m *Module) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) {
	state := types.DefaultGenesis()
	if len(bz) > 0 {
		var supplied types.GenesisState
		if err := supplied.UnmarshalJSON(bz); err != nil {
			panic(fmt.Errorf("identity: unmarshal genesis: %w", err))
		}
		state = &supplied
	}

	if err := state.Validate(); err != nil {
		panic(fmt.Errorf("identity: validate genesis: %w", err))
	}

	m.keeper.InitGenesisState(ctx, state)
}

func (m *Module) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	state := m.keeper.ExportGenesisState(ctx)
	bz, err := state.MarshalJSON()
	if err != nil {
		panic(fmt.Errorf("identity: marshal genesis: %w", err))
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
