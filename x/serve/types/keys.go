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
	FingerprintLen = 32
	ServePubKeyLen = 32
	ServeSigLen    = 64
)

type ServeReceipt struct {
	OrgID           string
	ContentHash     []byte
	ServeKey        string
	ServeKeyPubkey  []byte
	ContributorID   string
	Epoch           uint64
	Fingerprint     []byte
	IsSelfServe     bool
	ModelID         string
	TurnCount       uint32
	MatchedKeywords []string
}

func NewServeReceipt(orgID string, contentHash []byte, serveKey string, serveKeyPubkey []byte, contributorID string, epoch uint64, fingerprint []byte, isSelfServe bool, modelID string, turnCount uint32, matchedKeywords []string) *ServeReceipt {
	return &ServeReceipt{
		OrgID:           orgID,
		ContentHash:     contentHash,
		ServeKey:        serveKey,
		ServeKeyPubkey:  serveKeyPubkey,
		ContributorID:   contributorID,
		Epoch:           epoch,
		Fingerprint:     fingerprint,
		IsSelfServe:     isSelfServe,
		ModelID:         modelID,
		TurnCount:       turnCount,
		MatchedKeywords: matchedKeywords,
	}
}

func (sr *ServeReceipt) Validate() error {
	if sr.OrgID == "" {
		return ErrInvalidOrgID
	}
	if len(sr.ContentHash) != ContentHashLen {
		return ErrInvalidContentHash
	}
	if sr.ServeKey == "" {
		return ErrInvalidServeKey
	}
	if sr.ContributorID == "" {
		return ErrInvalidContributor
	}
	if len(sr.ServeKeyPubkey) != ServePubKeyLen {
		return ErrInvalidServeKeyPubkey
	}
	if len(sr.Fingerprint) != FingerprintLen {
		return ErrInvalidServeFingerprint
	}
	return nil
}

type EpochServeStats struct {
	OrgID                string
	Epoch                uint64
	TotalServes          uint64
	TotalDenials         uint64
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
		TotalDenials:         0,
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

func ServeReceiptToStored(sr *ServeReceipt) *StoredServeReceipt {
	return &StoredServeReceipt{
		OrgId:             sr.OrgID,
		MemoryContentHash: sr.ContentHash,
		ContributorId:     sr.ContributorID,
		Epoch:             sr.Epoch,
		IsSelfServe:       sr.IsSelfServe,
		ModelId:           sr.ModelID,
		TurnCount:         sr.TurnCount,
		MatchedKeywords:   sr.MatchedKeywords,
		ServeKeyPubkey:    sr.ServeKeyPubkey,
		Fingerprint:       sr.Fingerprint,
	}
}

func StoredToServeReceipt(stored StoredServeReceipt) ServeReceipt {
	return ServeReceipt{
		OrgID:           stored.OrgId,
		ContentHash:     stored.MemoryContentHash,
		ServeKey:        hex.EncodeToString(stored.ServeKeyPubkey),
		ServeKeyPubkey:  stored.ServeKeyPubkey,
		ContributorID:   stored.ContributorId,
		Epoch:           stored.Epoch,
		Fingerprint:     stored.Fingerprint,
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
		TotalDenials:         es.TotalDenials,
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
		TotalDenials:         stored.TotalDenials,
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
