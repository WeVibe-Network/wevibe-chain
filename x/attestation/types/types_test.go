package types_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wevibe-network/wevibe-chain/x/attestation/types"
)

func validSubmitMsg() *types.MsgSubmitSessionAttestation {
	return &types.MsgSubmitSessionAttestation{
		Signer:        "signer",
		OrgId:         "org-1",
		SessionHash:   make([]byte, types.SessionHashLen),
		ModelId:       "qwen3:4b",
		TurnCount:     4,
		TokenCount:    800,
		ProviderType:  types.ProviderType_PROVIDER_TYPE_LOCAL,
		ContributorId: "contributor-1",
		Epoch:         1,
	}
}

func TestMsgSubmitSessionAttestation_ValidateBasic_Valid(t *testing.T) {
	require.NoError(t, validSubmitMsg().ValidateBasic())

	// Cloud provider type is also valid.
	m := validSubmitMsg()
	m.ProviderType = types.ProviderType_PROVIDER_TYPE_CLOUD
	require.NoError(t, m.ValidateBasic())
}

func TestMsgSubmitSessionAttestation_ValidateBasic_Failures(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*types.MsgSubmitSessionAttestation)
		wantErr error // nil => only assert that an error occurs
	}{
		{"empty signer", func(m *types.MsgSubmitSessionAttestation) { m.Signer = "" }, nil},
		{"empty org", func(m *types.MsgSubmitSessionAttestation) { m.OrgId = "" }, types.ErrInvalidOrgID},
		{"nil session hash", func(m *types.MsgSubmitSessionAttestation) { m.SessionHash = nil }, types.ErrInvalidSessionHash},
		{"short session hash", func(m *types.MsgSubmitSessionAttestation) { m.SessionHash = make([]byte, 31) }, types.ErrInvalidSessionHash},
		{"long session hash", func(m *types.MsgSubmitSessionAttestation) { m.SessionHash = make([]byte, 33) }, types.ErrInvalidSessionHash},
		{"empty session hash", func(m *types.MsgSubmitSessionAttestation) { m.SessionHash = []byte{} }, types.ErrInvalidSessionHash},
		{"empty model", func(m *types.MsgSubmitSessionAttestation) { m.ModelId = "" }, types.ErrInvalidModelID},
		{"empty contributor", func(m *types.MsgSubmitSessionAttestation) { m.ContributorId = "" }, types.ErrInvalidContributor},
		{"unspecified provider", func(m *types.MsgSubmitSessionAttestation) {
			m.ProviderType = types.ProviderType_PROVIDER_TYPE_UNSPECIFIED
		}, types.ErrInvalidProviderType},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := validSubmitMsg()
			tc.mutate(m)
			err := m.ValidateBasic()
			require.Error(t, err)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}

func TestMsgSubmitSessionAttestation_Route(t *testing.T) {
	require.Equal(t, types.RouterKey, validSubmitMsg().Route())
	require.Equal(t, types.ModuleName, validSubmitMsg().Route())
}

func TestMsgSubmitSessionAttestation_GetSigners_InvalidBech32(t *testing.T) {
	// A non-bech32 signer yields a nil/empty address; GetSigners must not panic.
	m := validSubmitMsg()
	signers := m.GetSigners()
	require.Len(t, signers, 1)
	require.True(t, signers[0].Empty())
}

func TestMsgUpdateParams_ValidateBasic_Valid(t *testing.T) {
	p := types.DefaultParams()
	m := &types.MsgUpdateParams{Authority: "gov-authority", Params: &p}
	require.NoError(t, m.ValidateBasic())
}

func TestMsgUpdateParams_ValidateBasic_EmptyAuthority(t *testing.T) {
	p := types.DefaultParams()
	m := &types.MsgUpdateParams{Authority: "", Params: &p}
	require.Error(t, m.ValidateBasic())
}

func TestMsgUpdateParams_Route(t *testing.T) {
	require.Equal(t, types.RouterKey, (&types.MsgUpdateParams{}).Route())
}

func TestMsgUpdateParams_GetSigners_InvalidBech32(t *testing.T) {
	m := &types.MsgUpdateParams{Authority: "gov-authority"}
	signers := m.GetSigners()
	require.Len(t, signers, 1)
	require.True(t, signers[0].Empty())
}

func TestDefaultParams(t *testing.T) {
	p := types.DefaultParams()
	require.Equal(t, uint64(10000), p.MaxAttestationsPerEpoch)
	require.False(t, p.RequireAttestationForServe)
	require.NoError(t, p.Validate())
}

func TestParamsValidate(t *testing.T) {
	// Validate currently accepts any params (including the zero value).
	p := types.Params{}
	require.NoError(t, p.Validate())

	custom := types.Params{MaxAttestationsPerEpoch: 5000, RequireAttestationForServe: true}
	require.NoError(t, custom.Validate())
}

