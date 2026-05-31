package types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wevibe-network/wevibe-chain/x/reputation/types"
)

// ---------------------------------------------------------------------------
// MsgUpdateReputation.ValidateBasic
// ---------------------------------------------------------------------------

func TestTypesMsgUpdateReputation_ValidateBasic_Valid(t *testing.T) {
	msg := &types.MsgUpdateReputation{
		Signer:     "cosmos1abc",
		Developer:  []byte("cosmos1dev"),
		MemoryCid:  "cid123",
		Difficulty: 5,
		Quality:    7,
	}
	require.NoError(t, msg.ValidateBasic())
}

func TestTypesMsgUpdateReputation_ValidateBasic_Branches(t *testing.T) {
	base := func() *types.MsgUpdateReputation {
		return &types.MsgUpdateReputation{
			Signer:     "cosmos1abc",
			Developer:  []byte("cosmos1dev"),
			MemoryCid:  "cid123",
			Difficulty: 5,
			Quality:    7,
		}
	}

	t.Run("empty signer", func(t *testing.T) {
		m := base()
		m.Signer = ""
		require.Error(t, m.ValidateBasic())
	})

	t.Run("empty developer", func(t *testing.T) {
		m := base()
		m.Developer = nil
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidDeveloper)
	})

	t.Run("empty developer zero-length slice", func(t *testing.T) {
		m := base()
		m.Developer = []byte{}
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidDeveloper)
	})

	t.Run("empty memory cid", func(t *testing.T) {
		m := base()
		m.MemoryCid = ""
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidMemory)
	})

	t.Run("difficulty over max", func(t *testing.T) {
		m := base()
		m.Difficulty = 11
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidDifficulty)
	})

	t.Run("quality over max", func(t *testing.T) {
		m := base()
		m.Quality = 11
		require.ErrorIs(t, m.ValidateBasic(), types.ErrInvalidQuality)
	})

	t.Run("difficulty boundary 10 ok", func(t *testing.T) {
		m := base()
		m.Difficulty = 10
		require.NoError(t, m.ValidateBasic())
	})

	t.Run("quality boundary 10 ok", func(t *testing.T) {
		m := base()
		m.Quality = 10
		require.NoError(t, m.ValidateBasic())
	})

	t.Run("difficulty and quality zero ok", func(t *testing.T) {
		m := base()
		m.Difficulty = 0
		m.Quality = 0
		require.NoError(t, m.ValidateBasic())
	})
}

// ---------------------------------------------------------------------------
// MsgUpdateParams.ValidateBasic
// ---------------------------------------------------------------------------

func TestTypesMsgUpdateParams_ValidateBasic_Valid(t *testing.T) {
	p := types.DefaultParams()
	msg := &types.MsgUpdateParams{
		Authority: "cosmos1authority",
		Params:    &p,
	}
	require.NoError(t, msg.ValidateBasic())
}

func TestTypesMsgUpdateParams_ValidateBasic_EmptyAuthority(t *testing.T) {
	p := types.DefaultParams()
	msg := &types.MsgUpdateParams{
		Authority: "",
		Params:    &p,
	}
	require.Error(t, msg.ValidateBasic())
}

func TestTypesMsgUpdateParams_ValidateBasic_InvalidParams(t *testing.T) {
	msg := &types.MsgUpdateParams{
		Authority: "cosmos1authority",
		Params:    &types.Params{MaxDifficulty: 0, MaxQuality: 10},
	}
	require.Error(t, msg.ValidateBasic())
}

// ---------------------------------------------------------------------------
// MsgIncrementContribution.ValidateBasic
// ---------------------------------------------------------------------------

func TestTypesMsgIncrementContribution_ValidateBasic_Valid(t *testing.T) {
	msg := &types.MsgIncrementContribution{
		Authority:     "cosmos1authority",
		ContributorId: "contrib1",
	}
	require.NoError(t, msg.ValidateBasic())
}

