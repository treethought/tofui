package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/treethought/castr/api"
	"github.com/treethought/castr/config"
	"github.com/treethought/castr/db"
	"github.com/treethought/castr/ui"
)

func main() {

	cfg, err := config.ReadConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()
	db := db.GetDB()
	defer db.Close()
	client := api.NewClient(cfg)

	signerReady := make(chan struct{})
	signer := api.GetSigner()
	if signer == nil {
		fmt.Println("no signer found, visit http://localhost:8000/signin to sign in")
		api.StartSigninServer(cfg, func(fid uint64, uuid string) {
			signer := &api.Signer{FID: fid, UUID: uuid}
			if user, err := client.GetUserByFID(fid); err == nil {
				signer.Username = user.Username
				signer.DisplayName = user.DisplayName
			}
			api.SetSigner(signer)
			log.Println("signed in!")
			close(signerReady)
		})
	} else {
		log.Println("signer found, FID: ", signer.FID)
		close(signerReady)
	}

	<-signerReady
	app := ui.NewApp()

	feed := ui.NewFeedView(client, ui.DefaultFeedParams())
	app.Register("feed", feed)
	app.SetFocus("feed")

	castDetails := ui.NewCastView(nil)
	app.Register("cast", castDetails)

	profile := ui.NewProfile()
	app.Register("profile", profile)

	log.Println("starting app")
	// start the app
	p := tea.NewProgram(app, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}
