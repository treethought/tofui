package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/treethought/castr/api"
	"github.com/treethought/castr/ui"
)

var API_KEY = os.Getenv("API_KEY")

const HUB_URL = "https://api.neynar.com/v2/farcaster"

func main() {
	client := api.NewClient(HUB_URL, API_KEY)
	app := ui.NewApp()

	feed := ui.NewFeedView(client)
	app.Register("feed", feed)
	app.SetFocus("feed")

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	log.Println("starting app")
	// start the app
	p := tea.NewProgram(app, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}
