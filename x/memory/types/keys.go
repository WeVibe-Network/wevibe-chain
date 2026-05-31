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
	ContentHashLen = 32
)

type PendingCommitment struct {
	OrgID              string
	ContentHash        []byte
	Keywords           []*KeywordWeight
	Contributor        string
	ContributorAddress string
	Epoch              uint64
	SubmittedAt        uint64
	MemoryType         MemoryType
}

func NewPendingCommitment(orgID string, contentHash []byte, keywords []*KeywordWeight, contributor string, epoch, submittedAt uint64, memoryType MemoryType) *PendingCommitment {
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
		if kw == nil || kw.Keyword == "" {
			return ErrInvalidKeyword
		}
	}
	return nil
}

type MemoryCommitment struct {
	OrgID              string
	ContentHash        []byte
	EncryptedBlob      []byte
	Keywords           []*KeywordWeight
	Contributor        string
	ContributorAddress string
	Epoch              uint64
	CommittedAtHeight  uint64
	CommittingLeader   string
	State              MemoryState
	LastActiveEpoch    uint64
	WrappedDekEnc      []byte
	PlaintextHash      []byte
	Salt               []byte
	CiphertextHash     []byte
	WrappedDekHash     []byte
	ContributorSig     []byte
	MemoryType         MemoryType
	ApprovedAtEpoch    uint64
	ServeCountTotal    uint64
	DenialCountTotal   uint64
	ArchivedEpoch      uint64
}

func NewMemoryCommitment(orgID string, contentHash, encryptedBlob []byte, keywords []*KeywordWeight, contributor string, epoch, committedAtHeight uint64, committingLeader string, state MemoryState, lastActiveEpoch uint64, memoryType MemoryType, approvedAtEpoch uint64) *MemoryCommitment {
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
		if kw == nil || kw.Keyword == "" {
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

func pendingToStored(pc *PendingCommitment) *StoredPendingCommitment {
	return &StoredPendingCommitment{
		OrgId:              pc.OrgID,
		ContentHash:        pc.ContentHash,
		Keywords:           pc.Keywords,
		ContributorId:      pc.Contributor,
		Epoch:              pc.Epoch,
		SubmittedAtHeight:  pc.SubmittedAt,
		MemoryType:         pc.MemoryType,
		ContributorAddress: pc.ContributorAddress,
	}
}

func storedToPending(stored StoredPendingCommitment) PendingCommitment {
	return PendingCommitment{
		OrgID:              stored.OrgId,
		ContentHash:        stored.ContentHash,
		Keywords:           stored.Keywords,
		Contributor:        stored.ContributorId,
		ContributorAddress: stored.ContributorAddress,
		Epoch:              stored.Epoch,
		SubmittedAt:        stored.SubmittedAtHeight,
		MemoryType:         stored.MemoryType,
	}
}

func memoryToStored(mc *MemoryCommitment) *StoredMemoryCommitment {
	return &StoredMemoryCommitment{
		OrgId:                  mc.OrgID,
		ContentHash:            mc.ContentHash,
		EncryptedBlob:          mc.EncryptedBlob,
		Keywords:               mc.Keywords,
		ContributorPubkey:      mc.Contributor,
		ContributorAddress:     mc.ContributorAddress,
		Epoch:                  mc.Epoch,
		CommittedAtHeight:      mc.CommittedAtHeight,
		CommittingLeaderPubkey: mc.CommittingLeader,
		State:                  mc.State,
		LastActiveEpoch:        mc.LastActiveEpoch,
		WrappedDekEnc:          mc.WrappedDekEnc,
		PlaintextHash:          mc.PlaintextHash,
		Salt:                   mc.Salt,
		CiphertextHash:         mc.CiphertextHash,
		WrappedDekHash:         mc.WrappedDekHash,
		ContributorSig:         mc.ContributorSig,
		MemoryType:             mc.MemoryType,
		ApprovedAtEpoch:        mc.ApprovedAtEpoch,
		ServeCountTotal:        mc.ServeCountTotal,
		DenialCountTotal:       mc.DenialCountTotal,
		ArchivedEpoch:          mc.ArchivedEpoch,
	}
}

func storedToMemory(stored StoredMemoryCommitment) MemoryCommitment {
	return MemoryCommitment{
		OrgID:              stored.OrgId,
		ContentHash:        stored.ContentHash,
		EncryptedBlob:      stored.EncryptedBlob,
		Keywords:           stored.Keywords,
		Contributor:        stored.ContributorPubkey,
		ContributorAddress: stored.ContributorAddress,
		Epoch:              stored.Epoch,
		CommittedAtHeight:  stored.CommittedAtHeight,
		CommittingLeader:   stored.CommittingLeaderPubkey,
		State:              stored.State,
		LastActiveEpoch:    stored.LastActiveEpoch,
		WrappedDekEnc:      stored.WrappedDekEnc,
		PlaintextHash:      stored.PlaintextHash,
		Salt:               stored.Salt,
		CiphertextHash:     stored.CiphertextHash,
		WrappedDekHash:     stored.WrappedDekHash,
		ContributorSig:     stored.ContributorSig,
		MemoryType:         stored.MemoryType,
		ApprovedAtEpoch:    stored.ApprovedAtEpoch,
		ServeCountTotal:    stored.ServeCountTotal,
		DenialCountTotal:   stored.DenialCountTotal,
		ArchivedEpoch:      stored.ArchivedEpoch,
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
