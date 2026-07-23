package types_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wevibe-network/wevibe-chain/x/memory/types"
)

// validHash returns a 32-byte content hash for use in valid test cases.
func validHash() []byte {
	return bytes.Repeat([]byte{0xAB}, types.ContentHashLen)
}

// -----------------------------------------------------------------------------
// MsgSubmitCommitment.ValidateBasic
// -----------------------------------------------------------------------------

func TestMsgSubmitCommitment_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *types.MsgSubmitCommitment
		wantErr error
	}{
		{
			name: "valid legacy unattested",
			msg: &types.MsgSubmitCommitment{
				Signer:        "signer",
				OrgId:         "org1",
				ContentHash:   validHash(),
				ContributorId: "contributor",
			},
			wantErr: nil,
		},
		{
			name: "valid self declared provenance",
			msg: &types.MsgSubmitCommitment{
				Signer:          "signer",
				OrgId:           "org1",
				ContentHash:     validHash(),
				ContributorId:   "contributor",
				ProducerModelId: "openai/gpt-5.2",
			},
			wantErr: nil,
		},
		{
			name: "valid canonical provenance",
			msg: &types.MsgSubmitCommitment{
				Signer:                 "signer",
				OrgId:                  "org1",
				ContentHash:            validHash(),
				ContributorId:          "contributor",
				ProducerModelId:        "openai/gpt-5.2",
				AttestationSessionHash: bytes.Repeat([]byte{0x01}, types.ContentHashLen),
			},
			wantErr: nil,
		},
		{
			name: "empty signer",
			msg: &types.MsgSubmitCommitment{
				Signer:        "",
				OrgId:         "org1",
				ContentHash:   validHash(),
				ContributorId: "contributor",
			},
			wantErr: errSentinel, // non-sentinel fmt error, asserted as "any error"
		},
		{
			name: "empty org",
			msg: &types.MsgSubmitCommitment{
				Signer:        "signer",
				OrgId:         "",
				ContentHash:   validHash(),
				ContributorId: "contributor",
			},
			wantErr: types.ErrInvalidOrgID,
		},
		{
			name: "content hash too short",
			msg: &types.MsgSubmitCommitment{
				Signer:        "signer",
				OrgId:         "org1",
				ContentHash:   bytes.Repeat([]byte{0x01}, types.ContentHashLen-1),
				ContributorId: "contributor",
			},
			wantErr: types.ErrInvalidContentHash,
		},
		{
			name: "content hash too long",
			msg: &types.MsgSubmitCommitment{
				Signer:        "signer",
				OrgId:         "org1",
				ContentHash:   bytes.Repeat([]byte{0x01}, types.ContentHashLen+1),
				ContributorId: "contributor",
			},
			wantErr: types.ErrInvalidContentHash,
		},
		{
			name: "content hash empty",
			msg: &types.MsgSubmitCommitment{
				Signer:        "signer",
				OrgId:         "org1",
				ContentHash:   []byte{},
				ContributorId: "contributor",
			},
			wantErr: types.ErrInvalidContentHash,
		},
		{
			name: "empty contributor",
			msg: &types.MsgSubmitCommitment{
				Signer:        "signer",
				OrgId:         "org1",
				ContentHash:   validHash(),
				ContributorId: "",
			},
			wantErr: types.ErrInvalidContributor,
		},
		{
			name: "attestation session hash wrong length",
			msg: &types.MsgSubmitCommitment{
				Signer:                 "signer",
				OrgId:                  "org1",
				ContentHash:            validHash(),
				ContributorId:          "contributor",
				ProducerModelId:        "openai/gpt-5.2",
				AttestationSessionHash: bytes.Repeat([]byte{0x02}, 16),
			},
			wantErr: errSentinel,
		},
		{
			name: "attestation session hash set without producer model id",
			msg: &types.MsgSubmitCommitment{
				Signer:                 "signer",
				OrgId:                  "org1",
				ContentHash:            validHash(),
				ContributorId:          "contributor",
				AttestationSessionHash: bytes.Repeat([]byte{0x03}, types.ContentHashLen),
			},
			wantErr: errSentinel,
		},
		{
			name: "producer model id too long",
			msg: &types.MsgSubmitCommitment{
				Signer:          "signer",
				OrgId:           "org1",
				ContentHash:     validHash(),
				ContributorId:   "contributor",
				ProducerModelId: strings.Repeat("a", types.MaxProducerModelIdLen+1),
			},
			wantErr: errSentinel,
		},
		{
			name: "producer model id whitespace only",
			msg: &types.MsgSubmitCommitment{
				Signer:          "signer",
				OrgId:           "org1",
				ContentHash:     validHash(),
				ContributorId:   "contributor",
				ProducerModelId: "   \t",
			},
			wantErr: errSentinel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			assertErr(t, err, tt.wantErr)
		})
	}
}

