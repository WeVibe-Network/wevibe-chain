package types

import "errors"

var (
	ErrInvalidEmissionPool    = errors.New("invalid emission pool")
	ErrInvalidOperatorID      = errors.New("invalid operator ID")
	ErrInvalidOrgID           = errors.New("invalid org ID")
	ErrInvalidReward          = errors.New("invalid reward amount")
	ErrNoEmissionPool         = errors.New("no emission pool found")
	ErrNoWorkScore            = errors.New("no work score found")
	ErrOperatorNotFound       = errors.New("operator not found")
	ErrValidatorNotFound      = errors.New("validator not found")
	ErrInsufficientBalance    = errors.New("insufficient balance")
	ErrBootstrapExpired       = errors.New("bootstrap period expired")
	ErrInvalidBootstrapCredit = errors.New("invalid bootstrap credit")
	ErrNoBootstrapPool        = errors.New("no bootstrap pool found")
	ErrNoPendingReward        = errors.New("no pending reward for operator")
	ErrUnauthorized           = errors.New("unauthorized")
)
