package types

import "errors"

var (
	ErrInvalidOrgID         = errors.New("invalid org ID")
	ErrInvalidLeader        = errors.New("invalid leader pubkey")
	ErrInvalidDomain        = errors.New("invalid domain")
	ErrOrgAlreadyExists     = errors.New("org already exists")
	ErrLeaderAlreadyOwnsOrg = errors.New("leader already owns an org")
	ErrOrgNotFound          = errors.New("org not found")
	ErrMemberNotFound       = errors.New("member not found")
	ErrMemberExists         = errors.New("member already exists")
	ErrNotLeader            = errors.New("caller is not the org leader")
	ErrInvalidRole          = errors.New("invalid role")
	ErrInvalidQuota         = errors.New("invalid quota value")
	ErrOrgNotActive         = errors.New("org is not active")
	ErrInsufficientFund     = errors.New("insufficient funds for registration")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrInsufficientTreasury = errors.New("insufficient treasury balance")
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrInvalidRepTier       = errors.New("invalid rep tier configuration")
	ErrRepTierOverlap       = errors.New("rep tier ranges overlap")
	ErrTreasuryNotFound     = errors.New("treasury not found")
	ErrInvalidRecipient     = errors.New("invalid recipient address")
	ErrFeegrantUnavailable  = errors.New("feegrant module is not configured")
)