func TestTypesMsgIncrementContribution_ValidateBasic_Branches(t *testing.T) {
	t.Run("empty authority", func(t *testing.T) {
		msg := &types.MsgIncrementContribution{Authority: "", ContributorId: "contrib1"}
		require.Error(t, msg.ValidateBasic())
	})
	t.Run("empty contributor_id", func(t *testing.T) {
		msg := &types.MsgIncrementContribution{Authority: "cosmos1authority", ContributorId: ""}
		require.Error(t, msg.ValidateBasic())
	})
}

// ---------------------------------------------------------------------------
// MsgIncrementServe.ValidateBasic
// ---------------------------------------------------------------------------

func TestTypesMsgIncrementServe_ValidateBasic_Valid(t *testing.T) {
	msg := &types.MsgIncrementServe{
		Authority:     "cosmos1authority",
		ContributorId: "contrib1",
	}
	require.NoError(t, msg.ValidateBasic())
}

func TestTypesMsgIncrementServe_ValidateBasic_Branches(t *testing.T) {
	t.Run("empty authority", func(t *testing.T) {
		msg := &types.MsgIncrementServe{Authority: "", ContributorId: "contrib1"}
		require.Error(t, msg.ValidateBasic())
	})
	t.Run("empty contributor_id", func(t *testing.T) {
		msg := &types.MsgIncrementServe{Authority: "cosmos1authority", ContributorId: ""}
		require.Error(t, msg.ValidateBasic())
	})
}

// ---------------------------------------------------------------------------
// MsgRecordBan.ValidateBasic
// ---------------------------------------------------------------------------

func TestTypesMsgRecordBan_ValidateBasic_Valid(t *testing.T) {
	msg := &types.MsgRecordBan{
		Authority:     "cosmos1authority",
		ContributorId: "contrib1",
	}
	require.NoError(t, msg.ValidateBasic())
}

func TestTypesMsgRecordBan_ValidateBasic_Branches(t *testing.T) {
	t.Run("empty authority", func(t *testing.T) {
		msg := &types.MsgRecordBan{Authority: "", ContributorId: "contrib1"}
		require.Error(t, msg.ValidateBasic())
	})
	t.Run("empty contributor_id", func(t *testing.T) {
		msg := &types.MsgRecordBan{Authority: "cosmos1authority", ContributorId: ""}
		require.Error(t, msg.ValidateBasic())
	})
}

// ---------------------------------------------------------------------------
// Params.Validate / DefaultParams
// ---------------------------------------------------------------------------

func TestTypesDefaultParams(t *testing.T) {
	p := types.DefaultParams()
	require.True(t, p.Active)
	require.Equal(t, uint32(10), p.MaxDifficulty)
	require.Equal(t, uint32(10), p.MaxQuality)
	require.Equal(t, uint64(5), p.ServeXpPerServe)
	require.Equal(t, uint64(2), p.SelfServeXpPerServe)
	require.NoError(t, p.Validate())
}

func TestTypesParams_Validate_Branches(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p := types.Params{MaxDifficulty: 1, MaxQuality: 1}
		require.NoError(t, p.Validate())
	})
	t.Run("zero max_difficulty", func(t *testing.T) {
		p := types.Params{MaxDifficulty: 0, MaxQuality: 10}
		require.Error(t, p.Validate())
	})
	t.Run("zero max_quality", func(t *testing.T) {
		p := types.Params{MaxDifficulty: 10, MaxQuality: 0}
		require.Error(t, p.Validate())
	})
	t.Run("both zero reports difficulty first", func(t *testing.T) {
		p := types.Params{MaxDifficulty: 0, MaxQuality: 0}
		require.Error(t, p.Validate())
	})
}

// ---------------------------------------------------------------------------
// ReputationStats: NewReputationStats / Validate / AddMemory
// ---------------------------------------------------------------------------

