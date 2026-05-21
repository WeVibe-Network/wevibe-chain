package types

type ContributorProfile struct {
	ContributorId      string `json:"contributor_id"`
	OrgId              string `json:"org_id"`
	ContributionCount  uint64 `json:"contribution_count"`
	ServeCount         uint64 `json:"serve_count"`
	ReportUpheldCount  uint64 `json:"report_upheld_count"`
	FirstSeenEpoch     uint64 `json:"first_seen_epoch"`
	FirstSeenTimestamp int64  `json:"first_seen_timestamp"`
}