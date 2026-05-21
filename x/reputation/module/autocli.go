package reputation

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

func (m *Module) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: "wevibe.reputation.v1.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "GetReputation",
					Use:       "get-reputation [developer-hex]",
					Short:     "Query developer reputation",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "developer"},
					},
				},
				{
					RpcMethod: "GetXP",
					Use:       "get-xp [developer-hex]",
					Short:     "Query developer XP",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "developer"},
					},
				},
				{
					RpcMethod: "IsActive",
					Use:       "is-active",
					Short:     "Check if reputation module is active",
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query reputation module params",
				},
				{
					RpcMethod: "GetServeStats",
					Use:       "get-serve-stats [developer-hex]",
					Short:     "Query serve-based reputation stats",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "developer"},
					},
				},
				{
					RpcMethod: "GetContributorOrgSet",
					Use:       "get-contributor-org-set [developer-hex]",
					Short:     "Query org set for contributor",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "developer"},
					},
				},
				{
					RpcMethod: "GetCrossOrgProfile",
					Use:       "get-cross-org-profile [developer-hex]",
					Short:     "Query full cross-org profile",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "developer"},
					},
				},
				{
					RpcMethod: "ModeratorProfile",
					Use:       "moderator-profile [moderator-pubkey] [org-id]",
					Short:     "Query moderator profile by pubkey and org",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "moderator_pubkey"},
						{ProtoField: "org_id"},
					},
				},
				{
					RpcMethod: "LeaderProfile",
					Use:       "leader-profile [leader-pubkey] [org-id]",
					Short:     "Query leader profile by pubkey and org",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "leader_pubkey"},
						{ProtoField: "org_id"},
					},
				},
				{
					RpcMethod: "UpheldReportsByContributor",
					Use:       "upheld-reports-by-contributor [contributor-id]",
					Short:     "Query upheld reports filed against a contributor",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "contributor_id"},
					},
				},
				{
					RpcMethod: "UpheldReportsByModerator",
					Use:       "upheld-reports-by-moderator [moderator-pubkey]",
					Short:     "Query upheld reports where moderator approved something later deleted",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "moderator_pubkey"},
					},
				},
				{
					RpcMethod: "UpheldReportsByLeader",
					Use:       "upheld-reports-by-leader [leader-pubkey]",
					Short:     "Query upheld reports committed by a leader",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "leader_pubkey"},
					},
				},
				{
					RpcMethod: "VerifyUpheldReport",
					Use:       "verify-upheld-report [org-id] [content-hash-hex]",
					Short:     "Verify cryptographic triplet for an upheld report",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "content_hash"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: "wevibe.reputation.v1.Msg",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateReputation",
					Use:       "update-reputation [developer-hex] [memory-cid]",
					Short:     "Update developer reputation (requires attestation)",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "developer"},
						{ProtoField: "memory_cid"},
					},
				},
				{
					RpcMethod: "UpdateParams",
					Use:       "update-params",
					Short:     "Update reputation module params (governance only)",
				},
			},
		},
	}
}