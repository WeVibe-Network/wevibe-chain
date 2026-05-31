package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wevibe-network/wevibe-chain/x/serve/types"
)

func hash32(b byte) []byte {
	h := make([]byte, types.ContentHashLen)
	for i := range h {
		h[i] = b
	}
	return h
}

func nullifier32(b byte) []byte {
	n := make([]byte, types.NullifierLen)
	for i := range n {
		n[i] = b
	}
	return n
}

func validServeEntry() *types.ServeEntry {
	return &types.ServeEntry{
		MemoryContentHash: hash32(0x01),
		ServeKey:          "serve-key",
		ContributorId:     "contrib",
		Nullifier:         nullifier32(0x01),
		MatchedKeywords:   []string{"alpha"},
	}
}

func TestMsgSubmitServeBatch_ValidateBasic(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		msg := &types.MsgSubmitServeBatch{
			Signer: "s", OrgId: "org", Epoch: 1,
			Serves: []*types.ServeEntry{validServeEntry()},
		}
		require.NoError(t, msg.ValidateBasic())
	})

	cases := []struct {
		name   string
		mutate func(*types.MsgSubmitServeBatch)
	}{
		{"empty signer", func(m *types.MsgSubmitServeBatch) { m.Signer = "" }},
		{"empty org", func(m *types.MsgSubmitServeBatch) { m.OrgId = "" }},
		{"empty batch", func(m *types.MsgSubmitServeBatch) { m.Serves = nil }},
		{"short hash", func(m *types.MsgSubmitServeBatch) { m.Serves[0].MemoryContentHash = []byte{1, 2} }},
		{"empty serve key", func(m *types.MsgSubmitServeBatch) { m.Serves[0].ServeKey = "" }},
		{"empty contributor", func(m *types.MsgSubmitServeBatch) { m.Serves[0].ContributorId = "" }},
		{"short nullifier", func(m *types.MsgSubmitServeBatch) { m.Serves[0].Nullifier = []byte{1} }},
		{"nil matched keywords", func(m *types.MsgSubmitServeBatch) { m.Serves[0].MatchedKeywords = nil }},
		{"empty keyword string", func(m *types.MsgSubmitServeBatch) { m.Serves[0].MatchedKeywords = []string{""} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := &types.MsgSubmitServeBatch{
				Signer: "s", OrgId: "org", Epoch: 1,
				Serves: []*types.ServeEntry{validServeEntry()},
			}
			tc.mutate(msg)
			require.Error(t, msg.ValidateBasic())
		})
	}
}

func TestMsgSubmitDenialBatch_ValidateBasic(t *testing.T) {
	valid := func() *types.MsgSubmitDenialBatch {
		return &types.MsgSubmitDenialBatch{
			Signer: "s", OrgId: "org", Epoch: 1,
			Entries: []*types.DenialEntry{{MemoryHash: hash32(0x02), Nullifier: nullifier32(0x02), DenyKey: "dk"}},
		}
	}
	require.NoError(t, valid().ValidateBasic())

	cases := []struct {
		name   string
		mutate func(*types.MsgSubmitDenialBatch)
	}{
		{"empty signer", func(m *types.MsgSubmitDenialBatch) { m.Signer = "" }},
		{"empty org", func(m *types.MsgSubmitDenialBatch) { m.OrgId = "" }},
		{"empty entries", func(m *types.MsgSubmitDenialBatch) { m.Entries = nil }},
		{"short hash", func(m *types.MsgSubmitDenialBatch) { m.Entries[0].MemoryHash = []byte{1} }},
		{"short nullifier", func(m *types.MsgSubmitDenialBatch) { m.Entries[0].Nullifier = []byte{1} }},
		{"empty deny key", func(m *types.MsgSubmitDenialBatch) { m.Entries[0].DenyKey = "" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := valid()
			tc.mutate(m)
			require.Error(t, m.ValidateBasic())
		})
	}
}

func TestMsgUpdateParams_ValidateBasic(t *testing.T) {
	require.NoError(t, (&types.MsgUpdateParams{Authority: "gov"}).ValidateBasic())
	require.Error(t, (&types.MsgUpdateParams{Authority: ""}).ValidateBasic())
}

