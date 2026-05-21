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

const (
	defaultOrgID  = "org1"
	defaultLeader = "leader"
)

func storeMemory(t *testing.T, k *Keeper, ctx context.Context, orgID string, contentHash []byte, state types.MemoryState) string {
	return storeMemoryWithKeywords(t, k, ctx, orgID, contentHash, state, 0)
}

func storeMemoryWithKeywords(t *testing.T, k *Keeper, ctx context.Context, orgID string, contentHash []byte, state types.MemoryState, lastActiveEpoch uint64) string {
	memory := &types.MemoryCommitment{
		OrgID:               orgID,
		ContentHash:         contentHash,
		EncryptedBlob:       []byte("blob"),
		Keywords:            []*types.KeywordWeight{{Keyword: "kw", Weight: "1.0", ServeCount: 0, DenialCount: 0}},
		Contributor:         "contrib",
		Epoch:               1,
		CommittedAtHeight:   1,
		Approvers:           []string{"leader"},
		CommittingLeader:    "leader",
		State:               state,
		LastActiveEpoch:     lastActiveEpoch,
		MemoryType:          types.MemoryType_MEMORY_TYPE_CORRECT_IMPLEMENTATION,
	}
	require.NoError(t, k.saveMemoryCommitment(ctx, memory))
	return types.ContentHashToHex(contentHash)
}
