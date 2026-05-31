package keeper

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/gogoproto/proto"
	"github.com/wevibe-network/wevibe-chain/x/emissions/types"
)

type Keeper struct {
	storeService     store.KVStoreService
	logger           log.Logger
	authority        string
	serveKeeper      types.ServeKeeper
	memoryKeeper     types.MemoryKeeper
	orgKeeper        types.OrgKeeper
	reputationKeeper types.ReputationKeeper
}

func NewKeeper(
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,
	serveKeeper types.ServeKeeper,
	memoryKeeper types.MemoryKeeper,
	orgKeeper types.OrgKeeper,
	reputationKeeper types.ReputationKeeper,
) *Keeper {
	return &Keeper{
		storeService:     storeService,
		logger:           logger,
		authority:        authority,
		serveKeeper:      serveKeeper,
		memoryKeeper:     memoryKeeper,
		orgKeeper:        orgKeeper,
		reputationKeeper: reputationKeeper,
	}
}

func (k *Keeper) getStore(ctx context.Context) store.KVStore {
	return k.storeService.OpenKVStore(ctx)
}

const (
	KeyPrefixPool          = "pool/"
	KeyPrefixEmission      = "emission/"
	KeyPrefixOpReward      = "opreward/"
	KeyPrefixValReward     = "valreward/"
	KeyPrefixWorkScore     = "workscore/"
	KeyPrefixGate          = "gate/"
	KeyPrefixBootstrap     = "bootstrap/"
	KeyPrefixBootstrapPool = "bootstrappool/"
	KeyPrefixContribReward = "contribreward/"
)

func poolKey() []byte {
	return []byte(KeyPrefixPool)
}

func emissionKey(epoch uint64) []byte {
	return []byte(fmt.Sprintf("%s%s", KeyPrefixEmission, strconv.FormatUint(epoch, 10)))
}

func opRewardKey(operatorID string) []byte {
	return []byte(KeyPrefixOpReward + operatorID)
}

func valRewardKey(validatorID string) []byte {
	return []byte(KeyPrefixValReward + validatorID)
}

func workScoreKey(operatorID, orgID string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("%s%s/%s/%s", KeyPrefixWorkScore, operatorID, orgID, strconv.FormatUint(epoch, 10)))
}

func gateKey(operatorID, orgID string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("%s%s/%s/%s", KeyPrefixGate, operatorID, orgID, strconv.FormatUint(epoch, 10)))
}

func bootstrapKey(operatorID string) []byte {
	return []byte(KeyPrefixBootstrap + operatorID)
}

func bootstrapPoolKey() []byte {
	return []byte(KeyPrefixBootstrapPool)
}

func contribRewardKey(addr string) []byte {
	return []byte(KeyPrefixContribReward + addr)
}

const ParamsKey = "params"

func (k *Keeper) SetParams(ctx context.Context, params types.Params) error {
	store := k.getStore(ctx)
	bz, err := proto.Marshal(&params)
	if err != nil {
		return fmt.Errorf("marshal params: %w", err)
	}
	return store.Set([]byte(ParamsKey), bz)
}

func (k *Keeper) GetParams(ctx context.Context) (types.Params, error) {
	store := k.getStore(ctx)
	bz, err := store.Get([]byte(ParamsKey))
	if err != nil {
		return types.Params{}, fmt.Errorf("get params: %w", err)
	}
	if bz == nil {
		return types.DefaultParams(), nil
	}
	var params types.Params
	if err := proto.Unmarshal(bz, &params); err != nil {
		return types.Params{}, fmt.Errorf("unmarshal params: %w", err)
	}
	return params, nil
}

func (k *Keeper) GetEmissionPool(ctx context.Context) (*types.EmissionPool, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(poolKey())
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrNoEmissionPool
	}

	var stored types.StoredEmissionPool
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal emission pool: %w", err)
	}
	return &types.EmissionPool{
		TotalSupply:    stored.TotalSupply,
		DailyMint:      stored.DailyMint,
		OperatorShare:  stored.OperatorShare,
		ValidatorShare: stored.ValidatorShare,
		Epoch:          stored.Epoch,

		ValidatorPoolRemainingUvibe:   stored.ValidatorPoolRemainingUvibe,
		ContributorPoolRemainingUvibe: stored.ContributorPoolRemainingUvibe,
		ContributorRolloverUvibe:      stored.ContributorRolloverUvibe,
		StartEpoch:                    stored.StartEpoch,
		TotalEpochsElapsed:            stored.TotalEpochsElapsed,
	}, nil
}

