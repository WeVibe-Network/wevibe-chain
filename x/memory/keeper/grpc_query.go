package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/wevibe-network/wevibe-chain/x/memory/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type queryServer struct {
	types.UnimplementedQueryServer
	keeper *Keeper
}

var _ types.QueryServer = (*queryServer)(nil)

func NewQueryServerImpl(k *Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

func (q *queryServer) GetMemory(ctx context.Context, req *types.QueryGetMemoryRequest) (*types.QueryGetMemoryResponse, error) {
	memory, err := q.keeper.GetApprovedMemory(ctx, req.OrgId, req.ContentHash)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetMemoryResponse{
		Memory: &types.StoredMemoryCommitment{
			OrgId:                  memory.OrgID,
			ContentHash:            memory.ContentHash,
			EncryptedBlob:          memory.EncryptedBlob,
			Keywords:               memory.Keywords,
			ContributorPubkey:      memory.Contributor,
			ContributorAddress:     memory.ContributorAddress,
			ProducerModelId:        memory.ProducerModelId,
			AttestationSessionHash: memory.AttestationSessionHash,
			PlaintextHash:          memory.PlaintextHash,
			Salt:                   memory.Salt,
			CiphertextHash:         memory.CiphertextHash,
			WrappedDekHash:         memory.WrappedDekHash,
			ContributorSig:         memory.ContributorSig,
			Epoch:                  memory.Epoch,
			CommittedAtHeight:      memory.CommittedAtHeight,
			CommittingLeaderPubkey: memory.CommittingLeader,
			State:                  memory.State,
			LastActiveEpoch:        memory.LastActiveEpoch,
			WrappedDekEnc:          memory.WrappedDekEnc,
			McVersion:              memory.McVersion,
			MemoryType:             memory.MemoryType,
			ApprovedAtEpoch:        memory.ApprovedAtEpoch,
			ServeCountTotal:        memory.ServeCountTotal,
			DenialCountTotal:       memory.DenialCountTotal,
			ArchivedEpoch:          memory.ArchivedEpoch,
		},
	}, nil
}

func (q *queryServer) GetPendingCommitments(ctx context.Context, req *types.QueryGetPendingCommitmentsRequest) (*types.QueryGetPendingCommitmentsResponse, error) {
	pending, err := q.keeper.GetAllPendingForOrg(ctx, req.OrgId)
	if err != nil {
		return nil, err
	}
	var stored []*types.StoredPendingCommitment
	for _, p := range pending {
		stored = append(stored, &types.StoredPendingCommitment{
			OrgId:             p.OrgID,
			ContentHash:       p.ContentHash,
			Keywords:          p.Keywords,
			ContributorId:     p.Contributor,
			Epoch:             p.Epoch,
			SubmittedAtHeight: p.SubmittedAt,
			MemoryType:        p.MemoryType,
		})
	}
	return &types.QueryGetPendingCommitmentsResponse{Commitments: stored}, nil
}

func (q *queryServer) GetMemoryCount(ctx context.Context, req *types.QueryGetMemoryCountRequest) (*types.QueryGetMemoryCountResponse, error) {
	count, err := q.keeper.GetApprovedCount(ctx, req.OrgId)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetMemoryCountResponse{Count: count}, nil
}

func (q *queryServer) GetEpochMerkleRoot(ctx context.Context, req *types.QueryGetEpochMerkleRootRequest) (*types.QueryGetEpochMerkleRootResponse, error) {
	merkle, err := q.keeper.GetEpochMerkleRoot(ctx, req.OrgId, req.Epoch)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetEpochMerkleRootResponse{
		MerkleRoot:  merkle.MerkleRoot,
		MemoryCount: merkle.MemoryCount,
	}, nil
}

func (q *queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := q.keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: &params}, nil
}

