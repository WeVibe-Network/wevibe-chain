package emissions

import (
	"context"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"

	"github.com/wevibe-network/wevibe-chain/x/emissions/keeper"
	"github.com/wevibe-network/wevibe-chain/x/emissions/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

type Module struct {
	keeper *keeper.Keeper
}

var (
	_ depinject.OnePerModuleType = (*Module)(nil)
	_ appmodule.AppModule       = (*Module)(nil)
	_ module.HasServices        = (*Module)(nil)
)

func NewModule(k *keeper.Keeper) *Module {
	return &Module{keeper: k}
}

func (m *Module) IsOnePerModuleType() {}
func (m *Module) IsAppModule()       {}

func (m *Module) DefaultGenesis() []byte {
	return []byte("{}")
}

func (m *Module) InitGenesis(ctx context.Context, bz []byte) error {
	return nil
}

func (m *Module) ExportGenesis(ctx context.Context) ([]byte, error) {
	_, err := m.keeper.ExportGenesis(ctx)
	if err != nil {
		return nil, err
	}
	return []byte("{}"), nil
}

func (m *Module) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(m.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServerImpl(m.keeper))
}