func (k *Keeper) SetEmissionPool(ctx context.Context, pool *types.EmissionPool) error {
	if pool == nil {
		return types.ErrInvalidEmissionPool
	}
	if err := pool.Validate(); err != nil {
		return err
	}

	stored := types.StoredEmissionPool{
		TotalSupply:    pool.TotalSupply,
		DailyMint:      pool.DailyMint,
		OperatorShare:  pool.OperatorShare,
		ValidatorShare: pool.ValidatorShare,
		Epoch:          pool.Epoch,

		ValidatorPoolRemainingUvibe:   pool.ValidatorPoolRemainingUvibe,
		ContributorPoolRemainingUvibe: pool.ContributorPoolRemainingUvibe,
		ContributorRolloverUvibe:      pool.ContributorRolloverUvibe,
		StartEpoch:                    pool.StartEpoch,
		TotalEpochsElapsed:            pool.TotalEpochsElapsed,
	}
	bz, err := proto.Marshal(&stored)
	if err != nil {
		return fmt.Errorf("marshal emission pool: %w", err)
	}

	store := k.getStore(ctx)
	store.Set(poolKey(), bz)

	k.logger.Info("emission pool set",
		"total_supply", pool.TotalSupply,
		"daily_mint", pool.DailyMint,
		"operator_share", pool.OperatorShare,
		"validator_share", pool.ValidatorShare,
	)
	return nil
}