// -----------------------------------------------------------------------------
// MsgApproveMemory.ValidateBasic
// -----------------------------------------------------------------------------

func TestMsgApproveMemory_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *types.MsgApproveMemory
		wantErr error
	}{
		{
			name: "valid",
			msg: &types.MsgApproveMemory{
				Signer:        "signer",
				OrgId:         "org1",
				ContentHash:   validHash(),
				EncryptedBlob: []byte("blob"),
			},
			wantErr: nil,
		},
		{
			name: "empty signer",
			msg: &types.MsgApproveMemory{
				Signer:        "",
				OrgId:         "org1",
				ContentHash:   validHash(),
				EncryptedBlob: []byte("blob"),
			},
			wantErr: errSentinel,
		},
		{
			name: "empty org",
			msg: &types.MsgApproveMemory{
				Signer:        "signer",
				OrgId:         "",
				ContentHash:   validHash(),
				EncryptedBlob: []byte("blob"),
			},
			wantErr: types.ErrInvalidOrgID,
		},
		{
			name: "content hash wrong length",
			msg: &types.MsgApproveMemory{
				Signer:        "signer",
				OrgId:         "org1",
				ContentHash:   []byte("short"),
				EncryptedBlob: []byte("blob"),
			},
			wantErr: types.ErrInvalidContentHash,
		},
		{
			name: "empty blob",
			msg: &types.MsgApproveMemory{
				Signer:        "signer",
				OrgId:         "org1",
				ContentHash:   validHash(),
				EncryptedBlob: []byte{},
			},
			wantErr: types.ErrInvalidBlob,
		},
		{
			name: "nil blob",
			msg: &types.MsgApproveMemory{
				Signer:        "signer",
				OrgId:         "org1",
				ContentHash:   validHash(),
				EncryptedBlob: nil,
			},
			wantErr: types.ErrInvalidBlob,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			assertErr(t, err, tt.wantErr)
		})
	}
}

// -----------------------------------------------------------------------------
// MsgUpdateParams.ValidateBasic
// -----------------------------------------------------------------------------

