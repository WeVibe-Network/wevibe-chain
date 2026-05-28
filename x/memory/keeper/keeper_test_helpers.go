package keeper

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wevibe-network/wevibe-chain/x/memory/types"
	orgtypes "github.com/wevibe-network/wevibe-chain/x/org/types"
)

type testOrgKeeper struct {
	orgs       map[string]bool
	leaders    map[string]string
	moderators map[string]map[string]bool
	configs    map[string]*orgtypes.OrgConfig
	configErrs map[string]error
}

func newTestOrgKeeper() *testOrgKeeper {
	return &testOrgKeeper{
		orgs:       make(map[string]bool),
		leaders:    make(map[string]string),
		moderators: make(map[string]map[string]bool),
		configs:    make(map[string]*orgtypes.OrgConfig),
		configErrs: make(map[string]error),
	}
}

func (t *testOrgKeeper) HasOrg(ctx context.Context, orgID string) (bool, error) {
	return t.orgs[orgID], nil
}

func (t *testOrgKeeper) IsLeader(ctx context.Context, orgID, memberPubkey string) (bool, error) {
	return t.leaders[orgID] == memberPubkey, nil
}

func (t *testOrgKeeper) IsModerator(ctx context.Context, orgID, memberPubkey string) (bool, error) {
	if mods, ok := t.moderators[orgID]; ok {
		return mods[memberPubkey], nil
	}
	return false, nil
}

func (t *testOrgKeeper) GetOrgConfig(ctx context.Context, orgID string) (*orgtypes.OrgConfig, error) {
	if err, ok := t.configErrs[orgID]; ok && err != nil {
		return nil, err
	}
	if cfg, ok := t.configs[orgID]; ok {
		return cfg, nil
	}
	return &orgtypes.OrgConfig{OrgID: orgID}, nil
}

func (t *testOrgKeeper) setOrg(orgID, leader string) {
	if orgID == "" {
		return
	}
	t.orgs[orgID] = true
	t.leaders[orgID] = leader
}

func (t *testOrgKeeper) setModerator(orgID, moderator string, value bool) {
	if _, ok := t.moderators[orgID]; !ok {
		t.moderators[orgID] = make(map[string]bool)
	}
	t.moderators[orgID][moderator] = value
}

func (t *testOrgKeeper) setConfig(orgID string, cfg *orgtypes.OrgConfig) {
	if cfg == nil {
		delete(t.configs, orgID)
		return
	}
	copy := *cfg
	t.configs[orgID] = &copy
}

func (t *testOrgKeeper) setConfigErr(orgID string, err error) {
	if err == nil {
		delete(t.configErrs, orgID)
		return
	}
	t.configErrs[orgID] = err
}

type testServeKeeper struct {
	counts  map[string]uint64
	denials map[string]uint64
}

func newTestServeKeeper() *testServeKeeper {
	return &testServeKeeper{
		counts:  make(map[string]uint64),
		denials: make(map[string]uint64),
	}
}

func (s *testServeKeeper) GetMemoryServeCountForEpoch(ctx context.Context, orgID, cid string, epoch uint64) (uint64, error) {
	return s.counts[fmt.Sprintf("%s:%s:%d", orgID, cid, epoch)], nil
}

func (s *testServeKeeper) GetMemoryDenialCountForEpoch(ctx context.Context, orgID, cid string, epoch uint64) (uint64, error) {
	return s.denials[fmt.Sprintf("%s:%s:%d", orgID, cid, epoch)], nil
}

type memoryFixtureOption func(*types.MemoryCommitment)

type serveFixtureOption func(*mockServeKeeper, string, string)

type serveEpochKey struct {
	orgID string
	cid   string
	epoch uint64
}

type mockServeKeeper struct {
	servesByEpoch  map[serveEpochKey]uint64
	denialsByEpoch map[serveEpochKey]uint64
	matchedByEpoch map[serveEpochKey]map[string]bool
}

func newMockServeKeeper() *mockServeKeeper {
	return &mockServeKeeper{
		servesByEpoch:  make(map[serveEpochKey]uint64),
		denialsByEpoch: make(map[serveEpochKey]uint64),
		matchedByEpoch: make(map[serveEpochKey]map[string]bool),
	}
}

func makeServeEpochKey(orgID, cid string, epoch uint64) serveEpochKey {
	return serveEpochKey{orgID: orgID, cid: cid, epoch: epoch}
}

