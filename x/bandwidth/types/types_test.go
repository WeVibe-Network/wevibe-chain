package types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wevibe-network/wevibe-chain/x/bandwidth/types"
)

// ---------------------------------------------------------------------------
// MsgSetBandwidthOverride.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgSetBandwidthOverrideValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		msg     *types.MsgSetBandwidthOverride
		wantErr error
	}{
		{
			name: "valid both caps",
			msg: &types.MsgSetBandwidthOverride{
				Signer:    "signer",
				OrgId:     "test-org",
				MemoryCap: 10000,
				ServeCap:  50000,
			},
			wantErr: nil,
		},
		{
			name: "valid memory cap only",
			msg: &types.MsgSetBandwidthOverride{
				Signer:    "signer",
				OrgId:     "test-org",
				MemoryCap: 1,
				ServeCap:  0,
			},
			wantErr: nil,
		},
		{
			name: "valid serve cap only",
			msg: &types.MsgSetBandwidthOverride{
				Signer:    "signer",
				OrgId:     "test-org",
				MemoryCap: 0,
				ServeCap:  1,
			},
			wantErr: nil,
		},
		{
			name: "empty signer",
			msg: &types.MsgSetBandwidthOverride{
				Signer:    "",
				OrgId:     "test-org",
				MemoryCap: 10000,
				ServeCap:  50000,
			},
			// signer branch returns a plain fmt error, not a sentinel
			wantErr: nil, // sentinel comparison skipped; checked as generic error below
		},
		{
			name: "empty org id",
			msg: &types.MsgSetBandwidthOverride{
				Signer:    "signer",
				OrgId:     "",
				MemoryCap: 10000,
				ServeCap:  50000,
			},
			wantErr: types.ErrInvalidOrgID,
		},
		{
			name: "both caps zero",
			msg: &types.MsgSetBandwidthOverride{
				Signer:    "signer",
				OrgId:     "test-org",
				MemoryCap: 0,
				ServeCap:  0,
			},
			wantErr: types.ErrInvalidCap,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			switch {
			case tt.name == "valid both caps" || tt.name == "valid memory cap only" || tt.name == "valid serve cap only":
				require.NoError(t, err)
			case tt.name == "empty signer":
				require.Error(t, err)
				require.Contains(t, err.Error(), "signer cannot be empty")
			default:
				require.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// MsgUpdateParams.ValidateBasic
// ---------------------------------------------------------------------------

func TestMsgUpdateParamsValidateBasic(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		msg := &types.MsgUpdateParams{
			Authority: "gov",
			Params: &types.Params{
				DefaultMemoryCapPerEpoch: 10000,
				DefaultServeCapPerEpoch:  50000,
			},
		}
		require.NoError(t, msg.ValidateBasic())
	})

	t.Run("empty authority", func(t *testing.T) {
		msg := &types.MsgUpdateParams{
			Authority: "",
			Params: &types.Params{
				DefaultMemoryCapPerEpoch: 10000,
				DefaultServeCapPerEpoch:  50000,
			},
		}
		err := msg.ValidateBasic()
		require.Error(t, err)
		require.Contains(t, err.Error(), "authority cannot be empty")
	})

	t.Run("nil params still valid because Params.Validate has no deref", func(t *testing.T) {
		// (*Params).Validate() only returns nil and does not dereference its
		// receiver, so a nil Params pointer is safe and authority is the only
		// failing branch.
		msg := &types.MsgUpdateParams{
			Authority: "gov",
			Params:    nil,
		}
		require.NoError(t, msg.ValidateBasic())
	})

	t.Run("zero-value params valid", func(t *testing.T) {
		msg := &types.MsgUpdateParams{
			Authority: "gov",
			Params:    &types.Params{},
		}
		require.NoError(t, msg.ValidateBasic())
	})
}

// ---------------------------------------------------------------------------
// Params.Validate
// ---------------------------------------------------------------------------

func TestParamsValidate(t *testing.T) {
	// Validate currently has no failing branches; assert it accepts both
	// the default and edge (zero / max) values.
	p := types.DefaultParams()
	require.NoError(t, p.Validate())

	zero := types.Params{}
	require.NoError(t, zero.Validate())

	max := types.Params{
		DefaultMemoryCapPerEpoch: ^uint64(0),
		DefaultServeCapPerEpoch:  ^uint64(0),
	}
	require.NoError(t, max.Validate())
}

func TestDefaultParams(t *testing.T) {
	p := types.DefaultParams()
	require.Equal(t, uint64(10000), p.DefaultMemoryCapPerEpoch)
	require.Equal(t, uint64(50000), p.DefaultServeCapPerEpoch)
}

// ---------------------------------------------------------------------------
// BandwidthState.Validate
// ---------------------------------------------------------------------------

func TestBandwidthStateValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		bs := types.NewBandwidthState("test-org", 1, 10000, 50000)
		require.NoError(t, bs.Validate())
	})

	t.Run("valid zero epoch and caps", func(t *testing.T) {
		bs := types.NewBandwidthState("test-org", 0, 0, 0)
		require.NoError(t, bs.Validate())
	})

	t.Run("empty org id", func(t *testing.T) {
		bs := types.NewBandwidthState("", 1, 10000, 50000)
		require.ErrorIs(t, bs.Validate(), types.ErrInvalidOrgID)
	})
}

