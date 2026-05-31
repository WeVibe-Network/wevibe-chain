package types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wevibe-network/wevibe-chain/x/emissions/types"
)

// ---------------------------------------------------------------------------
// MsgMintDailyEmission.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgMintDailyEmission_ValidateBasic_Valid(t *testing.T) {
	msg := &types.MsgMintDailyEmission{Authority: "cosmos1abc", Epoch: 1}
	require.NoError(t, msg.ValidateBasic())
}

func TestMsgMintDailyEmission_ValidateBasic_EmptyAuthority(t *testing.T) {
	msg := &types.MsgMintDailyEmission{Authority: "", Epoch: 1}
	require.Error(t, msg.ValidateBasic())
}

func TestMsgMintDailyEmission_ValidateBasic_ZeroEpoch(t *testing.T) {
	msg := &types.MsgMintDailyEmission{Authority: "cosmos1abc", Epoch: 0}
	require.Error(t, msg.ValidateBasic())
}

// ---------------------------------------------------------------------------
// MsgDistributeOperatorRewards.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgDistributeOperatorRewards_ValidateBasic_Valid(t *testing.T) {
	msg := &types.MsgDistributeOperatorRewards{
		Signer:  "cosmos1abc",
		Rewards: []*types.OperatorRewardEntry{{OperatorId: "op1", Amount: 100}},
		Epoch:   1,
	}
	require.NoError(t, msg.ValidateBasic())
}

func TestMsgDistributeOperatorRewards_ValidateBasic_EmptySigner(t *testing.T) {
	msg := &types.MsgDistributeOperatorRewards{
		Signer:  "",
		Rewards: []*types.OperatorRewardEntry{{OperatorId: "op1", Amount: 100}},
		Epoch:   1,
	}
	require.Error(t, msg.ValidateBasic())
}

func TestMsgDistributeOperatorRewards_ValidateBasic_EmptyRewards(t *testing.T) {
	// nil rewards slice (len == 0)
	msg := &types.MsgDistributeOperatorRewards{Signer: "cosmos1abc", Rewards: nil, Epoch: 1}
	require.Error(t, msg.ValidateBasic())

	// explicitly empty (non-nil) rewards slice (len == 0)
	msg2 := &types.MsgDistributeOperatorRewards{
		Signer:  "cosmos1abc",
		Rewards: []*types.OperatorRewardEntry{},
		Epoch:   1,
	}
	require.Error(t, msg2.ValidateBasic())
}

// ---------------------------------------------------------------------------
// MsgUpdateParams.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgUpdateParams_ValidateBasic_Valid(t *testing.T) {
	p := types.DefaultParams()
	msg := &types.MsgUpdateParams{Authority: "cosmos1abc", Params: &p}
	require.NoError(t, msg.ValidateBasic())
}

func TestMsgUpdateParams_ValidateBasic_EmptyAuthority(t *testing.T) {
	p := types.DefaultParams()
	msg := &types.MsgUpdateParams{Authority: "", Params: &p}
	require.Error(t, msg.ValidateBasic())
}

func TestMsgUpdateParams_ValidateBasic_InvalidParams(t *testing.T) {
	// authority set, but params fail Params.Validate (shares do not sum to 100)
	p := types.DefaultParams()
	p.OperatorSharePercent = 50
	p.ValidatorSharePercent = 49
	msg := &types.MsgUpdateParams{Authority: "cosmos1abc", Params: &p}
	require.Error(t, msg.ValidateBasic())
}

// ---------------------------------------------------------------------------
// Params.Validate / DefaultParams
// ---------------------------------------------------------------------------