func TestGenesisStateJSONRoundtrip(t *testing.T) {
	sessionHash := make([]byte, types.SessionHashLen)
	sessionHash[0] = 0xAB
	gs := &types.GenesisState{
		Attestations: []*types.StoredSessionAttestation{
			{
				OrgId:         "org-1",
				SessionHash:   sessionHash,
				ModelId:       "claude-sonnet-4-20250514",
				TurnCount:     10,
				TokenCount:    3000,
				ProviderType:  types.ProviderType_PROVIDER_TYPE_CLOUD,
				ContributorId: "contributor-2",
				Epoch:         5,
			},
		},
	}

	bz, err := gs.MarshalJSON()
	require.NoError(t, err)

	var out types.GenesisState
	require.NoError(t, out.UnmarshalJSON(bz))
	require.Len(t, out.Attestations, 1)
	require.Equal(t, "org-1", out.Attestations[0].OrgId)
	require.Equal(t, "claude-sonnet-4-20250514", out.Attestations[0].ModelId)
	require.Equal(t, types.ProviderType_PROVIDER_TYPE_CLOUD, out.Attestations[0].ProviderType)
	require.Equal(t, sessionHash, out.Attestations[0].SessionHash)
}

func TestGenesisStateEmptyRoundtrip(t *testing.T) {
	gs := &types.GenesisState{}
	bz, err := gs.MarshalJSON()
	require.NoError(t, err)

	var out types.GenesisState
	require.NoError(t, out.UnmarshalJSON(bz))
	require.Len(t, out.Attestations, 0)
}

func TestKeysConstants(t *testing.T) {
	require.Equal(t, "attestation", types.ModuleName)
	require.Equal(t, "attestation", types.StoreKey)
	require.Equal(t, types.ModuleName, types.RouterKey)
	require.Equal(t, "params", types.ParamsKey)
	require.Equal(t, 32, types.SessionHashLen)
}

func TestContentHashToHex(t *testing.T) {
	hash := []byte{0x00, 0x0f, 0xff, 0xab}
	require.Equal(t, hex.EncodeToString(hash), types.ContentHashToHex(hash))
	require.Equal(t, "000fffab", types.ContentHashToHex(hash))

	// Empty input encodes to empty string.
	require.Equal(t, "", types.ContentHashToHex([]byte{}))
	require.Equal(t, "", types.ContentHashToHex(nil))
}

func TestAttestationKey(t *testing.T) {
	sessionHash := []byte{0x01, 0x02, 0x03}
	key := types.AttestationKey("org-1", sessionHash)
	expected := "attestation/org-1/" + hex.EncodeToString(sessionHash)
	require.Equal(t, expected, string(key))

	// Distinct session hashes produce distinct keys.
	key2 := types.AttestationKey("org-1", []byte{0x09})
	require.NotEqual(t, string(key), string(key2))

	// Distinct orgs produce distinct keys.
	key3 := types.AttestationKey("org-2", sessionHash)
	require.NotEqual(t, string(key), string(key3))
}

func TestAttestationPrefix(t *testing.T) {
	prefix := types.AttestationPrefix("org-1", 7)
	require.Equal(t, fmt.Sprintf("attestation/%s/%d/", "org-1", uint64(7)), string(prefix))
}

func TestAttestationByEpochKey(t *testing.T) {
	sessionHash := []byte{0xde, 0xad, 0xbe, 0xef}
	key := types.AttestationByEpochKey("org-1", 3, sessionHash)
	expected := fmt.Sprintf("session_epoch/%s/%d/%s", "org-1", uint64(3), hex.EncodeToString(sessionHash))
	require.Equal(t, expected, string(key))

	// The epoch-key is prefixed by the epoch-prefix for the same org/epoch.
	prefix := types.AttestationByEpochPrefix("org-1", 3)
	require.True(t, len(key) >= len(prefix))
	require.Equal(t, string(prefix), string(key[:len(prefix)]))
}

func TestAttestationByEpochPrefix(t *testing.T) {
	prefix := types.AttestationByEpochPrefix("org-9", 42)
	require.Equal(t, fmt.Sprintf("session_epoch/%s/%d/", "org-9", uint64(42)), string(prefix))

	// Different epochs yield different prefixes.
	require.NotEqual(t, string(prefix), string(types.AttestationByEpochPrefix("org-9", 43)))
}

func TestProviderTypeEnumValues(t *testing.T) {
	require.Equal(t, types.ProviderType(0), types.ProviderType_PROVIDER_TYPE_UNSPECIFIED)
	require.Equal(t, types.ProviderType(1), types.ProviderType_PROVIDER_TYPE_LOCAL)
	require.Equal(t, types.ProviderType(2), types.ProviderType_PROVIDER_TYPE_CLOUD)
}