func (k *Keeper) MintDailyEmission(ctx context.Context, epoch uint64) (*types.DailyEmission, error) {
	pool, err := k.GetEmissionPool(ctx)
	if err != nil {
		return nil, types.ErrNoEmissionPool
	}

	if epoch <= pool.Epoch {
		return nil, fmt.Errorf("epoch must be greater than last emission epoch")
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	remainingEpochs := uint64(1)
	if params.ScheduleDurationDays > pool.TotalEpochsElapsed {
		remainingEpochs = params.ScheduleDurationDays - pool.TotalEpochsElapsed
	}

	validatorEmission := uint64(0)
	if pool.ValidatorPoolRemainingUvibe > 0 {
		validatorEmission = pool.ValidatorPoolRemainingUvibe / remainingEpochs
		if validatorEmission > pool.ValidatorPoolRemainingUvibe {
			validatorEmission = pool.ValidatorPoolRemainingUvibe
		}
		pool.ValidatorPoolRemainingUvibe -= validatorEmission
	}

	contributorBudget := uint64(0)
	if pool.ContributorPoolRemainingUvibe > 0 {
		contributorBudget = pool.ContributorPoolRemainingUvibe / remainingEpochs
		epochCap := params.ContributorAnnualCapUvibe / types.EpochsPerYear
		if epochCap > 0 && contributorBudget > epochCap {
			contributorBudget = epochCap
		}
		if contributorBudget > pool.ContributorPoolRemainingUvibe {
			contributorBudget = pool.ContributorPoolRemainingUvibe
		}
		pool.ContributorPoolRemainingUvibe -= contributorBudget
	}

	counts, err := k.memoryKeeper.GetContributorsWithApprovalsInEpoch(ctx, epoch)
	if err != nil {
		k.logger.Info("contributor approvals query failed; treating as none", "epoch", epoch, "error", err)
		counts = map[string]uint64{}
	}

	qualifying := make([]string, 0, len(counts))
	for addr, count := range counts {
		if addr != "" && count >= params.ContributorQualifyThreshold {
			qualifying = append(qualifying, addr)
		}
	}
	sort.Strings(qualifying)

	distributedToContributors := uint64(0)
	if len(qualifying) == 0 {
		pool.ContributorRolloverUvibe += contributorBudget
	} else {
		total := contributorBudget + pool.ContributorRolloverUvibe
		perContributor := total / uint64(len(qualifying))
		remainder := total % uint64(len(qualifying))
		for _, addr := range qualifying {
			if perContributor > 0 {
				if err := k.AddContributorReward(ctx, addr, perContributor); err != nil {
					return nil, err
				}
			}
		}
		distributedToContributors = perContributor * uint64(len(qualifying))
		pool.ContributorRolloverUvibe = remainder // carry integer remainder forward to avoid token loss
	}

	totalEmitted := validatorEmission + distributedToContributors

	emission := types.NewDailyEmission(epoch, totalEmitted, 0, validatorEmission)

	storedEmission := types.StoredDailyEmission{
		Epoch:          emission.Epoch,
		TotalEmitted:   emission.TotalEmitted,
		OperatorShare:  0,
		ValidatorShare: validatorEmission,
	}
	emissionBz, err := proto.Marshal(&storedEmission)
	if err != nil {
		return nil, fmt.Errorf("marshal daily emission: %w", err)
	}

	store := k.getStore(ctx)
	store.Set(emissionKey(epoch), emissionBz)

	if pool.StartEpoch == 0 && pool.TotalEpochsElapsed == 0 {
		pool.StartEpoch = epoch
	}
	pool.TotalEpochsElapsed++
	pool.Epoch = epoch
	pool.TotalSupply += totalEmitted
	if err := k.SetEmissionPool(ctx, pool); err != nil {
		return nil, err
	}

	k.logger.Info("epoch emission minted",
		"epoch", epoch,
		"validator_emission", validatorEmission,
		"contributor_distributed", distributedToContributors,
		"qualifying", len(qualifying),
		"rollover", pool.ContributorRolloverUvibe,
	)

	return emission, nil
}

func (k *Keeper) AddContributorReward(ctx context.Context, addr string, amount uint64) error {
	store := k.getStore(ctx)
	current := uint64(0)
	bz, err := store.Get(contribRewardKey(addr))
	if err != nil {
		return err
	}
	if bz != nil {
		current, err = strconv.ParseUint(string(bz), 10, 64)
		if err != nil {
			return fmt.Errorf("parse contributor reward: %w", err)
		}
	}
	current += amount
	return store.Set(contribRewardKey(addr), []byte(strconv.FormatUint(current, 10)))
}

func (k *Keeper) GetContributorReward(ctx context.Context, addr string) (uint64, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(contribRewardKey(addr))
	if err != nil {
		return 0, err
	}
	if bz == nil {
		return 0, nil
	}
	amount, err := strconv.ParseUint(string(bz), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse contributor reward: %w", err)
	}
	return amount, nil
}

func (k *Keeper) GetAllContributorRewards(ctx context.Context) (map[string]uint64, error) {
	store := k.getStore(ctx)
	prefix := []byte(KeyPrefixContribReward)
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	rewards := make(map[string]uint64)
	for ; iter.Valid(); iter.Next() {
		amount, err := strconv.ParseUint(string(iter.Value()), 10, 64)
		if err != nil {
			continue
		}
		key := string(iter.Key())
		addr := key[len(KeyPrefixContribReward):]
		rewards[addr] = amount
	}
	return rewards, nil
}

func (k *Keeper) GetDailyEmission(ctx context.Context, epoch uint64) (*types.DailyEmission, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(emissionKey(epoch))
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, fmt.Errorf("daily emission not found for epoch")
	}

	var stored types.StoredDailyEmission
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal daily emission: %w", err)
	}
	return types.StoredToDailyEmission(&stored), nil
}

func (k *Keeper) SetAvgRetrievalVolume(ctx context.Context, avg uint64) {
	store := k.getStore(ctx)
	store.Set([]byte("avgresrievalvolume"), []byte(strconv.FormatUint(avg, 10)))
}

