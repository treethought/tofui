package cmd

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/treethought/tofui/api"
	"github.com/treethought/tofui/db"
	"github.com/treethought/tofui/ui"
)

var castCmd = &cobra.Command{
	Use:   "cast",
	Short: "publish a cast",
	Run: func(cmd *cobra.Command, args []string) {
		defer logFile.Close()
		defer db.GetDB().Close()
		signer := api.GetSigner("local")
		if signer != nil {
			log.Println("logged in as: ", signer.Username)
		}
		if signer == nil {
			fmt.Println("please sign in to use this command by running `tofui`")
			return
		}

		app := ui.NewLocalApp(cfg, true)
		p := tea.NewProgram(app, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(castCmd)
}