func TestDefaultParams_IsValid(t *testing.T) {
	p := types.DefaultParams()
	require.NoError(t, p.Validate())

	require.Equal(t, uint64(1000000000), p.DailyMintAmount)
	require.Equal(t, uint64(80), p.OperatorSharePercent)
	require.Equal(t, uint64(20), p.ValidatorSharePercent)
	require.Equal(t, uint64(30), p.StorageWeightPercent)
	require.Equal(t, uint64(70), p.RetrievalWeightPercent)
	require.Equal(t, "3.0", p.RarityMultiplierCap)
	require.Equal(t, uint64(365), p.BootstrapDurationEpochs)
}

func TestParams_Validate_SharesSumTo100(t *testing.T) {
	p := types.Params{
		OperatorSharePercent:   60,
		ValidatorSharePercent:  40,
		StorageWeightPercent:   30,
		RetrievalWeightPercent: 70,
	}
	require.NoError(t, p.Validate())
}

func TestParams_Validate_OperatorValidatorShareNot100(t *testing.T) {
	p := types.Params{
		OperatorSharePercent:   80,
		ValidatorSharePercent:  19, // sums to 99
		StorageWeightPercent:   30,
		RetrievalWeightPercent: 70,
	}
	require.Error(t, p.Validate())
}

func TestParams_Validate_StorageRetrievalWeightNot100(t *testing.T) {
	p := types.Params{
		OperatorSharePercent:   80,
		ValidatorSharePercent:  20,
		StorageWeightPercent:   30,
		RetrievalWeightPercent: 71, // sums to 101
	}
	require.Error(t, p.Validate())
}

func TestParams_Validate_AllZero(t *testing.T) {
	// 0+0 != 100, so this must fail on the first branch.
	p := types.Params{}
	require.Error(t, p.Validate())
}

func TestParams_Validate_BoundaryShares(t *testing.T) {
	// 100/0 split is a valid boundary for both pairs.
	p := types.Params{
		OperatorSharePercent:   100,
		ValidatorSharePercent:  0,
		StorageWeightPercent:   0,
		RetrievalWeightPercent: 100,
	}
	require.NoError(t, p.Validate())
}

// ---------------------------------------------------------------------------
// EmissionPool helpers + Validate
// ---------------------------------------------------------------------------

func TestNewEmissionPool(t *testing.T) {
	p := types.NewEmissionPool(1000, 10, 80, 20, 5)
	require.Equal(t, uint64(1000), p.TotalSupply)
	require.Equal(t, uint64(10), p.DailyMint)
	require.Equal(t, uint64(80), p.OperatorShare)
	require.Equal(t, uint64(20), p.ValidatorShare)
	require.Equal(t, uint64(5), p.Epoch)
}

func TestEmissionPool_Validate_Valid(t *testing.T) {
	p := types.NewEmissionPool(1000, 10, 80, 20, 0)
	require.NoError(t, p.Validate())
}

func TestEmissionPool_Validate_SharesNot100(t *testing.T) {
	p := types.NewEmissionPool(1000, 10, 80, 21, 0)
	require.ErrorIs(t, p.Validate(), types.ErrInvalidEmissionPool)
}

func TestEmissionPool_Validate_ZeroShares(t *testing.T) {
	p := types.NewEmissionPool(0, 0, 0, 0, 0)
	require.ErrorIs(t, p.Validate(), types.ErrInvalidEmissionPool)
}

// ---------------------------------------------------------------------------
// WorkScore helpers + Validate
// ---------------------------------------------------------------------------

func TestNewWorkScore_ComputesDerivedScores(t *testing.T) {
	w := types.NewWorkScore("op1", "org1", 1.5, 0.9, 100, 7)
	require.Equal(t, "op1", w.OperatorID)
	require.Equal(t, "org1", w.OrgID)
	require.Equal(t, 1.5, w.RarityMultiplier)
	require.Equal(t, 0.9, w.AvailabilityScore)
	require.Equal(t, uint64(100), w.RetrievalVolume)
	require.Equal(t, uint64(7), w.Epoch)

	expectedStorage := 0.3 * 0.9
	expectedRetrieval := 0.7 * 100.0
	require.Equal(t, expectedStorage, w.StorageScore)
	require.Equal(t, expectedRetrieval, w.RetrievalScore)
	require.Equal(t, 1.5*(expectedStorage+expectedRetrieval), w.TotalScore)
}

