package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

func (m *Module) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: "wevibe.attestation.v1.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "GetSessionAttestation",
					Use:       "get-session [org-id] [session-hash]",
					Short:     "Get a session attestation by org and session hash",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "session_hash"},
					},
				},
				{
					RpcMethod: "ListSessionAttestations",
					Use:       "list-sessions [org-id] [epoch]",
					Short:     "List session attestations for an org in a specific epoch",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "epoch"},
					},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query attestation module params",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: "wevibe.attestation.v1.Msg",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "SubmitSessionAttestation",
					Use:       "submit-session [org-id] [session-hash]",
					Short:     "Submit a session attestation",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "org_id"},
						{ProtoField: "session_hash"},
					},
				},
				{
					RpcMethod: "UpdateParams",
					Use:       "update-params",
					Short:     "Update attestation module params (governance only)",
				},
			},
		},
	}
}