package keeper_test

import (
	"context"
	"errors"

	memorytypes "github.com/wevibe-network/wevibe-chain/x/memory/types"
)

// mockOrgKeeper implements types.OrgKeeper.
type mockOrgKeeper struct {
	orgs map[string]bool
	err  error
}

func newMockOrgKeeper(orgs ...string) *mockOrgKeeper {
	m := &mockOrgKeeper{orgs: make(map[string]bool)}
	for _, o := range orgs {
		m.orgs[o] = true
	}
	return m
}

func (m *mockOrgKeeper) HasOrg(ctx context.Context, orgID string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.orgs[orgID], nil
}

// mockMemoryKeeper implements types.MemoryKeeper.
type mockMemoryKeeper struct {
	// approved maps orgID|hexHash -> commitment; absence => GetApprovedMemory errors.
	approved map[string]*memorytypes.MemoryCommitment
	// validEpoch maps orgID|cid|epoch -> validity flag. Default true when present in approved.
	validEpoch map[string]bool
	// validErr forces IsValidInEpoch to error.
	validErr error
	// boostCalls / decayCalls record side-effect invocations.
	boostCalls int
	decayCalls int
	// boostErr / decayErr force the boost/decay side effects to fail (non-fatal path).
	boostErr error
	decayErr error
}

func newMockMemoryKeeper() *mockMemoryKeeper {
	return &mockMemoryKeeper{
		approved:   make(map[string]*memorytypes.MemoryCommitment),
		validEpoch: make(map[string]bool),
	}
}

func memKey(orgID, hexHash string) string { return orgID + "|" + hexHash }

func (m *mockMemoryKeeper) approve(orgID string, hash []byte) {
	m.approved[memKey(orgID, hexEncode(hash))] = &memorytypes.MemoryCommitment{
		OrgID:       orgID,
		ContentHash: hash,
	}
}

func (m *mockMemoryKeeper) GetApprovedMemory(ctx context.Context, orgID string, contentHash []byte) (*memorytypes.MemoryCommitment, error) {
	c, ok := m.approved[memKey(orgID, hexEncode(contentHash))]
	if !ok {
		return nil, errors.New("approved memory not found")
	}
	return c, nil
}

func (m *mockMemoryKeeper) IsValidInEpoch(ctx context.Context, orgID string, cid string, epoch uint64) (bool, error) {
	if m.validErr != nil {
		return false, m.validErr
	}
	key := orgID + "|" + cid
	if v, ok := m.validEpoch[key]; ok {
		return v, nil
	}
	// Default: valid if the memory is approved.
	_, ok := m.approved[memKey(orgID, cid)]
	return ok, nil
}

func (m *mockMemoryKeeper) ApplyServeBoost(ctx context.Context, orgID string, contentHash []byte, epoch uint64) error {
	m.boostCalls++
	return m.boostErr
}

func (m *mockMemoryKeeper) ApplyDenialDecay(ctx context.Context, orgID string, contentHash []byte, epoch uint64) error {
	m.decayCalls++
	return m.decayErr
}

// mockBandwidthKeeper implements types.BandwidthKeeper.
type mockBandwidthKeeper struct {
	consumed map[string]uint64
	err      error
}

func newMockBandwidthKeeper() *mockBandwidthKeeper {
	return &mockBandwidthKeeper{consumed: make(map[string]uint64)}
}

func (m *mockBandwidthKeeper) ConsumeServeBandwidth(ctx context.Context, orgID string, epoch uint64, count uint64) error {
	if m.err != nil {
		return m.err
	}
	m.consumed[orgID] += count
	return nil
}

// mockReputationKeeper implements types.ReputationKeeper.
type mockReputationKeeper struct {
	serveCalls int
	serveErr   error
}

func newMockReputationKeeper() *mockReputationKeeper {
	return &mockReputationKeeper{}
}

func (m *mockReputationKeeper) RecordServe(ctx context.Context, contributorWallet []byte, orgID string, epoch uint64, isSelfServe bool) error {
	m.serveCalls++
	return m.serveErr
}

func (m *mockReputationKeeper) IncrementContribution(ctx context.Context, contributorWallet, orgID, memoryCID string) error {
	return nil
}

func (m *mockReputationKeeper) IncrementServe(ctx context.Context, contributorWallet, orgID, memoryCID string, count uint64) error {
	return nil
}

func (m *mockReputationKeeper) RecordBan(ctx context.Context, contributorWallet, orgID, memoryCID string) error {
	return nil
}
