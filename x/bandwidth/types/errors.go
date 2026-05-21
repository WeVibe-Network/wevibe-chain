package types

import "errors"

var (
	ErrInvalidOrgID             = errors.New("invalid org ID")
	ErrBandwidthExhausted       = errors.New("bandwidth exhausted for this epoch")
	ErrMemoryBandwidthExhausted = errors.New("memory bandwidth exhausted for this epoch")
	ErrServeBandwidthExhausted  = errors.New("serve bandwidth exhausted for this epoch")
	ErrInvalidCap               = errors.New("bandwidth cap must be greater than zero")
	ErrOverrideNotFound         = errors.New("no bandwidth override for org")
	ErrUnauthorized             = errors.New("unauthorized")
	ErrOrgNotFound             = errors.New("org not found")
)