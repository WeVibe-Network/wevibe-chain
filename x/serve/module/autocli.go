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
					RpcMethod: "ListEvents",
					Use:       "list-events [org-id] [epoch]",
					Short:     "List immutable recall-pivot events for an org epoch",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "epoch"},
					},
				},
				{
					RpcMethod: "GetPolicyAnchor",
					Use:       "get-policy-anchor [policy-version]",
					Short:     "Get an anchored edge-policy version hash",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "policy_version"},
					},
				},
				{
					RpcMethod: "GetLatestPolicyAnchor",
					Use:       "get-latest-policy-anchor",
					Short:     "Get the latest anchored edge-policy version hash",
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
					Short:     "Submit a batch of serve receipts",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "epoch"},
					},
				},
				{
					RpcMethod: "SubmitDenialBatch",
					Use:       "submit-denials [org-id] [epoch]",
					Short:     "Submit a batch of denial receipts",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "epoch"},
					},
				},
				{
					RpcMethod: "SubmitEventBatch",
					Use:       "submit-events [org-id] [epoch]",
					Short:     "Submit a batch of consumer-signed recall-pivot events",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "epoch"},
					},
				},
				{
					RpcMethod: "AnchorPolicyVersion",
					Use:       "anchor-policy-version [policy-version] [policy-hash]",
					Short:     "Anchor a published edge-policy version hash (governance only)",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "policy_version"},
						{ProtoField: "policy_hash"},
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