func (k *Keeper) ComputeWorkScore(ctx context.Context, operatorID, orgID string, rarityMultiplier, availabilityScore float64, retrievalVolume, epoch uint64) (*types.WorkScore, error) {
	if operatorID == "" {
		return nil, types.ErrInvalidOperatorID
	}
	if orgID == "" {
		return nil, types.ErrInvalidOrgID
	}

	score := types.NewWorkScore(operatorID, orgID, rarityMultiplier, availabilityScore, retrievalVolume, epoch)

	storedScore := types.StoredWorkScore{
		OperatorId:        score.OperatorID,
		OrgId:             score.OrgID,
		RarityMultiplier:  score.RarityMultiplier,
		AvailabilityScore: score.AvailabilityScore,
		RetrievalVolume:   score.RetrievalVolume,
		StorageScore:      score.StorageScore,
		RetrievalScore:    score.RetrievalScore,
		TotalScore:        score.TotalScore,
		Epoch:             score.Epoch,
	}
	bz, err := proto.Marshal(&storedScore)
	if err != nil {
		return nil, fmt.Errorf("marshal work score: %w", err)
	}

	store := k.getStore(ctx)
	if err := store.Set(workScoreKey(operatorID, orgID, epoch), bz); err != nil {
		return nil, err
	}

	k.logger.Info("work score computed",
		"operator_id", operatorID,
		"org_id", orgID,
		"rarity_multiplier", rarityMultiplier,
		"availability_score", availabilityScore,
		"retrieval_volume", retrievalVolume,
		"total_score", score.TotalScore,
	)

	return score, nil
}

func (k *Keeper) GetWorkScore(ctx context.Context, operatorID, orgID string, epoch uint64) (*types.WorkScore, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(workScoreKey(operatorID, orgID, epoch))
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrNoWorkScore
	}

	var stored types.StoredWorkScore
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal work score: %w", err)
	}
	return &types.WorkScore{
		OperatorID:        stored.OperatorId,
		OrgID:             stored.OrgId,
		RarityMultiplier:  stored.RarityMultiplier,
		AvailabilityScore: stored.AvailabilityScore,
		RetrievalVolume:   stored.RetrievalVolume,
		StorageScore:      stored.StorageScore,
		RetrievalScore:    stored.RetrievalScore,
		TotalScore:        stored.TotalScore,
		Epoch:             stored.Epoch,
	}, nil
}

func (k *Keeper) GetOperatorWorkScores(ctx context.Context, operatorID string, epoch uint64) ([]*types.WorkScore, error) {
	store := k.getStore(ctx)
	prefix := []byte(fmt.Sprintf("%s%s/", KeyPrefixWorkScore, operatorID))
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var scores []*types.WorkScore
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredWorkScore
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		if stored.Epoch == epoch {
			scores = append(scores, &types.WorkScore{
				OperatorID:        stored.OperatorId,
				OrgID:             stored.OrgId,
				RarityMultiplier:  stored.RarityMultiplier,
				AvailabilityScore: stored.AvailabilityScore,
				RetrievalVolume:   stored.RetrievalVolume,
				StorageScore:      stored.StorageScore,
				RetrievalScore:    stored.RetrievalScore,
				TotalScore:        stored.TotalScore,
				Epoch:             stored.Epoch,
			})
		}
	}
	return scores, nil
}

func (k *Keeper) ComputeTotalWorkScore(ctx context.Context, operatorID string, epoch uint64) (float64, error) {
	scores, err := k.GetOperatorWorkScores(ctx, operatorID, epoch)
	if err != nil {
		return 0, err
	}

	var total float64
	for _, score := range scores {
		total += score.TotalScore
	}
	return total, nil
}

func (k *Keeper) DistributeOperatorRewards(ctx context.Context, rewards map[string]uint64, epoch uint64) error {
	_, err := k.GetEmissionPool(ctx)
	if err != nil {
		return types.ErrNoEmissionPool
	}

	emission, err := k.GetDailyEmission(ctx, epoch)
	if err != nil {
		return fmt.Errorf("daily emission not found for epoch")
	}

	store := k.getStore(ctx)
	for opID, amount := range rewards {
		bz := []byte(strconv.FormatUint(amount, 10))
		store.Set(opRewardKey(opID), bz)

		if emission.OperatorRewards == nil {
			emission.OperatorRewards = make(map[string]uint64)
		}
		emission.OperatorRewards[opID] = amount

		k.logger.Info("operator reward distributed",
			"operator_id", opID,
			"amount", amount,
			"epoch", epoch,
		)
	}

	emissionBz, err := proto.Marshal(types.DailyEmissionToStored(emission))
	if err != nil {
		return fmt.Errorf("marshal emission: %w", err)
	}
	if err := store.Set(emissionKey(epoch), emissionBz); err != nil {
		return err
	}

	return nil
}