func TestServeAttestation_Validate(t *testing.T) {
	valid := types.NewServeAttestation("org", hash32(0x03), "sk", "c", 1, nullifier32(0x03), false, "model", 2, []string{"k"})
	require.NoError(t, valid.Validate())

	require.Error(t, (&types.ServeAttestation{}).Validate())                                                   // empty org
	require.Error(t, (&types.ServeAttestation{OrgID: "o", ContentHash: []byte{1}}).Validate())                 // bad hash
	require.Error(t, (&types.ServeAttestation{OrgID: "o", ContentHash: hash32(1)}).Validate())                 // empty serve key
	require.Error(t, (&types.ServeAttestation{OrgID: "o", ContentHash: hash32(1), ServeKey: "s"}).Validate())  // empty contributor
}

func TestContributorEpochServes_AddOrgID_Dedup(t *testing.T) {
	cs := types.NewContributorEpochServes("c", 1)
	cs.AddOrgID("org-a")
	cs.AddOrgID("org-a") // duplicate ignored
	cs.AddOrgID("org-b")
	require.ElementsMatch(t, []string{"org-a", "org-b"}, cs.OrgIDs)
}

func TestServeAttestation_StoredRoundtrip(t *testing.T) {
	sa := types.NewServeAttestation("org", hash32(0x04), "sk", "c", 7, nullifier32(0x04), true, "m", 3, []string{"x", "y"})
	stored := types.ServeAttestationToStored(sa)
	back := types.StoredToServeAttestation(*stored)
	require.Equal(t, sa.OrgID, back.OrgID)
	require.Equal(t, sa.ContentHash, back.ContentHash)
	require.Equal(t, sa.IsSelfServe, back.IsSelfServe)
	require.Equal(t, sa.MatchedKeywords, back.MatchedKeywords)
}

func TestEpochServeStats_StoredRoundtrip(t *testing.T) {
	es := types.NewEpochServeStats("org", 2)
	es.TotalServes = 5
	es.SelfServes = 2
	es.ModelBreakdown["m"] = 3
	back := types.StoredToEpochServeStats(*types.EpochServeStatsToStored(es))
	require.Equal(t, uint64(5), back.TotalServes)
	require.Equal(t, uint64(2), back.SelfServes)
	require.Equal(t, uint64(3), back.ModelBreakdown["m"])
}

func TestContributorEpochServes_StoredRoundtrip(t *testing.T) {
	cs := types.NewContributorEpochServes("c", 3)
	cs.ServeCount = 4
	cs.SelfServeCount = 1
	cs.TotalTurns = 10
	cs.AddOrgID("org-a")
	back := types.StoredToContributorEpochServes(*types.ContributorEpochServesToStored(cs))
	require.Equal(t, uint64(4), back.ServeCount)
	require.Equal(t, uint64(1), back.SelfServeCount)
	require.Equal(t, uint64(10), back.TotalTurns)
	require.Equal(t, []string{"org-a"}, back.OrgIDs)
}

func TestDefaultParamsAndValidate(t *testing.T) {
	p := types.DefaultParams()
	require.NoError(t, p.Validate())
	require.Equal(t, uint32(500), p.MaxServesPerBatch)
	require.Equal(t, uint32(100), p.MaxServesPerMemoryPerEpoch)
}

func TestContentHashToHex(t *testing.T) {
	require.Equal(t, "", types.ContentHashToHex(nil))
	require.Equal(t, "0102", types.ContentHashToHex([]byte{0x01, 0x02}))
}

func TestGenesisJSONRoundtrip(t *testing.T) {
	gs := types.NewGenesisState(
		[]*types.ServeAttestation{types.NewServeAttestation("org", hash32(5), "sk", "c", 1, nullifier32(5), false, "m", 1, []string{"k"})},
		nil, nil, nil,
	)
	bz, err := gs.MarshalJSON()
	require.NoError(t, err)
	var decoded types.GenesisState
	require.NoError(t, decoded.UnmarshalJSON(bz))
	require.Len(t, decoded.Attestations, 1)
	require.Equal(t, "org", decoded.Attestations[0].OrgID)
}
