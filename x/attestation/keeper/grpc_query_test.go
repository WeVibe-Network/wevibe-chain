package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wevibe-network/wevibe-chain/x/attestation/keeper"
	"github.com/wevibe-network/wevibe-chain/x/attestation/types"
)

// queryStoredAttestation builds a StoredSessionAttestation for query-test fixtures.
func queryStoredAttestation(seed byte, epoch uint64) *types.StoredSessionAttestation {
	hash := make([]byte, types.SessionHashLen)
	for i := range hash {
		hash[i] = seed
	}
	return &types.StoredSessionAttestation{
		OrgId:             "org-1",
		SessionHash:       hash,
		ModelId:           "qwen3:4b",
		TurnCount:         3,
		TokenCount:        600,
		ProviderType:      types.ProviderType_PROVIDER_TYPE_LOCAL,
		ContributorId:     "contributor-1",
		Epoch:             epoch,
		SubmittedAtHeight: 10,
	}
}

func TestQueryGetSessionAttestation_Success(t *testing.T) {
	k, ctx := setupKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	att := queryStoredAttestation(0x11, 1)
	require.NoError(t, k.SetSessionAttestation(ctx, att))

	resp, err := qs.GetSessionAttestation(ctx, &types.QueryGetSessionAttestationRequest{
		OrgId:       att.OrgId,
		SessionHash: att.SessionHash,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Attestation)
	require.Equal(t, att.OrgId, resp.Attestation.OrgId)
	require.Equal(t, att.ModelId, resp.Attestation.ModelId)
	require.Equal(t, att.SessionHash, resp.Attestation.SessionHash)
}

func TestQueryGetSessionAttestation_NotFound(t *testing.T) {
	k, ctx := setupKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	resp, err := qs.GetSessionAttestation(ctx, &types.QueryGetSessionAttestationRequest{
		OrgId:       "org-1",
		SessionHash: make([]byte, types.SessionHashLen),
	})
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrAttestationNotFound)
	require.Nil(t, resp)
}

func TestQueryListSessionAttestations_Success(t *testing.T) {
	k, ctx := setupKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	for i := byte(0); i < 3; i++ {
		require.NoError(t, k.SetSessionAttestation(ctx, queryStoredAttestation(0x20+i, 1)))
	}

	resp, err := qs.ListSessionAttestations(ctx, &types.QueryListSessionAttestationsRequest{
		OrgId: "org-1",
		Epoch: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Attestations, 3)
}

func TestQueryListSessionAttestations_Empty(t *testing.T) {
	k, ctx := setupKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	// No attestations stored for this org/epoch.
	resp, err := qs.ListSessionAttestations(ctx, &types.QueryListSessionAttestationsRequest{
		OrgId: "org-1",
		Epoch: 99,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Attestations, 0)
}

func TestQueryParams_Success(t *testing.T) {
	k, ctx := setupKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	resp, err := qs.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Params)
	require.Equal(t, uint64(10000), resp.Params.MaxAttestationsPerEpoch)
	require.False(t, resp.Params.RequireAttestationForServe)
}

func TestQueryParams_ReflectsUpdate(t *testing.T) {
	k, ctx := setupKeeper(t)
	qs := keeper.NewQueryServerImpl(k)

	require.NoError(t, k.SetParams(ctx, types.Params{
		MaxAttestationsPerEpoch:    777,
		RequireAttestationForServe: true,
	}))

	resp, err := qs.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp.Params)
	require.Equal(t, uint64(777), resp.Params.MaxAttestationsPerEpoch)
	require.True(t, resp.Params.RequireAttestationForServe)
}