func TestMsgUpdateParams_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *types.MsgUpdateParams
		wantErr bool
	}{
		{
			name: "valid",
			msg: func() *types.MsgUpdateParams {
				p := types.DefaultParams()
				return &types.MsgUpdateParams{Authority: "gov", Params: &p}
			}(),
			wantErr: false,
		},
		{
			name: "empty authority",
			msg: func() *types.MsgUpdateParams {
				p := types.DefaultParams()
				return &types.MsgUpdateParams{Authority: "", Params: &p}
			}(),
			wantErr: true,
		},
		{
			name: "invalid params (grace epochs zero)",
			msg: func() *types.MsgUpdateParams {
				p := types.DefaultParams()
				p.GraceEpochs = 0
				return &types.MsgUpdateParams{Authority: "gov", Params: &p}
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// MsgReportMemory.ValidateBasic
// -----------------------------------------------------------------------------

func TestMsgReportMemory_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *types.MsgReportMemory
		wantErr error
	}{
		{
			name: "valid",
			msg: &types.MsgReportMemory{
				Signer:         "signer",
				OrgId:          "org1",
				ContentHash:    validHash(),
				ReporterPubkey: "reporter",
				Reason:         "spam",
			},
			wantErr: nil,
		},
		{
			name: "empty signer",
			msg: &types.MsgReportMemory{
				Signer:         "",
				OrgId:          "org1",
				ContentHash:    validHash(),
				ReporterPubkey: "reporter",
				Reason:         "spam",
			},
			wantErr: errSentinel,
		},
		{
			name: "empty org",
			msg: &types.MsgReportMemory{
				Signer:         "signer",
				OrgId:          "",
				ContentHash:    validHash(),
				ReporterPubkey: "reporter",
				Reason:         "spam",
			},
			wantErr: types.ErrInvalidOrgID,
		},
		{
			name: "content hash wrong length",
			msg: &types.MsgReportMemory{
				Signer:         "signer",
				OrgId:          "org1",
				ContentHash:    []byte("bad"),
				ReporterPubkey: "reporter",
				Reason:         "spam",
			},
			wantErr: types.ErrInvalidContentHash,
		},
		{
			name: "empty reporter pubkey",
			msg: &types.MsgReportMemory{
				Signer:         "signer",
				OrgId:          "org1",
				ContentHash:    validHash(),
				ReporterPubkey: "",
				Reason:         "spam",
			},
			wantErr: types.ErrInvalidReporter,
		},
		{
			name: "empty reason",
			msg: &types.MsgReportMemory{
				Signer:         "signer",
				OrgId:          "org1",
				ContentHash:    validHash(),
				ReporterPubkey: "reporter",
				Reason:         "",
			},
			wantErr: types.ErrInvalidReportReason,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			assertErr(t, err, tt.wantErr)
		})
	}
}

// -----------------------------------------------------------------------------
// PendingCommitment.Validate
// -----------------------------------------------------------------------------

func TestPendingCommitment_Validate(t *testing.T) {
	validKeywords := []*types.KeywordWeight{{Keyword: "kw", Weight: "1.0"}}

	tests := []struct {
		name    string
		pc      *types.PendingCommitment
		wantErr error
	}{
		{
			name: "valid",
			pc: &types.PendingCommitment{
				OrgID:       "org1",
				ContentHash: validHash(),
				Keywords:    validKeywords,
				Contributor: "contributor",
				MemoryType:  types.MemoryType_MEMORY_TYPE_MEMORY,
			},
			wantErr: nil,
		},
		{
			name: "valid with no keywords",
			pc: &types.PendingCommitment{
				OrgID:       "org1",
				ContentHash: validHash(),
				Keywords:    nil,
				Contributor: "contributor",
				MemoryType:  types.MemoryType_MEMORY_TYPE_MEMORY,
			},
			wantErr: nil,
		},
		{
			name: "empty org",
			pc: &types.PendingCommitment{
				OrgID:       "",
				ContentHash: validHash(),
				Contributor: "contributor",
				MemoryType:  types.MemoryType_MEMORY_TYPE_MEMORY,
			},
			wantErr: types.ErrInvalidOrgID,
		},
		{
			name: "bad content hash",
			pc: &types.PendingCommitment{
				OrgID:       "org1",
				ContentHash: []byte("short"),
				Contributor: "contributor",
				MemoryType:  types.MemoryType_MEMORY_TYPE_MEMORY,
			},
			wantErr: types.ErrInvalidContentHash,
		},
		{
			name: "empty contributor",
			pc: &types.PendingCommitment{
				OrgID:       "org1",
				ContentHash: validHash(),
				Contributor: "",
				MemoryType:  types.MemoryType_MEMORY_TYPE_MEMORY,
			},
			wantErr: types.ErrInvalidContributor,
		},
		{
			name: "invalid memory type unspecified",
			pc: &types.PendingCommitment{
				OrgID:       "org1",
				ContentHash: validHash(),
				Contributor: "contributor",
				MemoryType:  types.MemoryType_MEMORY_TYPE_UNSPECIFIED,
			},
			wantErr: types.ErrInvalidMemoryType,
		},
		{
			name: "nil keyword",
			pc: &types.PendingCommitment{
				OrgID:       "org1",
				ContentHash: validHash(),
				Contributor: "contributor",
				MemoryType:  types.MemoryType_MEMORY_TYPE_MEMORY,
				Keywords:    []*types.KeywordWeight{nil},
			},
			wantErr: types.ErrInvalidKeyword,
		},
		{
			name: "empty keyword string",
			pc: &types.PendingCommitment{
				OrgID:       "org1",
				ContentHash: validHash(),
				Contributor: "contributor",
				MemoryType:  types.MemoryType_MEMORY_TYPE_MEMORY,
				Keywords:    []*types.KeywordWeight{{Keyword: ""}},
			},
			wantErr: types.ErrInvalidKeyword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pc.Validate()
			assertErr(t, err, tt.wantErr)
		})
	}
}

