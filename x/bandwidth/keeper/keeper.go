package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log/v2"

	"github.com/wevibe-network/wevibe-chain/x/bandwidth/types"
	"github.com/cosmos/gogoproto/proto"
)

type Keeper struct {
	storeService store.KVStoreService
	logger      log.Logger
	authority   string
	orgKeeper   types.OrgKeeper
}

func NewKeeper(storeService store.KVStoreService, logger log.Logger, authority string, orgKeeper types.OrgKeeper) *Keeper {
	return &Keeper{
		storeService: storeService,
		logger:       logger,
		authority:    authority,
		orgKeeper:    orgKeeper,
	}
}

func (k *Keeper) getStore(ctx context.Context) store.KVStore {
	return k.storeService.OpenKVStore(ctx)
}

func stateKey(orgID string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("state/%s/%d", orgID, epoch))
}

func overrideKey(orgID string) []byte {
	return []byte(fmt.Sprintf("override/%s", orgID))
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

func (k *Keeper) GetOrInitBandwidthState(ctx context.Context, orgID string, epoch uint64) (*types.BandwidthState, error) {
	store := k.getStore(ctx)
	key := stateKey(orgID, epoch)

	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}

	if bz != nil {
		var stored types.StoredBandwidthState
		if err := proto.Unmarshal(bz, &stored); err != nil {
			return nil, fmt.Errorf("unmarshal bandwidth state: %w", err)
		}
		state := storedToState(stored)
		return &state, nil
	}

	var memoryCap, serveCap uint64
	overrideStored, err := store.Get(overrideKey(orgID))
	if err != nil {
		return nil, err
	}
	if overrideStored != nil {
		var override types.StoredBandwidthOverride
		if err := proto.Unmarshal(overrideStored, &override); err != nil {
			return nil, fmt.Errorf("unmarshal override: %w", err)
		}
		memoryCap = override.MemoryCap
		serveCap = override.ServeCap
	} else {
		params, err := k.GetParams(ctx)
		if err != nil {
			return nil, err
		}
		memoryCap = params.DefaultMemoryCapPerEpoch
		serveCap = params.DefaultServeCapPerEpoch
	}

	state := types.NewBandwidthState(orgID, epoch, memoryCap, serveCap)
	bz, err = proto.Marshal(stateToStored(state))
	if err != nil {
		return nil, fmt.Errorf("marshal bandwidth state: %w", err)
	}
	if err := store.Set(key, bz); err != nil {
		return nil, err
	}

	return state, nil
}

func (k *Keeper) GetOrInitBandwidthStateRaw(ctx context.Context, orgID string, epoch uint64) (uint64, uint64, uint64, uint64, error) {
	state, err := k.GetOrInitBandwidthState(ctx, orgID, epoch)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return state.MemoryUsed, state.MemoryCap, state.ServeUsed, state.ServeCap, nil
}

func (k *Keeper) GetBandwidthState(ctx context.Context, orgID string, epoch uint64) (*types.BandwidthState, error) {
	store := k.getStore(ctx)
	key := stateKey(orgID, epoch)

	bz, err := store.Get(key)
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, nil
	}

	var stored types.StoredBandwidthState
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal bandwidth state: %w", err)
	}
	state := storedToState(stored)
	return &state, nil
}

func (k *Keeper) ConsumeMemoryBandwidth(ctx context.Context, orgID string, epoch uint64) error {
	state, err := k.GetOrInitBandwidthState(ctx, orgID, epoch)
	if err != nil {
		return err
	}

	if state.MemoryUsed >= state.MemoryCap {
		return types.ErrMemoryBandwidthExhausted
	}

	state.MemoryUsed++

	store := k.getStore(ctx)
	bz, err := proto.Marshal(stateToStored(state))
	if err != nil {
		return fmt.Errorf("marshal bandwidth state: %w", err)
	}
	if err := store.Set(stateKey(orgID, epoch), bz); err != nil {
		return err
	}

	return nil
}

