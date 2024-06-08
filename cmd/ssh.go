package cmd

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"text/template"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	wlog "github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/accesscontrol"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

	"github.com/treethought/tofui/api"
	"github.com/treethought/tofui/ui"
)

var (
	//go:embed siwn.html
	sinwhtml []byte
	host     = "0.0.0.0"
	port     = "42069"
	// port = "22"
	addr = host + ":" + port
)

type Server struct {
	// user may have more than one session
	// map of pk to active programs
	prgmSessions map[string][]*tea.Program
	mux          sync.Mutex
}

// sshCmd represents the ssh command
var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "serve tofui over ssh",
	Run: func(cmd *cobra.Command, args []string) {
		sv := &Server{
			prgmSessions: make(map[string][]*tea.Program),
		}
		go sv.startSigninHTTPServer()
		sv.runSSHServer()
	},
}

func (sv *Server) runSSHServer() {
	s, err := wish.NewServer(
		wish.WithAddress(addr),
		wish.WithHostKeyPath(".ssh/tofui_ed25519"),
		// Accept any public key.
		ssh.PublicKeyAuth(func(ssh.Context, ssh.PublicKey) bool { return true }),
		// Do not accept password auth.
		ssh.PasswordAuth(func(ssh.Context, string) bool { return false }),
		wish.WithMiddleware(
			sv.teaMiddleware(),
			activeterm.Middleware(),
			logging.Middleware(),
			accesscontrol.Middleware(),
		),
	)
	if err != nil {
		wlog.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	wlog.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			wlog.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	wlog.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		wlog.Error("Could not stop server", "error", err)
	}
}

func (sv *Server) teaMiddleware() wish.Middleware {
	// var p *tea.Program
	// newProg := func(m tea.Model, opts ...tea.ProgramOption) *tea.Program {
	// 	p = tea.NewProgram(m, opts...)
	// 	return p
	// }
	teaHandler := func(s ssh.Session) *tea.Program {
		_, _, active := s.Pty()
		if !active {
			wish.Fatalln(s, "no active terminal, skipping")
			return nil
		}

		renderer := bubbletea.MakeRenderer(s)
		app, err := ui.NewSSHApp(cfg, s, renderer)
		if err != nil {
			wlog.Error("failed to create app", "error", err)
			return nil
		}
		if app.PublicKey() == "" {
			log.Fatal("new app's public key is nil")
		}

		p := tea.NewProgram(app, append(bubbletea.MakeOptions(s), tea.WithAltScreen())...)

		sv.mux.Lock()
		sv.prgmSessions[app.PublicKey()] = append(sv.prgmSessions[app.PublicKey()], p)
		sv.mux.Unlock()
		log.Println("new app session added: ", app.PublicKey())
		return p
	}
	return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}

func init() {
	rootCmd.AddCommand(sshCmd)

}

func (sv *Server) HttpHandleSignin(w http.ResponseWriter, r *http.Request) {
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
	query := r.URL.Query()
	pk := query.Get("pk")
	if pk == "" {
		w.Write([]byte("error: missing pk"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	data.PublicKey = pk
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("failed to execute template: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (sv *Server) HttpHandleSigninSuccess(w http.ResponseWriter, r *http.Request) {
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
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	signerUUid := query.Get("signer_uuid")

	sv.signinCallback(fid, signerUUid, pk)
	w.Write([]byte("success, you may now close the window and return to your terminal."))
}

func (sv *Server) signinCallback(fid uint64, uuid, pk string) {
	client := api.NewClient(cfg)
	signer := &api.Signer{FID: fid, UUID: uuid, PublicKey: pk}
	if user, err := client.GetUserByFID(fid, fid); err == nil {
		signer.Username = user.Username
		signer.DisplayName = user.DisplayName
	}
	api.SetSigner(signer)

	var prgms []*tea.Program
	var ok bool
	sv.mux.Lock()
	prgms, ok = sv.prgmSessions[pk]
	sv.mux.Unlock()
	if !ok || len(prgms) == 0 {
		log.Println("failed to send signin msg, session not found")
		return
	}
	for _, p := range prgms {
		if p == nil {
			log.Println("nil program")
			continue
		}
		// TODO send to tea program as msg instead of directly calling
		p.Send(&ui.UpdateSignerMsg{Signer: signer})
	}
	fmt.Println("signed in as:", signer.Username)
}

func (sv *Server) startSigninHTTPServer() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mux := http.NewServeMux()
	mux.HandleFunc("/signin", sv.HttpHandleSignin)
	mux.HandleFunc("/signin/success", sv.HttpHandleSigninSuccess)

	srv := &http.Server{
		Addr:    "0.0.0.0:8000",
		Handler: mux,
	}
	log.Println("listening on :8000")
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	<-ctx.Done()
	srv.Shutdown(context.Background())

}
