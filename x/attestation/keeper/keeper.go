package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/gogoproto/proto"

	"github.com/wevibe-network/wevibe-chain/x/attestation/types"
)

type Keeper struct {
	storeService store.KVStoreService
	logger       log.Logger
	authority    string
	orgKeeper    types.OrgKeeper
}

func NewKeeper(
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,
	orgKeeper types.OrgKeeper,
) *Keeper {
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

func prefixEndBytes(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}
	end := make([]byte, len(prefix))
	copy(end, prefix)
	for i := len(end) - 1; i >= 0; i-- {
		end[i] = end[i] + 1
		if end[i] != 0 {
			return end
		}
	}
	return nil
}

func (k *Keeper) SetSessionAttestation(ctx context.Context, att *types.StoredSessionAttestation) error {
	store := k.getStore(ctx)
	bz, err := proto.Marshal(att)
	if err != nil {
		return fmt.Errorf("marshal attestation: %w", err)
	}
	if err := store.Set(types.AttestationKey(att.OrgId, att.SessionHash), bz); err != nil {
		return err
	}
	return store.Set(types.AttestationByEpochKey(att.OrgId, att.Epoch, att.SessionHash), bz)
}

func (k *Keeper) GetSessionAttestation(ctx context.Context, orgID string, sessionHash []byte) (*types.StoredSessionAttestation, error) {
	store := k.getStore(ctx)
	bz, err := store.Get(types.AttestationKey(orgID, sessionHash))
	if err != nil {
		return nil, err
	}
	if bz == nil {
		return nil, types.ErrAttestationNotFound
	}
	var att types.StoredSessionAttestation
	if err := proto.Unmarshal(bz, &att); err != nil {
		return nil, fmt.Errorf("unmarshal attestation: %w", err)
	}
	return &att, nil
}

func (k *Keeper) HasSessionAttestation(ctx context.Context, orgID string, sessionHash []byte) bool {
	store := k.getStore(ctx)
	has, err := store.Has(types.AttestationKey(orgID, sessionHash))
	if err != nil {
		return false
	}
	return has
}

func (k *Keeper) ListSessionAttestations(ctx context.Context, orgID string, epoch uint64) ([]*types.StoredSessionAttestation, error) {
	store := k.getStore(ctx)
	prefix := types.AttestationByEpochPrefix(orgID, epoch)
	iter, err := store.Iterator(prefix, prefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var attestations []*types.StoredSessionAttestation
	for ; iter.Valid(); iter.Next() {
		var att types.StoredSessionAttestation
		if err := proto.Unmarshal(iter.Value(), &att); err != nil {
			continue
		}
		attestations = append(attestations, &att)
	}
	return attestations, nil
}

func (k *Keeper) VerifyCommitLLMReceipt(ctx context.Context, receiptHash []byte) (bool, string) {
	return false, "unverified: commitllm integration pending"
}

func (k *Keeper) VerifyCloudProviderSignature(ctx context.Context, signatureHash []byte) (bool, string) {
	return false, "unverified: cloud provider attestation pending"
}

func (k *Keeper) SetParams(ctx context.Context, params types.Params) error {
	store := k.getStore(ctx)
	bz, err := proto.Marshal(&params)
	if err != nil {
		return fmt.Errorf("marshal params: %w", err)
	}
	return store.Set([]byte(types.ParamsKey), bz)
}

func (k *Keeper) GetParams(ctx context.Context) (types.Params, error) {
	store := k.getStore(ctx)
	bz, err := store.Get([]byte(types.ParamsKey))
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

func (k *Keeper) InitGenesis(ctx context.Context, state *types.GenesisState) error {
	for _, att := range state.Attestations {
		if err := k.SetSessionAttestation(ctx, att); err != nil {
			return err
		}
	}
	return nil
}

func (k *Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	store := k.getStore(ctx)
	prefix := []byte("attestation/")
	iter, err := store.Iterator(prefix, prefixEndBytes(prefix))
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var attestations []*types.StoredSessionAttestation
	for ; iter.Valid(); iter.Next() {
		var att types.StoredSessionAttestation
		if err := proto.Unmarshal(iter.Value(), &att); err != nil {
			continue
		}
		attestations = append(attestations, &att)
	}
	return &types.GenesisState{Attestations: attestations}, nil
}

func (k *Keeper) Logger(ctx context.Context) log.Logger {
	return k.logger.With("module", "x/attestation")
}
