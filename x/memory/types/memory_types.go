package types

func ValidMemoryType(mt MemoryType) bool {
	return mt == MemoryType_MEMORY_TYPE_CORRECT_IMPLEMENTATION ||
		mt == MemoryType_MEMORY_TYPE_NEGATIVE_SIGNAL
}

const (
	RelationContradicts = RelationType_RELATION_TYPE_CONTRADICTS
	RelationReplaces    = RelationType_RELATION_TYPE_REPLACES
	RelationDeprecates  = RelationType_RELATION_TYPE_DEPRECATES
	RelationSupersedes  = RelationType_RELATION_TYPE_SUPERSEDES
)

type MemoryRelationship struct {
	SourceCID    string
	TargetCID    string
	RelationType RelationType
	OrgID        string
	Proposer     string
	Approved     bool
	Epoch        uint64
}

type ValidityMetadata struct {
	ValidAfterEpoch uint64
	ValidUntilEpoch uint64
	ScopeTags       map[string]string
}