func (m *mockServeKeeper) GetMemoryServeCountForEpoch(ctx context.Context, orgID, cid string, epoch uint64) (uint64, error) {
	return m.servesByEpoch[makeServeEpochKey(orgID, cid, epoch)], nil
}

func (m *mockServeKeeper) GetMemoryDenialCountForEpoch(ctx context.Context, orgID, cid string, epoch uint64) (uint64, error) {
	return m.denialsByEpoch[makeServeEpochKey(orgID, cid, epoch)], nil
}

func (m *mockServeKeeper) GetMatchedKeywordsForEpoch(ctx context.Context, orgID, cid string, epoch uint64) (map[string]bool, error) {
	matches, ok := m.matchedByEpoch[makeServeEpochKey(orgID, cid, epoch)]
	if !ok {
		return map[string]bool{}, nil
	}
	out := make(map[string]bool, len(matches))
	for keyword, matched := range matches {
		if matched {
			out[keyword] = true
		}
	}
	return out, nil
}

func withServeCount(epoch, count uint64) serveFixtureOption {
	return func(m *mockServeKeeper, orgID, cid string) {
		m.servesByEpoch[makeServeEpochKey(orgID, cid, epoch)] = count
	}
}

func withDenialCount(epoch, count uint64) serveFixtureOption {
	return func(m *mockServeKeeper, orgID, cid string) {
		m.denialsByEpoch[makeServeEpochKey(orgID, cid, epoch)] = count
	}
}

func withMatchedKeywords(epoch uint64, keywords ...string) serveFixtureOption {
	return func(m *mockServeKeeper, orgID, cid string) {
		key := makeServeEpochKey(orgID, cid, epoch)
		matched := make(map[string]bool, len(keywords))
		for _, keyword := range keywords {
			if keyword == "" {
				continue
			}
			matched[keyword] = true
		}
		m.matchedByEpoch[key] = matched
	}
}

func attachMockServeKeeper(k *Keeper, orgID, cid string, options ...serveFixtureOption) *mockServeKeeper {
	mock := newMockServeKeeper()
	for _, option := range options {
		option(mock, orgID, cid)
	}
	k.SetServeKeeper(mock)
	return mock
}

func withMemoryServeTotal(total uint64) memoryFixtureOption {
	return func(memory *types.MemoryCommitment) {
		memory.ServeCountTotal = total
	}
}

func withMemoryDenialTotal(total uint64) memoryFixtureOption {
	return func(memory *types.MemoryCommitment) {
		memory.DenialCountTotal = total
	}
}

func withKeywordWeight(keyword, weight string) memoryFixtureOption {
	return func(memory *types.MemoryCommitment) {
		for i := range memory.Keywords {
			if memory.Keywords[i].Keyword == keyword {
				memory.Keywords[i].Weight = weight
			}
		}
	}
}

func withKeywords(keywords ...*types.KeywordWeight) memoryFixtureOption {
	return func(memory *types.MemoryCommitment) {
		memory.Keywords = keywords
	}
}

const (
	defaultOrgID  = "org1"
	defaultLeader = "leader"
)

func storeMemory(t *testing.T, k *Keeper, ctx context.Context, orgID string, contentHash []byte, state types.MemoryState) string {
	return storeMemoryWithKeywords(t, k, ctx, orgID, contentHash, state, 0)
}

func storeMemoryWithKeywords(t *testing.T, k *Keeper, ctx context.Context, orgID string, contentHash []byte, state types.MemoryState, lastActiveEpoch uint64, options ...memoryFixtureOption) string {
	memory := &types.MemoryCommitment{
		OrgID:             orgID,
		ContentHash:       contentHash,
		EncryptedBlob:     []byte("blob"),
		Keywords:          []*types.KeywordWeight{{Keyword: "kw", Weight: "1.0", ServeCount: 0, DenialCount: 0}},
		Contributor:       "contrib",
		Epoch:             1,
		CommittedAtHeight: 1,
		CommittingLeader:  "leader",
		State:             state,
		LastActiveEpoch:   lastActiveEpoch,
		MemoryType:        types.MemoryType_MEMORY_TYPE_CORRECT_IMPLEMENTATION,
	}
	for _, option := range options {
		option(memory)
	}
	require.NoError(t, k.saveMemoryCommitment(ctx, memory))
	return types.ContentHashToHex(contentHash)
}
