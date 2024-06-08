package api

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"text/template"

	"github.com/dgraph-io/badger/v4"

	"github.com/treethought/castr/config"
	"github.com/treethought/castr/db"
)

var (
	//go:embed siwn.html
	sinwhtml []byte
	once     sync.Once
	sdb      *badger.DB

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

func StartSigninServer(cfg *config.Config, f func(fid uint64, signerUUid, pk string)) {
	tmpl, err := template.New("signin").Parse(string(sinwhtml))
	if err != nil {
		log.Fatal("failed to parse template: ", err)
	}
	data := struct {
		ClientID  string
		PublicKey string
	}{
		ClientID: cfg.Neynar.ClientID,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mux := http.NewServeMux()
	mux.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		pk := query.Get("pk")
		if pk == "" {
			w.Write([]byte("error: missing pk"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		data.PublicKey = pk
		err := tmpl.Execute(w, data)
		if err != nil {
			log.Println("failed to execute template: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/signin/success", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		pk := query.Get("pk")
		if pk == "" {
			w.Write([]byte("error: missing pk"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fid, err := strconv.ParseUint(query.Get("fid"), 10, 64)
		if err != nil {
			w.Write([]byte("error: missing fid"))
			cancel()
			return
		}
		signerUUid := query.Get("signer_uuid")

		f(fid, signerUUid, pk)
		w.Write([]byte("success, you may now close the window"))
		cancel()
	})

	srv := &http.Server{
		Addr:    "0.0.0.0:8000",
		Handler: mux,
	}
	// listener, err := net.Listen("tcp", srv.Addr)
	log.Println("listening on :8000")
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	<-ctx.Done()
	srv.Shutdown(context.Background())

}
