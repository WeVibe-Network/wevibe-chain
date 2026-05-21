package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

func (m *Module) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: "wevibe.memory.v1.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "GetMemory",
					Use:       "get-memory [org-id] [content-hash]",
					Short:     "Query approved memory by org and content hash",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "content_hash"},
					},
				},
				{
					RpcMethod: "GetPendingCommitments",
					Use:       "get-pending [org-id]",
					Short:     "List pending commitments for an org",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
					},
				},
				{
					RpcMethod: "GetMemoryCount",
					Use:       "get-count [org-id]",
					Short:     "Get approved memory count for an org",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
					},
				},
				{
					RpcMethod: "GetEpochMerkleRoot",
					Use:       "get-merkle-root [org-id] [epoch]",
					Short:     "Get Merkle root for an org at a specific epoch",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "epoch"},
					},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query memory module params",
				},
				{
					RpcMethod: "GetMemoriesBatch",
					Use:       "get-memories-batch [org-id] [content-hash...]",
					Short:     "Batch query approved memories by org and content hashes (up to 50)",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "content_hashes", Varargs: true},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: "wevibe.memory.v1.Msg",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "SubmitCommitment",
					Use:       "submit-commitment [org-id] [content-hash] [contributor-id]",
					Short:     "Submit a pending memory commitment",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "content_hash"},
						{ProtoField: "contributor_id"},
					},
				},
				{
					RpcMethod: "ApproveMemory",
					Use:       "approve-memory [org-id] [content-hash]",
					Short:     "Approve a pending memory and store the encrypted blob",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "content_hash"},
					},
				},
				{
					RpcMethod: "RejectMemory",
					Use:       "reject-memory [org-id] [content-hash]",
					Short:     "Reject a pending memory",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "content_hash"},
					},
				},
				{
					RpcMethod: "PurgeExpired",
					Use:       "purge-expired [org-id]",
					Short:     "Purge expired pending commitments",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
					},
				},
				{
					RpcMethod: "UpdateParams",
					Use:       "update-params",
					Short:     "Update memory module params (governance only)",
				},
			},
		},
	}
}