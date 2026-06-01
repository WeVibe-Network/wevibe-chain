package types

func ValidMemoryType(mt MemoryType) bool {
	return mt == MemoryType_MEMORY_TYPE_MEMORY
}

// CanonicalMemoryType is the single, shared canonicalization of a MemoryType
// into the string emitted in the contributor-submission canonical body. It is
// the one and only implementation (R-ONE-PATH): both the memory keeper (signer
// path) and the reputation keeper (verification path) call this function so the
// signed bytes and the verified bytes are identical. With the single-type model
// (D-5.1), the only valid value canonicalizes to "memory".
func CanonicalMemoryType(mt MemoryType) string {
	switch mt {
	case MemoryType_MEMORY_TYPE_MEMORY:
		return "memory"
	default:
		return ""
	}
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