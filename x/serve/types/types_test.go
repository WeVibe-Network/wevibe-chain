package types_test

import (
	"crypto/sha256"
	"strconv"
	"strings"
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

func fingerprint32(b byte) []byte {
	n := make([]byte, types.FingerprintLen)
	for i := range n {
		n[i] = b
	}
	return n
}

func sig64(b byte) []byte {
	s := make([]byte, types.ServeSigLen)
	for i := range s {
		s[i] = b
	}
	return s
}

func validServeEntry() *types.ServeEntry {
	return &types.ServeEntry{
		MemoryContentHash: hash32(0x01),
		ServeKeyPubkey:    hash32(0x10),
		ServeSig:          sig64(0x20),
		ContributorId:     "contrib",
		Nonce:             []byte{0x01},
		EpisodeRef:        []byte{0x11, 0x22, 0x33},
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
		{"short serve pubkey", func(m *types.MsgSubmitServeBatch) { m.Serves[0].ServeKeyPubkey = []byte{1} }},
		{"short serve signature", func(m *types.MsgSubmitServeBatch) { m.Serves[0].ServeSig = []byte{1} }},
		{"empty contributor", func(m *types.MsgSubmitServeBatch) { m.Serves[0].ContributorId = "" }},
		{"empty episode ref", func(m *types.MsgSubmitServeBatch) { m.Serves[0].EpisodeRef = nil }},
		{"oversize episode ref", func(m *types.MsgSubmitServeBatch) { m.Serves[0].EpisodeRef = make([]byte, 65) }},
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

	t.Run("episode ref boundary", func(t *testing.T) {
		// A well-formed episode_ref of exactly 64 bytes is accepted.
		entry := validServeEntry()
		entry.EpisodeRef = make([]byte, 64)
		msg := &types.MsgSubmitServeBatch{Signer: "s", OrgId: "org", Epoch: 1, Serves: []*types.ServeEntry{entry}}
		require.NoError(t, msg.ValidateBasic())

		// 64-byte accepted, 65-byte rejected (asserted in the reject table above).
		entry65 := validServeEntry()
		entry65.EpisodeRef = make([]byte, 65)
		msg65 := &types.MsgSubmitServeBatch{Signer: "s", OrgId: "org", Epoch: 1, Serves: []*types.ServeEntry{entry65}}
		require.Error(t, msg65.ValidateBasic())
	})
}

func hexRepeat(b byte, count int) string {
	return strings.Repeat(string([]byte{hexDigit(b >> 4), hexDigit(b & 0x0f)}), count)
}

func hexDigit(n byte) byte {
	if n < 10 {
		return '0' + n
	}
	return 'a' + (n - 10)
}

func validOutcomeEvent() *types.EventEntry {
	return &types.EventEntry{
		EventType:         types.EventType_EVENT_TYPE_OUTCOME,
		MemoryContentHash: hash32(0x01),
		SignerPubkey:      hash32(0x02),
		Nonce:             []byte{0x03},
		Signature:         sig64(0x04),
		Body: &types.EventEntry_Outcome{Outcome: &types.OutcomeEventBody{
			EpisodeRef:  []byte{0x10},
			Resolution:  types.OutcomeResolution_OUTCOME_RESOLUTION_WORKED,
			Source:      types.OutcomeSource_OUTCOME_SOURCE_HARVESTED,
			EvidenceRef: []byte{0x11},
			ServeRef:    fingerprint32(0x12),
		}},
	}
}

func TestCanonicalServeBodyV3_GoldenVector(t *testing.T) {
	// episodeRef: 32 bytes of a fixed repeating pattern.
	episodeRef := make([]byte, 32)
	for i := range episodeRef {
		episodeRef[i] = 0xAB
	}
	body := string(types.CanonicalServeBody("org-a", hash32(0x01), 7, hash32(0x02), []byte{0x03, 0x04}, episodeRef))
	expected := "wevibe-serve-v3\n" +
		"org-a\n" +
		hexRepeat(0x01, 32) + "\n" +
		"7\n" +
		hexRepeat(0x02, 32) + "\n" +
		"0304\n" +
		hexRepeat(0xAB, 32)
	require.Equal(t, expected, body)
	require.True(t, strings.HasPrefix(body, "wevibe-serve-v3"))
	require.NotContains(t, body, ",")
	require.Equal(t, 6, strings.Count(body, "\n"))
	// episode_ref must be present in the v3 preimage, as its hex-encoded line.
	require.Contains(t, body, "\n"+hexRepeat(0xAB, 32))
}

func TestCanonicalEventBody_GoldenVectors(t *testing.T) {
	memory := hash32(0x01)
	pubkey := hash32(0x02)
	nonce := []byte{0x03, 0x04}

	tests := []struct {
		name      string
		eventType types.EventType
		entry     *types.EventEntry
		expected  string
	}{
		{
			name:      "outcome",
			eventType: types.EventType_EVENT_TYPE_OUTCOME,
			entry: &types.EventEntry{Body: &types.EventEntry_Outcome{Outcome: &types.OutcomeEventBody{
				EpisodeRef: []byte{0x10, 0x11}, Resolution: types.OutcomeResolution_OUTCOME_RESOLUTION_WORKED, Source: types.OutcomeSource_OUTCOME_SOURCE_HARVESTED, EvidenceRef: []byte{0x12}, ServeRef: []byte{0x13, 0x14},
			}}},
			expected: "wevibe-event-v1\noutcome\norg-a\n" + hexRepeat(0x01, 32) + "\n7\n" + hexRepeat(0x02, 32) + "\n1011\n12\n1314\nresolution=worked\nsource=harvested\n0304",
		},
		{
			name:      "validity predicate",
			eventType: types.EventType_EVENT_TYPE_VALIDITY_PREDICATE,
			entry: &types.EventEntry{Body: &types.EventEntry_ValidityPredicate{ValidityPredicate: &types.ValidityPredicateEventBody{
				PredicateId: []byte{0x20}, Result: types.PredicateResult_PREDICATE_RESULT_ABSENT, EvidenceRef: []byte{0x21, 0x22},
			}}},
			expected: "wevibe-event-v1\nvalidity_predicate\norg-a\n" + hexRepeat(0x01, 32) + "\n7\n" + hexRepeat(0x02, 32) + "\n20\nresult=absent\n2122\n0304",
		},
		{
			name:      "cost to discover",
			eventType: types.EventType_EVENT_TYPE_COST_TO_DISCOVER,
			entry: &types.EventEntry{Body: &types.EventEntry_CostToDiscover{CostToDiscover: &types.CostToDiscoverEventBody{
				Cycles: 5, ToolCalls: 6, AttemptsToGreen: 7, EvidenceRef: []byte{0x30},
			}}},
			expected: "wevibe-event-v1\ncost_to_discover\norg-a\n" + hexRepeat(0x01, 32) + "\n7\n" + hexRepeat(0x02, 32) + "\ncycles=5\ntool_calls=6\nattempts_to_green=7\n30\n0304",
		},
		{
			name:      "convergence",
			eventType: types.EventType_EVENT_TYPE_CONVERGENCE,
			entry: &types.EventEntry{Body: &types.EventEntry_Convergence{Convergence: &types.ConvergenceEventBody{
				ConvergenceRef: []byte{0x40, 0x41},
			}}},
			expected: "wevibe-event-v1\nconvergence\norg-a\n" + hexRepeat(0x01, 32) + "\n7\n" + hexRepeat(0x02, 32) + "\n4041\n0304",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body, err := types.CanonicalEventBody(tc.eventType, "org-a", memory, 7, pubkey, nonce, tc.entry)
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(body))
		})
	}

	_, err := types.CanonicalEventBody(types.EventType_EVENT_TYPE_CONTEST, "org-a", memory, 7, pubkey, nonce, &types.EventEntry{})
	require.Error(t, err)
	_, err = types.CanonicalEventBody(types.EventType_EVENT_TYPE_OUTCOME, "org-a", memory, 7, pubkey, nonce, &types.EventEntry{Body: &types.EventEntry_Convergence{Convergence: &types.ConvergenceEventBody{}}})
	require.Error(t, err)
	_, err = types.CanonicalEventBody(types.EventType_EVENT_TYPE_VALIDITY_PREDICATE, "org-a", memory, 7, pubkey, nonce, &types.EventEntry{Body: &types.EventEntry_ValidityPredicate{ValidityPredicate: &types.ValidityPredicateEventBody{}}})
	require.Error(t, err)
}

