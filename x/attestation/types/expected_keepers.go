package types

import "context"

type OrgKeeper interface {
	HasOrg(ctx context.Context, orgID string) (bool, error)
}