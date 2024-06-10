package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/treethought/tofui/config"
	"github.com/treethought/tofui/db"
	"github.com/treethought/tofui/ui"
)

var (
	configPath = os.Getenv("CONFIG_FILE")
	cfg        *config.Config
	logFile    *os.File
)

var rootCmd = &cobra.Command{
	Use:   "tofui",
	Short: "terminally on farcaster user interface",
	Run: func(cmd *cobra.Command, args []string) {
		runLocal()

	},
}

func runLocal() {
	defer logFile.Close()
	defer db.GetDB().Close()
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
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config file (default is $HOME/.tofui.yaml)")
}

func initConfig() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		return
	}
	var err error
	if configPath == "" {
		if _, err := os.Stat("config.yaml"); err == nil {
			configPath = "config.yaml"
		} else {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Fatal("failed to find default config file: ", err)
			}
			configPath = filepath.Join(homeDir, ".tofui", "config.yaml")
		}
	}
	cfg, err = config.ReadConfig(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatal("failed to fing config file, run `tofui init` to create one")
		}
		log.Fatal("failed to read config: ", err)
	}

	lf := cfg.Log.Path
	if lf == "" {
		lf = "tofui.log"
	}
	dir := filepath.Dir(lf)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	logFile, err = tea.LogToFile(lf, "debug")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	log.Println("loaded config: ", configPath)
	db.InitDB(cfg)

}
