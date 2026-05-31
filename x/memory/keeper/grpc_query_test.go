package keeper

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/wevibe-network/wevibe-chain/testutil/keeper"
	"github.com/wevibe-network/wevibe-chain/x/memory/types"
)

// queryTestServer wires up a keeper with the standard test fixtures and returns
// the query server implementation alongside the keeper for direct state setup.
func queryTestServer(t *testing.T) (types.QueryServer, *Keeper, context.Context) {
	t.Helper()
	k, _, _ := makeTestKeeper(t)
	return NewQueryServerImpl(k), k, context.Background()
}

// seedApprovedMemory stores an approved (committed) memory and returns its CID.
func seedApprovedMemory(t *testing.T, k *Keeper, ctx context.Context, orgID string, contentHash []byte) string {
	t.Helper()
	return storeMemory(t, k, ctx, orgID, contentHash, types.MemoryState_MEMORY_STATE_COMMITTED)
}

// -----------------------------------------------------------------------------
// GetMemory
// -----------------------------------------------------------------------------

func TestQueryGetMemory_Success(t *testing.T) {
	qs, k, ctx := queryTestServer(t)

	contentHash := []byte("12345678901234567890123456789012")
	seedApprovedMemory(t, k, ctx, "test-org", contentHash)

	resp, err := qs.GetMemory(ctx, &types.QueryGetMemoryRequest{
		OrgId:       "test-org",
		ContentHash: contentHash,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Memory)
	require.Equal(t, "test-org", resp.Memory.OrgId)
	require.Equal(t, contentHash, resp.Memory.ContentHash)
	require.Equal(t, types.MemoryState_MEMORY_STATE_COMMITTED, resp.Memory.State)
}

func TestQueryGetMemory_NotFound(t *testing.T) {
	qs, _, ctx := queryTestServer(t)

	resp, err := qs.GetMemory(ctx, &types.QueryGetMemoryRequest{
		OrgId:       "test-org",
		ContentHash: []byte("00000000000000000000000000000000"),
	})
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrMemoryNotFound)
	require.Nil(t, resp)
}

// -----------------------------------------------------------------------------
// GetPendingCommitments
// -----------------------------------------------------------------------------