func (k *Keeper) DistributeValidatorRewards(ctx context.Context, rewards map[string]uint64, epoch uint64) error {
	_, err := k.GetEmissionPool(ctx)
	if err != nil {
		return types.ErrNoEmissionPool
	}

	emission, err := k.GetDailyEmission(ctx, epoch)
	if err != nil {
		return fmt.Errorf("daily emission not found for epoch")
	}

	store := k.getStore(ctx)
	for valID, amount := range rewards {
		bz := []byte(strconv.FormatUint(amount, 10))
		store.Set(valRewardKey(valID), bz)

		if emission.ValidatorRewards == nil {
			emission.ValidatorRewards = make(map[string]uint64)
		}
		emission.ValidatorRewards[valID] = amount

		k.logger.Info("validator reward distributed",
			"validator_id", valID,
			"amount", amount,
			"epoch", epoch,
		)
	}

	emissionBz, err := proto.Marshal(types.DailyEmissionToStored(emission))
	if err != nil {
		return fmt.Errorf("marshal emission: %w", err)
	}
	if err := store.Set(emissionKey(epoch), emissionBz); err != nil {
		return err
	}

	return nil
}

func (k *Keeper) GetOperatorReward(ctx context.Context, operatorID string) (uint64, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(opRewardKey(operatorID))
	if err != nil {
		return 0, err
	}
	if bz == nil {
		return 0, types.ErrNoPendingReward
	}

	amount, err := strconv.ParseUint(string(bz), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("unmarshal operator reward: %w", err)
	}
	return amount, nil
}

func (k *Keeper) SetAsymmetricGate(ctx context.Context, gate *types.AsymmetricGate) error {
	if gate == nil {
		return fmt.Errorf("invalid asymmetric gate")
	}

	bz, err := proto.Marshal(types.AsymmetricGateToStored(gate))
	if err != nil {
		return fmt.Errorf("marshal asymmetric gate: %w", err)
	}

	store := k.getStore(ctx)
	if err := store.Set(gateKey(gate.OperatorID, gate.OrgID, gate.Epoch), bz); err != nil {
		return err
	}

	k.logger.Info("asymmetric gate set",
		"operator_id", gate.OperatorID,
		"org_id", gate.OrgID,
		"storage_passed", gate.StoragePassed,
		"retrieval_allowed", gate.RetrievalAllowed,
		"epoch", gate.Epoch,
	)

	return nil
}

func (k *Keeper) GetAsymmetricGate(ctx context.Context, operatorID, orgID string, epoch uint64) (*types.AsymmetricGate, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(gateKey(operatorID, orgID, epoch))
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return types.NewAsymmetricGate(operatorID, orgID, true, epoch), nil
	}

	var stored types.StoredAsymmetricGate
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal asymmetric gate: %w", err)
	}
	return types.StoredToAsymmetricGate(&stored), nil
}

func (k *Keeper) CheckRetrievalAllowed(ctx context.Context, operatorID, orgID string, epoch uint64) (bool, error) {
	gate, err := k.GetAsymmetricGate(ctx, operatorID, orgID, epoch)
	if err != nil {
		return false, err
	}

	if !gate.StoragePassed {
		k.logger.Info("retrieval blocked by asymmetric gate",
			"operator_id", operatorID,
			"org_id", orgID,
			"epoch", epoch,
		)
	}

	return gate.RetrievalAllowed, nil
}

func (k *Keeper) SetBootstrapCredits(ctx context.Context, credits uint64) {
	store := k.getStore(ctx)
	if err := store.Set(bootstrapPoolKey(), []byte(strconv.FormatUint(credits, 10))); err != nil {
		return
	}
	k.logger.Info("bootstrap credits pool set", "credits", credits)
}