func TestNewWorkScore_ZeroValues(t *testing.T) {
	w := types.NewWorkScore("op1", "org1", 0, 0, 0, 0)
	require.Equal(t, 0.0, w.StorageScore)
	require.Equal(t, 0.0, w.RetrievalScore)
	require.Equal(t, 0.0, w.TotalScore)
}

func TestWorkScore_Validate_Valid(t *testing.T) {
	w := types.NewWorkScore("op1", "org1", 1.0, 1.0, 1, 1)
	require.NoError(t, w.Validate())
}

func TestWorkScore_Validate_EmptyOperatorID(t *testing.T) {
	w := &types.WorkScore{OperatorID: "", OrgID: "org1"}
	require.ErrorIs(t, w.Validate(), types.ErrInvalidOperatorID)
}

func TestWorkScore_Validate_EmptyOrgID(t *testing.T) {
	w := &types.WorkScore{OperatorID: "op1", OrgID: ""}
	require.ErrorIs(t, w.Validate(), types.ErrInvalidOrgID)
}

// ---------------------------------------------------------------------------
// DailyEmission / OperatorReward / ValidatorReward constructors
// ---------------------------------------------------------------------------

func TestNewDailyEmission_InitializesMaps(t *testing.T) {
	e := types.NewDailyEmission(1, 10000, 8000, 2000)
	require.Equal(t, uint64(1), e.Epoch)
	require.Equal(t, uint64(10000), e.TotalEmitted)
	require.Equal(t, uint64(8000), e.OperatorShare)
	require.Equal(t, uint64(2000), e.ValidatorShare)
	require.NotNil(t, e.OperatorRewards)
	require.NotNil(t, e.ValidatorRewards)
	require.Empty(t, e.OperatorRewards)
	require.Empty(t, e.ValidatorRewards)
}

func TestNewOperatorReward(t *testing.T) {
	r := types.NewOperatorReward("op1", "org1", 500, 3)
	require.Equal(t, "op1", r.OperatorID)
	require.Equal(t, "org1", r.OrgID)
	require.Equal(t, uint64(500), r.Amount)
	require.Equal(t, uint64(3), r.Epoch)
}

func TestNewValidatorReward(t *testing.T) {
	r := types.NewValidatorReward("val1", 250, 4)
	require.Equal(t, "val1", r.ValidatorID)
	require.Equal(t, uint64(250), r.Amount)
	require.Equal(t, uint64(4), r.Epoch)
}

// ---------------------------------------------------------------------------
// BootstrapCredit constructor + CanRedeem + Redeem
// ---------------------------------------------------------------------------

func TestNewBootstrapCredit(t *testing.T) {
	b := types.NewBootstrapCredit("op1", 1000)
	require.Equal(t, "op1", b.OperatorID)
	require.Equal(t, uint64(1000), b.Credits)
	require.Equal(t, uint64(0), b.Redeemed)
}

func TestBootstrapCredit_CanRedeem(t *testing.T) {
	b := types.NewBootstrapCredit("op1", 1000)
	require.True(t, b.CanRedeem(0))    // boundary: zero
	require.True(t, b.CanRedeem(1000)) // boundary: exact balance
	require.False(t, b.CanRedeem(1001))
}

func TestBootstrapCredit_Redeem_Success(t *testing.T) {
	b := types.NewBootstrapCredit("op1", 1000)
	require.NoError(t, b.Redeem(400))
	require.Equal(t, uint64(400), b.Redeemed)

	// redeem remaining exactly to zero (boundary)
	require.NoError(t, b.Redeem(600))
	require.Equal(t, uint64(1000), b.Redeemed)
	require.False(t, b.CanRedeem(1))
}

