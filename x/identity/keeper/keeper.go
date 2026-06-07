package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/wevibe-network/wevibe-chain/x/identity/types"
)

type Keeper struct {
	storeService store.KVStoreService
	logger       log.Logger
	authority    string
}

func NewKeeper(storeService store.KVStoreService, logger log.Logger, authority string) *Keeper {
	return &Keeper{
		storeService: storeService,
		logger:       logger,
		authority:    authority,
	}
}

func (k *Keeper) getStore(ctx context.Context) store.KVStore {
	return k.storeService.OpenKVStore(ctx)
}

const paramsKey = "params"

func (k *Keeper) setParams(ctx context.Context, params types.Params) error {
	store := k.getStore(ctx)
	bz, err := proto.Marshal(&params)
	if err != nil {
		return fmt.Errorf("marshal params: %w", err)
	}
	return store.Set([]byte(paramsKey), bz)
}

func (k *Keeper) getParams(ctx context.Context) (types.Params, error) {
	store := k.getStore(ctx)
	bz, err := store.Get([]byte(paramsKey))
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

func (k *Keeper) SetAlias(ctx context.Context, alias *types.StoredIdentityAlias) error {
	if alias == nil {
		return fmt.Errorf("alias cannot be nil")
	}

	store := k.getStore(ctx)
	bz, err := proto.Marshal(alias)
	if err != nil {
		return fmt.Errorf("marshal alias: %w", err)
	}
	return store.Set(types.AliasKey(alias.PasskeyPubkey), bz)
}

func (k *Keeper) GetAlias(ctx context.Context, passkeyPubkey string) (*types.StoredIdentityAlias, bool, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(types.AliasKey(passkeyPubkey))
	if err != nil {
		return nil, false, err
	}
	if bz == nil {
		return nil, false, nil
	}

	var alias types.StoredIdentityAlias
	if err := proto.Unmarshal(bz, &alias); err != nil {
		return nil, false, fmt.Errorf("unmarshal alias: %w", err)
	}
	return &alias, true, nil
}

func (k *Keeper) ResolveIdentity(ctx context.Context, passkeyPubkey string) (walletAddress string, isMigrated bool, found bool, err error) {
	alias, found, err := k.GetAlias(ctx, passkeyPubkey)
	if err != nil {
		return "", false, false, err
	}
	if !found {
		return "", false, false, nil
	}

	return alias.WalletAddress, alias.IsMigrated, true, nil
}

func (k *Keeper) IterateAliases(ctx context.Context, cb func(*types.StoredIdentityAlias) bool) error {
	store := k.getStore(ctx)
	iter, err := store.Iterator(types.KeyPrefixAlias, storetypes.PrefixEndBytes(types.KeyPrefixAlias))
	if err != nil {
		return err
	}
	defer iter.Close()

	stopCallback := false
	for ; iter.Valid(); iter.Next() {
		var alias types.StoredIdentityAlias
		if err := proto.Unmarshal(iter.Value(), &alias); err != nil {
			return fmt.Errorf("unmarshal alias: %w", err)
		}
		if stopCallback {
			continue
		}
		if !cb(&alias) {
			stopCallback = true
		}
	}

	return nil
}

func (k *Keeper) InitGenesisState(ctx context.Context, state *types.GenesisState) {
	if state == nil {
		state = types.DefaultGenesis()
	}
	if err := state.Validate(); err != nil {
		panic(fmt.Errorf("identity: validate genesis: %w", err))
	}

	if err := k.setParams(ctx, state.Params); err != nil {
		panic(fmt.Errorf("identity: set params: %w", err))
	}

	for _, alias := range state.Aliases {
		if alias == nil {
			continue
		}
		if err := k.SetAlias(ctx, alias); err != nil {
			panic(fmt.Errorf("identity: set alias: %w", err))
		}
	}
}

func (k *Keeper) ExportGenesisState(ctx context.Context) *types.GenesisState {
	params, err := k.getParams(ctx)
	if err != nil {
		panic(fmt.Errorf("identity: get params: %w", err))
	}

	aliases := make([]*types.StoredIdentityAlias, 0)
	if err := k.IterateAliases(ctx, func(alias *types.StoredIdentityAlias) bool {
		aliasCopy := *alias
		aliases = append(aliases, &aliasCopy)
		return true
	}); err != nil {
		panic(fmt.Errorf("identity: iterate aliases: %w", err))
	}

	return &types.GenesisState{
		Aliases: aliases,
		Params:  params,
	}
}
