package db

import (
	"os"
	"sync"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	log "github.com/sirupsen/logrus"
)

var (
	db     *DB
	once   sync.Once
	dbPath = os.Getenv("DB_PATH")
)

func GetDB() *DB {
	once.Do(func() {
		db = NewDB()
	})
	return db
}

type DB struct {
	db *badger.DB
	lf *os.File
}

func NewDB() *DB {
	path := dbPath
	if path == "" {
		path = "/tmp/castr"
	}

	err := os.MkdirAll(path, 0755)
	if err != nil {
		log.Fatalf("failed to create db directory: %v", err)
	}

	lfPath := path + "/db.log"

	lf, err := os.Create(lfPath)
	if err != nil {
		log.Fatalf("failed to create db log file: %v", err)
	}

	logger := log.New()
	logger.SetOutput(lf)
	opts := badger.DefaultOptions(path)
	opts.Logger = logger

	b, err := badger.Open(opts)
	if err != nil {
		log.Fatal("failed to open db: ", err)
	}
	db := &DB{db: b, lf: lf}
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
	db.lf.Close()
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