func TestQueryGetPendingCommitments_Success(t *testing.T) {
	qs, k, ctx := queryTestServer(t)

	pc := newPendingCommitment("test-org", []byte("11111111111111111111111111111111"), []string{"kw1"}, "contributor", 1, 100)
	require.NoError(t, k.SubmitCommitment(ctx, pc))

	resp, err := qs.GetPendingCommitments(ctx, &types.QueryGetPendingCommitmentsRequest{OrgId: "test-org"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Commitments, 1)
	require.Equal(t, "test-org", resp.Commitments[0].OrgId)
	require.Equal(t, "contributor", resp.Commitments[0].ContributorId)
}

func TestQueryGetPendingCommitments_Empty(t *testing.T) {
	qs, _, ctx := queryTestServer(t)

	resp, err := qs.GetPendingCommitments(ctx, &types.QueryGetPendingCommitmentsRequest{OrgId: "empty-org"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.Commitments)
}

// -----------------------------------------------------------------------------
// GetMemoryCount
// -----------------------------------------------------------------------------

func TestQueryGetMemoryCount_Success(t *testing.T) {
	qs, k, ctx := queryTestServer(t)

	contentHash := []byte("12345678901234567890123456789012")
	commitment := newPendingCommitment("test-org", contentHash, []string{"kw1"}, "contributor", 1, 100)
	require.NoError(t, k.SubmitCommitment(ctx, commitment))
	require.NoError(t, approveMemory(k, ctx, "test-org", contentHash, []byte("blob"), "leader-pubkey", nil))

	resp, err := qs.GetMemoryCount(ctx, &types.QueryGetMemoryCountRequest{OrgId: "test-org"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(1), resp.Count)
}

func TestQueryGetMemoryCount_EmptyOrgIsZero(t *testing.T) {
	qs, _, ctx := queryTestServer(t)

	resp, err := qs.GetMemoryCount(ctx, &types.QueryGetMemoryCountRequest{OrgId: "no-such-org"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(0), resp.Count)
}

// -----------------------------------------------------------------------------
// GetEpochMerkleRoot
// -----------------------------------------------------------------------------

func TestQueryGetEpochMerkleRoot_Success(t *testing.T) {
	qs, k, ctx := queryTestServer(t)

	contentHash := []byte("12345678901234567890123456789012")
	commitment := newPendingCommitment("test-org", contentHash, []string{"kw1"}, "contributor", 1, 100)
	require.NoError(t, k.SubmitCommitment(ctx, commitment))
	require.NoError(t, approveMemory(k, ctx, "test-org", contentHash, []byte("blob"), "leader-pubkey", nil))
	require.NoError(t, k.ComputeAndStoreEpochMerkleRoot(ctx, "test-org", 1))

	resp, err := qs.GetEpochMerkleRoot(ctx, &types.QueryGetEpochMerkleRootRequest{OrgId: "test-org", Epoch: 1})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(1), resp.MemoryCount)
	require.Len(t, resp.MerkleRoot, 32)
}

func TestQueryGetEpochMerkleRoot_NotFound(t *testing.T) {
	qs, _, ctx := queryTestServer(t)

	resp, err := qs.GetEpochMerkleRoot(ctx, &types.QueryGetEpochMerkleRootRequest{OrgId: "test-org", Epoch: 999})
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrMerkleRootNotFound)
	require.Nil(t, resp)
}

// -----------------------------------------------------------------------------
// Params
// -----------------------------------------------------------------------------

func TestQueryParams_Success(t *testing.T) {
	qs, k, ctx := queryTestServer(t)

	want := types.DefaultParams()
	require.NoError(t, k.SetParams(ctx, want))

	resp, err := qs.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Params)
	require.Equal(t, want, *resp.Params)
}

func TestQueryParams_DefaultWhenUnset(t *testing.T) {
	qs, _, ctx := queryTestServer(t)

	resp, err := qs.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Params)
	require.NoError(t, resp.Params.Validate())
}

// -----------------------------------------------------------------------------
// ListRelationships
// -----------------------------------------------------------------------------