func (q *queryServer) ListRelationships(ctx context.Context, req *types.QueryListRelationshipsRequest) (*types.QueryListRelationshipsResponse, error) {
	rels, err := q.keeper.ListRelationshipsForMemory(ctx, req.OrgId, req.Cid)
	if err != nil {
		return nil, err
	}
	stored := make([]*types.StoredMemoryRelationship, 0, len(rels))
	for _, rel := range rels {
		stored = append(stored, &types.StoredMemoryRelationship{
			OrgId:        rel.OrgID,
			SourceCid:    rel.SourceCID,
			TargetCid:    rel.TargetCID,
			RelationType: rel.RelationType,
			Proposer:     rel.Proposer,
			Approved:     rel.Approved,
			Epoch:        rel.Epoch,
		})
	}
	return &types.QueryListRelationshipsResponse{Relationships: stored}, nil
}

func (q *queryServer) GetValidity(ctx context.Context, req *types.QueryGetValidityRequest) (*types.QueryGetValidityResponse, error) {
	metadata, found, err := q.keeper.GetValidityMetadata(ctx, req.OrgId, req.Cid)
	if err != nil {
		return nil, err
	}
	if !found {
		return &types.QueryGetValidityResponse{Found: false}, nil
	}
	scopeBz, err := json.Marshal(metadata.ScopeTags)
	if err != nil {
		return nil, fmt.Errorf("marshal scope tags: %w", err)
	}
	stored := &types.StoredValidityMetadata{
		OrgId:           req.OrgId,
		MemoryCid:       req.Cid,
		ValidAfterEpoch: metadata.ValidAfterEpoch,
		ValidUntilEpoch: metadata.ValidUntilEpoch,
		ScopeTagsBz:     scopeBz,
	}
	return &types.QueryGetValidityResponse{Metadata: stored, Found: true}, nil
}

func (q *queryServer) GetMemoriesBatch(ctx context.Context, req *types.QueryGetMemoriesBatchRequest) (*types.QueryGetMemoriesBatchResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.OrgId == "" {
		return nil, status.Error(codes.InvalidArgument, "org_id required")
	}
	if len(req.ContentHashes) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one content_hash required")
	}
	if len(req.ContentHashes) > 50 {
		return nil, status.Error(codes.InvalidArgument, "max 50 content hashes per batch")
	}

	var memories []*types.StoredMemoryCommitment
	var notFound [][]byte

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	for _, hash := range req.ContentHashes {
		memory, err := q.keeper.GetApprovedMemory(sdkCtx, req.OrgId, hash)
		if err != nil {
			notFound = append(notFound, hash)
			continue
		}
		memories = append(memories, &types.StoredMemoryCommitment{
			OrgId:                  memory.OrgID,
			ContentHash:            memory.ContentHash,
			EncryptedBlob:          memory.EncryptedBlob,
			Keywords:               memory.Keywords,
			ContributorPubkey:      memory.Contributor,
			ContributorAddress:     memory.ContributorAddress,
			ProducerModelId:        memory.ProducerModelId,
			AttestationSessionHash: memory.AttestationSessionHash,
			PlaintextHash:          memory.PlaintextHash,
			Salt:                   memory.Salt,
			CiphertextHash:         memory.CiphertextHash,
			WrappedDekHash:         memory.WrappedDekHash,
			ContributorSig:         memory.ContributorSig,
			Epoch:                  memory.Epoch,
			CommittedAtHeight:      memory.CommittedAtHeight,
			CommittingLeaderPubkey: memory.CommittingLeader,
			State:                  memory.State,
			LastActiveEpoch:        memory.LastActiveEpoch,
			WrappedDekEnc:          memory.WrappedDekEnc,
			McVersion:              memory.McVersion,
			MemoryType:             memory.MemoryType,
			ApprovedAtEpoch:        memory.ApprovedAtEpoch,
			ServeCountTotal:        memory.ServeCountTotal,
			DenialCountTotal:       memory.DenialCountTotal,
			ArchivedEpoch:          memory.ArchivedEpoch,
		})
	}

	return &types.QueryGetMemoriesBatchResponse{
		Memories: memories,
		NotFound: notFound,
	}, nil
}
