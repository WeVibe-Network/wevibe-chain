package types

const (
	ModuleName = "bandwidth"
	StoreKey   = "bandwidth"
)

type BandwidthState struct {
	OrgID        string
	Epoch        uint64
	MemoryUsed   uint64
	MemoryCap    uint64
	ServeUsed    uint64
	ServeCap     uint64
}

func NewBandwidthState(orgID string, epoch, memoryCap, serveCap uint64) *BandwidthState {
	return &BandwidthState{
		OrgID:        orgID,
		Epoch:        epoch,
		MemoryUsed:   0,
		MemoryCap:    memoryCap,
		ServeUsed:    0,
		ServeCap:     serveCap,
	}
}

func (bs *BandwidthState) Validate() error {
	if bs.OrgID == "" {
		return ErrInvalidOrgID
	}
	return nil
}

type BandwidthOverride struct {
	OrgID      string
	MemoryCap  uint64
	ServeCap   uint64
}

func NewBandwidthOverride(orgID string, memoryCap, serveCap uint64) *BandwidthOverride {
	return &BandwidthOverride{
		OrgID:     orgID,
		MemoryCap: memoryCap,
		ServeCap:  serveCap,
	}
}

func (bo *BandwidthOverride) Validate() error {
	if bo.OrgID == "" {
		return ErrInvalidOrgID
	}
	return nil
}

func stateToStored(s *BandwidthState) *StoredBandwidthState {
	return &StoredBandwidthState{
		OrgId:       s.OrgID,
		Epoch:       s.Epoch,
		MemoryUsed:  s.MemoryUsed,
		MemoryCap:   s.MemoryCap,
		ServeUsed:   s.ServeUsed,
		ServeCap:    s.ServeCap,
	}
}

func storedToState(stored StoredBandwidthState) BandwidthState {
	return BandwidthState{
		OrgID:        stored.OrgId,
		Epoch:        stored.Epoch,
		MemoryUsed:   stored.MemoryUsed,
		MemoryCap:    stored.MemoryCap,
		ServeUsed:    stored.ServeUsed,
		ServeCap:     stored.ServeCap,
	}
}

func overrideToStored(o *BandwidthOverride) *StoredBandwidthOverride {
	return &StoredBandwidthOverride{
		OrgId:     o.OrgID,
		MemoryCap: o.MemoryCap,
		ServeCap:  o.ServeCap,
	}
}

func storedToOverride(stored StoredBandwidthOverride) BandwidthOverride {
	return BandwidthOverride{
		OrgID:     stored.OrgId,
		MemoryCap: stored.MemoryCap,
		ServeCap:  stored.ServeCap,
	}
}