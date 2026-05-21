package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

func (m *Module) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: "wevibe.bandwidth.v1.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "GetBandwidthState",
					Use:       "get-state [org-id] [epoch]",
					Short:     "Query bandwidth state for an org at an epoch",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "epoch"},
					},
				},
				{
					RpcMethod: "GetBandwidthOverride",
					Use:       "get-override [org-id]",
					Short:     "Query bandwidth override for an org",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
					},
				},
				{
					RpcMethod: "GetRemainingBandwidth",
					Use:       "get-remaining [org-id] [epoch]",
					Short:     "Query remaining bandwidth for an org at an epoch",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "epoch"},
					},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query bandwidth module params",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: "wevibe.bandwidth.v1.Msg",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "SetBandwidthOverride",
					Use:       "set-override [org-id] [memory-cap] [serve-cap]",
					Short:     "Set custom bandwidth caps for an org",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "memory_cap"},
						{ProtoField: "serve_cap"},
					},
				},
				{
					RpcMethod: "UpdateParams",
					Use:       "update-params",
					Short:     "Update bandwidth module params (governance only)",
				},
			},
		},
	}
}