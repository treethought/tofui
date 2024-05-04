package api

import (
	"context"
	_ "embed"
	"log"
	"net/http"
	"strconv"
)

var (
	//go:embed siwn.html
	sinwhtml []byte
)

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
