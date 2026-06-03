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
	Leader          string    `json:"leader"`
	Domain          string    `json:"domain"`
	CreatedAt       uint64    `json:"created_at"`
	RenewalHeight   uint64    `json:"renewal_height"`
	StorageQuota    uint64    `json:"storage_quota"`
	RetrievalBudget uint64    `json:"retrieval_budget"`
	Status          OrgStatus `json:"status"`
	// CO-044 (D-S32-CO044-KEY-SEPARATION): per-org hub serving key chain address
	// (the only signer allowed to submit serve/denial batches) and the leader's
	// chain wallet address (the on-chain authority for org decisions and serving
	// key rotation).
	HubServingAddress   string `json:"hub_serving_address"`
	LeaderWalletAddress string `json:"leader_wallet_address"`
}

func NewOrg(orgID, leader, domain string, storageQuota, retrievalBudget uint64) *Org {
	return &Org{
		OrgID:           orgID,
		Leader:          leader,
		Domain:          domain,
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
	OrgID  string `json:"org_id"`
	Pubkey string `json:"pubkey"`
	Role   string `json:"role"`
}

func NewMemberRecord(orgID, pubkey, role string) *MemberRecord {
	return &MemberRecord{
		OrgID:  orgID,
		Pubkey: pubkey,
		Role:   role,
	}
}

type DynamicPrice struct {
	Price         uint64 `json:"price"`
	LastCreation  uint64 `json:"last_creation"`
	CreationCount uint64 `json:"creation_count"`
}

type Treasury struct {
	OrgID   string `json:"org_id"`
	Balance string `json:"balance"`
}

func treasuryToStored(treasury *Treasury) *StoredTreasury {
	return &StoredTreasury{
		OrgId:   treasury.OrgID,
		Balance: treasury.Balance,
	}
}

func storedToTreasury(stored StoredTreasury) Treasury {
	return Treasury{
		OrgID:   stored.OrgId,
		Balance: stored.Balance,
	}
}

type RepTierConfig struct {
	OrgID string           `json:"org_id"`
	Tiers []*RepTierRecord `json:"tiers"`
}

type RepTierRecord struct {
	MinReputation            uint64 `json:"min_reputation"`
	MaxReputation            uint64 `json:"max_reputation"`
	MaxContributionsPerEpoch uint64 `json:"max_contributions_per_epoch"`
	PayoutPerMemory          string `json:"payout_per_memory"`
}

func repTierConfigToStored(cfg *RepTierConfig) *StoredRepTierConfig {
	tiers := make([]*RepTier, len(cfg.Tiers))
	for i, t := range cfg.Tiers {
		tiers[i] = &RepTier{
			MinReputation:            t.MinReputation,
			MaxReputation:            t.MaxReputation,
			MaxContributionsPerEpoch: t.MaxContributionsPerEpoch,
			PayoutPerMemory:          t.PayoutPerMemory,
		}
	}
	return &StoredRepTierConfig{
		OrgId: cfg.OrgID,
		Tiers: tiers,
	}
}

func storedToRepTierConfig(stored StoredRepTierConfig) RepTierConfig {
	tiers := make([]*RepTierRecord, len(stored.Tiers))
	for i, t := range stored.Tiers {
		tiers[i] = &RepTierRecord{
			MinReputation:            t.MinReputation,
			MaxReputation:            t.MaxReputation,
			MaxContributionsPerEpoch: t.MaxContributionsPerEpoch,
			PayoutPerMemory:          t.PayoutPerMemory,
		}
	}
	return RepTierConfig{
		OrgID: stored.OrgId,
		Tiers: tiers,
	}
}

type OrgConfig struct {
	OrgID                    string `json:"org_id"`
	ServeAttestationRequired bool   `json:"serve_attestation_required"`
	ContestStakeVibe         uint64 `json:"contest_stake_vibe"`
	MinContributionsPerEpoch uint64 `json:"min_contributions_per_epoch"`
}

func orgConfigToStored(cfg *OrgConfig) *StoredOrgConfig {
	return &StoredOrgConfig{
		OrgId:                    cfg.OrgID,
		ServeAttestationRequired: cfg.ServeAttestationRequired,
		ContestStakeVibe:         cfg.ContestStakeVibe,
		MinContributionsPerEpoch: cfg.MinContributionsPerEpoch,
	}
}

func storedToOrgConfig(stored StoredOrgConfig) OrgConfig {
	return OrgConfig{
		OrgID:                    stored.OrgId,
		ServeAttestationRequired: stored.ServeAttestationRequired,
		ContestStakeVibe:         stored.ContestStakeVibe,
		MinContributionsPerEpoch: stored.MinContributionsPerEpoch,
	}
}
