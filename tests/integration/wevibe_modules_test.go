package integration

import (
	"testing"

	"github.com/stretchr/testify/require"

	attestationkeeper "github.com/wevibe-network/wevibe-chain/x/attestation/keeper"
	attestationtypes "github.com/wevibe-network/wevibe-chain/x/attestation/types"
	bandwidthkeeper "github.com/wevibe-network/wevibe-chain/x/bandwidth/keeper"
	bandwidthtypes "github.com/wevibe-network/wevibe-chain/x/bandwidth/types"
	memorykeeper "github.com/wevibe-network/wevibe-chain/x/memory/keeper"
	memorytypes "github.com/wevibe-network/wevibe-chain/x/memory/types"
	orgtypes "github.com/wevibe-network/wevibe-chain/x/org/types"
	servekeeper "github.com/wevibe-network/wevibe-chain/x/serve/keeper"
	servetypes "github.com/wevibe-network/wevibe-chain/x/serve/types"
)

// registerOrg is a shared helper that registers an org through the live
// message router so module integration tests can exercise org-gated paths.
func registerOrg(t *testing.T, suite *TestSuite) string {
	t.Helper()
	_, err := suite.DeliverMsg(&orgtypes.MsgRegisterOrg{
		Signer:          suite.UserAddr.String(),
		Leader:          suite.UserAddr.String(),
		StorageQuota:    1000,
		RetrievalBudget: 500,
	})
	require.NoError(t, err)
	return orgtypes.DeriveOrgID(suite.UserAddr)
}

// ---------------------------------------------------------------------------
// Attestation — full message flow through the live app router
// ---------------------------------------------------------------------------

func sessionHash(seed byte) []byte {
	h := make([]byte, attestationtypes.SessionHashLen)
	for i := range h {
		h[i] = seed
	}
	return h
}

func TestAttestation_SubmitAndQuery_Integration(t *testing.T) {
	suite := NewTestSuite(t)
	orgID := registerOrg(t, suite)

	sh := sessionHash(0x42)
	_, err := suite.DeliverMsg(&attestationtypes.MsgSubmitSessionAttestation{
		Signer:        suite.UserAddr.String(),
		OrgId:         orgID,
		SessionHash:   sh,
		ModelId:       "qwen3:4b",
		TurnCount:     5,
		TokenCount:    1200,
		ProviderType:  attestationtypes.ProviderType_PROVIDER_TYPE_LOCAL,
		ContributorId: "contrib-1",
		Epoch:         1,
	})
	require.NoError(t, err)

	q := attestationkeeper.NewQueryServerImpl(suite.AttestationKeeper)
	resp, err := q.GetSessionAttestation(suite.Ctx, &attestationtypes.QueryGetSessionAttestationRequest{
		OrgId:       orgID,
		SessionHash: sh,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	list, err := q.ListSessionAttestations(suite.Ctx, &attestationtypes.QueryListSessionAttestationsRequest{
		OrgId: orgID,
		Epoch: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, list)
}

func TestAttestation_DuplicateRejected_Integration(t *testing.T) {
	suite := NewTestSuite(t)
	orgID := registerOrg(t, suite)

	msg := &attestationtypes.MsgSubmitSessionAttestation{
		Signer:        suite.UserAddr.String(),
		OrgId:         orgID,
		SessionHash:   sessionHash(0x43),
		ModelId:       "qwen3:4b",
		TurnCount:     1,
		TokenCount:    100,
		ProviderType:  attestationtypes.ProviderType_PROVIDER_TYPE_LOCAL,
		ContributorId: "contrib-1",
		Epoch:         1,
	}
	_, err := suite.DeliverMsg(msg)
	require.NoError(t, err)

	_, err = suite.DeliverMsg(msg)
	require.Error(t, err)
}

func TestAttestation_OrgNotFound_Integration(t *testing.T) {
	suite := NewTestSuite(t)
	// No org registered.
	_, err := suite.DeliverMsg(&attestationtypes.MsgSubmitSessionAttestation{
		Signer:        suite.UserAddr.String(),
		OrgId:         "ghost-org",
		SessionHash:   sessionHash(0x44),
		ModelId:       "qwen3:4b",
		ProviderType:  attestationtypes.ProviderType_PROVIDER_TYPE_LOCAL,
		ContributorId: "contrib-1",
		Epoch:         1,
	})
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Bandwidth — query wiring through the live app
// ---------------------------------------------------------------------------

func TestBandwidth_QueryWiring_Integration(t *testing.T) {
	suite := NewTestSuite(t)
	orgID := registerOrg(t, suite)

	q := bandwidthkeeper.NewQueryServerImpl(suite.BandwidthKeeper)

	// Params query returns defaults on a fresh chain.
	pResp, err := q.Params(suite.Ctx, &bandwidthtypes.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, pResp.Params)

	// Remaining bandwidth for a fresh org/epoch must not error.
	rResp, err := q.GetRemainingBandwidth(suite.Ctx, &bandwidthtypes.QueryGetRemainingBandwidthRequest{
		OrgId: orgID,
		Epoch: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, rResp)
}

// ---------------------------------------------------------------------------
// Memory — query wiring through the live app
// ---------------------------------------------------------------------------

func TestMemory_QueryWiring_Integration(t *testing.T) {
	suite := NewTestSuite(t)
	orgID := registerOrg(t, suite)

	q := memorykeeper.NewQueryServerImpl(suite.MemoryKeeper)

	pResp, err := q.Params(suite.Ctx, &memorytypes.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, pResp.Params)

	// A fresh org has zero approved memories.
	cResp, err := q.GetMemoryCount(suite.Ctx, &memorytypes.QueryGetMemoryCountRequest{OrgId: orgID})
	require.NoError(t, err)
	require.Equal(t, uint64(0), cResp.Count)
}

// ---------------------------------------------------------------------------
// Serve — query wiring through the live app
// ---------------------------------------------------------------------------

func TestServe_QueryWiring_Integration(t *testing.T) {
	suite := NewTestSuite(t)
	orgID := registerOrg(t, suite)

	q := servekeeper.NewQueryServerImpl(suite.ServeKeeper)

	// Params query returns defaults on a fresh chain.
	pResp, err := q.Params(suite.Ctx, &servetypes.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, pResp.Params)

	// No serves yet => epoch stats not found.
	_, err = q.GetEpochServeStats(suite.Ctx, &servetypes.QueryGetEpochServeStatsRequest{
		OrgId: orgID,
		Epoch: 1,
	})
	require.Error(t, err)
}
