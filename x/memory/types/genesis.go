package types

type GenesisState struct {
	PendingCommitments []*PendingCommitment      `json:"pending_commitments"`
	MemoryCommitments  []*MemoryCommitment       `json:"memory_commitments"`
	Relationships      []*MemoryRelationship     `json:"relationships"`
	ValidityMetadata   []*StoredValidityMetadata `json:"validity_metadata"`
	MerkleRoots        []*EpochMerkleRoot        `json:"merkle_roots"`
	Params             Params                    `json:"params"`
}

func NewGenesisState(pending []*PendingCommitment, commitments []*MemoryCommitment, relationships []*MemoryRelationship, validity []*StoredValidityMetadata, merkle []*EpochMerkleRoot, params Params) *GenesisState {
	return &GenesisState{
		PendingCommitments: pending,
		MemoryCommitments:  commitments,
		Relationships:      relationships,
		ValidityMetadata:   validity,
		MerkleRoots:        merkle,
		Params:             params,
	}
}
