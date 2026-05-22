package keeper_test

import (
	"context"
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/wevibe-network/wevibe-chain/testutil/keeper"
	"github.com/wevibe-network/wevibe-chain/x/reputation/keeper"
	"github.com/wevibe-network/wevibe-chain/x/reputation/types"
)

func setupMsgServer(t *testing.T) (types.MsgServer, context.Context) {
	storeKey := storetypes.NewKVStoreKey("reputation")
	storeService, _ := testkeeper.NewTestStoreService(t, storeKey)
	logger := testkeeper.NewTestLogger()
	k := keeper.NewKeeper(storeService, logger, "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9ry")

	k.Activate(context.Background())

	return keeper.NewMsgServerImpl(k), context.Background()
}

func TestMsgUpdateReputation_ValidateBasic(t *testing.T) {
	msg := &types.MsgUpdateReputation{}
	require.Error(t, msg.ValidateBasic())

	msg.Signer = "cosmos1abc"
	msg.Developer = []byte("cosmos1dev")
	msg.MemoryCid = "memory1"
	msg.Difficulty = 5
	msg.Quality = 5
	require.NoError(t, msg.ValidateBasic())
}

func TestMsgUpdateReputation_InvalidDifficulty(t *testing.T) {
	msg := &types.MsgUpdateReputation{
		Signer:     "cosmos1abc",
		Developer:  []byte("cosmos1dev"),
		MemoryCid:  "memory1",
		Difficulty: 15,
		Quality:    5,
	}
	require.Error(t, msg.ValidateBasic())
}

func TestMsgUpdateReputation_InvalidQuality(t *testing.T) {
	msg := &types.MsgUpdateReputation{
		Signer:     "cosmos1abc",
		Developer:  []byte("cosmos1dev"),
		MemoryCid:  "memory1",
		Difficulty: 5,
		Quality:    15,
	}
	require.Error(t, msg.ValidateBasic())
}

func TestMsgUpdateReputation_Success(t *testing.T) {
	srv, ctx := setupMsgServer(t)

	msg := &types.MsgUpdateReputation{
		Signer:     "cosmos1abc",
		Developer:  []byte("cosmos1dev"),
		MemoryCid:  "memory1",
		Difficulty: 5,
		Quality:    5,
		DomainTags: []string{"tag1", "tag2"},
		Provenance: "attested",
	}

	resp, err := srv.UpdateReputation(ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(25), resp.Xp)
}
