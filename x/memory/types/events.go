package types

const (
	EventTypeCommitmentSubmitted = "commitment_submitted"
	EventTypeMemoryApproved      = "memory_approved"
	EventTypeMemoryRejected      = "memory_rejected"
	EventTypeExpiredPurged       = "expired_purged"

	AttributeKeyOrgID        = "org_id"
	AttributeKeyContentHash  = "content_hash"
	AttributeKeyContributor  = "contributor_id"
	AttributeKeyEpoch        = "epoch"
	AttributeKeyPurgedCount  = "purged_count"
	AttributeKeyBlockHeight  = "block_height"
)