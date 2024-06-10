package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/treethought/tofui/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init tofui config",
	Run: func(cmd *cobra.Command, args []string) {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("failed to get user home directory")
		}
		fmt.Println("To use tofui locally, you will need to create a Neynar app")
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter Client ID (found at https://dev.neynar.com/app): \n")
		clientID, _ := reader.ReadString('\n')
		clientID = clientID[:len(clientID)-1] // Trim newline character

		fmt.Print("Enter API Key (found at https://dev.neynar.com/): \n")
		apiKey, _ := reader.ReadString('\n')
		apiKey = apiKey[:len(apiKey)-1] // Trim newline character

		cfg := &config.Config{}
		cfg.Neynar.ClientID = clientID
		cfg.Neynar.APIKey = apiKey
		cfg.Neynar.BaseUrl = "https://api.neynar.com/v2/farcaster"
		cfg.Server.Host = "localhost"
		cfg.Server.HTTPPort = 4200
		cfg.DB.Dir = filepath.Join(home, ".tofui", "db")
		cfg.Log.Path = filepath.Join(home, ".tofui", "debug.log")

		path := filepath.Join(home, ".tofui", "config.yaml")

		data, err := yaml.Marshal(cfg)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		err = os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			log.Fatalf("failed to create config directory: %v", err)
		}
		if err = os.WriteFile(path, data, 0644); err != nil {
			log.Fatalf("error: %v", err)
		}
		fmt.Printf("Wrote config file created at %s\n", path)
		fmt.Println("Note: You must add 'http://localhost:4200' to your Neynar app's Authorzed origins to sign in!!")
		fmt.Println("\nYou can now run `tofui` to start the app")

	},
}

func init() {
	rootCmd.AddCommand(initCmd)

}
