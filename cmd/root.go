package cmd

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/treethought/castr/api"
	"github.com/treethought/castr/config"
	"github.com/treethought/castr/db"
	"github.com/treethought/castr/ui"
)

var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "castr",
	Short: "terminally on farcaster user interface",
	Run: func(cmd *cobra.Command, args []string) {
		runLocal()

	},
}

func runLocal() {
	client := api.NewClient(cfg)
	app := ui.NewApp()

	feed := ui.NewFeedView(client, ui.DefaultFeedParams())
	app.Register("feed", feed)
	app.SetFocus("feed")

	castDetails := ui.NewCastView(nil)
	app.Register("cast", castDetails)

	profile := ui.NewProfile()
	app.Register("profile", profile)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func Execute() {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	db := db.GetDB()
	defer db.Close()

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	os.MkdirAll("/tmp/castr", 0755)
	var err error
	cfg, err = config.ReadConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
}