func TestBootstrapCredit_Redeem_Insufficient(t *testing.T) {
	b := types.NewBootstrapCredit("op1", 100)
	require.ErrorIs(t, b.Redeem(101), types.ErrInsufficientBalance)
	require.Equal(t, uint64(0), b.Redeemed) // unchanged on failure
}

// ---------------------------------------------------------------------------
// AsymmetricGate constructor
// ---------------------------------------------------------------------------

func TestNewAsymmetricGate_RetrievalMirrorsStorage(t *testing.T) {
	passed := types.NewAsymmetricGate("op1", "org1", true, 1)
	require.True(t, passed.StoragePassed)
	require.True(t, passed.RetrievalAllowed)

	failed := types.NewAsymmetricGate("op1", "org1", false, 1)
	require.False(t, failed.StoragePassed)
	require.False(t, failed.RetrievalAllowed)
}

// ---------------------------------------------------------------------------
// Stored <-> domain conversion round trips (keys.go helpers)
// ---------------------------------------------------------------------------

func TestEmissionPool_StoredRoundTrip(t *testing.T) {
	require.Nil(t, types.EmissionPoolToStored(nil))
	require.Nil(t, types.StoredToEmissionPool(nil))

	p := types.NewEmissionPool(1000, 10, 80, 20, 5)
	got := types.StoredToEmissionPool(types.EmissionPoolToStored(p))
	require.Equal(t, p, got)
}

func TestDailyEmission_StoredRoundTrip(t *testing.T) {
	require.Nil(t, types.DailyEmissionToStored(nil))
	require.Nil(t, types.StoredToDailyEmission(nil))

	e := types.NewDailyEmission(1, 10000, 8000, 2000)
	e.OperatorRewards["op1"] = 5000
	e.ValidatorRewards["val1"] = 1000
	got := types.StoredToDailyEmission(types.DailyEmissionToStored(e))
	require.Equal(t, e.Epoch, got.Epoch)
	require.Equal(t, e.TotalEmitted, got.TotalEmitted)
	require.Equal(t, e.OperatorShare, got.OperatorShare)
	require.Equal(t, e.ValidatorShare, got.ValidatorShare)
	require.Equal(t, e.OperatorRewards, got.OperatorRewards)
	require.Equal(t, e.ValidatorRewards, got.ValidatorRewards)
}

func TestWorkScore_StoredRoundTrip(t *testing.T) {
	require.Nil(t, types.WorkScoreToStored(nil))
	require.Nil(t, types.StoredToWorkScore(nil))

	w := types.NewWorkScore("op1", "org1", 1.5, 0.9, 100, 7)
	got := types.StoredToWorkScore(types.WorkScoreToStored(w))
	require.Equal(t, w, got)
}

func TestBootstrapCredit_StoredRoundTrip(t *testing.T) {
	require.Nil(t, types.BootstrapCreditToStored(nil))
	require.Nil(t, types.StoredToBootstrapCredit(nil))

	b := types.NewBootstrapCredit("op1", 1000)
	require.NoError(t, b.Redeem(250))
	got := types.StoredToBootstrapCredit(types.BootstrapCreditToStored(b))
	require.Equal(t, b, got)
}

func TestAsymmetricGate_StoredRoundTrip(t *testing.T) {
	require.Nil(t, types.AsymmetricGateToStored(nil))
	require.Nil(t, types.StoredToAsymmetricGate(nil))

	g := types.NewAsymmetricGate("op1", "org1", true, 2)
	got := types.StoredToAsymmetricGate(types.AsymmetricGateToStored(g))
	require.Equal(t, g, got)
}

// ---------------------------------------------------------------------------
// GenesisState helpers + JSON codec
// ---------------------------------------------------------------------------

