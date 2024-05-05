package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/treethought/castr/api"
	"github.com/treethought/castr/db"
	"github.com/treethought/castr/ui"
)

var API_KEY = os.Getenv("API_KEY")

const HUB_URL = "https://api.neynar.com/v2/farcaster"

func main() {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()
	db := db.GetDB()
	defer db.Close()
	client := api.NewClient(HUB_URL, API_KEY)
	// update channel mappings
	// go func() {
	// 	lastloaded, err := db.Get([]byte("channelsloaded"))
	// 	// convert lastloaded int64
	// 	update := false
	// 	lastSec, err := strconv.ParseInt(string(lastloaded), 10, 64)
	// 	if err != nil {
	// 		update = true
	// 	}
	// 	if update || time.Since(time.Unix(int64(lastSec), 0)) > time.Hour {
	// 		log.Println("fetching channels")
	// 		err := client.FetchAllChannels()
	// 		if err != nil {
	// 			log.Fatal("error fetching channels: ", err)
	// 		}
	// 	}
	// 	log.Println("channels loaded")
	// }()

	signer := api.GetSigner()
	if signer == nil {
		fmt.Println("no signer found, visit http://localhost:8000/signin to sign in")
		api.StartSigninServer(func(fid uint64, uuid string) {
			signer := &api.Signer{FID: fid, UUID: uuid}
			if user, err := client.GetUserByFID(fid); err == nil {
				signer.Username = user.Username
				signer.DisplayName = user.DisplayName
			}
			api.SetSigner(signer)
			log.Println("signed in!")
		})
	}

	log.Println("signer found, FID: ", signer.FID)

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
