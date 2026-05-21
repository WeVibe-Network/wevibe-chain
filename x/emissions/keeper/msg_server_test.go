package keeper_test

import (
	"context"
	"testing"

	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/wevibe-network/wevibe-chain/testutil/keeper"
	"github.com/wevibe-network/wevibe-chain/x/emissions/keeper"
	"github.com/wevibe-network/wevibe-chain/x/emissions/types"
)

func setupMsgServer(t *testing.T) (types.MsgServer, context.Context) {
	storeKey := storetypes.NewKVStoreKey("emissions")
	storeService, _ := testkeeper.NewTestStoreService(t, storeKey)
	logger := testkeeper.NewTestLogger()
	k := keeper.NewKeeper(storeService, logger, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry", nil, nil, nil, newMockReputationKeeper())
	return keeper.NewMsgServerImpl(k), context.Background()
}

func TestMsgMintDailyEmission_ValidateBasic(t *testing.T) {
	msg := &types.MsgMintDailyEmission{}
	require.Error(t, msg.ValidateBasic())

	msg.Authority = "cosmos1abc"
	msg.Epoch = 1
	require.NoError(t, msg.ValidateBasic())
}

func TestMsgMintDailyEmission_NoPool(t *testing.T) {
	srv, ctx := setupMsgServer(t)

	msg := &types.MsgMintDailyEmission{
		Authority: "cosmos1abc",
		Epoch:     1,
	}

	_, err := srv.MintDailyEmission(ctx, msg)
	require.Error(t, err)
}

func TestMsgDistributeOperatorRewards_ValidateBasic(t *testing.T) {
	msg := &types.MsgDistributeOperatorRewards{}
	require.Error(t, msg.ValidateBasic())

	msg.Signer = "cosmos1abc"
	msg.Rewards = []*types.OperatorRewardEntry{{OperatorId: "op1", Amount: 100}}
	require.NoError(t, msg.ValidateBasic())
}

func TestMsgDistributeOperatorRewards_NoEmission(t *testing.T) {
	srv, ctx := setupMsgServer(t)

	msg := &types.MsgDistributeOperatorRewards{
		Signer:  "cosmos1abc",
		Rewards: []*types.OperatorRewardEntry{{OperatorId: "op1", Amount: 100}},
		Epoch:   1,
	}

	_, err := srv.DistributeOperatorRewards(ctx, msg)
	require.Error(t, err)
}
