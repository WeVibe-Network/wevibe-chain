package keeper_test

import (
	"context"
	"testing"

	logv2 "cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/wevibe-network/wevibe-chain/testutil/keeper"
	"github.com/wevibe-network/wevibe-chain/x/attestation/keeper"
	"github.com/wevibe-network/wevibe-chain/x/attestation/types"
)

type mockOrgKeeper struct {
	orgs map[string]bool
}

func (m *mockOrgKeeper) HasOrg(ctx context.Context, orgID string) (bool, error) {
	return m.orgs[orgID], nil
}

func setupKeeper(t *testing.T) (*keeper.Keeper, context.Context) {
	key := storetypes.NewKVStoreKey("attestation")
	storeService, _ := testkeeper.NewTestStoreService(t, key)
	logger := logv2.NewNopLogger()
	orgKeeper := &mockOrgKeeper{orgs: map[string]bool{"org-1": true}}
	k := keeper.NewKeeper(storeService, logger, "gov-authority", orgKeeper)
	return k, context.Background()
}

func TestSetGetSessionAttestation(t *testing.T) {
	k, ctx := setupKeeper(t)

	sessionHash := make([]byte, 32)
	for i := range sessionHash {
		sessionHash[i] = byte(i)
	}

	att := &types.StoredSessionAttestation{
		OrgId:             "org-1",
		SessionHash:       sessionHash,
		ModelId:           "qwen3:4b",
		TurnCount:         5,
		TokenCount:        1200,
		ProviderType:      types.ProviderType_PROVIDER_TYPE_LOCAL,
		ContributorId:     "contributor-1",
		Epoch:             1,
		SubmittedAtHeight: 100,
	}

	err := k.SetSessionAttestation(ctx, att)
	require.NoError(t, err)

	got, err := k.GetSessionAttestation(ctx, "org-1", sessionHash)
	require.NoError(t, err)
	require.Equal(t, att.OrgId, got.OrgId)
	require.Equal(t, att.ModelId, got.ModelId)
	require.Equal(t, att.TurnCount, got.TurnCount)
	require.Equal(t, att.TokenCount, got.TokenCount)
	require.Equal(t, att.ProviderType, got.ProviderType)
	require.Equal(t, att.ContributorId, got.ContributorId)

	require.True(t, k.HasSessionAttestation(ctx, "org-1", sessionHash))
	require.False(t, k.HasSessionAttestation(ctx, "org-1", make([]byte, 32)))
}

func TestListSessionAttestations(t *testing.T) {
	k, ctx := setupKeeper(t)

	for i := 0; i < 3; i++ {
		sessionHash := make([]byte, 32)
		sessionHash[0] = byte(i)
		att := &types.StoredSessionAttestation{
			OrgId:             "org-1",
			SessionHash:       sessionHash,
			ModelId:           "qwen3:4b",
			TurnCount:         uint32(i + 1),
			TokenCount:        uint32((i + 1) * 500),
			ProviderType:      types.ProviderType_PROVIDER_TYPE_LOCAL,
			ContributorId:     "contributor-1",
			Epoch:             1,
			SubmittedAtHeight: uint64(100 + i),
		}
		err := k.SetSessionAttestation(ctx, att)
		require.NoError(t, err)
	}

	list, err := k.ListSessionAttestations(ctx, "org-1", 1)
	require.NoError(t, err)
	require.Len(t, list, 3)

	list2, err := k.ListSessionAttestations(ctx, "org-1", 2)
	require.NoError(t, err)
	require.Len(t, list2, 0)
}

func TestNotFound(t *testing.T) {
	k, ctx := setupKeeper(t)
	_, err := k.GetSessionAttestation(ctx, "org-1", make([]byte, 32))
	require.ErrorIs(t, err, types.ErrAttestationNotFound)
}

func TestParams(t *testing.T) {
	k, ctx := setupKeeper(t)

	params, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(10000), params.MaxAttestationsPerEpoch)
	require.False(t, params.RequireAttestationForServe)

	custom := types.Params{
		MaxAttestationsPerEpoch:    5000,
		RequireAttestationForServe: true,
	}
	err = k.SetParams(ctx, custom)
	require.NoError(t, err)

	got, err := k.GetParams(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(5000), got.MaxAttestationsPerEpoch)
	require.True(t, got.RequireAttestationForServe)
}

func TestGenesisRoundtrip(t *testing.T) {
	k, ctx := setupKeeper(t)

	sessionHash := make([]byte, 32)
	sessionHash[0] = 0xAB
	att := &types.StoredSessionAttestation{
		OrgId:             "org-1",
		SessionHash:       sessionHash,
		ModelId:           "claude-sonnet-4-20250514",
		TurnCount:         10,
		TokenCount:        3000,
		ProviderType:      types.ProviderType_PROVIDER_TYPE_CLOUD,
		ContributorId:     "contributor-2",
		Epoch:             5,
		SubmittedAtHeight: 500,
	}
	err := k.SetSessionAttestation(ctx, att)
	require.NoError(t, err)

	state, err := k.ExportGenesis(ctx)
	require.NoError(t, err)
	require.Len(t, state.Attestations, 1)

	bz, err := state.MarshalJSON()
	require.NoError(t, err)
	var state2 types.GenesisState
	err = state2.UnmarshalJSON(bz)
	require.NoError(t, err)
	require.Len(t, state2.Attestations, 1)

	k2, ctx2 := setupKeeper(t)
	err = k2.InitGenesis(ctx2, &state2)
	require.NoError(t, err)

	got, err := k2.GetSessionAttestation(ctx2, "org-1", sessionHash)
	require.NoError(t, err)
	require.Equal(t, "claude-sonnet-4-20250514", got.ModelId)
	require.Equal(t, types.ProviderType_PROVIDER_TYPE_CLOUD, got.ProviderType)
}

func TestVerificationStubs(t *testing.T) {
	k, ctx := setupKeeper(t)

	verified, status := k.VerifyCommitLLMReceipt(ctx, make([]byte, 32))
	require.False(t, verified)
	require.Contains(t, status, "unverified")

	verified2, status2 := k.VerifyCloudProviderSignature(ctx, make([]byte, 32))
	require.False(t, verified2)
	require.Contains(t, status2, "unverified")
}
