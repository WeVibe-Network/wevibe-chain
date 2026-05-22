package keeper

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/wevibe-network/wevibe-chain/x/memory/types"
)

const currentEpochKey = "current_epoch"

type dec struct {
	value *big.Rat
}

func newDec(s string) dec {
	r := new(big.Rat)
	r.SetString(s)
	return dec{value: r}
}

func (d dec) Add(other dec) dec {
	r := new(big.Rat).Add(d.value, other.value)
	return dec{value: r}
}

func (d dec) Sub(other dec) dec {
	r := new(big.Rat).Sub(d.value, other.value)
	return dec{value: r}
}

func (d dec) GT(other dec) bool {
	return d.value.Cmp(other.value) > 0
}

func (d dec) IsNegative() bool {
	return d.value.Sign() < 0
}

func (d dec) IsPositive() bool {
	return d.value.Sign() > 0
}

func (d dec) String() string {
	return d.value.FloatString(4)
}

var zeroDec = dec{value: big.NewRat(0, 1)}
var oneDec = dec{value: big.NewRat(1, 1)}

func parseWeight(s string) dec {
	if s == "" {
		return newDec("1.0")
	}
	s = strings.TrimSpace(s)
	if !strings.Contains(s, ".") && !strings.Contains(s, "/") {
		if i, ok := new(big.Int).SetString(s, 10); ok {
			return dec{value: new(big.Rat).SetFrac(i, big.NewInt(1))}
		}
	}
	return newDec(s)
}

func (k *Keeper) ArchiveMemory(ctx context.Context, orgID, cid, actor string) error {
	isLeader, err := k.orgKeeper.IsLeader(ctx, orgID, actor)
	if err != nil {
		return err
	}
	if !isLeader {
		return types.ErrUnauthorized
	}

	memory, err := k.loadMemoryByCID(ctx, orgID, cid)
	if err != nil {
		return err
	}

	if memory.State == types.MemoryState_MEMORY_STATE_ARCHIVED {
		return nil
	}

	memory.State = types.MemoryState_MEMORY_STATE_ARCHIVED
	return k.saveMemoryCommitment(ctx, memory)
}

func (k *Keeper) saveMemoryCommitment(ctx context.Context, memory *types.MemoryCommitment) error {
	store := k.getStore(ctx)
	bz, err := proto.Marshal(memoryToStored(memory))
	if err != nil {
		return fmt.Errorf("marshal memory: %w", err)
	}
	if err := store.Set(approvedKey(memory.OrgID, memory.ContentHash), bz); err != nil {
		return fmt.Errorf("persist memory: %w", err)
	}
	return nil
}

func (k *Keeper) loadMemoryByCID(ctx context.Context, orgID, cid string) (*types.MemoryCommitment, error) {
	contentHash, err := decodeCID(cid)
	if err != nil {
		return nil, err
	}

	store := k.getStore(ctx)
	bz, err := store.Get(approvedKey(orgID, contentHash))
	if err != nil {
		return nil, fmt.Errorf("get memory: %w", err)
	}
	if bz == nil {
		return nil, types.ErrMemoryNotFound
	}

	var stored types.StoredMemoryCommitment
	if err := proto.Unmarshal(bz, &stored); err != nil {
		return nil, fmt.Errorf("unmarshal memory: %w", err)
	}

	memory := storedToMemory(stored)
	return &memory, nil
}

func (k *Keeper) ApplyIdleDecay(ctx context.Context, orgID string, epoch uint64) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	store := k.getStore(ctx)
	prefix := []byte("approved/")
	iter, err := store.Iterator(prefix, storetypes.PrefixEndBytes(prefix))
	if err != nil {
		return fmt.Errorf("iterate approved memories: %w", err)
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var stored types.StoredMemoryCommitment
		if err := proto.Unmarshal(iter.Value(), &stored); err != nil {
			return fmt.Errorf("unmarshal memory commitment: %w", err)
		}

		memory := storedToMemory(stored)

		if !isDecayEligible(memory.State) {
			continue
		}

		cidHex := types.ContentHashToHex(memory.ContentHash)

		serveCount, err := k.getMemoryServeCount(ctx, memory.OrgID, cidHex, epoch)
		if err != nil {
			return err
		}
		denialCount, err := k.getMemoryDenialCount(ctx, memory.OrgID, cidHex, epoch)
		if err != nil {
			return err
		}

		if serveCount > 0 || denialCount > 0 {
			continue
		}

		if epoch <= memory.LastActiveEpoch {
			continue
		}

		if isInBootstrapGrace(epoch, memory.Epoch, params.BootstrapGraceEpochs) {
			continue
		}

		allZero := k.applyIdleDecayToMemory(&memory, params)

		if allZero {
			memory.State = types.MemoryState_MEMORY_STATE_ARCHIVED
		}

		if err := k.saveMemoryCommitment(ctx, &memory); err != nil {
			return err
		}
	}

	return nil
}

