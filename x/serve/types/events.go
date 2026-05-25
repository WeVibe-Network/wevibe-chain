package types

const (
	EventTypeServeBatchSubmitted  = "serve_batch_submitted"
	EventTypeServeRecorded        = "serve_recorded"
	EventTypeDenialBatchSubmitted = "denial_batch_submitted"

	AttributeKeyOrgID           = "org_id"
	AttributeKeyEpoch          = "epoch"
	AttributeKeyAccepted        = "accepted"
	AttributeKeyRejected        = "rejected"
	AttributeKeyContributor     = "contributor_id"
	AttributeKeySelfServe       = "is_self_serve"
	AttributeKeySubmitter       = "submitter"
	AttributeKeyAcceptedCount    = "accepted_count"
	AttributeKeyRejectedCount   = "rejected_count"
	AttributeKeyBlockHeight     = "block_height"
)