func TestNewGenesisState(t *testing.T) {
	pool := types.NewEmissionPool(1000, 10, 80, 20, 0)
	gs := types.NewGenesisState(
		pool,
		[]*types.DailyEmission{types.NewDailyEmission(1, 10000, 8000, 2000)},
		[]*types.OperatorReward{types.NewOperatorReward("op1", "org1", 500, 1)},
		[]*types.ValidatorReward{types.NewValidatorReward("val1", 250, 1)},
		[]*types.BootstrapCredit{types.NewBootstrapCredit("op1", 1000)},
		[]*types.WorkScore{types.NewWorkScore("op1", "org1", 1.0, 1.0, 1, 1)},
		[]*types.AsymmetricGate{types.NewAsymmetricGate("op1", "org1", true, 1)},
		42,
	)
	require.Equal(t, pool, gs.EmissionPool)
	require.Len(t, gs.DailyEmissions, 1)
	require.Len(t, gs.OperatorRewards, 1)
	require.Len(t, gs.ValidatorRewards, 1)
	require.Len(t, gs.BootstrapCredits, 1)
	require.Len(t, gs.WorkScores, 1)
	require.Len(t, gs.AsymmetricGates, 1)
	require.Equal(t, uint64(42), gs.BootstrapExpiry)
}

func TestNewGenesisState_EmptyCollections(t *testing.T) {
	gs := types.NewGenesisState(nil, nil, nil, nil, nil, nil, nil, 0)
	require.Nil(t, gs.EmissionPool)
	require.Empty(t, gs.DailyEmissions)
	require.Empty(t, gs.OperatorRewards)
	require.Empty(t, gs.ValidatorRewards)
	require.Empty(t, gs.BootstrapCredits)
	require.Empty(t, gs.WorkScores)
	require.Empty(t, gs.AsymmetricGates)
	require.Equal(t, uint64(0), gs.BootstrapExpiry)
}

func TestGenesisState_JSONRoundTrip(t *testing.T) {
	pool := types.NewEmissionPool(1000, 10, 80, 20, 3)
	gs := types.NewGenesisState(
		pool,
		[]*types.DailyEmission{types.NewDailyEmission(1, 10000, 8000, 2000)},
		nil, nil,
		[]*types.BootstrapCredit{types.NewBootstrapCredit("op1", 1000)},
		[]*types.WorkScore{types.NewWorkScore("op1", "org1", 1.0, 1.0, 1, 1)},
		nil,
		99,
	)

	data, err := json.Marshal(gs)
	require.NoError(t, err)

	var out types.GenesisState
	require.NoError(t, json.Unmarshal(data, &out))

	require.Equal(t, gs.BootstrapExpiry, out.BootstrapExpiry)
	require.Equal(t, gs.EmissionPool, out.EmissionPool)
	require.Len(t, out.DailyEmissions, 1)
	require.Len(t, out.BootstrapCredits, 1)
	require.Len(t, out.WorkScores, 1)
	require.Equal(t, "op1", out.BootstrapCredits[0].OperatorID)
}

func TestGenesisState_UnmarshalJSON_Invalid(t *testing.T) {
	var gs types.GenesisState
	require.Error(t, gs.UnmarshalJSON([]byte("{not valid json")))
}

// ---------------------------------------------------------------------------
// constants.go / events.go
// ---------------------------------------------------------------------------

func TestModuleNameConstant(t *testing.T) {
	require.Equal(t, "emissions", types.EmissionsModuleName)
}

func TestEventConstants(t *testing.T) {
	require.Equal(t, "emission_minted", types.EventEmissionMinted)
	require.Equal(t, "operator_reward_distributed", types.EventOperatorRewardDistributed)
	require.Equal(t, "validator_reward_distributed", types.EventValidatorRewardDistributed)
	require.Equal(t, "bootstrap_credit_redeemed", types.EventBootstrapCreditRedeemed)
	require.Equal(t, "work_score_computed", types.EventWorkScoreComputed)
	require.Equal(t, "asymmetric_gate_updated", types.EventAsymmetricGateUpdated)
}