func TestTypesNewReputationStats(t *testing.T) {
	r := types.NewReputationStats("dev1")
	require.Equal(t, "dev1", r.DeveloperID)
	require.NotNil(t, r.DomainTags)
	require.NotNil(t, r.ProvenanceBreakdown)
	require.Equal(t, uint64(0), r.ServeCount)
	require.Equal(t, uint64(0), r.SelfServeCount)
	require.Equal(t, uint64(0), r.OrgBreadth)
	require.Equal(t, uint64(0), r.FirstSeenEpoch)
	require.Equal(t, uint64(0), r.ServeXP)
	require.NoError(t, r.Validate())
}

func TestTypesReputationStats_Validate_EmptyDeveloper(t *testing.T) {
	r := types.NewReputationStats("")
	require.ErrorIs(t, r.Validate(), types.ErrInvalidDeveloper)
}

func TestTypesReputationStats_AddMemory(t *testing.T) {
	r := types.NewReputationStats("dev1")

	r.AddMemory(5, 7, []string{"golang", "cosmos"}, "commitllm")
	require.Equal(t, uint64(1), r.MemoryCount)
	require.Equal(t, uint64(1), r.DifficultyBucket[5])
	require.Equal(t, uint64(1), r.DomainTags["golang"])
	require.Equal(t, uint64(1), r.DomainTags["cosmos"])
	require.Equal(t, uint64(1), r.ProvenanceBreakdown["commitllm"])
	require.Equal(t, uint64(35), r.XP)

	// second memory accumulates
	r.AddMemory(3, 4, []string{"golang"}, "unattested")
	require.Equal(t, uint64(2), r.MemoryCount)
	require.Equal(t, uint64(2), r.DomainTags["golang"])
	require.Equal(t, uint64(1), r.DifficultyBucket[3])
	require.Equal(t, uint64(35+12), r.XP)
}

func TestTypesReputationStats_AddMemory_BoundaryAndEmpty(t *testing.T) {
	r := types.NewReputationStats("dev1")

	// difficulty boundary 10 is recorded
	r.AddMemory(10, 1, nil, "")
	require.Equal(t, uint64(1), r.DifficultyBucket[10])
	// empty provenance not counted
	require.Equal(t, 0, len(r.ProvenanceBreakdown))

	// difficulty > 10 is not bucketed (bucket has 11 slots: 0..10)
	r.AddMemory(11, 2, []string{"", "rust"}, "proxy-attested")
	require.Equal(t, uint64(2), r.MemoryCount)
	// empty tag skipped, "rust" counted
	require.Equal(t, uint64(0), r.DomainTags[""])
	require.Equal(t, uint64(1), r.DomainTags["rust"])
	require.Equal(t, uint64(1), r.ProvenanceBreakdown["proxy-attested"])

	// zero difficulty/quality => zero XP added, bucket[0] incremented
	before := r.XP
	r.AddMemory(0, 0, nil, "")
	require.Equal(t, before, r.XP)
	require.Equal(t, uint64(1), r.DifficultyBucket[0])
}

// ---------------------------------------------------------------------------
// AttestedMemory: NewAttestedMemory / Validate / GetXP
// ---------------------------------------------------------------------------

func TestTypesNewAttestedMemory(t *testing.T) {
	m := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, []string{"golang"}, "commitllm")
	require.Equal(t, []byte("dev1"), m.Developer)
	require.Equal(t, "cid1", m.MemoryCID)
	require.Equal(t, uint8(5), m.Difficulty)
	require.Equal(t, uint8(7), m.Quality)
	require.Equal(t, []string{"golang"}, m.DomainTags)
	require.Equal(t, "commitllm", m.Provenance)
}

func TestTypesAttestedMemory_Validate_Valid(t *testing.T) {
	m := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 7, nil, "")
	require.NoError(t, m.Validate())
}

