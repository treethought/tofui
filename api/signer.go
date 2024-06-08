package api

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/dgraph-io/badger/v4"

	"github.com/treethought/castr/db"
)

var (
	once sync.Once
	sdb  *badger.DB

	cache = make(map[string]*Signer)
	mu    sync.RWMutex
)

type Signer struct {
	FID         uint64
	UUID        string
	Username    string
	DisplayName string
	PublicKey   string
}

func SetSigner(s *Signer) {
	once.Do(func() {
		d, _ := json.Marshal(s)
		key := fmt.Sprintf("signer:%s", s.PublicKey)
		if err := db.GetDB().Set([]byte(key), d); err != nil {
			log.Fatal("failed to save signer: ", err)
		}
	})
}

func GetSigner(pk string) *Signer {
	mu.RLock()
	signer, ok := cache[pk]
	mu.RUnlock()
	if ok {
		return signer
	}
	key := fmt.Sprintf("signer:%s", pk)
	d, err := db.GetDB().Get([]byte(key))
	if err != nil {
		log.Println("no signer found in db")
		return nil
	}
	signer = &Signer{}
	if err = json.Unmarshal(d, signer); err != nil {
		log.Println("failed to unmarshal signer: ", err)
		return nil
	}
	mu.Lock()
	cache[pk] = signer
	mu.Unlock()
	return signer
}
