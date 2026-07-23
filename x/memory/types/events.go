package types

const (
	EventTypeCommitmentSubmitted = "commitment_submitted"
	EventTypeMemoryRejected      = "memory_rejected"
	EventTypeExpiredPurged       = "expired_purged"

	AttributeKeyOrgID                  = "org_id"
	AttributeKeyContentHash            = "content_hash"
	AttributeKeyContributor            = "contributor_id"
	AttributeKeyProducerModelId        = "producer_model_id"
	AttributeKeyAttestationSessionHash = "attestation_session_hash"
	AttributeKeyEpoch                  = "epoch"
	AttributeKeyPurgedCount            = "purged_count"
	AttributeKeyBlockHeight            = "block_height"
)
