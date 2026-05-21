package types

import "errors"

var (
	ErrInvalidDeveloper      = errors.New("invalid developer")
	ErrDeveloperNotFound    = errors.New("developer not found")
	ErrReputationNotActive  = errors.New("reputation module not active")
	ErrInvalidMemory        = errors.New("invalid memory")
	ErrInvalidDifficulty    = errors.New("invalid difficulty grade")
	ErrInvalidQuality       = errors.New("invalid quality grade")
	ErrNoStats              = errors.New("no stats found for developer")
	ErrUnauthorized         = errors.New("unauthorized")
)

type ReputationStats struct {
	DeveloperID         string            `json:"developer_id"`
	MemoryCount         uint64            `json:"memory_count"`
	DifficultyBucket    [11]uint64         `json:"difficulty_bucket"`
	DomainTags          map[string]uint64 `json:"domain_tags"`
	ProvenanceBreakdown map[string]uint64 `json:"provenance_breakdown"`
	XP                  uint64            `json:"xp"`
	ServeCount          uint64            `json:"serve_count"`
	SelfServeCount      uint64            `json:"self_serve_count"`
	OrgBreadth          uint64            `json:"org_breadth"`
	FirstSeenEpoch      uint64            `json:"first_seen_epoch"`
	ServeXP             uint64            `json:"serve_xp"`
}

func NewReputationStats(developerID string) *ReputationStats {
	return &ReputationStats{
		DeveloperID:         developerID,
		DomainTags:          make(map[string]uint64),
		ProvenanceBreakdown: make(map[string]uint64),
		ServeCount:          0,
		SelfServeCount:      0,
		OrgBreadth:          0,
		FirstSeenEpoch:      0,
		ServeXP:             0,
	}
}

func (r *ReputationStats) Validate() error {
	if r.DeveloperID == "" {
		return ErrInvalidDeveloper
	}
	return nil
}

func (r *ReputationStats) AddMemory(difficulty, quality uint8, domainTags []string, provenance string) {
	r.MemoryCount++

	if difficulty <= 10 {
		r.DifficultyBucket[difficulty]++
	}

	for _, tag := range domainTags {
		if tag != "" {
			r.DomainTags[tag]++
		}
	}

	if provenance != "" {
		r.ProvenanceBreakdown[provenance]++
	}

	r.XP += uint64(difficulty) * uint64(quality)
}

type AttestedMemory struct {
	Developer   []byte   `json:"developer"`
	MemoryCID   string   `json:"memory_cid"`
	Difficulty  uint8    `json:"difficulty"`
	Quality     uint8    `json:"quality"`
	DomainTags  []string `json:"domain_tags"`
	Provenance  string   `json:"provenance"`
}

func NewAttestedMemory(developer []byte, memoryCID string, difficulty, quality uint8, domainTags []string, provenance string) *AttestedMemory {
	return &AttestedMemory{
		Developer:   developer,
		MemoryCID:   memoryCID,
		Difficulty:  difficulty,
		Quality:     quality,
		DomainTags:  domainTags,
		Provenance:  provenance,
	}
}

func (a *AttestedMemory) Validate() error {
	if len(a.Developer) == 0 {
		return ErrInvalidDeveloper
	}
	if a.MemoryCID == "" {
		return ErrInvalidMemory
	}
	if a.Difficulty > 10 {
		return ErrInvalidDifficulty
	}
	if a.Quality > 10 {
		return ErrInvalidQuality
	}
	return nil
}

func (a *AttestedMemory) GetXP() uint64 {
	return uint64(a.Difficulty) * uint64(a.Quality)
}

type DifficultyHistogram struct {
	Developer  []byte          `json:"developer"`
	Buckets    [11]uint64      `json:"buckets"`
	TotalCount uint64          `json:"total_count"`
}

func NewDifficultyHistogram(developer []byte, buckets [11]uint64) *DifficultyHistogram {
	var total uint64
	for _, count := range buckets {
		total += count
	}
	return &DifficultyHistogram{
		Developer:  developer,
		Buckets:    buckets,
		TotalCount: total,
	}
}

type DomainExpertise struct {
	Developer  []byte            `json:"developer"`
	DomainTags map[string]uint64  `json:"domain_tags"`
	TotalTags  uint64             `json:"total_tags"`
}

func NewDomainExpertise(developer []byte, domainTags map[string]uint64) *DomainExpertise {
	var total uint64
	for _, count := range domainTags {
		total += count
	}
	return &DomainExpertise{
		Developer:  developer,
		DomainTags:  domainTags,
		TotalTags:   total,
	}
}

type ProvenanceStats struct {
	Developer     []byte `json:"developer"`
	Tier1Count    uint64 `json:"tier1_count"`
	Tier2Count    uint64 `json:"tier2_count"`
	UnattestedCount uint64 `json:"unattested_count"`
	TotalCount   uint64 `json:"total_count"`
}

func NewProvenanceStats(developer []byte, breakdown map[string]uint64) *ProvenanceStats {
	stats := &ProvenanceStats{
		Developer: developer,
	}

	for provenance, count := range breakdown {
		switch provenance {
		case "commitllm":
			stats.Tier1Count = count
		case "proxy-attested":
			stats.Tier2Count = count
		case "unattested":
			stats.UnattestedCount = count
		}
		stats.TotalCount += count
	}

	return stats
}

type GenesisState struct {
	Active              bool                        `json:"active"`
	Stats               []*ReputationStats          `json:"stats"`
	ContributorOrgSets []*StoredContributorOrgSet `json:"contributor_org_sets"`
}