func (k *Keeper) GetBootstrapCredits(ctx context.Context) (uint64, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(bootstrapPoolKey())
	if err != nil {
		return 0, err
	}
	if bz == nil {
		return 0, types.ErrNoBootstrapPool
	}

	credits, err := strconv.ParseUint(string(bz), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse bootstrap credits: %w", err)
	}
	return credits, nil
}

func (k *Keeper) RedeemBootstrapCredit(ctx context.Context, operatorID string, amount uint64) error {
	store := k.getStore(ctx)

	currentEpoch := k.getBootstrapExpiry(ctx)
	if currentEpoch == 0 {
		return types.ErrBootstrapExpired
	}

	bz, err := store.Get(bootstrapKey(operatorID))
	if err != nil {
		return err
	}
	var credit *types.BootstrapCredit
	if bz == nil {
		credit = types.NewBootstrapCredit(operatorID, 0)
	} else {
		var existing types.StoredBootstrapCredit
		if err := proto.Unmarshal(bz, &existing); err != nil {
			return fmt.Errorf("unmarshal bootstrap credit: %w", err)
		}
		credit = types.StoredToBootstrapCredit(&existing)
	}

	if err := credit.Redeem(amount); err != nil {
		return err
	}

	creditBz, err := proto.Marshal(types.BootstrapCreditToStored(credit))
	if err != nil {
		return fmt.Errorf("marshal bootstrap credit: %w", err)
	}
	if err := store.Set(bootstrapKey(operatorID), creditBz); err != nil {
		return err
	}

	bootstrapPool, _ := k.GetBootstrapCredits(ctx)
	if err := store.Set(bootstrapPoolKey(), []byte(strconv.FormatUint(bootstrapPool-amount, 10))); err != nil {
		return err
	}

	k.logger.Info("bootstrap credit redeemed",
		"operator_id", operatorID,
		"amount", amount,
		"remaining", credit.Credits-credit.Redeemed,
	)

	return nil
}

func (k *Keeper) GetBootstrapCredit(ctx context.Context, operatorID string) (*types.BootstrapCredit, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(bootstrapKey(operatorID))
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return types.NewBootstrapCredit(operatorID, 0), nil
	}

	var stored types.StoredBootstrapCredit
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal bootstrap credit: %w", err)
	}
	return types.StoredToBootstrapCredit(&stored), nil
}

func (k *Keeper) AddBootstrapCredit(ctx context.Context, operatorID string, credits uint64) {
	store := k.getStore(ctx)
	bz, err := store.Get(bootstrapKey(operatorID))
	if err != nil {
		return
	}

	var credit *types.BootstrapCredit
	if bz == nil {
		credit = types.NewBootstrapCredit(operatorID, credits)
	} else {
		var existing types.StoredBootstrapCredit
		if err := proto.Unmarshal(bz, &existing); err != nil {
			return
		}
		ecredit := types.StoredToBootstrapCredit(&existing)
		ecredit.Credits += credits
		credit = ecredit
	}

	creditBz, _ := proto.Marshal(types.BootstrapCreditToStored(credit))
	store.Set(bootstrapKey(operatorID), creditBz)
}

func (k *Keeper) SetBootstrapExpiry(ctx context.Context, expiry uint64) {
	store := k.getStore(ctx)
	if err := store.Set([]byte("bootstrapExpiry"), []byte(strconv.FormatUint(expiry, 10))); err != nil {
		return
	}
}

func (k *Keeper) getBootstrapExpiry(ctx context.Context) uint64 {
	store := k.getStore(ctx)
	bz, err := store.Get([]byte("bootstrapExpiry"))
	if err != nil {
		return 0
	}
	if bz == nil {
		return 0
	}
	expiry, _ := strconv.ParseUint(string(bz), 10, 64)
	return expiry
}

func (k *Keeper) IsBootstrapExpired(ctx context.Context, currentEpoch uint64) bool {
	return currentEpoch > k.getBootstrapExpiry(ctx)
}