func (k *Keeper) ConsumeServeBandwidth(ctx context.Context, orgID string, epoch uint64, count uint64) error {
	state, err := k.GetOrInitBandwidthState(ctx, orgID, epoch)
	if err != nil {
		return err
	}

	if state.ServeUsed+count > state.ServeCap {
		return types.ErrServeBandwidthExhausted
	}

	state.ServeUsed += count

	store := k.getStore(ctx)
	bz, err := proto.Marshal(stateToStored(state))
	if err != nil {
		return fmt.Errorf("marshal bandwidth state: %w", err)
	}
	if err := store.Set(stateKey(orgID, epoch), bz); err != nil {
		return err
	}

	return nil
}

func (k *Keeper) GetBandwidthOverride(ctx context.Context, orgID string) (*types.BandwidthOverride, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(overrideKey(orgID))
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrOverrideNotFound
	}

	var stored types.StoredBandwidthOverride
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal override: %w", err)
	}
	override := storedToOverride(stored)
	return &override, nil
}

func (k *Keeper) SetBandwidthOverride(ctx context.Context, orgID string, memoryCap, serveCap uint64) error {
	store := k.getStore(ctx)
	override := types.NewBandwidthOverride(orgID, memoryCap, serveCap)
	bz, err := proto.Marshal(overrideToStored(override))
	if err != nil {
		return fmt.Errorf("marshal override: %w", err)
	}
	return store.Set(overrideKey(orgID), bz)
}

func (k *Keeper) DeleteBandwidthOverride(ctx context.Context, orgID string) error {
	store := k.getStore(ctx)
	return store.Delete(overrideKey(orgID))
}

func (k *Keeper) GetRemainingBandwidth(ctx context.Context, orgID string, epoch uint64) (uint64, uint64, error) {
	state, err := k.GetOrInitBandwidthState(ctx, orgID, epoch)
	if err != nil {
		return 0, 0, err
	}
	return state.MemoryCap - state.MemoryUsed, state.ServeCap - state.ServeUsed, nil
}

func (k *Keeper) InitGenesis(ctx context.Context, state *types.GenesisState) error {
	store := k.getStore(ctx)

	for _, bs := range state.BandwidthStates {
		bz, err := proto.Marshal(stateToStored(bs))
		if err != nil {
			return err
		}
		if err := store.Set(stateKey(bs.OrgID, bs.Epoch), bz); err != nil {
			return err
		}
	}

	seenOverrides := make(map[string]bool)
	for _, bo := range state.BandwidthOverrides {
		if seenOverrides[bo.OrgID] {
			continue
		}
		seenOverrides[bo.OrgID] = true
		bz, err := proto.Marshal(overrideToStored(bo))
		if err != nil {
			return err
		}
		if err := store.Set(overrideKey(bo.OrgID), bz); err != nil {
			return err
		}
	}

	return nil
}

func (k *Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	store := k.getStore(ctx)

	var states []*types.BandwidthState
	statePrefix := []byte("state/")
	iter, err := store.Iterator(statePrefix, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var stored types.StoredBandwidthState
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			continue
		}
		state := storedToState(stored)
		states = append(states, &state)
	}

	var overrides []*types.BandwidthOverride
	overridePrefix := []byte("override/")
	overrideIter, err := store.Iterator(overridePrefix, nil)
	if err != nil {
		return nil, err
	}
	defer overrideIter.Close()
	seenOverrideOrgs := make(map[string]bool)
	for ; overrideIter.Valid(); overrideIter.Next() {
		var stored types.StoredBandwidthOverride
		if err := proto.Unmarshal(overrideIter.Value(), &stored); err != nil {
			continue
		}
		if seenOverrideOrgs[stored.OrgId] {
			continue
		}
		seenOverrideOrgs[stored.OrgId] = true
		override := storedToOverride(stored)
		overrides = append(overrides, &override)
	}

	return &types.GenesisState{
		BandwidthStates:    states,
		BandwidthOverrides: overrides,
	}, nil
}

func (k *Keeper) Logger(ctx context.Context) log.Logger {
	return k.logger.With("module", "x/bandwidth")
}