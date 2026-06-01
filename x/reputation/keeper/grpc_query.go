package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	query "github.com/cosmos/cosmos-sdk/types/query"
	memorytypes "github.com/wevibe-network/wevibe-chain/x/memory/types"
	"github.com/wevibe-network/wevibe-chain/x/reputation/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type queryServer struct {
	keeper *Keeper
}

var _ types.QueryServer = (*queryServer)(nil)

func NewQueryServerImpl(k *Keeper) types.QueryServer {
	return &queryServer{keeper: k}
}

func (q *queryServer) GetReputation(ctx context.Context, req *types.QueryGetReputationRequest) (*types.QueryGetReputationResponse, error) {
	stats, err := q.keeper.GetReputation(ctx, req.Developer)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetReputationResponse{
		DeveloperId: stats.DeveloperID,
		MemoryCount: stats.MemoryCount,
		Xp:          stats.XP,
	}, nil
}

func (q *queryServer) GetXP(ctx context.Context, req *types.QueryGetXPRequest) (*types.QueryGetXPResponse, error) {
	xp, err := q.keeper.GetXP(ctx, req.Developer)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetXPResponse{Xp: xp}, nil
}

func (q *queryServer) IsActive(ctx context.Context, req *types.QueryIsActiveRequest) (*types.QueryIsActiveResponse, error) {
	active := q.keeper.IsActive(ctx)
	return &types.QueryIsActiveResponse{Active: active}, nil
}

func (q *queryServer) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := q.keeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryParamsResponse{Params: &params}, nil
}

func (q *queryServer) GetServeStats(ctx context.Context, req *types.QueryGetServeStatsRequest) (*types.QueryGetServeStatsResponse, error) {
	stats, err := q.keeper.GetServeStats(ctx, req.Developer)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetServeStatsResponse{
		ServeCount:     stats.ServeCount,
		SelfServeCount: stats.SelfServeCount,
		OrgBreadth:     stats.OrgBreadth,
		ServeXp:        stats.ServeXP,
		FirstSeenEpoch: stats.FirstSeenEpoch,
	}, nil
}

func (q *queryServer) GetContributorOrgSet(ctx context.Context, req *types.QueryGetContributorOrgSetRequest) (*types.QueryGetContributorOrgSetResponse, error) {
	orgSet, err := q.keeper.GetContributorOrgSet(ctx, req.Developer)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetContributorOrgSetResponse{
		OrgIds: orgSet.OrgIds,
	}, nil
}

func (q *queryServer) GetCrossOrgProfile(ctx context.Context, req *types.QueryGetCrossOrgProfileRequest) (*types.QueryGetCrossOrgProfileResponse, error) {
	stats, err := q.keeper.GetReputation(ctx, req.Developer)
	if err != nil {
		return nil, err
	}
	orgSet, err := q.keeper.GetContributorOrgSet(ctx, req.Developer)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetCrossOrgProfileResponse{
		DeveloperId:    stats.DeveloperID,
		MemoryCount:    stats.MemoryCount,
		Xp:             stats.XP,
		ServeCount:     stats.ServeCount,
		SelfServeCount: stats.SelfServeCount,
		OrgBreadth:     stats.OrgBreadth,
		ServeXp:        stats.ServeXP,
		FirstSeenEpoch: stats.FirstSeenEpoch,
		OrgIds:         orgSet.OrgIds,
		DomainTags:     stats.DomainTags,
	}, nil
}