func TestComputeEventFingerprint(t *testing.T) {
	body := []byte("event-body")
	expected := sha256.Sum256(body)
	require.Equal(t, expected[:], types.ComputeEventFingerprint(body))
}

func TestOutcomeEventPreimageIncludesServeRefBeforeNonce(t *testing.T) {
	body, err := types.CanonicalEventBody(types.EventType_EVENT_TYPE_OUTCOME, "org-a", hash32(0x01), 7, hash32(0x02), []byte{0x03, 0x04}, &types.EventEntry{Body: &types.EventEntry_Outcome{Outcome: &types.OutcomeEventBody{
		EpisodeRef: []byte{0x10, 0x11}, Resolution: types.OutcomeResolution_OUTCOME_RESOLUTION_WORKED, Source: types.OutcomeSource_OUTCOME_SOURCE_HARVESTED, EvidenceRef: []byte{0x12}, ServeRef: fingerprint32(0x13),
	}}})
	require.NoError(t, err)
	require.Equal(t, 11, strings.Count(string(body), "\n"))
	require.Equal(t, "wevibe-event-v1\noutcome\norg-a\n"+hexRepeat(0x01, 32)+"\n7\n"+hexRepeat(0x02, 32)+"\n1011\n12\n"+hexRepeat(0x13, 32)+"\nresolution=worked\nsource=harvested\n0304", string(body))
}