func TestTypesAttestedMemory_Validate_Branches(t *testing.T) {
	t.Run("empty developer nil", func(t *testing.T) {
		m := types.NewAttestedMemory(nil, "cid1", 5, 7, nil, "")
		require.ErrorIs(t, m.Validate(), types.ErrInvalidDeveloper)
	})
	t.Run("empty developer zero-length", func(t *testing.T) {
		m := types.NewAttestedMemory([]byte{}, "cid1", 5, 7, nil, "")
		require.ErrorIs(t, m.Validate(), types.ErrInvalidDeveloper)
	})
	t.Run("empty cid", func(t *testing.T) {
		m := types.NewAttestedMemory([]byte("dev1"), "", 5, 7, nil, "")
		require.ErrorIs(t, m.Validate(), types.ErrInvalidMemory)
	})
	t.Run("difficulty over max", func(t *testing.T) {
		m := types.NewAttestedMemory([]byte("dev1"), "cid1", 11, 7, nil, "")
		require.ErrorIs(t, m.Validate(), types.ErrInvalidDifficulty)
	})
	t.Run("quality over max", func(t *testing.T) {
		m := types.NewAttestedMemory([]byte("dev1"), "cid1", 5, 11, nil, "")
		require.ErrorIs(t, m.Validate(), types.ErrInvalidQuality)
	})
	t.Run("difficulty boundary 10 ok", func(t *testing.T) {
		m := types.NewAttestedMemory([]byte("dev1"), "cid1", 10, 10, nil, "")
		require.NoError(t, m.Validate())
	})
	t.Run("zero difficulty and quality ok", func(t *testing.T) {
		m := types.NewAttestedMemory([]byte("dev1"), "cid1", 0, 0, nil, "")
		require.NoError(t, m.Validate())
	})
}

func TestTypesAttestedMemory_GetXP(t *testing.T) {
	require.Equal(t, uint64(35), types.NewAttestedMemory([]byte("d"), "c", 5, 7, nil, "").GetXP())
	require.Equal(t, uint64(0), types.NewAttestedMemory([]byte("d"), "c", 0, 9, nil, "").GetXP())
	require.Equal(t, uint64(0), types.NewAttestedMemory([]byte("d"), "c", 9, 0, nil, "").GetXP())
	require.Equal(t, uint64(100), types.NewAttestedMemory([]byte("d"), "c", 10, 10, nil, "").GetXP())
}

// ---------------------------------------------------------------------------
// DifficultyHistogram: NewDifficultyHistogram
// ---------------------------------------------------------------------------

func TestTypesNewDifficultyHistogram(t *testing.T) {
	var buckets [11]uint64
	buckets[3] = 2
	buckets[7] = 1
	h := types.NewDifficultyHistogram([]byte("dev1"), buckets)
	require.Equal(t, []byte("dev1"), h.Developer)
	require.Equal(t, buckets, h.Buckets)
	require.Equal(t, uint64(3), h.TotalCount)
}

func TestTypesNewDifficultyHistogram_Empty(t *testing.T) {
	var buckets [11]uint64
	h := types.NewDifficultyHistogram([]byte("dev1"), buckets)
	require.Equal(t, uint64(0), h.TotalCount)
}

// ---------------------------------------------------------------------------
// DomainExpertise: NewDomainExpertise
// ---------------------------------------------------------------------------

func TestTypesNewDomainExpertise(t *testing.T) {
	tags := map[string]uint64{"golang": 2, "rust": 1}
	d := types.NewDomainExpertise([]byte("dev1"), tags)
	require.Equal(t, []byte("dev1"), d.Developer)
	require.Equal(t, tags, d.DomainTags)
	require.Equal(t, uint64(3), d.TotalTags)
}

func TestTypesNewDomainExpertise_Empty(t *testing.T) {
	d := types.NewDomainExpertise([]byte("dev1"), map[string]uint64{})
	require.Equal(t, uint64(0), d.TotalTags)
}

// ---------------------------------------------------------------------------
// ProvenanceStats: NewProvenanceStats
// ---------------------------------------------------------------------------

