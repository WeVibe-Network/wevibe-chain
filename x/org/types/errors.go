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
	ErrInvalidHubEndpoints  = errors.New("invalid hub endpoints")
	ErrInvalidQuota         = errors.New("invalid quota value")
	ErrOrgNotActive         = errors.New("org is not active")
	ErrInsufficientFund     = errors.New("insufficient funds for registration")
	ErrSlotCapReached       = errors.New("org slot cap reached")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrFeegrantUnavailable  = errors.New("feegrant module is not configured")
)
