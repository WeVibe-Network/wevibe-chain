package types

import "errors"

var (
	ErrInvalidEmissionPool    = errors.New("invalid emission pool")
	ErrInvalidOperatorID      = errors.New("invalid operator ID")
	ErrInvalidOrgID           = errors.New("invalid org ID")
	ErrInvalidReward          = errors.New("invalid reward amount")
	ErrNoEmissionPool         = errors.New("no emission pool found")
	ErrOperatorNotFound       = errors.New("operator not found")
	ErrValidatorNotFound      = errors.New("validator not found")
	ErrInsufficientBalance    = errors.New("insufficient balance")
	ErrBootstrapExpired       = errors.New("bootstrap period expired")
	ErrInvalidBootstrapCredit = errors.New("invalid bootstrap credit")
	ErrNoBootstrapPool        = errors.New("no bootstrap pool found")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrNotMigrated            = errors.New("identity not migrated")
	ErrUnauthorizedClaim      = errors.New("unauthorized claim")
	ErrNothingToClaim         = errors.New("nothing to claim")
	ErrInvalidWalletAddress   = errors.New("invalid wallet address")
)
