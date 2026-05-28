package types

import (
	"encoding/hex"
)

const (
	ModuleName = "serve"
	StoreKey   = "serve"

	// Prefix that indexes matched keywords per memory per epoch.
	// Key: matched_keywords/{org_id}/{cid_hex}/{epoch}/{keyword}
	// Value: 1 (presence indicator; absence means not matched)
	MatchedKeywordsPrefix = "matched_keywords/"
)

const (
	ContentHashLen = 32
	NullifierLen   = 32
)

type ServeAttestation struct {
	OrgID           string
	ContentHash     []byte
	ServeKey        string
	ContributorID   string
	Epoch           uint64
	Nullifier       []byte
	IsSelfServe     bool
	ModelID         string
	TurnCount       uint32
	MatchedKeywords []string
}

func NewServeAttestation(orgID string, contentHash []byte, serveKey, contributorID string, epoch uint64, nullifier []byte, isSelfServe bool, modelID string, turnCount uint32, matchedKeywords []string) *ServeAttestation {
	return &ServeAttestation{
		OrgID:           orgID,
		ContentHash:     contentHash,
		ServeKey:        serveKey,
		ContributorID:   contributorID,
		Epoch:           epoch,
		Nullifier:       nullifier,
		IsSelfServe:     isSelfServe,
		ModelID:         modelID,
		TurnCount:       turnCount,
		MatchedKeywords: matchedKeywords,
	}
}

func (sa *ServeAttestation) Validate() error {
	if sa.OrgID == "" {
		return ErrInvalidOrgID
	}
	if len(sa.ContentHash) != ContentHashLen {
		return ErrInvalidContentHash
	}
	if sa.ServeKey == "" {
		return ErrInvalidServeKey
	}
	if sa.ContributorID == "" {
		return ErrInvalidContributor
	}
	if len(sa.Nullifier) != NullifierLen {
		return ErrInvalidNullifier
	}
	return nil
}

type EpochServeStats struct {
	OrgID                string
	Epoch                uint64
	TotalServes          uint64
	UniqueMemoriesServed uint64
	UniqueServeKeys      uint64
	SelfServes           uint64
	ModelBreakdown       map[string]uint64
}

func NewEpochServeStats(orgID string, epoch uint64) *EpochServeStats {
	return &EpochServeStats{
		OrgID:                orgID,
		Epoch:                epoch,
		TotalServes:          0,
		UniqueMemoriesServed: 0,
		UniqueServeKeys:      0,
		SelfServes:           0,
		ModelBreakdown:       make(map[string]uint64),
	}
}

type ContributorEpochServes struct {
	ContributorID  string
	Epoch          uint64
	ServeCount     uint64
	SelfServeCount uint64
	OrgIDs         []string
	TotalTurns     uint64
}

func NewContributorEpochServes(contributorID string, epoch uint64) *ContributorEpochServes {
	return &ContributorEpochServes{
		ContributorID:  contributorID,
		Epoch:          epoch,
		ServeCount:     0,
		SelfServeCount: 0,
		OrgIDs:         []string{},
		TotalTurns:     0,
	}
}

func (cs *ContributorEpochServes) AddOrgID(orgID string) {
	for _, id := range cs.OrgIDs {
		if id == orgID {
			return
		}
	}
	cs.OrgIDs = append(cs.OrgIDs, orgID)
}

func ServeAttestationToStored(sa *ServeAttestation) *StoredServeAttestation {
	return &StoredServeAttestation{
		OrgId:             sa.OrgID,
		MemoryContentHash: sa.ContentHash,
		ServeKey:          sa.ServeKey,
		ContributorId:     sa.ContributorID,
		Epoch:             sa.Epoch,
		Nullifier:         sa.Nullifier,
		IsSelfServe:       sa.IsSelfServe,
		ModelId:           sa.ModelID,
		TurnCount:         sa.TurnCount,
		MatchedKeywords:   sa.MatchedKeywords,
	}
}

func StoredToServeAttestation(stored StoredServeAttestation) ServeAttestation {
	return ServeAttestation{
		OrgID:           stored.OrgId,
		ContentHash:     stored.MemoryContentHash,
		ServeKey:        stored.ServeKey,
		ContributorID:   stored.ContributorId,
		Epoch:           stored.Epoch,
		Nullifier:       stored.Nullifier,
		IsSelfServe:     stored.IsSelfServe,
		ModelID:         stored.ModelId,
		TurnCount:       stored.TurnCount,
		MatchedKeywords: stored.MatchedKeywords,
	}
}

func EpochServeStatsToStored(es *EpochServeStats) *StoredEpochServeStats {
	return &StoredEpochServeStats{
		OrgId:                es.OrgID,
		Epoch:                es.Epoch,
		TotalServes:          es.TotalServes,
		UniqueMemoriesServed: es.UniqueMemoriesServed,
		UniqueServeKeys:      es.UniqueServeKeys,
		SelfServes:           es.SelfServes,
		ModelBreakdown:       es.ModelBreakdown,
	}
}

func StoredToEpochServeStats(stored StoredEpochServeStats) EpochServeStats {
	return EpochServeStats{
		OrgID:                stored.OrgId,
		Epoch:                stored.Epoch,
		TotalServes:          stored.TotalServes,
		UniqueMemoriesServed: stored.UniqueMemoriesServed,
		UniqueServeKeys:      stored.UniqueServeKeys,
		SelfServes:           stored.SelfServes,
		ModelBreakdown:       stored.ModelBreakdown,
	}
}

func ContributorEpochServesToStored(cs *ContributorEpochServes) *StoredContributorEpochServes {
	return &StoredContributorEpochServes{
		ContributorId:  cs.ContributorID,
		Epoch:          cs.Epoch,
		ServeCount:     cs.ServeCount,
		SelfServeCount: cs.SelfServeCount,
		OrgIds:         cs.OrgIDs,
		TotalTurns:     cs.TotalTurns,
	}
}

func StoredToContributorEpochServes(stored StoredContributorEpochServes) ContributorEpochServes {
	return ContributorEpochServes{
		ContributorID:  stored.ContributorId,
		Epoch:          stored.Epoch,
		ServeCount:     stored.ServeCount,
		SelfServeCount: stored.SelfServeCount,
		OrgIDs:         stored.OrgIds,
		TotalTurns:     stored.TotalTurns,
	}
}

func ContentHashToHex(hash []byte) string {
	return hex.EncodeToString(hash)
}
