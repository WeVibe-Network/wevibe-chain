package keeper

import (
	"context"
	"testing"

	"github.com/wevibe-network/wevibe-chain/x/bandwidth/types"
)

func TestMsgSetBandwidthOverride_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *types.MsgSetBandwidthOverride
		wantErr bool
	}{
		{
			name: "empty signer",
			msg: &types.MsgSetBandwidthOverride{
				Signer:    "",
				OrgId:     "test-org",
				MemoryCap: 10000,
				ServeCap:  50000,
			},
			wantErr: true,
		},
		{
			name: "empty org_id",
			msg: &types.MsgSetBandwidthOverride{
				Signer:    "signer",
				OrgId:     "",
				MemoryCap: 10000,
				ServeCap:  50000,
			},
			wantErr: true,
		},
		{
			name: "zero caps",
			msg: &types.MsgSetBandwidthOverride{
				Signer:    "signer",
				OrgId:     "test-org",
				MemoryCap: 0,
				ServeCap:  0,
			},
			wantErr: true,
		},
		{
			name: "valid with memory cap only",
			msg: &types.MsgSetBandwidthOverride{
				Signer:    "signer",
				OrgId:     "test-org",
				MemoryCap: 10000,
				ServeCap:  0,
			},
			wantErr: false,
		},
		{
			name: "valid with serve cap only",
			msg: &types.MsgSetBandwidthOverride{
				Signer:    "signer",
				OrgId:     "test-org",
				MemoryCap: 0,
				ServeCap:  50000,
			},
			wantErr: false,
		},
		{
			name: "valid with both caps",
			msg: &types.MsgSetBandwidthOverride{
				Signer:    "signer",
				OrgId:     "test-org",
				MemoryCap: 10000,
				ServeCap:  50000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMsgSetBandwidthOverride_Success(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()
	msgServer := NewMsgServerImpl(k)

	msg := &types.MsgSetBandwidthOverride{
		Signer:    "leader-pubkey",
		OrgId:     "test-org",
		MemoryCap: 20000,
		ServeCap:  100000,
	}

	_, err := msgServer.SetBandwidthOverride(ctx, msg)
	if err != nil {
		t.Fatalf("SetBandwidthOverride failed: %v", err)
	}

	state, err := k.GetOrInitBandwidthState(ctx, "test-org", 1)
	if err != nil {
		t.Fatalf("GetOrInitBandwidthState failed: %v", err)
	}
	if state.MemoryCap != 20000 {
		t.Errorf("MemoryCap mismatch: got %d, want 20000", state.MemoryCap)
	}
	if state.ServeCap != 100000 {
		t.Errorf("ServeCap mismatch: got %d, want 100000", state.ServeCap)
	}
}

func TestMsgSetBandwidthOverride_NotLeader(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()
	msgServer := NewMsgServerImpl(k)

	msg := &types.MsgSetBandwidthOverride{
		Signer:    "not-leader-pubkey",
		OrgId:     "test-org",
		MemoryCap: 20000,
		ServeCap:  100000,
	}

	_, err := msgServer.SetBandwidthOverride(ctx, msg)
	if err != types.ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestMsgSetBandwidthOverride_OrgNotFound(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()
	msgServer := NewMsgServerImpl(k)

	msg := &types.MsgSetBandwidthOverride{
		Signer:    "leader-pubkey",
		OrgId:     "nonexistent-org",
		MemoryCap: 20000,
		ServeCap:  100000,
	}

	_, err := msgServer.SetBandwidthOverride(ctx, msg)
	if err != types.ErrOrgNotFound {
		t.Fatalf("expected ErrOrgNotFound, got: %v", err)
	}
}

func TestMsgUpdateParams_Success(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()
	msgServer := NewMsgServerImpl(k)

	msg := &types.MsgUpdateParams{
		Authority: "gov",
		Params: &types.Params{
			DefaultMemoryCapPerEpoch: 50000,
			DefaultServeCapPerEpoch:  100000,
		},
	}

	_, err := msgServer.UpdateParams(ctx, msg)
	if err != nil {
		t.Fatalf("UpdateParams failed: %v", err)
	}

	params, _ := k.GetParams(ctx)
	if params.DefaultMemoryCapPerEpoch != 50000 {
		t.Errorf("DefaultMemoryCapPerEpoch mismatch: got %d, want 50000", params.DefaultMemoryCapPerEpoch)
	}
	if params.DefaultServeCapPerEpoch != 100000 {
		t.Errorf("DefaultServeCapPerEpoch mismatch: got %d, want 100000", params.DefaultServeCapPerEpoch)
	}
}

func TestMsgUpdateParams_Unauthorized(t *testing.T) {
	k, _ := makeTestKeeper(t)
	ctx := context.Background()
	msgServer := NewMsgServerImpl(k)

	msg := &types.MsgUpdateParams{
		Authority: "not-gov",
		Params: &types.Params{
			DefaultMemoryCapPerEpoch: 50000,
			DefaultServeCapPerEpoch:  100000,
		},
	}

	_, err := msgServer.UpdateParams(ctx, msg)
	if err != types.ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}