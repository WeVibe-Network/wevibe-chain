package emissions

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

func (m *Module) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: "wevibe.emissions.v1.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "GetEmissionPool",
					Use:       "pool",
					Short:     "Query emission pool state",
				},
				{
					RpcMethod: "GetWorkScore",
					Use:       "work-score [operator-id] [org-id] [epoch]",
					Short:     "Query operator work score",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "operator_id"},
						{ProtoField: "org_id"},
						{ProtoField: "epoch"},
					},
				},
				{
					RpcMethod: "GetOperatorReward",
					Use:       "operator-reward [operator-id]",
					Short:     "Query pending operator reward",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "operator_id"},
					},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query emissions module params",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: "wevibe.emissions.v1.Msg",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "MintDailyEmission",
					Use:       "mint-emission [epoch]",
					Short:     "Manually trigger daily emission (governance only)",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "epoch"},
					},
				},
				{
					RpcMethod: "UpdateParams",
					Use:       "update-params",
					Short:     "Update emissions module params (governance only)",
				},
			},
		},
	}
}