func (k *Keeper) applyIdleDecayToMemory(memory *types.MemoryCommitment, params types.Params) bool {
	idleDecay := float64(params.IdleDecayRateBps) / 10000.0

	for i := range memory.Keywords {
		weight := parseWeight(memory.Keywords[i].Weight)
		newWeight := weight.Sub(dec{value: new(big.Rat).SetFloat64(idleDecay)})
		if newWeight.IsNegative() {
			newWeight = zeroDec
		}
		memory.Keywords[i].Weight = newWeight.String()
	}

	allZero := true
	for _, kw := range memory.Keywords {
		w := parseWeight(kw.Weight)
		if w.IsPositive() {
			allZero = false
			break
		}
	}

	return allZero
}

func (k *Keeper) ApplyServeBoost(ctx context.Context, orgID string, contentHash []byte) error {
	memory, err := k.GetApprovedMemory(ctx, orgID, contentHash)
	if err != nil {
		return err
	}

	if memory.State == types.MemoryState_MEMORY_STATE_ARCHIVED {
		return nil
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	currentEpoch := k.getCurrentEpoch(ctx)
	if isInBootstrapGrace(currentEpoch, memory.Epoch, params.BootstrapGraceEpochs) {
		return nil
	}

	cidHex := types.ContentHashToHex(contentHash)
	serveCount, err := k.getMemoryServeCount(ctx, orgID, cidHex, currentEpoch)
	if err != nil {
		return err
	}

	if serveCount > params.MaxServeBoostPerEpoch {
		return nil
	}

	boost := float64(params.ServeBoostBps) / 10000.0

	for i := range memory.Keywords {
		weight := parseWeight(memory.Keywords[i].Weight)
		newWeight := weight.Add(dec{value: new(big.Rat).SetFloat64(boost)})
		if newWeight.GT(oneDec) {
			newWeight = oneDec
		}
		memory.Keywords[i].Weight = newWeight.String()
		memory.Keywords[i].ServeCount++
	}

	memory.LastActiveEpoch = currentEpoch

	return k.saveMemoryCommitment(ctx, memory)
}

func (k *Keeper) ApplyDenialDecay(ctx context.Context, orgID string, contentHash []byte) error {
	memory, err := k.GetApprovedMemory(ctx, orgID, contentHash)
	if err != nil {
		return err
	}

	if memory.State == types.MemoryState_MEMORY_STATE_ARCHIVED {
		return nil
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	currentEpoch := k.getCurrentEpoch(ctx)
	if isInBootstrapGrace(currentEpoch, memory.Epoch, params.BootstrapGraceEpochs) {
		return nil
	}

	decay := float64(params.DenialDecayBps) / 10000.0

	for i := range memory.Keywords {
		weight := parseWeight(memory.Keywords[i].Weight)
		newWeight := weight.Sub(dec{value: new(big.Rat).SetFloat64(decay)})
		if newWeight.IsNegative() {
			newWeight = zeroDec
		}
		memory.Keywords[i].Weight = newWeight.String()
		memory.Keywords[i].DenialCount++
	}

	memory.LastActiveEpoch = currentEpoch

	allZero := true
	for _, kw := range memory.Keywords {
		w := parseWeight(kw.Weight)
		if w.IsPositive() {
			allZero = false
			break
		}
	}

	if allZero {
		memory.State = types.MemoryState_MEMORY_STATE_ARCHIVED
	}

	return k.saveMemoryCommitment(ctx, memory)
}

func (k *Keeper) getCurrentEpoch(ctx context.Context) uint64 {
	store := k.getStore(ctx)
	bz, err := store.Get([]byte(currentEpochKey))
	if err != nil || bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (k *Keeper) setCurrentEpoch(ctx context.Context, epoch uint64) error {
	store := k.getStore(ctx)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, epoch)
	if err := store.Set([]byte(currentEpochKey), buf); err != nil {
		return fmt.Errorf("set current epoch: %w", err)
	}
	return nil
}

func (k *Keeper) getMemoryServeCount(ctx context.Context, orgID, cid string, epoch uint64) (uint64, error) {
	if k.serveKeeper == nil {
		return 0, nil
	}
	count, err := k.serveKeeper.GetMemoryServeCountForEpoch(ctx, orgID, cid, epoch)
	if err != nil {
		return 0, fmt.Errorf("serve keeper: %w", err)
	}
	return count, nil
}

func (k *Keeper) getMemoryDenialCount(ctx context.Context, orgID, cid string, epoch uint64) (uint64, error) {
	if k.serveKeeper == nil {
		return 0, nil
	}
	count, err := k.serveKeeper.GetMemoryDenialCountForEpoch(ctx, orgID, cid, epoch)
	if err != nil {
		return 0, fmt.Errorf("serve keeper denial count: %w", err)
	}
	return count, nil
}

func isDecayEligible(state types.MemoryState) bool {
	return state == types.MemoryState_MEMORY_STATE_COMMITTED
}

func isInBootstrapGrace(epoch, memoryEpoch, bootstrapGraceEpochs uint64) bool {
	if bootstrapGraceEpochs == 0 {
		return false
	}
	if epoch < memoryEpoch {
		return true
	}
	return epoch-memoryEpoch < bootstrapGraceEpochs
}

func decodeCID(cid string) ([]byte, error) {
	contentHash, err := hex.DecodeString(cid)
	if err != nil {
		return nil, fmt.Errorf("decode cid: %w", err)
	}
	if len(contentHash) != types.ContentHashLen {
		return nil, types.ErrInvalidContentHash
	}
	return contentHash, nil
}
