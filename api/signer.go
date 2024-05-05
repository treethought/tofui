package api

import (
	"context"
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/treethought/castr/db"
)

var (
	//go:embed siwn.html
	sinwhtml []byte
	signer   *Signer
	once     sync.Once
)

type Signer struct {
	FID         uint64
	UUID        string
	Username    string
	DisplayName string
}

func SetSigner(s *Signer) {
	once.Do(func() {
		signer = s
		d, _ := json.Marshal(s)
		if err := db.GetDB().Set([]byte("signer"), d); err != nil {
			log.Fatal("failed to save signer: ", err)
		}
	})
}
func GetSigner() *Signer {
	if signer == nil {
		d, err := db.GetDB().Get([]byte("signer"))
		if err != nil {
			log.Println("no signer found in db")
			return nil
		}
		signer = &Signer{}
		if err = json.Unmarshal(d, signer); err != nil {
			log.Println("failed to unmarshal signer: ", err)
			return nil
		}
	}
	return signer
}

func StartSigninServer(f func(fid uint64, signerUUid string)) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mux := http.NewServeMux()
	mux.HandleFunc("/signin", func(w http.ResponseWriter, r *http.Request) {
		w.Write(sinwhtml)
	})
	mux.HandleFunc("/signin/success", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		fid, err := strconv.ParseUint(query.Get("fid"), 10, 64)
		if err != nil {
			w.Write([]byte("error: missing fid"))
			cancel()
			return
		}
		signerUUid := query.Get("signer_uuid")

		f(fid, signerUUid)
		w.Write([]byte("success, you may now close the window"))
		cancel()
	})

	srv := &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}
	// listener, err := net.Listen("tcp", srv.Addr)
	log.Println("listening on :8000")
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	srv.Shutdown(context.Background())

}