// -----------------------------------------------------------------------------
// MemoryCommitment.Validate
// -----------------------------------------------------------------------------

func TestMemoryCommitment_Validate(t *testing.T) {
	validKeywords := []*types.KeywordWeight{{Keyword: "kw", Weight: "1.0"}}

	tests := []struct {
		name    string
		mc      *types.MemoryCommitment
		wantErr error
	}{
		{
			name: "valid",
			mc: &types.MemoryCommitment{
				OrgID:         "org1",
				ContentHash:   validHash(),
				EncryptedBlob: []byte("blob"),
				Keywords:      validKeywords,
				Contributor:   "contributor",
				MemoryType:    types.MemoryType_MEMORY_TYPE_MEMORY,
			},
			wantErr: nil,
		},
		{
			name: "empty org",
			mc: &types.MemoryCommitment{
				OrgID:         "",
				ContentHash:   validHash(),
				EncryptedBlob: []byte("blob"),
				Contributor:   "contributor",
				MemoryType:    types.MemoryType_MEMORY_TYPE_MEMORY,
			},
			wantErr: types.ErrInvalidOrgID,
		},
		{
			name: "bad content hash",
			mc: &types.MemoryCommitment{
				OrgID:         "org1",
				ContentHash:   bytes.Repeat([]byte{0x01}, 16),
				EncryptedBlob: []byte("blob"),
				Contributor:   "contributor",
				MemoryType:    types.MemoryType_MEMORY_TYPE_MEMORY,
			},
			wantErr: types.ErrInvalidContentHash,
		},
		{
			name: "empty blob",
			mc: &types.MemoryCommitment{
				OrgID:         "org1",
				ContentHash:   validHash(),
				EncryptedBlob: nil,
				Contributor:   "contributor",
				MemoryType:    types.MemoryType_MEMORY_TYPE_MEMORY,
			},
			wantErr: types.ErrInvalidBlob,
		},
		{
			name: "empty contributor",
			mc: &types.MemoryCommitment{
				OrgID:         "org1",
				ContentHash:   validHash(),
				EncryptedBlob: []byte("blob"),
				Contributor:   "",
				MemoryType:    types.MemoryType_MEMORY_TYPE_MEMORY,
			},
			wantErr: types.ErrInvalidContributor,
		},
		{
			name: "invalid memory type",
			mc: &types.MemoryCommitment{
				OrgID:         "org1",
				ContentHash:   validHash(),
				EncryptedBlob: []byte("blob"),
				Contributor:   "contributor",
				MemoryType:    types.MemoryType_MEMORY_TYPE_UNSPECIFIED,
			},
			wantErr: types.ErrInvalidMemoryType,
		},
		{
			name: "nil keyword entry",
			mc: &types.MemoryCommitment{
				OrgID:         "org1",
				ContentHash:   validHash(),
				EncryptedBlob: []byte("blob"),
				Contributor:   "contributor",
				MemoryType:    types.MemoryType_MEMORY_TYPE_MEMORY,
				Keywords:      []*types.KeywordWeight{nil},
			},
			wantErr: types.ErrInvalidKeyword,
		},
		{
			name: "empty keyword string",
			mc: &types.MemoryCommitment{
				OrgID:         "org1",
				ContentHash:   validHash(),
				EncryptedBlob: []byte("blob"),
				Contributor:   "contributor",
				MemoryType:    types.MemoryType_MEMORY_TYPE_MEMORY,
				Keywords:      []*types.KeywordWeight{{Keyword: ""}},
			},
			wantErr: types.ErrInvalidKeyword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mc.Validate()
			assertErr(t, err, tt.wantErr)
		})
	}
}

