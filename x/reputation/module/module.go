package reputation

import (
	"context"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"

	"github.com/wevibe-network/wevibe-chain/x/reputation/keeper"
	"github.com/wevibe-network/wevibe-chain/x/reputation/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

type Module struct {
	keeper *keeper.Keeper
}

var (
	_ depinject.OnePerModuleType = (*Module)(nil)
	_ appmodule.AppModule       = (*Module)(nil)
	_ module.HasServices       = (*Module)(nil)
)

func NewModule(k *keeper.Keeper) *Module {
	return &Module{keeper: k}
}

func (m *Module) IsOnePerModuleType() {}
func (m *Module) IsAppModule()       {}

func (m *Module) DefaultGenesis() []byte {
	state := &types.GenesisState{Active: false}
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
	return m.keeper.InitGenesisState(ctx, &state)
}

func (m *Module) ExportGenesis(ctx context.Context) ([]byte, error) {
	state, err := m.keeper.ExportGenesisState(ctx)
	if err != nil {
		return nil, err
	}
	return state.MarshalJSON()
}

func (m *Module) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(m.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(m.keeper))
}