func TestQueryListRelationships_Success(t *testing.T) {
	qs, k, ctx := queryTestServer(t)

	rel := &types.MemoryRelationship{
		OrgID:        "test-org",
		SourceCID:    "cid-source",
		TargetCID:    "cid-target",
		RelationType: types.RelationContradicts,
		Proposer:     "proposer",
		Approved:     true,
		Epoch:        2,
	}
	require.NoError(t, k.saveRelationship(ctx, "test-org", rel))

	resp, err := qs.ListRelationships(ctx, &types.QueryListRelationshipsRequest{
		OrgId: "test-org",
		Cid:   "cid-source",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Relationships, 1)
	require.Equal(t, "cid-source", resp.Relationships[0].SourceCid)
	require.Equal(t, "cid-target", resp.Relationships[0].TargetCid)
	require.Equal(t, types.RelationType_RELATION_TYPE_CONTRADICTS, resp.Relationships[0].RelationType)
}

func TestQueryListRelationships_Empty(t *testing.T) {
	qs, _, ctx := queryTestServer(t)

	resp, err := qs.ListRelationships(ctx, &types.QueryListRelationshipsRequest{
		OrgId: "test-org",
		Cid:   "no-such-cid",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.Relationships)
}

// -----------------------------------------------------------------------------
// GetValidity
// -----------------------------------------------------------------------------

// seedValidity persists a StoredValidityMetadata record directly into the
// keeper store so the query path can read it back.
func seedValidity(t *testing.T, k *Keeper, ctx context.Context, orgID, cid string, after, until uint64, scope map[string]string) {
	t.Helper()
	var scopeBz []byte
	if len(scope) > 0 {
		bz, err := json.Marshal(scope)
		require.NoError(t, err)
		scopeBz = bz
	}
	stored := &types.StoredValidityMetadata{
		OrgId:           orgID,
		MemoryCid:       cid,
		ValidAfterEpoch: after,
		ValidUntilEpoch: until,
		ScopeTagsBz:     scopeBz,
	}
	bz, err := proto.Marshal(stored)
	require.NoError(t, err)
	require.NoError(t, k.getStore(ctx).Set(validityKey(orgID, cid), bz))
}

func TestQueryGetValidity_Success(t *testing.T) {
	qs, k, ctx := queryTestServer(t)

	seedValidity(t, k, ctx, "test-org", "cid-1", 5, 20, map[string]string{"region": "us"})

	resp, err := qs.GetValidity(ctx, &types.QueryGetValidityRequest{
		OrgId: "test-org",
		Cid:   "cid-1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.True(t, resp.Found)
	require.NotNil(t, resp.Metadata)
	require.Equal(t, "test-org", resp.Metadata.OrgId)
	require.Equal(t, "cid-1", resp.Metadata.MemoryCid)
	require.Equal(t, uint64(5), resp.Metadata.ValidAfterEpoch)
	require.Equal(t, uint64(20), resp.Metadata.ValidUntilEpoch)
	require.NotEmpty(t, resp.Metadata.ScopeTagsBz)
}

func TestQueryGetValidity_NotFound(t *testing.T) {
	qs, _, ctx := queryTestServer(t)

	resp, err := qs.GetValidity(ctx, &types.QueryGetValidityRequest{
		OrgId: "test-org",
		Cid:   "missing-cid",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.False(t, resp.Found)
	require.Nil(t, resp.Metadata)
}

// -----------------------------------------------------------------------------
// GetMemoriesBatch
// -----------------------------------------------------------------------------

func TestQueryGetMemoriesBatch_Success(t *testing.T) {
	// GetMemoriesBatch calls sdk.UnwrapSDKContext, so it requires a real SDK
	// context rather than context.Background(). Build one mirroring the
	// msg_server test setup.
	storeKey := storetypes.NewKVStoreKey("memory")
	storeService, cms := keeper.NewTestStoreService(t, storeKey)
	logger := keeper.NewTestLogger()
	sdk.DefaultBondDenom = "uvibe"
	mockOrg := &mockOrgKeeper{
		orgs:    map[string]bool{"test-org": true},
		leaders: map[string]string{"test-org": "leader-pubkey"},
	}
	k := NewKeeper(storeService, logger, "gov", mockOrg, &mockReputationKeeper{})
	sdkCtx := sdk.NewContext(cms, tmproto.Header{Height: 1, Time: time.Now().UTC()}, false, logger)
	ctx := sdk.WrapSDKContext(sdkCtx)
	qs := NewQueryServerImpl(k)

	hashFound := []byte("11111111111111111111111111111111")
	hashMissing := []byte("22222222222222222222222222222222")
	seedApprovedMemory(t, k, ctx, "test-org", hashFound)

	resp, err := qs.GetMemoriesBatch(ctx, &types.QueryGetMemoriesBatchRequest{
		OrgId:         "test-org",
		ContentHashes: [][]byte{hashFound, hashMissing},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Memories, 1)
	require.Equal(t, hashFound, resp.Memories[0].ContentHash)
	require.Len(t, resp.NotFound, 1)
	require.Equal(t, hashMissing, resp.NotFound[0])
}

func TestQueryGetMemoriesBatch_EmptyHashesIsInvalid(t *testing.T) {
	qs, _, ctx := queryTestServer(t)

	resp, err := qs.GetMemoriesBatch(ctx, &types.QueryGetMemoriesBatchRequest{
		OrgId:         "test-org",
		ContentHashes: nil,
	})
	require.Error(t, err)
	require.Nil(t, resp)
}

func TestQueryGetMemoriesBatch_MissingOrgIsInvalid(t *testing.T) {
	qs, _, ctx := queryTestServer(t)

	resp, err := qs.GetMemoriesBatch(ctx, &types.QueryGetMemoriesBatchRequest{
		OrgId:         "",
		ContentHashes: [][]byte{[]byte("11111111111111111111111111111111")},
	})
	require.Error(t, err)
	require.Nil(t, resp)
}