func TestOutcomeServeRefValidationRequiresSha256Size(t *testing.T) {
	for _, size := range []int{0, 31, 33} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			msg := &types.MsgSubmitEventBatch{Signer: "s", OrgId: "org", Epoch: 1, Events: []*types.EventEntry{validOutcomeEvent()}}
			msg.Events[0].GetOutcome().ServeRef = make([]byte, size)
			require.EqualError(t, msg.ValidateBasic(), "outcome serve_ref must be exactly 32 bytes")
		})
	}
}

func TestComputeServeFingerprintBindsServePubkey(t *testing.T) {
	memory := hash32(0x01)
	epoch := uint64(7)
	serveRefA := types.ComputeServeFingerprint(memory, hash32(0x02), epoch)
	serveRefB := types.ComputeServeFingerprint(memory, hash32(0x03), epoch)
	require.NotEqual(t, serveRefA, serveRefB)
}

func TestMsgSubmitDenialBatch_ValidateBasic(t *testing.T) {
	valid := func() *types.MsgSubmitDenialBatch {
		return &types.MsgSubmitDenialBatch{
			Signer: "s", OrgId: "org", Epoch: 1,
			Entries: []*types.DenialEntry{{
				MemoryHash:       hash32(0x02),
				ServeKeyPubkey:   hash32(0x03),
				ServeSig:         sig64(0x04),
				ServeFingerprint: fingerprint32(0x05),
			}},
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
		{"short serve pubkey", func(m *types.MsgSubmitDenialBatch) { m.Entries[0].ServeKeyPubkey = []byte{1} }},
		{"short serve signature", func(m *types.MsgSubmitDenialBatch) { m.Entries[0].ServeSig = []byte{1} }},
		{"short serve fingerprint", func(m *types.MsgSubmitDenialBatch) { m.Entries[0].ServeFingerprint = []byte{1} }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := valid()
			tc.mutate(m)
			require.Error(t, m.ValidateBasic())
		})
	}
}

func TestMsgSubmitDenialBatch_NegAnchorInert(t *testing.T) {
	msg := &types.MsgSubmitDenialBatch{
		Signer: "s", OrgId: "org", Epoch: 1,
		Entries: []*types.DenialEntry{{
			MemoryHash:       hash32(0x02),
			ServeKeyPubkey:   hash32(0x03),
			ServeSig:         sig64(0x04),
			ServeFingerprint: fingerprint32(0x05),
			NegAnchor:        []byte{0x01},
		}},
	}
	require.Error(t, msg.ValidateBasic())
}

