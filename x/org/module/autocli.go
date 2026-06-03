package org

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

func (m *Module) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: "wevibe.org.v1.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "GetOrg",
					Use:       "get-org [org-id]",
					Short:     "Query org details",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
					},
				},
				{
					RpcMethod: "GetMembers",
					Use:       "get-members [org-id]",
					Short:     "List org members",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
					},
				},
				{
					RpcMethod: "IsMember",
					Use:       "is-member [org-id] [pubkey]",
					Short:     "Check if pubkey is an org member",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "pubkey"},
					},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query org module params",
				},
				{
					RpcMethod: "GetOrgConfig",
					Use:       "get-org-config [org-id]",
					Short:     "Query org configuration",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: "wevibe.org.v1.Msg",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "RegisterOrg",
					Use:       "register-org [org-id] [leader-address]",
					Short:     "Register a new org",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "leader"},
					},
				},
				{
					RpcMethod: "AddMember",
					Use:       "add-member [org-id] [pubkey] [role]",
					Short:     "Add a member to an org",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "pubkey"},
						{ProtoField: "role"},
					},
				},
				{
					RpcMethod: "RemoveMember",
					Use:       "remove-member [org-id] [pubkey]",
					Short:     "Remove a member from an org",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "pubkey"},
					},
				},
				{
					RpcMethod: "UpdateParams",
					Use:       "update-params",
					Short:     "Update org module params (governance only)",
				},
				{
					RpcMethod: "SetOrgConfig",
					Use:       "set-org-config [org-id]",
					Short:     "Set org configuration",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
					},
				},
				{
					RpcMethod: "GrantTrialAllowance",
					Use:       "grant-trial-allowance [org-id] [grantee] [daily-submissions] [trial-days]",
					Short:     "Grant a temporary fee allowance to an org member",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "grantee"},
						{ProtoField: "daily_submissions"},
						{ProtoField: "trial_days"},
					},
				},
			},
		},
	}
}
