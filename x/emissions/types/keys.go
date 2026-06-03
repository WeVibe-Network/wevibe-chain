package types

type EmissionPool struct {
	TotalSupply    uint64 `json:"total_supply"`
	DailyMint      uint64 `json:"daily_mint"`
	OperatorShare  uint64 `json:"operator_share"`
	ValidatorShare uint64 `json:"validator_share"`
	Epoch          uint64 `json:"epoch"`

	ValidatorPoolRemainingUvibe   uint64 `json:"validator_pool_remaining_uvibe"`
	ContributorPoolRemainingUvibe uint64 `json:"contributor_pool_remaining_uvibe"`
	ContributorRolloverUvibe      uint64 `json:"contributor_rollover_uvibe"`
	StartEpoch                    uint64 `json:"start_epoch"`
	TotalEpochsElapsed            uint64 `json:"total_epochs_elapsed"`
}

func NewEmissionPool(totalSupply, dailyMint, operatorShare, validatorShare, epoch uint64) *EmissionPool {
	return &EmissionPool{
		TotalSupply:    totalSupply,
		DailyMint:      dailyMint,
		OperatorShare:  operatorShare,
		ValidatorShare: validatorShare,
		Epoch:          epoch,
	}
}

func (p *EmissionPool) Validate() error {
	if p.OperatorShare+p.ValidatorShare != 100 {
		return ErrInvalidEmissionPool
	}
	return nil
}

type DailyEmission struct {
	Epoch          uint64 `json:"epoch"`
	TotalEmitted   uint64 `json:"total_emitted"`
	ValidatorShare uint64 `json:"validator_share"`
}

func NewDailyEmission(epoch, totalEmitted, validatorShare uint64) *DailyEmission {
	return &DailyEmission{
		Epoch:          epoch,
		TotalEmitted:   totalEmitted,
		ValidatorShare: validatorShare,
	}
}

type ValidatorReward struct {
	ValidatorID string `json:"validator_id"`
	Amount      uint64 `json:"amount"`
	Epoch       uint64 `json:"epoch"`
}

func NewValidatorReward(validatorID string, amount, epoch uint64) *ValidatorReward {
	return &ValidatorReward{
		ValidatorID: validatorID,
		Amount:      amount,
		Epoch:       epoch,
	}
}

type BootstrapCredit struct {
	OperatorID string `json:"operator_id"`
	Credits    uint64 `json:"credits"`
	Redeemed   uint64 `json:"redeemed"`
}

func NewBootstrapCredit(operatorID string, credits uint64) *BootstrapCredit {
	return &BootstrapCredit{
		OperatorID: operatorID,
		Credits:    credits,
		Redeemed:   0,
	}
}

func (b *BootstrapCredit) CanRedeem(amount uint64) bool {
	return b.Credits-b.Redeemed >= amount
}

func (b *BootstrapCredit) Redeem(amount uint64) error {
	if !b.CanRedeem(amount) {
		return ErrInsufficientBalance
	}
	b.Redeemed += amount
	return nil
}

type AsymmetricGate struct {
	OperatorID       string `json:"operator_id"`
	OrgID            string `json:"org_id"`
	StoragePassed    bool   `json:"storage_passed"`
	RetrievalAllowed bool   `json:"retrieval_allowed"`
	Epoch            uint64 `json:"epoch"`
}

func NewAsymmetricGate(operatorID, orgID string, storagePassed bool, epoch uint64) *AsymmetricGate {
	return &AsymmetricGate{
		OperatorID:       operatorID,
		OrgID:            orgID,
		StoragePassed:    storagePassed,
		RetrievalAllowed: storagePassed,
		Epoch:            epoch,
	}
}

func EmissionPoolToStored(p *EmissionPool) *StoredEmissionPool {
	if p == nil {
		return nil
	}
	return &StoredEmissionPool{
		TotalSupply:    p.TotalSupply,
		DailyMint:      p.DailyMint,
		OperatorShare:  p.OperatorShare,
		ValidatorShare: p.ValidatorShare,
		Epoch:          p.Epoch,

		ValidatorPoolRemainingUvibe:   p.ValidatorPoolRemainingUvibe,
		ContributorPoolRemainingUvibe: p.ContributorPoolRemainingUvibe,
		ContributorRolloverUvibe:      p.ContributorRolloverUvibe,
		StartEpoch:                    p.StartEpoch,
		TotalEpochsElapsed:            p.TotalEpochsElapsed,
	}
}

func StoredToEmissionPool(s *StoredEmissionPool) *EmissionPool {
	if s == nil {
		return nil
	}
	return &EmissionPool{
		TotalSupply:    s.TotalSupply,
		DailyMint:      s.DailyMint,
		OperatorShare:  s.OperatorShare,
		ValidatorShare: s.ValidatorShare,
		Epoch:          s.Epoch,

		ValidatorPoolRemainingUvibe:   s.ValidatorPoolRemainingUvibe,
		ContributorPoolRemainingUvibe: s.ContributorPoolRemainingUvibe,
		ContributorRolloverUvibe:      s.ContributorRolloverUvibe,
		StartEpoch:                    s.StartEpoch,
		TotalEpochsElapsed:            s.TotalEpochsElapsed,
	}
}

func DailyEmissionToStored(e *DailyEmission) *StoredDailyEmission {
	if e == nil {
		return nil
	}
	return &StoredDailyEmission{
		Epoch:          e.Epoch,
		TotalEmitted:   e.TotalEmitted,
		ValidatorShare: e.ValidatorShare,
	}
}

func StoredToDailyEmission(s *StoredDailyEmission) *DailyEmission {
	if s == nil {
		return nil
	}
	return &DailyEmission{
		Epoch:          s.Epoch,
		TotalEmitted:   s.TotalEmitted,
		ValidatorShare: s.ValidatorShare,
	}
}

func BootstrapCreditToStored(b *BootstrapCredit) *StoredBootstrapCredit {
	if b == nil {
		return nil
	}
	return &StoredBootstrapCredit{
		OperatorId: b.OperatorID,
		Credits:    b.Credits,
		Redeemed:   b.Redeemed,
	}
}

func StoredToBootstrapCredit(s *StoredBootstrapCredit) *BootstrapCredit {
	if s == nil {
		return nil
	}
	return &BootstrapCredit{
		OperatorID: s.OperatorId,
		Credits:    s.Credits,
		Redeemed:   s.Redeemed,
	}
}

func AsymmetricGateToStored(g *AsymmetricGate) *StoredAsymmetricGate {
	if g == nil {
		return nil
	}
	return &StoredAsymmetricGate{
		OperatorId:       g.OperatorID,
		OrgId:            g.OrgID,
		StoragePassed:    g.StoragePassed,
		RetrievalAllowed: g.RetrievalAllowed,
		Epoch:            g.Epoch,
	}
}

func StoredToAsymmetricGate(s *StoredAsymmetricGate) *AsymmetricGate {
	if s == nil {
		return nil
	}
	return &AsymmetricGate{
		OperatorID:       s.OperatorId,
		OrgID:            s.OrgId,
		StoragePassed:    s.StoragePassed,
		RetrievalAllowed: s.RetrievalAllowed,
		Epoch:            s.Epoch,
	}
}