func (q *queryServer) GetContributorProfile(ctx context.Context, req *types.QueryGetContributorProfileRequest) (*types.QueryGetContributorProfileResponse, error) {
	if req.ContributorId == "" {
		return nil, types.ErrInvalidDeveloper
	}

	devBytes := []byte(req.ContributorId)

	stats, err := q.keeper.GetReputation(ctx, devBytes)
	if err != nil {
		return nil, err
	}

	topDomains, _ := q.keeper.GetTopDomains(ctx, devBytes, 10)

	orgSet, _ := q.keeper.GetContributorOrgSet(ctx, devBytes)
	var orgIDs []string
	if orgSet != nil {
		orgIDs = orgSet.OrgIds
	}

	histogram := make([]uint64, len(stats.DifficultyBucket))
	for i, v := range stats.DifficultyBucket {
		histogram[i] = v
	}

	resp := &types.QueryGetContributorProfileResponse{
		ContributorId:       req.ContributorId,
		Xp:                  stats.XP,
		ServeXp:             stats.ServeXP,
		MemoryCount:         stats.MemoryCount,
		ServeCount:          stats.ServeCount,
		SelfServeCount:      stats.SelfServeCount,
		OrgBreadth:          stats.OrgBreadth,
		FirstSeenEpoch:      stats.FirstSeenEpoch,
		DifficultyHistogram: histogram,
		TopDomains:          topDomains,
		ProvenanceBreakdown: stats.ProvenanceBreakdown,
		OrgIds:              orgIDs,
	}

	if q.keeper.serveKeeper != nil && req.Epoch > 0 {
		serveCount, selfServeCount, totalTurns, _, err := q.keeper.serveKeeper.GetContributorEpochServesRaw(ctx, req.ContributorId, req.Epoch)
		if err == nil {
			resp.EpochServes = serveCount
			resp.EpochSelfServes = selfServeCount
			resp.EpochTotalTurns = totalTurns
		}
	}

	return resp, nil
}

func (q *queryServer) LeaderProfile(ctx context.Context, req *types.QueryLeaderProfileRequest) (*types.QueryLeaderProfileResponse, error) {
	if req.LeaderPubkey == "" || req.OrgId == "" {
		return nil, types.ErrInvalidDeveloper
	}
	profile, err := q.keeper.GetLeaderProfile(ctx, req.LeaderPubkey, req.OrgId)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, types.ErrInvalidDeveloper
	}
	return &types.QueryLeaderProfileResponse{Profile: profile}, nil
}

func (q *queryServer) ModeratorProfile(ctx context.Context, req *types.QueryModeratorProfileRequest) (*types.QueryModeratorProfileResponse, error) {
	if req.ModeratorPubkey == "" || req.OrgId == "" {
		return nil, types.ErrInvalidDeveloper
	}
	profile, err := q.keeper.GetModeratorProfile(ctx, req.ModeratorPubkey, req.OrgId)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, types.ErrInvalidDeveloper
	}
	return &types.QueryModeratorProfileResponse{Profile: profile}, nil
}

func (q *queryServer) UpheldReportsByContributor(ctx context.Context, req *types.QueryUpheldReportsByContributorRequest) (*types.QueryUpheldReportsByContributorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ContributorId == "" {
		return nil, status.Error(codes.InvalidArgument, "contributor_id required")
	}

	var matching []*types.UpheldReportEntry
	err := q.keeper.memoryKeeper.IterateUpheldReports(ctx, func(report *memorytypes.StoredMemoryReport) bool {
		if report.ContributorPubkey == req.ContributorId {
			matching = append(matching, &types.UpheldReportEntry{
				OrgId:               report.OrgId,
				ContentHash:         report.ContentHash,
				ContributorPubkey:   report.ContributorPubkey,
				ApprovingModerators: report.ApprovingModerators,
				UpholdingModerators: report.UpholdingModerators,
				ReporterPubkey:      report.ReporterPubkey,
				Reason:              report.Reason,
				UpheldAtEpoch:       report.UpheldAtEpoch,
			})
		}
		return true
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "iterate upheld reports: %v", err)
	}

	paginated, pageRes, err := paginateReports(matching, req.Pagination)
	if err != nil {
		return nil, err
	}

	return &types.QueryUpheldReportsByContributorResponse{
		Reports:    paginated,
		Pagination: pageRes,
	}, nil
}