func (k *Keeper) GetAllOperatorRewards(ctx context.Context) (map[string]uint64, error) {
	store := k.getStore(ctx)
	prefix := []byte(KeyPrefixOpReward)
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	rewards := make(map[string]uint64)
	for ; iter.Valid(); iter.Next() {
		amount, err := strconv.ParseUint(string(iter.Value()), 10, 64)
		if err != nil {
			continue
		}
		key := string(iter.Key())
		opID := key[len(KeyPrefixOpReward):]
		rewards[opID] = amount
	}
	return rewards, nil
}

func (k *Keeper) GetAllValidatorRewards(ctx context.Context) (map[string]uint64, error) {
	store := k.getStore(ctx)
	prefix := []byte(KeyPrefixValReward)
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	rewards := make(map[string]uint64)
	for ; iter.Valid(); iter.Next() {
		amount, err := strconv.ParseUint(string(iter.Value()), 10, 64)
		if err != nil {
			continue
		}
		key := string(iter.Key())
		valID := key[len(KeyPrefixValReward):]
		rewards[valID] = amount
	}
	return rewards, nil
}

func (k *Keeper) GetAllWorkScores(ctx context.Context, epoch uint64) ([]*types.WorkScore, error) {
	store := k.getStore(ctx)
	prefix := []byte(KeyPrefixWorkScore)
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var scores []*types.WorkScore
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredWorkScore
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		score := types.StoredToWorkScore(&stored)
		if score.Epoch == epoch {
			scores = append(scores, score)
		}
	}
	return scores, nil
}

func (k *Keeper) ComputeRarityMultiplier(ctx context.Context, orgID string, totalOperators, activeOperators uint64) float64 {
	if totalOperators == 0 {
		return 1.0
	}

	ratio := float64(totalOperators) / float64(activeOperators)
	if ratio < 1.0 {
		ratio = 1.0
	}

	multiplier := 1.0 + (ratio-1.0)*0.5
	if multiplier > 3.0 {
		multiplier = 3.0
	}

	return multiplier
}

func (k *Keeper) InitGenesis(ctx context.Context, state *types.GenesisState) error {
	store := k.getStore(ctx)

	if state.EmissionPool != nil {
		bz, err := proto.Marshal(types.EmissionPoolToStored(state.EmissionPool))
		if err != nil {
			return err
		}
		if err := store.Set(poolKey(), bz); err != nil {
			return err
		}
	}

	for _, emission := range state.DailyEmissions {
		bz, err := proto.Marshal(types.DailyEmissionToStored(emission))
		if err != nil {
			return err
		}
		if err := store.Set(emissionKey(emission.Epoch), bz); err != nil {
			return err
		}
	}

	for _, reward := range state.OperatorRewards {
		bz := []byte(strconv.FormatUint(reward.Amount, 10))
		if err := store.Set(opRewardKey(reward.OperatorID), bz); err != nil {
			return err
		}
	}

	for _, reward := range state.ValidatorRewards {
		bz := []byte(strconv.FormatUint(reward.Amount, 10))
		if err := store.Set(valRewardKey(reward.ValidatorID), bz); err != nil {
			return err
		}
	}

	for _, credit := range state.BootstrapCredits {
		bz, err := proto.Marshal(types.BootstrapCreditToStored(credit))
		if err != nil {
			return err
		}
		if err := store.Set(bootstrapKey(credit.OperatorID), bz); err != nil {
			return err
		}
	}

	for _, score := range state.WorkScores {
		bz, err := proto.Marshal(types.WorkScoreToStored(score))
		if err != nil {
			return err
		}
		if err := store.Set(workScoreKey(score.OperatorID, score.OrgID, score.Epoch), bz); err != nil {
			return err
		}
	}

	for _, gate := range state.AsymmetricGates {
		bz, err := proto.Marshal(types.AsymmetricGateToStored(gate))
		if err != nil {
			return err
		}
		if err := store.Set(gateKey(gate.OperatorID, gate.OrgID, gate.Epoch), bz); err != nil {
			return err
		}
	}

	if state.BootstrapExpiry > 0 {
		if err := store.Set([]byte("bootstrapExpiry"), []byte(strconv.FormatUint(state.BootstrapExpiry, 10))); err != nil {
			return err
		}
	}

	return nil
}

