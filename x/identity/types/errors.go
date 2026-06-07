package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrAliasAlreadyMigrated    = errorsmod.Register(ModuleName, 1, "alias already migrated")
	ErrInvalidPasskeySignature = errorsmod.Register(ModuleName, 2, "invalid passkey signature")
	ErrInvalidWalletAddress    = errorsmod.Register(ModuleName, 3, "invalid wallet address")
	ErrInvalidPasskeyPubkey    = errorsmod.Register(ModuleName, 4, "invalid passkey pubkey")
)