func (q *queryServer) UpheldReportsByModerator(ctx context.Context, req *types.QueryUpheldReportsByModeratorRequest) (*types.QueryUpheldReportsByModeratorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ModeratorPubkey == "" {
		return nil, status.Error(codes.InvalidArgument, "moderator_pubkey required")
	}

	var matching []*types.UpheldReportEntry
	err := q.keeper.memoryKeeper.IterateUpheldReports(ctx, func(report *memorytypes.StoredMemoryReport) bool {
		for _, approver := range report.ApprovingModerators {
			if approver == req.ModeratorPubkey {
				matching = append(matching, &types.UpheldReportEntry{
					OrgId:               report.OrgId,
					ContentHash:         report.ContentHash,
					ContributorPubkey:   report.ContributorPubkey,
					ApprovingModerators: report.ApprovingModerators,
					UpholdingModerators: report.UpholdingModerators,
					ReporterPubkey:      report.ReporterPubkey,
					Reason:              report.Reason,
					UpheldAtEpoch:       report.UpheldAtEpoch,
				})
				break
			}
		}
		return true
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "iterate upheld reports: %v", err)
	}

	paginated, pageRes, err := paginateReports(matching, req.Pagination)
	if err != nil {
		return nil, err
	}

	return &types.QueryUpheldReportsByModeratorResponse{
		Reports:    paginated,
		Pagination: pageRes,
	}, nil
}

func (q *queryServer) UpheldReportsByLeader(ctx context.Context, req *types.QueryUpheldReportsByLeaderRequest) (*types.QueryUpheldReportsByLeaderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.LeaderPubkey == "" {
		return nil, status.Error(codes.InvalidArgument, "leader_pubkey required")
	}

	var matching []*types.UpheldReportEntry
	err := q.keeper.memoryKeeper.IterateUpheldReports(ctx, func(report *memorytypes.StoredMemoryReport) bool {
		if report.CommittingLeaderPubkey == req.LeaderPubkey {
			matching = append(matching, &types.UpheldReportEntry{
				OrgId:               report.OrgId,
				ContentHash:         report.ContentHash,
				ContributorPubkey:   report.ContributorPubkey,
				ApprovingModerators: report.ApprovingModerators,
				UpholdingModerators: report.UpholdingModerators,
				ReporterPubkey:      report.ReporterPubkey,
				Reason:              report.Reason,
				UpheldAtEpoch:       report.UpheldAtEpoch,
			})
		}
		return true
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "iterate upheld reports: %v", err)
	}

	paginated, pageRes, err := paginateReports(matching, req.Pagination)
	if err != nil {
		return nil, err
	}

	return &types.QueryUpheldReportsByLeaderResponse{
		Reports:    paginated,
		Pagination: pageRes,
	}, nil
}