func TestNewBandwidthState(t *testing.T) {
	bs := types.NewBandwidthState("test-org", 7, 111, 222)
	require.Equal(t, "test-org", bs.OrgID)
	require.Equal(t, uint64(7), bs.Epoch)
	require.Equal(t, uint64(0), bs.MemoryUsed)
	require.Equal(t, uint64(111), bs.MemoryCap)
	require.Equal(t, uint64(0), bs.ServeUsed)
	require.Equal(t, uint64(222), bs.ServeCap)
}

// ---------------------------------------------------------------------------
// BandwidthOverride.Validate
// ---------------------------------------------------------------------------

func TestBandwidthOverrideValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		bo := types.NewBandwidthOverride("test-org", 20000, 100000)
		require.NoError(t, bo.Validate())
	})

	t.Run("valid zero caps", func(t *testing.T) {
		bo := types.NewBandwidthOverride("test-org", 0, 0)
		require.NoError(t, bo.Validate())
	})

	t.Run("empty org id", func(t *testing.T) {
		bo := types.NewBandwidthOverride("", 20000, 100000)
		require.ErrorIs(t, bo.Validate(), types.ErrInvalidOrgID)
	})
}

func TestNewBandwidthOverride(t *testing.T) {
	bo := types.NewBandwidthOverride("test-org", 333, 444)
	require.Equal(t, "test-org", bo.OrgID)
	require.Equal(t, uint64(333), bo.MemoryCap)
	require.Equal(t, uint64(444), bo.ServeCap)
}

// ---------------------------------------------------------------------------
// GenesisState helpers
// ---------------------------------------------------------------------------

func TestNewGenesisState(t *testing.T) {
	states := []*types.BandwidthState{
		types.NewBandwidthState("test-org", 1, 10000, 50000),
	}
	overrides := []*types.BandwidthOverride{
		types.NewBandwidthOverride("test-org", 20000, 100000),
	}

	gs := types.NewGenesisState(states, overrides)
	require.Len(t, gs.BandwidthStates, 1)
	require.Len(t, gs.BandwidthOverrides, 1)
	require.Equal(t, "test-org", gs.BandwidthStates[0].OrgID)
	require.Equal(t, "test-org", gs.BandwidthOverrides[0].OrgID)
}

func TestNewGenesisStateEmpty(t *testing.T) {
	gs := types.NewGenesisState(nil, nil)
	require.Empty(t, gs.BandwidthStates)
	require.Empty(t, gs.BandwidthOverrides)
}

func TestGenesisStateJSONRoundTrip(t *testing.T) {
	gs := types.NewGenesisState(
		[]*types.BandwidthState{
			{
				OrgID:      "test-org",
				Epoch:      3,
				MemoryUsed: 12,
				MemoryCap:  10000,
				ServeUsed:  34,
				ServeCap:   50000,
			},
		},
		[]*types.BandwidthOverride{
			{OrgID: "test-org", MemoryCap: 20000, ServeCap: 100000},
		},
	)

	raw, err := json.Marshal(gs)
	require.NoError(t, err)

	var decoded types.GenesisState
	require.NoError(t, json.Unmarshal(raw, &decoded))

	require.Len(t, decoded.BandwidthStates, 1)
	require.Len(t, decoded.BandwidthOverrides, 1)

	gotState := decoded.BandwidthStates[0]
	require.Equal(t, "test-org", gotState.OrgID)
	require.Equal(t, uint64(3), gotState.Epoch)
	require.Equal(t, uint64(12), gotState.MemoryUsed)
	require.Equal(t, uint64(10000), gotState.MemoryCap)
	require.Equal(t, uint64(34), gotState.ServeUsed)
	require.Equal(t, uint64(50000), gotState.ServeCap)

	gotOverride := decoded.BandwidthOverrides[0]
	require.Equal(t, "test-org", gotOverride.OrgID)
	require.Equal(t, uint64(20000), gotOverride.MemoryCap)
	require.Equal(t, uint64(100000), gotOverride.ServeCap)
}

func TestGenesisStateUnmarshalEmptyObject(t *testing.T) {
	var gs types.GenesisState
	require.NoError(t, json.Unmarshal([]byte(`{}`), &gs))
	require.Empty(t, gs.BandwidthStates)
	require.Empty(t, gs.BandwidthOverrides)
}

// ---------------------------------------------------------------------------
// keys.go module constants
// ---------------------------------------------------------------------------

func TestModuleConstants(t *testing.T) {
	require.Equal(t, "bandwidth", types.ModuleName)
	require.Equal(t, "bandwidth", types.StoreKey)
	require.Equal(t, types.ModuleName, types.RouterKey)
}
