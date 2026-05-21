package keeper

import (
	"github.com/wevibe-network/wevibe-chain/x/bandwidth/types"
)

func stateToStored(s *types.BandwidthState) *types.StoredBandwidthState {
	return &types.StoredBandwidthState{
		OrgId:       s.OrgID,
		Epoch:       s.Epoch,
		MemoryUsed:  s.MemoryUsed,
		MemoryCap:   s.MemoryCap,
		ServeUsed:   s.ServeUsed,
		ServeCap:    s.ServeCap,
	}
}

func storedToState(stored types.StoredBandwidthState) types.BandwidthState {
	return types.BandwidthState{
		OrgID:        stored.OrgId,
		Epoch:        stored.Epoch,
		MemoryUsed:   stored.MemoryUsed,
		MemoryCap:    stored.MemoryCap,
		ServeUsed:    stored.ServeUsed,
		ServeCap:     stored.ServeCap,
	}
}

func overrideToStored(o *types.BandwidthOverride) *types.StoredBandwidthOverride {
	return &types.StoredBandwidthOverride{
		OrgId:     o.OrgID,
		MemoryCap: o.MemoryCap,
		ServeCap:  o.ServeCap,
	}
}

func storedToOverride(stored types.StoredBandwidthOverride) types.BandwidthOverride {
	return types.BandwidthOverride{
		OrgID:     stored.OrgId,
		MemoryCap: stored.MemoryCap,
		ServeCap:  stored.ServeCap,
	}
}