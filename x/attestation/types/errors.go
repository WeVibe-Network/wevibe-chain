package types

import "errors"

var (
	ErrInvalidOrgID            = errors.New("invalid org ID")
	ErrInvalidSessionHash      = errors.New("session hash must be 32 bytes")
	ErrInvalidContributor      = errors.New("invalid contributor ID")
	ErrInvalidModelID          = errors.New("invalid model ID")
	ErrInvalidProviderType     = errors.New("invalid provider type")
	ErrDuplicateAttestation    = errors.New("session attestation already exists")
	ErrAttestationNotFound     = errors.New("session attestation not found")
	ErrOrgNotFound             = errors.New("org not found")
	ErrMaxAttestationsExceeded = errors.New("max attestations per epoch exceeded")
	ErrUnauthorized            = errors.New("unauthorized")
)