func TestMsgSubmitEventBatch_ValidateBasic(t *testing.T) {
	validMsg := func() *types.MsgSubmitEventBatch {
		return &types.MsgSubmitEventBatch{
			Signer: "s", OrgId: "org", Epoch: 1,
			Events: []*types.EventEntry{validOutcomeEvent()},
		}
	}
	require.NoError(t, validMsg().ValidateBasic())

	cases := []struct {
		name   string
		mutate func(*types.MsgSubmitEventBatch)
	}{
		{"parked contest", func(m *types.MsgSubmitEventBatch) { m.Events[0].EventType = types.EventType_EVENT_TYPE_CONTEST }},
		{"parked sponsorship", func(m *types.MsgSubmitEventBatch) { m.Events[0].EventType = types.EventType_EVENT_TYPE_SPONSORSHIP }},
		{"serve receipt path", func(m *types.MsgSubmitEventBatch) { m.Events[0].EventType = types.EventType_EVENT_TYPE_SERVE }},
		{"block receipt path", func(m *types.MsgSubmitEventBatch) { m.Events[0].EventType = types.EventType_EVENT_TYPE_BLOCK }},
		{"oversize episode ref", func(m *types.MsgSubmitEventBatch) { m.Events[0].GetOutcome().EpisodeRef = make([]byte, 65) }},
		{"bad hash len", func(m *types.MsgSubmitEventBatch) { m.Events[0].MemoryContentHash = []byte{1} }},
		{"bad pubkey len", func(m *types.MsgSubmitEventBatch) { m.Events[0].SignerPubkey = []byte{1} }},
		{"bad sig len", func(m *types.MsgSubmitEventBatch) { m.Events[0].Signature = []byte{1} }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := validMsg()
			tc.mutate(msg)
			require.Error(t, msg.ValidateBasic())
		})
	}
}

func TestMsgAnchorPolicyVersion_ValidateBasic(t *testing.T) {
	valid := &types.MsgAnchorPolicyVersion{Authority: "gov", PolicyVersion: "v1", PolicyHash: hash32(0x01)}
	require.NoError(t, valid.ValidateBasic())
	require.Error(t, (&types.MsgAnchorPolicyVersion{PolicyVersion: "v1", PolicyHash: hash32(0x01)}).ValidateBasic())
	require.Error(t, (&types.MsgAnchorPolicyVersion{Authority: "gov", PolicyHash: hash32(0x01)}).ValidateBasic())
	require.Error(t, (&types.MsgAnchorPolicyVersion{Authority: "gov", PolicyVersion: strings.Repeat("a", 129), PolicyHash: hash32(0x01)}).ValidateBasic())
	require.Error(t, (&types.MsgAnchorPolicyVersion{Authority: "gov", PolicyVersion: "v1", PolicyHash: []byte{1}}).ValidateBasic())
}

func TestMsgUpdateParams_ValidateBasic(t *testing.T) {
	require.NoError(t, (&types.MsgUpdateParams{Authority: "gov"}).ValidateBasic())
	require.Error(t, (&types.MsgUpdateParams{Authority: ""}).ValidateBasic())
}

func TestServeReceipt_Validate(t *testing.T) {
	valid := types.NewServeReceipt("org", hash32(0x03), "sk", hash32(0x30), "c", 1, fingerprint32(0x03), false, "model", 2, []byte{0xaa})
	require.NoError(t, valid.Validate())

	require.Error(t, (&types.ServeReceipt{}).Validate())                                   // empty org
	require.Error(t, (&types.ServeReceipt{OrgID: "o", ContentHash: []byte{1}}).Validate()) // bad hash
	require.Error(t, (&types.ServeReceipt{OrgID: "o", ContentHash: hash32(1)}).Validate()) // empty serve key
	require.Error(t, (&types.ServeReceipt{OrgID: "o", ContentHash: hash32(1), ServeKey: "s", ServeKeyPubkey: hash32(0x31)}).Validate())
	require.Error(t, (&types.ServeReceipt{OrgID: "o", ContentHash: hash32(1), ServeKey: "s", ServeKeyPubkey: []byte{1}, ContributorID: "c", Fingerprint: fingerprint32(0x32)}).Validate())
	require.Error(t, (&types.ServeReceipt{OrgID: "o", ContentHash: hash32(1), ServeKey: "s", ServeKeyPubkey: hash32(0x33), ContributorID: "c", Fingerprint: []byte{1}}).Validate())
}

