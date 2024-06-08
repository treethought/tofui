package cmd

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/treethought/tofui/config"
	"github.com/treethought/tofui/db"
	"github.com/treethought/tofui/ui"
)

var (
	configPath = os.Getenv("CONFIG_FILE")
	cfg        *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "tofui",
	Short: "terminally on farcaster user interface",
	Run: func(cmd *cobra.Command, args []string) {
		runLocal()

	},
}

func runLocal() {
	sv := &Server{
		prgmSessions: make(map[string][]*tea.Program),
	}
	go sv.startSigninHTTPServer()
	app := ui.NewLocalApp(cfg)
	p := tea.NewProgram(app, tea.WithAltScreen())
	sv.prgmSessions["local"] = append(sv.prgmSessions["local"], p)
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
	os.MkdirAll("/tmp/tofui", 0755)
	var err error
	if configPath == "" {
		configPath = "config.yaml"
	}
	cfg, err = config.ReadConfig(configPath)
	if err != nil {
		log.Fatal("failed to read config: ", err)
	}
}
