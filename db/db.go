package db

import (
	"log"
	"sync"
	"time"

	badger "github.com/dgraph-io/badger/v4"
)

var (
	db   *DB
	once sync.Once
)

func GetDB() *DB {
	once.Do(func() {
		db = NewDB()
	})
	return db
}

type DB struct {
	db *badger.DB
}

func NewDB() *DB {
	b, err := badger.Open(badger.DefaultOptions("/tmp/castr"))
	if err != nil {
		log.Fatal(err)
	}
	db := &DB{db: b}
	go db.runGC()
	return db
}

func (db *DB) runGC() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
	again:
		err := db.db.RunValueLogGC(0.7)
		if err == nil {
			goto again
		}
	}
}

func (db *DB) Close() {
	db.db.Close()
}

func (db *DB) Set(key, value []byte) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func (db *DB) Get(key []byte) ([]byte, error) {
	var value []byte
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		value, err = item.ValueCopy(nil)
		return err
	})
	return value, err
}

func (db *DB) Delete(key []byte) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (db *DB) GetKeys(prefix []byte) ([][]byte, error) {
	keys := make([][]byte, 0)
	err := db.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.KeyCopy(nil)
			keys = append(keys, k)
		}
		return nil
	})
	return keys, err
}
