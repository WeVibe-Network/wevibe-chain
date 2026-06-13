package types

import "cosmossdk.io/collections"

const ModuleName = "org"

var (
	OrgKeyPrefix    = collections.NewPrefixedPairRange[string, *Org]("")
	MemberKeyPrefix = collections.NewPrefixedPairRange[MemberKey, []byte](MemberKey{})
)

type OrgKey struct {
	OrgID string
}

type MemberKey struct {
	OrgID  string
	Member string
}

type Org struct {
	OrgID           string    `json:"org_id"`
	Slot            uint64    `json:"slot"`
	Leader          string    `json:"leader"`
	Domain          string    `json:"domain"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	TechStack       string    `json:"tech_stack"`
	FocusAreas      string    `json:"focus_areas"`
	CreatedAt       uint64    `json:"created_at"`
	RenewalHeight   uint64    `json:"renewal_height"`
	StorageQuota    uint64    `json:"storage_quota"`
	RetrievalBudget uint64    `json:"retrieval_budget"`
	Status          OrgStatus `json:"status"`
	// CO-044 (D-S32-CO044-KEY-SEPARATION): per-org hub serving key chain address
	// (the only signer allowed to submit serve/denial batches) and the leader's
	// chain wallet address (the on-chain authority for org decisions and serving
	// key rotation).
	HubServingAddress   string   `json:"hub_serving_address"`
	HubEndpoints        []string `json:"hub_endpoints"`
	HubResponsePubkey   string   `json:"hub_response_pubkey"`
	LeaderWalletAddress string   `json:"leader_wallet_address"`
	AccountAddress      string   `json:"account_address"`
}

func NewOrg(orgID, leader, domain, description, techStack, focusAreas string, storageQuota, retrievalBudget uint64) *Org {
	return &Org{
		OrgID:           orgID,
		Leader:          leader,
		Domain:          domain,
		Description:     description,
		TechStack:       techStack,
		FocusAreas:      focusAreas,
		CreatedAt:       0,
		RenewalHeight:   0,
		StorageQuota:    storageQuota,
		RetrievalBudget: retrievalBudget,
		Status:          OrgStatus_ACTIVE,
	}
}

func (o *Org) Validate() error {
	if o.OrgID == "" {
		return ErrInvalidOrgID
	}
	if o.Leader == "" {
		return ErrInvalidLeader
	}
	if len(o.Description) > 500 || containsASCIIControl(o.Description) {
		return ErrInvalidDescription
	}
	if len(o.TechStack) > 200 || containsASCIIControl(o.TechStack) {
		return ErrInvalidTechStack
	}
	if len(o.FocusAreas) > 200 || containsASCIIControl(o.FocusAreas) {
		return ErrInvalidFocusAreas
	}
	if len(o.Name) > 100 || containsASCIIControl(o.Name) {
		return ErrInvalidName
	}
	if len(o.Domain) > 128 {
		return ErrInvalidDomain
	}
	for _, c := range o.Domain {
		if (c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == ' ' ||
			c == ',' ||
			c == '.' ||
			c == '-' ||
			c == '_' ||
			c == '/' ||
			c == '+' ||
			c == '&' {
			continue
		}

		return ErrInvalidDomain
	}
	return nil
}

func containsASCIIControl(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 {
			return true
		}
	}

	return false
}

func (o *Org) IsActive() bool {
	return o.Status == OrgStatus_ACTIVE
}

func (m *MemberRecord) MemberKey() MemberKey {
	return MemberKey{
		OrgID:  m.OrgID,
		Member: m.Pubkey,
	}
}

type OrgStatus int32

const (
	OrgStatus_ACTIVE    OrgStatus = 0
	OrgStatus_DORMANT   OrgStatus = 1
	OrgStatus_SUSPENDED OrgStatus = 2
	OrgStatus_CLOSED    OrgStatus = 3
)

type MemberRecord struct {
	OrgID         string `json:"org_id"`
	Pubkey        string `json:"pubkey"`
	Role          string `json:"role"`
	X25519Pubkey  string `json:"x25519_pubkey"`
	CanContribute bool   `json:"can_contribute"`
	CanModerate   bool   `json:"can_moderate"`
}

func NewMemberRecord(orgID, pubkey, role, x25519Pubkey string) *MemberRecord {
	return &MemberRecord{
		OrgID:        orgID,
		Pubkey:       pubkey,
		Role:         role,
		X25519Pubkey: x25519Pubkey,
	}
}

type OrgConfig struct {
	OrgID                    string `json:"org_id"`
	ServeAttestationRequired bool   `json:"serve_attestation_required"`
	ContestStakeVibe         uint64 `json:"contest_stake_vibe"`
	VocabHash                string `json:"vocab_hash"`
	EmbeddingModelID         string `json:"embedding_model_id"`
	MinContributionsPerEpoch uint64 `json:"min_contributions_per_epoch"`
}

func orgConfigToStored(cfg *OrgConfig) *StoredOrgConfig {
	return &StoredOrgConfig{
		OrgId:                    cfg.OrgID,
		ServeAttestationRequired: cfg.ServeAttestationRequired,
		ContestStakeVibe:         cfg.ContestStakeVibe,
		VocabHash:                cfg.VocabHash,
		EmbeddingModelId:         cfg.EmbeddingModelID,
		MinContributionsPerEpoch: cfg.MinContributionsPerEpoch,
	}
}

func storedToOrgConfig(stored StoredOrgConfig) OrgConfig {
	return OrgConfig{
		OrgID:                    stored.OrgId,
		ServeAttestationRequired: stored.ServeAttestationRequired,
		ContestStakeVibe:         stored.ContestStakeVibe,
		VocabHash:                stored.VocabHash,
		EmbeddingModelID:         stored.EmbeddingModelId,
		MinContributionsPerEpoch: stored.MinContributionsPerEpoch,
	}
}
