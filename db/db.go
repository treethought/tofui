package db

import (
	slog "log"
	"os"
	"sync"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	log "github.com/sirupsen/logrus"

	"github.com/treethought/tofui/config"
)

var (
	db   *DB
	once sync.Once
)

func GetDB() *DB {
	return db
}

type DB struct {
	db *badger.DB
	lf *os.File
}

func InitDB(cfg *config.Config) {
	once.Do(func() {
		path := cfg.DB.Dir
		if path == "" {
			path = ".tofui/db"
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
		slog.Print("opening db:", path)

		logger := log.New()
		logger.SetOutput(lf)
		opts := badger.DefaultOptions(path)
		opts.Logger = logger

		b, err := badger.Open(opts)
		if err != nil {
			log.Fatal("failed to open db: ", err)
		}
		d := &DB{db: b, lf: lf}
		db = d
		go db.runGC()
	})
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
	slog.Println("closing db")
	if db.db != nil {
		db.db.Close()
		db.lf.Close()
	}
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
