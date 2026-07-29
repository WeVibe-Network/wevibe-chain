package types

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
)

const (
	ModuleName = "memory"
	StoreKey   = "memory"
)

var (
	RelationshipKeyPrefix = []byte{0x10}
	ValidityKeyPrefix     = []byte{0x12}
	ParamsKey             = []byte{0x13}
	ReportKeyPrefix       = []byte{0x14}
)

const (
	ContentHashLen        = 32
	MaxProducerModelIdLen = 128
)

type ProvenanceStatus string

const (
	ProvenanceUnattested        ProvenanceStatus = "UNATTESTED"
	ProvenanceSelfDeclared      ProvenanceStatus = "SELF_DECLARED"
	ProvenanceSessionReferenced ProvenanceStatus = "SESSION_REFERENCED"
)

func DeriveProvenanceStatus(producerModelID string, attestationSessionHash []byte) ProvenanceStatus {
	if producerModelID == "" && len(attestationSessionHash) == 0 {
		return ProvenanceUnattested
	}
	if len(attestationSessionHash) == 0 {
		return ProvenanceSelfDeclared
	}
	return ProvenanceSessionReferenced
}

type PendingCommitment struct {
	OrgID                  string
	ContentHash            []byte
	Keywords               []string
	Contributor            string
	ContributorAddress     string
	ProducerModelId        string
	AttestationSessionHash []byte
	Epoch                  uint64
	SubmittedAt            uint64
	MemoryType             MemoryType
	McVersion              uint32
}

func NewPendingCommitment(orgID string, contentHash []byte, keywords []string, contributor string, epoch, submittedAt uint64, memoryType MemoryType) *PendingCommitment {
	return &PendingCommitment{
		OrgID:       orgID,
		ContentHash: contentHash,
		Keywords:    keywords,
		Contributor: contributor,
		Epoch:       epoch,
		SubmittedAt: submittedAt,
		MemoryType:  memoryType,
	}
}

func (pc *PendingCommitment) Validate() error {
	if pc.OrgID == "" {
		return ErrInvalidOrgID
	}
	if len(pc.ContentHash) != ContentHashLen {
		return ErrInvalidContentHash
	}
	if pc.Contributor == "" {
		return ErrInvalidContributor
	}
	if !ValidMemoryType(pc.MemoryType) {
		return ErrInvalidMemoryType
	}
	for _, kw := range pc.Keywords {
		if kw == "" {
			return ErrInvalidKeyword
		}
	}
	return nil
}

type MemoryCommitment struct {
	OrgID                  string
	ContentHash            []byte
	EncryptedBlob          []byte
	Keywords               []string
	Contributor            string
	ContributorAddress     string
	ProducerModelId        string
	AttestationSessionHash []byte
	Epoch                  uint64
	CommittedAtHeight      uint64
	CommittingLeader       string
	State                  MemoryState
	LastActiveEpoch        uint64
	WrappedDekEnc          []byte
	PlaintextHash          []byte
	Salt                   []byte
	CiphertextHash         []byte
	WrappedDekHash         []byte
	ContributorSig         []byte
	MemoryType             MemoryType
	ApprovedAtEpoch        uint64
	McVersion              uint32
}

func (mc *MemoryCommitment) ProvenanceStatus() ProvenanceStatus {
	if mc == nil {
		return ProvenanceUnattested
	}
	return DeriveProvenanceStatus(mc.ProducerModelId, mc.AttestationSessionHash)
}

func NewMemoryCommitment(orgID string, contentHash, encryptedBlob []byte, keywords []string, contributor string, epoch, committedAtHeight uint64, committingLeader string, state MemoryState, lastActiveEpoch uint64, memoryType MemoryType, approvedAtEpoch uint64) *MemoryCommitment {
	return &MemoryCommitment{
		OrgID:             orgID,
		ContentHash:       contentHash,
		EncryptedBlob:     encryptedBlob,
		Keywords:          keywords,
		Contributor:       contributor,
		Epoch:             epoch,
		CommittedAtHeight: committedAtHeight,
		CommittingLeader:  committingLeader,
		State:             state,
		LastActiveEpoch:   lastActiveEpoch,
		MemoryType:        memoryType,
		ApprovedAtEpoch:   approvedAtEpoch,
	}
}

func (mc *MemoryCommitment) Validate() error {
	if mc.OrgID == "" {
		return ErrInvalidOrgID
	}
	if len(mc.ContentHash) != ContentHashLen {
		return ErrInvalidContentHash
	}
	if len(mc.EncryptedBlob) == 0 {
		return ErrInvalidBlob
	}
	if mc.Contributor == "" {
		return ErrInvalidContributor
	}
	if !ValidMemoryType(mc.MemoryType) {
		return ErrInvalidMemoryType
	}
	for _, kw := range mc.Keywords {
		if kw == "" {
			return ErrInvalidKeyword
		}
	}
	return nil
}

type EpochMerkleRoot struct {
	OrgID       string
	Epoch       uint64
	MerkleRoot  []byte
	MemoryCount uint64
}

func NewEpochMerkleRoot(orgID string, epoch uint64, merkleRoot []byte, memoryCount uint64) *EpochMerkleRoot {
	return &EpochMerkleRoot{
		OrgID:       orgID,
		Epoch:       epoch,
		MerkleRoot:  merkleRoot,
		MemoryCount: memoryCount,
	}
}

func ComputeMerkleRoot(contentHashes [][]byte) []byte {
	if len(contentHashes) == 0 {
		return nil
	}
	slices.SortFunc(contentHashes, func(a, b []byte) int {
		return slices.Compare(a, b)
	})
	var concatenated []byte
	for _, h := range contentHashes {
		concatenated = append(concatenated, h...)
	}
	hash := sha256.Sum256(concatenated)
	return hash[:]
}

func ContentHashToHex(hash []byte) string {
	return hex.EncodeToString(hash)
}
