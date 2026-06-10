package types

import "errors"

var (
	ErrInvalidOrgID               = errors.New("invalid org ID")
	ErrInvalidContentHash         = errors.New("content hash must be 32 bytes")
	ErrInvalidServeKey            = errors.New("invalid serve key")
	ErrInvalidServeKeyPubkey      = errors.New("serve key pubkey must be 32 bytes")
	ErrInvalidServeSignature      = errors.New("serve signature must be 64 bytes")
	ErrInvalidServeFingerprint    = errors.New("serve fingerprint must be 32 bytes")
	ErrInvalidContributor         = errors.New("invalid contributor ID")
	ErrDuplicateServeFingerprint  = errors.New("serve fingerprint already exists")
	ErrDuplicateDenialFingerprint = errors.New("denial fingerprint already exists")
	ErrBatchTooLarge              = errors.New("serve batch exceeds max size")
	ErrBatchEmpty                 = errors.New("serve batch is empty")
	ErrMemoryNotFound             = errors.New("approved memory not found")
	ErrOrgNotFound                = errors.New("org not found")
	ErrOrgTooYoung                = errors.New("org has not reached minimum age")
	ErrMaxServesExceeded          = errors.New("max serves per memory per epoch exceeded")
	ErrStatsNotFound              = errors.New("epoch serve stats not found")
	ErrContributorNotFound        = errors.New("contributor serves not found")
	ErrUnauthorized               = errors.New("unauthorized")
)
