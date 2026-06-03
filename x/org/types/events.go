package types

const (
	EventTypeOrgRegistered     = "org_registered"
	EventTypeOrgRenewed       = "org_renewed"
	EventTypeOrgDormant       = "org_dormant"
	EventTypeMemberAdded      = "member_added"
	EventTypeMemberRemoved    = "member_removed"
	EventTypeOrgConfigSet     = "org_config_set"
	EventTypeOrgBurnPaid      = "org_burn_paid"
	AttributeKeyOrgID          = "org_id"
	AttributeKeyLeader         = "leader"
	AttributeKeyMember          = "member"
	AttributeKeyRole           = "role"
	AttributeKeyMemberPubkey    = "member_pubkey"
	AttributeKeyEpochIds       = "epoch_ids"
	AttributeKeyRemovedBy       = "removed_by"
	AttributeKeyBlockHeight     = "block_height"
	AttributeKeyStorageQuota   = "storage_quota"
	AttributeKeyBudget         = "retrieval_budget"
	AttributeKeyAmount         = "amount"
	AttributeKeyRecipient      = "recipient"
	AttributeKeyBurnPrice      = "burn_price"
)
