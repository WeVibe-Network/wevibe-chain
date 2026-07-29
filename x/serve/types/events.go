package types

const (
	EventTypeServeBatchSubmitted   = "serve_batch_submitted"
	EventTypeServeRecorded         = "serve_recorded"
	EventTypeDenialBatchSubmitted  = "denial_batch_submitted"
	EventTypeEventBatchSubmitted   = "event_batch_submitted"
	EventTypePolicyVersionAnchored = "policy_version_anchored"

	AttributeKeyOrgID                  = "org_id"
	AttributeKeyEpoch                  = "epoch"
	AttributeKeyAccepted               = "accepted"
	AttributeKeyRejected               = "rejected"
	AttributeKeyContributor            = "contributor_id"
	AttributeKeySelfServe              = "is_self_serve"
	AttributeKeySubmitter              = "submitter"
	AttributeKeyAcceptedCount          = "accepted_count"
	AttributeKeyRejectedCount          = "rejected_count"
	AttributeKeyRejectedDuplicateCount = "rejected_duplicate_count"
	AttributeKeyRejectedInvalidCount   = "rejected_invalid_count"
	AttributeKeyBlockHeight            = "block_height"
	AttributeKeyPolicyVersion          = "policy_version"
	AttributeKeyPolicyHash             = "policy_hash"
)