func (k *Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	store := k.getStore(ctx)

	emissionPool := new(types.EmissionPool)
	bz, err := store.Get(poolKey())
	if err != nil {
		return nil, err
	}
	if bz != nil {
		var stored types.StoredEmissionPool
		if err := proto.Unmarshal(bz, &stored); err != nil {
			return nil, err
		}
		emissionPool = types.StoredToEmissionPool(&stored)
	}

	var dailyEmissions []*types.DailyEmission
	{
		prefix := []byte(KeyPrefixEmission)
		iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
		if err != nil {
			return nil, err
		}
		defer iter.Close()
		for ; iter.Valid(); iter.Next() {
			var stored types.StoredDailyEmission
			if err := proto.Unmarshal(iter.Value(), &stored); err == nil {
				dailyEmissions = append(dailyEmissions, types.StoredToDailyEmission(&stored))
			}
		}
	}

	var operatorRewards []*types.OperatorReward
	{
		prefix := []byte(KeyPrefixOpReward)
		iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
		if err != nil {
			return nil, err
		}
		defer iter.Close()
		for ; iter.Valid(); iter.Next() {
			amount, err := strconv.ParseUint(string(iter.Value()), 10, 64)
			if err == nil {
				key := string(iter.Key())
				opID := key[len(KeyPrefixOpReward):]
				operatorRewards = append(operatorRewards, &types.OperatorReward{OperatorID: opID, Amount: amount})
			}
		}
	}

	var validatorRewards []*types.ValidatorReward
	{
		prefix := []byte(KeyPrefixValReward)
		iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
		if err != nil {
			return nil, err
		}
		defer iter.Close()
		for ; iter.Valid(); iter.Next() {
			amount, err := strconv.ParseUint(string(iter.Value()), 10, 64)
			if err == nil {
				key := string(iter.Key())
				valID := key[len(KeyPrefixValReward):]
				validatorRewards = append(validatorRewards, &types.ValidatorReward{ValidatorID: valID, Amount: amount})
			}
		}
	}

	var bootstrapCredits []*types.BootstrapCredit
	{
		prefix := []byte(KeyPrefixBootstrap)
		iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
		if err != nil {
			return nil, err
		}
		defer iter.Close()
		for ; iter.Valid(); iter.Next() {
			var stored types.StoredBootstrapCredit
			if err := proto.Unmarshal(iter.Value(), &stored); err == nil {
				bootstrapCredits = append(bootstrapCredits, types.StoredToBootstrapCredit(&stored))
			}
		}
	}

	var workScores []*types.WorkScore
	{
		prefix := []byte(KeyPrefixWorkScore)
		iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
		if err != nil {
			return nil, err
		}
		defer iter.Close()
		for ; iter.Valid(); iter.Next() {
			var stored types.StoredWorkScore
			if err := proto.Unmarshal(iter.Value(), &stored); err == nil {
				workScores = append(workScores, types.StoredToWorkScore(&stored))
			}
		}
	}

	var asymmetricGates []*types.AsymmetricGate
	{
		prefix := []byte(KeyPrefixGate)
		iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
		if err != nil {
			return nil, err
		}
		defer iter.Close()
		for ; iter.Valid(); iter.Next() {
			var stored types.StoredAsymmetricGate
			if err := proto.Unmarshal(iter.Value(), &stored); err == nil {
				asymmetricGates = append(asymmetricGates, types.StoredToAsymmetricGate(&stored))
			}
		}
	}

	bootstrapExpiry := k.getBootstrapExpiry(ctx)

	return types.NewGenesisState(
		emissionPool,
		dailyEmissions,
		operatorRewards,
		validatorRewards,
		bootstrapCredits,
		workScores,
		asymmetricGates,
		bootstrapExpiry,
	), nil
}

func (k *Keeper) Logger(ctx context.Context) log.Logger {
	return k.logger.With("module", "x/emissions")
}