func TestTypesNewProvenanceStats(t *testing.T) {
	breakdown := map[string]uint64{
		"commitllm":      3,
		"proxy-attested": 2,
		"unattested":     1,
	}
	s := types.NewProvenanceStats([]byte("dev1"), breakdown)
	require.Equal(t, []byte("dev1"), s.Developer)
	require.Equal(t, uint64(3), s.Tier1Count)
	require.Equal(t, uint64(2), s.Tier2Count)
	require.Equal(t, uint64(1), s.UnattestedCount)
	require.Equal(t, uint64(6), s.TotalCount)
}

func TestTypesNewProvenanceStats_UnknownProvenanceCountsTotalOnly(t *testing.T) {
	breakdown := map[string]uint64{
		"commitllm": 4,
		"mystery":   5,
	}
	s := types.NewProvenanceStats([]byte("dev1"), breakdown)
	require.Equal(t, uint64(4), s.Tier1Count)
	require.Equal(t, uint64(0), s.Tier2Count)
	require.Equal(t, uint64(0), s.UnattestedCount)
	// total accumulates every entry, even unknown buckets
	require.Equal(t, uint64(9), s.TotalCount)
}

func TestTypesNewProvenanceStats_Empty(t *testing.T) {
	s := types.NewProvenanceStats([]byte("dev1"), map[string]uint64{})
	require.Equal(t, uint64(0), s.TotalCount)
	require.Equal(t, uint64(0), s.Tier1Count)
}

// ---------------------------------------------------------------------------
// GenesisState JSON round-trip (genesis.go helpers)
// ---------------------------------------------------------------------------

func TestTypesGenesisState_JSONRoundTrip(t *testing.T) {
	g := &types.GenesisState{
		Active: true,
		Stats: []*types.ReputationStats{
			{DeveloperID: "dev1", MemoryCount: 2, XP: 35},
		},
		ContributorOrgSets: nil,
	}

	bz, err := json.Marshal(g)
	require.NoError(t, err)

	var out types.GenesisState
	require.NoError(t, json.Unmarshal(bz, &out))

	require.True(t, out.Active)
	require.Len(t, out.Stats, 1)
	require.Equal(t, "dev1", out.Stats[0].DeveloperID)
	require.Equal(t, uint64(2), out.Stats[0].MemoryCount)
	require.Equal(t, uint64(35), out.Stats[0].XP)
	require.Empty(t, out.ContributorOrgSets)
}

func TestTypesGenesisState_JSONRoundTrip_Inactive(t *testing.T) {
	g := &types.GenesisState{Active: false}
	bz, err := json.Marshal(g)
	require.NoError(t, err)

	var out types.GenesisState
	require.NoError(t, json.Unmarshal(bz, &out))
	require.False(t, out.Active)
	require.Empty(t, out.Stats)
}

func TestTypesGenesisState_UnmarshalInvalid(t *testing.T) {
	var out types.GenesisState
	require.Error(t, out.UnmarshalJSON([]byte("not-json")))
}

// ---------------------------------------------------------------------------
// constants.go
// ---------------------------------------------------------------------------

func TestTypesReputationModuleName(t *testing.T) {
	require.Equal(t, "reputation", types.ReputationModuleName)
}

// ---------------------------------------------------------------------------
// ContributorProfile (contributor_profile.go) — struct field wiring
// ---------------------------------------------------------------------------

func TestTypesContributorProfile_Fields(t *testing.T) {
	p := types.ContributorProfile{
		ContributorId:      "contrib1",
		OrgId:              "org1",
		ContributionCount:  4,
		ServeCount:         3,
		ReportUpheldCount:  1,
		FirstSeenEpoch:     5,
		FirstSeenTimestamp: 123456,
	}
	require.Equal(t, "contrib1", p.ContributorId)
	require.Equal(t, "org1", p.OrgId)
	require.Equal(t, uint64(4), p.ContributionCount)
	require.Equal(t, uint64(3), p.ServeCount)
	require.Equal(t, uint64(1), p.ReportUpheldCount)
	require.Equal(t, uint64(5), p.FirstSeenEpoch)
	require.Equal(t, int64(123456), p.FirstSeenTimestamp)
}