func TestContributorEpochServes_AddOrgID_Dedup(t *testing.T) {
	cs := types.NewContributorEpochServes("c", 1)
	cs.AddOrgID("org-a")
	cs.AddOrgID("org-a") // duplicate ignored
	cs.AddOrgID("org-b")
	require.ElementsMatch(t, []string{"org-a", "org-b"}, cs.OrgIDs)
}

func TestServeReceipt_StoredRoundtrip(t *testing.T) {
	sr := types.NewServeReceipt("org", hash32(0x04), "sk", hash32(0x41), "c", 7, fingerprint32(0x04), true, "m", 3, []byte{0xbb})
	stored := types.ServeReceiptToStored(sr)
	back := types.StoredToServeReceipt(*stored)
	require.Equal(t, sr.OrgID, back.OrgID)
	require.Equal(t, sr.ContentHash, back.ContentHash)
	require.Equal(t, sr.IsSelfServe, back.IsSelfServe)
	require.Equal(t, sr.ServeKeyPubkey, back.ServeKeyPubkey)
	require.Equal(t, sr.Fingerprint, back.Fingerprint)
}

func TestEpochServeStats_StoredRoundtrip(t *testing.T) {
	es := types.NewEpochServeStats("org", 2)
	es.TotalServes = 5
	es.TotalDenials = 3
	es.SelfServes = 2
	es.ModelBreakdown["m"] = 3
	back := types.StoredToEpochServeStats(*types.EpochServeStatsToStored(es))
	require.Equal(t, uint64(5), back.TotalServes)
	require.Equal(t, uint64(3), back.TotalDenials)
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
		[]*types.ServeReceipt{types.NewServeReceipt("org", hash32(5), "sk", hash32(0x51), "c", 1, fingerprint32(5), false, "m", 1, []byte{0xcc})},
		nil, nil, nil, nil, nil,
	)
	bz, err := gs.MarshalJSON()
	require.NoError(t, err)
	var decoded types.GenesisState
	require.NoError(t, decoded.UnmarshalJSON(bz))
	require.Len(t, decoded.ServeReceipts, 1)
	require.Equal(t, "org", decoded.ServeReceipts[0].OrgID)
}

func TestOutcomeResolutionSourceValidation(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(m *types.MsgSubmitEventBatch)
		wantErr string
	}{
		{"resolution unspecified", func(m *types.MsgSubmitEventBatch) {
			m.Events[0].GetOutcome().Resolution = types.OutcomeResolution_OUTCOME_RESOLUTION_UNSPECIFIED
		}, "outcome resolution must be worked, didnt_work, or unobserved"},
		{"source unspecified", func(m *types.MsgSubmitEventBatch) {
			m.Events[0].GetOutcome().Source = types.OutcomeSource_OUTCOME_SOURCE_UNSPECIFIED
		}, "outcome source must be harvested or user"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := &types.MsgSubmitEventBatch{Signer: "s", OrgId: "org", Epoch: 1, Events: []*types.EventEntry{validOutcomeEvent()}}
			tc.mutate(msg)
			require.EqualError(t, msg.ValidateBasic(), tc.wantErr)
		})
	}
}

func TestOutcomeUnobservedPreimageToken(t *testing.T) {
	body, err := types.CanonicalEventBody(types.EventType_EVENT_TYPE_OUTCOME, "org-a", hash32(0x01), 7, hash32(0x02), []byte{0x03, 0x04}, &types.EventEntry{Body: &types.EventEntry_Outcome{Outcome: &types.OutcomeEventBody{
		EpisodeRef: []byte{0x10, 0x11}, Resolution: types.OutcomeResolution_OUTCOME_RESOLUTION_UNOBSERVED, Source: types.OutcomeSource_OUTCOME_SOURCE_HARVESTED, EvidenceRef: []byte{0x12}, ServeRef: fingerprint32(0x13),
	}}})
	require.NoError(t, err)
	require.Equal(t, "wevibe-event-v1\noutcome\norg-a\n"+hexRepeat(0x01, 32)+"\n7\n"+hexRepeat(0x02, 32)+"\n1011\n12\n"+hexRepeat(0x13, 32)+"\nresolution=unobserved\nsource=harvested\n0304", string(body))
}
