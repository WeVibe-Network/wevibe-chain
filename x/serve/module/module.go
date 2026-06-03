package module

import (
	"context"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/client"
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
	_ interface {
		RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux)
	} = (*Module)(nil)
)

func NewModule(k *keeper.Keeper) *Module {
	return &Module{keeper: k}
}

func (m *Module) IsOnePerModuleType() {}
func (m *Module) IsAppModule()        {}

func (m *Module) DefaultGenesis() []byte {
	state := &types.GenesisState{}
	bz, _ := state.MarshalJSON()
	return bz
}

func (m *Module) InitGenesis(ctx context.Context, bz []byte) error {
	if len(bz) == 0 {
		return nil
	}
	var state types.GenesisState
	if err := state.UnmarshalJSON(bz); err != nil {
		return err
	}
	return m.keeper.InitGenesis(ctx, &state)
}

func (m *Module) ExportGenesis(ctx context.Context) ([]byte, error) {
	state, err := m.keeper.ExportGenesis(ctx)
	if err != nil {
		return nil, err
	}
	return state.MarshalJSON()
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