// -----------------------------------------------------------------------------
// Params.Validate and DefaultParams
// -----------------------------------------------------------------------------

func TestDefaultParams_Valid(t *testing.T) {
	p := types.DefaultParams()
	require.NoError(t, p.Validate())

	// Spot-check that documented defaults are wired through.
	require.Equal(t, uint64(1000), p.MaxPendingPerOrg)
	require.Equal(t, uint64(7), p.PendingRetentionEpochs)
	require.Equal(t, uint64(1048576), p.MaxBlobSizeBytes)
	require.Equal(t, uint32(20), p.MaxKeywordsPerMemory)
	require.Equal(t, types.DefaultRetrievalThresholdBps, p.RetrievalThresholdBps)
	require.Equal(t, types.DefaultGraceEpochs, p.GraceEpochs)
	require.Equal(t, types.DefaultIdleTrafficRefBps, p.IdleTrafficRefBpsPerMem)
	require.Equal(t, types.DefaultIdleTrafficFloorBps, p.IdleTrafficFloorBps)
	require.Equal(t, types.DefaultTrustMinServes, p.TrustMinServes)
	require.Equal(t, types.DefaultTrustMaxRateBps, p.TrustMaxRateBps)
}

func TestParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(p *types.Params)
		wantErr bool
	}{
		{
			name:    "default valid",
			mutate:  func(p *types.Params) {},
			wantErr: false,
		},
		{
			name:    "grace epochs zero",
			mutate:  func(p *types.Params) { p.GraceEpochs = 0 },
			wantErr: true,
		},
		{
			name:    "grace epochs one is boundary ok",
			mutate:  func(p *types.Params) { p.GraceEpochs = 1 },
			wantErr: false,
		},
		{
			name:    "trust min serves zero",
			mutate:  func(p *types.Params) { p.TrustMinServes = 0 },
			wantErr: true,
		},
		{
			name:    "trust min serves one is boundary ok",
			mutate:  func(p *types.Params) { p.TrustMinServes = 1 },
			wantErr: false,
		},
		{
			name:    "retrieval threshold over 10000",
			mutate:  func(p *types.Params) { p.RetrievalThresholdBps = 10001 },
			wantErr: true,
		},
		{
			name:    "retrieval threshold at 10000 boundary ok",
			mutate:  func(p *types.Params) { p.RetrievalThresholdBps = 10000 },
			wantErr: false,
		},
		{
			name:    "initial confidence over 10000",
			mutate:  func(p *types.Params) { p.InitialConfidenceBps = 10001 },
			wantErr: true,
		},
		{
			name:    "serve d bps over 10000",
			mutate:  func(p *types.Params) { p.ServeDBps = 10001 },
			wantErr: true,
		},
		{
			name:    "denial d bps over 10000",
			mutate:  func(p *types.Params) { p.DenialDBps = 10001 },
			wantErr: true,
		},
		{
			name:    "idle d bps over 10000",
			mutate:  func(p *types.Params) { p.IdleDBps = 10001 },
			wantErr: true,
		},
		{
			name:    "serve floor bps over 10000",
			mutate:  func(p *types.Params) { p.ServeFloorBps = 10001 },
			wantErr: true,
		},
		{
			name:    "denial floor bps over 10000",
			mutate:  func(p *types.Params) { p.DenialFloorBps = 10001 },
			wantErr: true,
		},
		{
			name:    "idle protect bps over 10000",
			mutate:  func(p *types.Params) { p.IdleProtectBps = 10001 },
			wantErr: true,
		},
		{
			name:    "idle untrusted bps over 10000",
			mutate:  func(p *types.Params) { p.IdleUntrustedBps = 10001 },
			wantErr: true,
		},
		{
			name:    "idle traffic ref bps over 10000",
			mutate:  func(p *types.Params) { p.IdleTrafficRefBpsPerMem = 10001 },
			wantErr: true,
		},
		{
			name:    "idle traffic floor bps over 10000",
			mutate:  func(p *types.Params) { p.IdleTrafficFloorBps = 10001 },
			wantErr: true,
		},
		{
			name:    "idle traffic ref bps zero",
			mutate:  func(p *types.Params) { p.IdleTrafficRefBpsPerMem = 0 },
			wantErr: true,
		},
		{
			name:    "trust max rate bps over 10000",
			mutate:  func(p *types.Params) { p.TrustMaxRateBps = 10001 },
			wantErr: true,
		},
		{
			name: "all bps at zero with min epochs ok",
			mutate: func(p *types.Params) {
				p.RetrievalThresholdBps = 0
				p.InitialConfidenceBps = 0
				p.ServeDBps = 0
				p.DenialDBps = 0
				p.IdleDBps = 0
				p.ServeFloorBps = 0
				p.DenialFloorBps = 0
				p.IdleProtectBps = 0
				p.IdleUntrustedBps = 0
				p.IdleTrafficRefBpsPerMem = 1
				p.IdleTrafficFloorBps = 0
				p.TrustMaxRateBps = 0
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := types.DefaultParams()
			tt.mutate(&p)
			err := p.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// memory_types.go pure logic
// -----------------------------------------------------------------------------

func TestValidMemoryType(t *testing.T) {
	require.True(t, types.ValidMemoryType(types.MemoryType_MEMORY_TYPE_MEMORY))
	require.False(t, types.ValidMemoryType(types.MemoryType_MEMORY_TYPE_UNSPECIFIED))
	require.False(t, types.ValidMemoryType(types.MemoryType(99)))
}

func TestRelationAliases(t *testing.T) {
	require.Equal(t, types.RelationType_RELATION_TYPE_CONTRADICTS, types.RelationContradicts)
	require.Equal(t, types.RelationType_RELATION_TYPE_REPLACES, types.RelationReplaces)
	require.Equal(t, types.RelationType_RELATION_TYPE_DEPRECATES, types.RelationDeprecates)
	require.Equal(t, types.RelationType_RELATION_TYPE_SUPERSEDES, types.RelationSupersedes)
}

// -----------------------------------------------------------------------------
// keys.go helpers and constructors
// -----------------------------------------------------------------------------

func TestNewPendingCommitment(t *testing.T) {
	hash := validHash()
	kws := []*types.KeywordWeight{{Keyword: "kw", Weight: "1.0"}}
	pc := types.NewPendingCommitment("org1", hash, kws, "contributor", 3, 42, types.MemoryType_MEMORY_TYPE_MEMORY)

	require.Equal(t, "org1", pc.OrgID)
	require.Equal(t, hash, pc.ContentHash)
	require.Equal(t, kws, pc.Keywords)
	require.Equal(t, "contributor", pc.Contributor)
	require.Equal(t, uint64(3), pc.Epoch)
	require.Equal(t, uint64(42), pc.SubmittedAt)
	require.Equal(t, types.MemoryType_MEMORY_TYPE_MEMORY, pc.MemoryType)
}

func TestNewMemoryCommitment(t *testing.T) {
	hash := validHash()
	blob := []byte("blob")
	kws := []*types.KeywordWeight{{Keyword: "kw", Weight: "1.0"}}
	mc := types.NewMemoryCommitment(
		"org1", hash, blob, kws, "contributor", 5, 100, "leader",
		types.MemoryState_MEMORY_STATE_COMMITTED, 5,
		types.MemoryType_MEMORY_TYPE_MEMORY, 5,
	)

	require.Equal(t, "org1", mc.OrgID)
	require.Equal(t, hash, mc.ContentHash)
	require.Equal(t, blob, mc.EncryptedBlob)
	require.Equal(t, kws, mc.Keywords)
	require.Equal(t, "contributor", mc.Contributor)
	require.Equal(t, uint64(5), mc.Epoch)
	require.Equal(t, uint64(100), mc.CommittedAtHeight)
	require.Equal(t, "leader", mc.CommittingLeader)
	require.Equal(t, types.MemoryState_MEMORY_STATE_COMMITTED, mc.State)
	require.Equal(t, uint64(5), mc.LastActiveEpoch)
	require.Equal(t, types.MemoryType_MEMORY_TYPE_MEMORY, mc.MemoryType)
	require.Equal(t, uint64(5), mc.ApprovedAtEpoch)
}

func TestDeriveProvenanceStatus(t *testing.T) {
	modelID := "openai/gpt-5.2"
	sessionHash := bytes.Repeat([]byte{0x22}, types.ContentHashLen)

	require.Equal(t, types.ProvenanceUnattested, types.DeriveProvenanceStatus("", nil))
	require.Equal(t, types.ProvenanceSelfDeclared, types.DeriveProvenanceStatus(modelID, nil))
	require.Equal(t, types.ProvenanceSessionReferenced, types.DeriveProvenanceStatus(modelID, sessionHash))

	// Honest semantics: there is no "VERIFIED" status in this on-chain helper.
	require.NotEqual(t, types.ProvenanceStatus("VERIFIED"), types.ProvenanceUnattested)
	require.NotEqual(t, types.ProvenanceStatus("VERIFIED"), types.ProvenanceSelfDeclared)
	require.NotEqual(t, types.ProvenanceStatus("VERIFIED"), types.ProvenanceSessionReferenced)

	modelOnlyMemory := &types.MemoryCommitment{ProducerModelId: modelID}
	require.Equal(t, types.ProvenanceSelfDeclared, modelOnlyMemory.ProvenanceStatus())
	require.NotEqual(t, types.ProvenanceSessionReferenced, modelOnlyMemory.ProvenanceStatus())
}

func TestNewEpochMerkleRoot(t *testing.T) {
	root := bytes.Repeat([]byte{0x07}, 32)
	emr := types.NewEpochMerkleRoot("org1", 9, root, 4)
	require.Equal(t, "org1", emr.OrgID)
	require.Equal(t, uint64(9), emr.Epoch)
	require.Equal(t, root, emr.MerkleRoot)
	require.Equal(t, uint64(4), emr.MemoryCount)
}

func TestContentHashToHex(t *testing.T) {
	// empty input -> empty hex
	require.Equal(t, "", types.ContentHashToHex(nil))
	require.Equal(t, "", types.ContentHashToHex([]byte{}))

	hash := []byte{0x00, 0x01, 0xAB, 0xFF}
	require.Equal(t, "0001abff", types.ContentHashToHex(hash))

	full := validHash()
	got := types.ContentHashToHex(full)
	require.Equal(t, hex.EncodeToString(full), got)
	require.Len(t, got, types.ContentHashLen*2)
}

func TestComputeMerkleRoot(t *testing.T) {
	// empty -> nil
	require.Nil(t, types.ComputeMerkleRoot(nil))
	require.Nil(t, types.ComputeMerkleRoot([][]byte{}))

	// single hash -> sha256 of that hash
	h1 := bytes.Repeat([]byte{0x01}, 32)
	expectedSingle := sha256.Sum256(h1)
	require.Equal(t, expectedSingle[:], types.ComputeMerkleRoot([][]byte{append([]byte(nil), h1...)}))

	// output always 32 bytes for non-empty input
	root := types.ComputeMerkleRoot([][]byte{
		bytes.Repeat([]byte{0x02}, 32),
		bytes.Repeat([]byte{0x03}, 32),
	})
	require.Len(t, root, 32)
}

func TestComputeMerkleRoot_OrderIndependent(t *testing.T) {
	a := bytes.Repeat([]byte{0x0A}, 32)
	b := bytes.Repeat([]byte{0x0B}, 32)

	root1 := types.ComputeMerkleRoot([][]byte{append([]byte(nil), a...), append([]byte(nil), b...)})
	root2 := types.ComputeMerkleRoot([][]byte{append([]byte(nil), b...), append([]byte(nil), a...)})

	require.Equal(t, root1, root2, "merkle root must be order-independent because inputs are sorted")
}

// -----------------------------------------------------------------------------
// genesis.go helpers
// -----------------------------------------------------------------------------

func TestNewGenesisState(t *testing.T) {
	pending := []*types.PendingCommitment{{OrgID: "org1", ContentHash: validHash()}}
	commitments := []*types.MemoryCommitment{{OrgID: "org1", ContentHash: validHash()}}
	relationships := []*types.MemoryRelationship{{OrgID: "org1", SourceCID: "a", TargetCID: "b"}}
	validity := []*types.StoredValidityMetadata{{OrgId: "org1", MemoryCid: "a"}}
	merkle := []*types.EpochMerkleRoot{{OrgID: "org1", Epoch: 1}}
	params := types.DefaultParams()

	gs := types.NewGenesisState(pending, commitments, relationships, validity, merkle, params)

	require.Equal(t, pending, gs.PendingCommitments)
	require.Equal(t, commitments, gs.MemoryCommitments)
	require.Equal(t, relationships, gs.Relationships)
	require.Equal(t, validity, gs.ValidityMetadata)
	require.Equal(t, merkle, gs.MerkleRoots)
	require.Equal(t, params, gs.Params)
}

func TestNewGenesisState_Empty(t *testing.T) {
	gs := types.NewGenesisState(nil, nil, nil, nil, nil, types.DefaultParams())
	require.Empty(t, gs.PendingCommitments)
	require.Empty(t, gs.MemoryCommitments)
	require.Empty(t, gs.Relationships)
	require.Empty(t, gs.ValidityMetadata)
	require.Empty(t, gs.MerkleRoots)
	require.NoError(t, gs.Params.Validate())
}

func TestGenesisState_JSONRoundTrip(t *testing.T) {
	gs := types.NewGenesisState(
		[]*types.PendingCommitment{{OrgID: "org1", ContentHash: validHash(), Contributor: "c"}},
		[]*types.MemoryCommitment{{OrgID: "org1", ContentHash: validHash()}},
		nil, nil, nil,
		types.DefaultParams(),
	)

	bz, err := gs.MarshalJSON()
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	var decoded types.GenesisState
	require.NoError(t, decoded.UnmarshalJSON(bz))
	require.Len(t, decoded.PendingCommitments, 1)
	require.Equal(t, "org1", decoded.PendingCommitments[0].OrgID)
	require.Len(t, decoded.MemoryCommitments, 1)
	require.Equal(t, types.DefaultParams(), decoded.Params)
}

// -----------------------------------------------------------------------------
// helpers
// -----------------------------------------------------------------------------

// errSentinel is a marker used in table tests to indicate "expect a non-nil
// error that is NOT one of the typed sentinel errors" (e.g. a plain fmt error).
var errSentinel = sentinelMarker{}

type sentinelMarker struct{}

func (sentinelMarker) Error() string { return "expect-any-error" }

func assertErr(t *testing.T, got, want error) {
	t.Helper()
	if want == nil {
		require.NoError(t, got)
		return
	}
	if want == errSentinel {
		require.Error(t, got)
		return
	}
	require.ErrorIs(t, got, want)
}
