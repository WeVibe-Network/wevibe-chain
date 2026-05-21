package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

func (m *Module) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: "wevibe.serve.v1.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "GetEpochServeStats",
					Use:       "get-stats [org-id] [epoch]",
					Short:     "Get epoch serve statistics for an org",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "epoch"},
					},
				},
				{
					RpcMethod: "GetContributorServes",
					Use:       "get-contributor [contributor-id] [epoch]",
					Short:     "Get serves by a contributor in a specific epoch",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "contributor_id"},
						{ProtoField: "epoch"},
					},
				},
				{
					RpcMethod: "GetMemoryServeCount",
					Use:       "get-memory-count [org-id] [content-hash] [epoch]",
					Short:     "Get serve count for a specific memory in an epoch",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "content_hash"},
						{ProtoField: "epoch"},
					},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query serve module params",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: "wevibe.serve.v1.Msg",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "SubmitServeBatch",
					Use:       "submit-batch [org-id] [epoch]",
					Short:     "Submit a batch of serve attestations",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "epoch"},
					},
				},
				{
					RpcMethod: "SubmitDenialBatch",
					Use:       "submit-denials [org-id] [epoch]",
					Short:     "Submit a batch of denial attestations",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "epoch"},
					},
				},
				{
					RpcMethod: "UpdateParams",
					Use:       "update-params",
					Short:     "Update serve module params (governance only)",
				},
			},
		},
	}
}
