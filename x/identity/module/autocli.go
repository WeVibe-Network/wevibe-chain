package identity

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

func (m *Module) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: "wevibe.identity.v1.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "ResolveIdentity",
					Use:       "resolve-identity [passkey-pubkey]",
					Short:     "Resolve a passkey public key to a wallet address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "passkey_pubkey"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: "wevibe.identity.v1.Msg",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "MigrateIdentity",
					Use:       "migrate-identity [passkey-pubkey] [passkey-signature] [nonce]",
					Short:     "Migrate an off-chain passkey identity to an on-chain wallet",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "passkey_pubkey"},
						{ProtoField: "passkey_signature"},
						{ProtoField: "nonce"},
					},
				},
			},
		},
	}
}
