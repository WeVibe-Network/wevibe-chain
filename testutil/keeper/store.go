package keeper

import (
	"context"
	"testing"

	corestore "cosmossdk.io/core/store"
	logv2 "cosmossdk.io/log/v2"
	dbm "github.com/cosmos/cosmos-db"
	storev2 "github.com/cosmos/cosmos-sdk/store/v2"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
)

type testStoreService struct {
	key *storetypes.KVStoreKey
	db  dbm.DB
	cms storetypes.CommitMultiStore
	kv  storetypes.KVStore
}

func (t *testStoreService) OpenKVStore(ctx context.Context) corestore.KVStore {
	return &wrapKVStore{kv: t.kv}
}

func (t *testStoreService) CommitMultiStore() storetypes.CommitMultiStore {
	return t.cms
}

type wrapKVStore struct {
	kv storetypes.KVStore
}

func (w wrapKVStore) Get(key []byte) ([]byte, error) {
	return w.kv.Get(key), nil
}

func (w wrapKVStore) Has(key []byte) (bool, error) {
	return w.kv.Has(key), nil
}

func (w wrapKVStore) Set(key, value []byte) error {
	w.kv.Set(key, value)
	return nil
}

func (w wrapKVStore) Delete(key []byte) error {
	w.kv.Delete(key)
	return nil
}

func (w wrapKVStore) Iterator(start, end []byte) (corestore.Iterator, error) {
	return w.kv.Iterator(start, end), nil
}

func (w wrapKVStore) ReverseIterator(start, end []byte) (corestore.Iterator, error) {
	return w.kv.ReverseIterator(start, end), nil
}

func NewTestStoreService(t *testing.T, key *storetypes.KVStoreKey) (corestore.KVStoreService, storetypes.CommitMultiStore) {
	t.Helper()
	db := dbm.NewMemDB()

	cms := storev2.NewCommitMultiStore(db, logv2.NewNopLogger())
	cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, nil)
	if err := cms.LoadLatestVersion(); err != nil {
		t.Fatalf("failed to load latest version: %v", err)
	}
	kv := cms.GetKVStore(key)
	return &testStoreService{key: key, db: db, cms: cms, kv: kv}, cms
}

func (t *testStoreService) GetDB() dbm.DB {
	return t.db
}

func NewTestStoreServiceWithCMS(t *testing.T, key *storetypes.KVStoreKey, cms storetypes.CommitMultiStore) corestore.KVStoreService {
	t.Helper()
	kv := cms.GetKVStore(key)
	return &testStoreService{key: key, cms: cms, kv: kv}
}

func NewTestStoreServiceWithDB(t *testing.T, key *storetypes.KVStoreKey, db dbm.DB) corestore.KVStoreService {
	t.Helper()
	cms := storev2.NewCommitMultiStore(db, logv2.NewNopLogger())
	cms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, nil)
	if err := cms.LoadLatestVersion(); err != nil {
		t.Fatalf("failed to load latest version: %v", err)
	}
	kv := cms.GetKVStore(key)
	return &testStoreService{key: key, db: db, cms: cms, kv: kv}
}

func NewTestLogger() logv2.Logger {
	return logv2.NewNopLogger()
}

func NewTestDB() dbm.DB {
	return dbm.NewMemDB()
}
