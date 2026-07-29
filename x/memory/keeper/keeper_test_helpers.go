package keeper

import (
	"context"
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

func (t *testOrgKeeper) GetMember(ctx context.Context, orgID, memberPubkey string) (*orgtypes.MemberRecord, error) {
	if t.leaders[orgID] == memberPubkey {
		return &orgtypes.MemberRecord{OrgID: orgID, Pubkey: memberPubkey, Role: "leader", X25519Pubkey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"}, nil
	}
	if mods, ok := t.moderators[orgID]; ok && mods[memberPubkey] {
		return &orgtypes.MemberRecord{OrgID: orgID, Pubkey: memberPubkey, Role: "moderator", X25519Pubkey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"}, nil
	}
	return nil, orgtypes.ErrMemberNotFound
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

func (t *testOrgKeeper) GetLeaderWallet(ctx context.Context, orgID string) (string, error) {
	return t.leaders[orgID], nil
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

type memoryFixtureOption func(*types.MemoryCommitment)

func withMemoryType(memoryType types.MemoryType) memoryFixtureOption {
	return func(memory *types.MemoryCommitment) {
		memory.MemoryType = memoryType
	}
}

func withKeywords(keywords ...string) memoryFixtureOption {
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
		Keywords:          []string{"kw"},
		Contributor:       "contrib",
		Epoch:             1,
		CommittedAtHeight: 1,
		CommittingLeader:  "leader",
		State:             state,
		LastActiveEpoch:   lastActiveEpoch,
		MemoryType:        types.MemoryType_MEMORY_TYPE_MEMORY,
	}
	for _, option := range options {
		option(memory)
	}
	require.NoError(t, k.saveMemoryCommitment(ctx, memory))
	return types.ContentHashToHex(contentHash)
}