func (q *queryServer) VerifyUpheldReport(ctx context.Context, req *types.QueryVerifyUpheldReportRequest) (*types.QueryVerifyUpheldReportResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.OrgId == "" {
		return nil, status.Error(codes.InvalidArgument, "org_id required")
	}
	if len(req.ContentHash) == 0 {
		return nil, status.Error(codes.InvalidArgument, "content_hash required")
	}

	memory, err := q.keeper.memoryKeeper.GetApprovedMemory(ctx, req.OrgId, req.ContentHash)
	if err != nil {
		if err == memorytypes.ErrMemoryNotFound {
			return nil, status.Error(codes.NotFound, "approved memory not found")
		}
		return nil, status.Errorf(codes.Internal, "lookup approved memory: %v", err)
	}

	report, err := q.keeper.memoryKeeper.GetUpheldReport(ctx, req.OrgId, req.ContentHash)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "lookup upheld report: %v", err)
	}

	if len(memory.PlaintextHash) == 0 || len(memory.Salt) == 0 || len(memory.CiphertextHash) == 0 || len(memory.ContributorSig) == 0 {
		return nil, status.Error(codes.Internal, "approved memory missing verification fields")
	}

	wrappedDekHash := sha256.Sum256(memory.WrappedDekEnc)
	if len(memory.WrappedDekHash) > 0 && !bytesEqual(memory.WrappedDekHash, wrappedDekHash[:]) {
		return nil, status.Error(codes.Internal, "stored wrapped_dek_hash mismatch")
	}

	hasher := sha256.New()
	_, _ = hasher.Write(memory.EncryptedBlob)
	_, _ = hasher.Write(memory.WrappedDekEnc)
	submissionHash := hasher.Sum(nil)
	if !bytesEqual(memory.ContentHash, submissionHash) {
		return nil, status.Error(codes.Internal, "stored content_hash mismatch")
	}

	canonicalBody := buildSubmitMemoryCanonicalBody(
		memory.CiphertextHash,
		memory.Contributor,
		memory.Epoch,
		memory.MemoryType,
		memory.OrgID,
		memory.PlaintextHash,
		memory.Salt,
		submissionHash,
		wrappedDekHash[:],
	)

	var plaintext, ciphertext, capsule []byte
	var plaintextOversized bool
	var approvingModerators, upholdingModerators []string
	var upheldAtEpoch uint64
	if report != nil {
		plaintext = report.Plaintext
		ciphertext = report.Ciphertext
		capsule = report.Capsule
		plaintextOversized = report.PlaintextOversized
		approvingModerators = report.ApprovingModerators
		upholdingModerators = report.UpholdingModerators
		upheldAtEpoch = report.UpheldAtEpoch
	}

	return &types.QueryVerifyUpheldReportResponse{
		Plaintext:           plaintext,
		Ciphertext:          ciphertext,
		Capsule:             capsule,
		PlaintextHash:       memory.PlaintextHash,
		PlaintextOversized:  plaintextOversized,
		ApprovingModerators: approvingModerators,
		UpholdingModerators: upholdingModerators,
		UpheldAtEpoch:       upheldAtEpoch,
		Salt:                memory.Salt,
		CiphertextHash:      memory.CiphertextHash,
		WrappedDekHash:      wrappedDekHash[:],
		ContributorSig:      memory.ContributorSig,
		ContributorPubkey:   memory.Contributor,
		EncryptedBlob:       memory.EncryptedBlob,
		WrappedDekEnc:       memory.WrappedDekEnc,
		ContentHash:         memory.ContentHash,
		OrgId:               memory.OrgID,
		Epoch:               memory.Epoch,
		MemoryType:          memorytypes.CanonicalMemoryType(memory.MemoryType),
		CanonicalBody:       canonicalBody,
	}, nil
}

func buildSubmitMemoryCanonicalBody(ciphertextHash []byte, contributorPubkey string, epochID uint64, memoryType memorytypes.MemoryType, orgID string, plaintextHash, salt, submissionHash, wrappedDekHash []byte) []byte {
	canonicalLines := []string{
		"wevibe.submit_memory.v1",
		"ciphertext_hash:" + hex.EncodeToString(ciphertextHash),
		"contributor_pubkey:" + contributorPubkey,
		fmt.Sprintf("epoch_id:%d", epochID),
		"memory_type:" + memorytypes.CanonicalMemoryType(memoryType),
		"org_id:" + orgID,
		"plaintext_hash:" + hex.EncodeToString(plaintextHash),
		"salt:" + hex.EncodeToString(salt),
		"submission_hash:" + hex.EncodeToString(submissionHash),
		"wrapped_dek_hash:" + hex.EncodeToString(wrappedDekHash),
	}

	return []byte(strings.Join(canonicalLines, "\n"))
}

func bytesEqual(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func paginateReports(reports []*types.UpheldReportEntry, page *query.PageRequest) ([]*types.UpheldReportEntry, *query.PageResponse, error) {
	if page == nil {
		page = &query.PageRequest{Limit: 100}
	}

	offset := int(page.Offset)
	limit := int(page.Limit)
	if limit == 0 {
		limit = 100
	}
	if offset > len(reports) {
		offset = len(reports)
	}
	end := offset + limit
	if end > len(reports) {
		end = len(reports)
	}

	return reports[offset:end], &query.PageResponse{Total: uint64(len(reports))}, nil
}
