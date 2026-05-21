package types

type EmissionPool struct {
	TotalSupply    uint64 `json:"total_supply"`
	DailyMint      uint64 `json:"daily_mint"`
	OperatorShare  uint64 `json:"operator_share"`
	ValidatorShare uint64 `json:"validator_share"`
	Epoch          uint64 `json:"epoch"`
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

type WorkScore struct {
	OperatorID          string  `json:"operator_id"`
	OrgID               string  `json:"org_id"`
	RarityMultiplier   float64 `json:"rarity_multiplier"`
	AvailabilityScore  float64 `json:"availability_score"`
	RetrievalVolume    uint64  `json:"retrieval_volume"`
	StorageScore        float64 `json:"storage_score"`
	RetrievalScore      float64 `json:"retrieval_score"`
	TotalScore          float64 `json:"total_score"`
	Epoch               uint64  `json:"epoch"`
}

func NewWorkScore(operatorID, orgID string, rarityMultiplier, availabilityScore float64, retrievalVolume, epoch uint64) *WorkScore {
	storageScore := 0.3 * availabilityScore
	retrievalScore := 0.7 * float64(retrievalVolume)
	totalScore := rarityMultiplier * (storageScore + retrievalScore)

	return &WorkScore{
		OperatorID:         operatorID,
		OrgID:              orgID,
		RarityMultiplier:   rarityMultiplier,
		AvailabilityScore:  availabilityScore,
		RetrievalVolume:     retrievalVolume,
		StorageScore:       storageScore,
		RetrievalScore:     retrievalScore,
		TotalScore:         totalScore,
		Epoch:              epoch,
	}
}

func (w *WorkScore) Validate() error {
	if w.OperatorID == "" {
		return ErrInvalidOperatorID
	}
	if w.OrgID == "" {
		return ErrInvalidOrgID
	}
	return nil
}

type DailyEmission struct {
	Epoch          uint64 `json:"epoch"`
	TotalEmitted   uint64 `json:"total_emitted"`
	OperatorShare  uint64 `json:"operator_share"`
	ValidatorShare uint64 `json:"validator_share"`
	OperatorRewards map[string]uint64 `json:"operator_rewards"`
	ValidatorRewards map[string]uint64 `json:"validator_rewards"`
}

func NewDailyEmission(epoch, totalEmitted, operatorShare, validatorShare uint64) *DailyEmission {
	return &DailyEmission{
		Epoch:             epoch,
		TotalEmitted:      totalEmitted,
		OperatorShare:     operatorShare,
		ValidatorShare:    validatorShare,
		OperatorRewards:   make(map[string]uint64),
		ValidatorRewards:  make(map[string]uint64),
	}
}

type OperatorReward struct {
	OperatorID string `json:"operator_id"`
	OrgID      string `json:"org_id"`
	Amount     uint64 `json:"amount"`
	Epoch      uint64 `json:"epoch"`
}

func NewOperatorReward(operatorID, orgID string, amount, epoch uint64) *OperatorReward {
	return &OperatorReward{
		OperatorID: operatorID,
		OrgID:      orgID,
		Amount:     amount,
		Epoch:      epoch,
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

type OperatorState struct {
	OperatorID         string  `json:"operator_id"`
	TotalPendingReward uint64  `json:"total_pending_reward"`
	WorkScores         []*WorkScore `json:"work_scores"`
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
	}
}

func DailyEmissionToStored(e *DailyEmission) *StoredDailyEmission {
	if e == nil {
		return nil
	}
	return &StoredDailyEmission{
		Epoch:            e.Epoch,
		TotalEmitted:     e.TotalEmitted,
		OperatorShare:    e.OperatorShare,
		ValidatorShare:   e.ValidatorShare,
		OperatorRewards:  e.OperatorRewards,
		ValidatorRewards: e.ValidatorRewards,
	}
}

func StoredToDailyEmission(s *StoredDailyEmission) *DailyEmission {
	if s == nil {
		return nil
	}
	return &DailyEmission{
		Epoch:            s.Epoch,
		TotalEmitted:     s.TotalEmitted,
		OperatorShare:    s.OperatorShare,
		ValidatorShare:   s.ValidatorShare,
		OperatorRewards:  s.OperatorRewards,
		ValidatorRewards: s.ValidatorRewards,
	}
}

func WorkScoreToStored(w *WorkScore) *StoredWorkScore {
	if w == nil {
		return nil
	}
	return &StoredWorkScore{
		OperatorId:         w.OperatorID,
		OrgId:              w.OrgID,
		RarityMultiplier:   w.RarityMultiplier,
		AvailabilityScore:  w.AvailabilityScore,
		RetrievalVolume:    w.RetrievalVolume,
		StorageScore:       w.StorageScore,
		RetrievalScore:     w.RetrievalScore,
		TotalScore:         w.TotalScore,
		Epoch:              w.Epoch,
	}
}

func StoredToWorkScore(s *StoredWorkScore) *WorkScore {
	if s == nil {
		return nil
	}
	return &WorkScore{
		OperatorID:         s.OperatorId,
		OrgID:              s.OrgId,
		RarityMultiplier:   s.RarityMultiplier,
		AvailabilityScore:  s.AvailabilityScore,
		RetrievalVolume:    s.RetrievalVolume,
		StorageScore:       s.StorageScore,
		RetrievalScore:     s.RetrievalScore,
		TotalScore:         s.TotalScore,
		Epoch:              s.Epoch,
	}
}

func BootstrapCreditToStored(b *BootstrapCredit) *StoredBootstrapCredit {
	if b == nil {
		return nil
	}
	return &StoredBootstrapCredit{
		OperatorId: b.OperatorID,
		Credits:     b.Credits,
		Redeemed:    b.Redeemed,
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