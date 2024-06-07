package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/treethought/castr/api"
)

var signInCmd = &cobra.Command{
	Use:   "signin",
	Short: "sign into farcaster",
	Long: `Sign in is performed with Neynar's SIWN.
  You will be prompted to visit a URL in your browser to begin the sign in flow.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("running signin")
		runSignin()
	},
}

func runSignin() {
	client := api.NewClient(cfg)

	signer := api.GetSigner()
	if signer == nil {
		fmt.Println("Please visit http://localhost:8000/signin to sign in")
		api.StartSigninServer(cfg, func(fid uint64, uuid string) {
			signer := &api.Signer{FID: fid, UUID: uuid}
			if user, err := client.GetUserByFID(fid); err == nil {
				signer.Username = user.Username
				signer.DisplayName = user.DisplayName
			}
			api.SetSigner(signer)
			fmt.Println("signed in as:", signer.Username)
		})
	} else {
		fmt.Println("already signed in as:", signer.Username)
	}

}

func init() {
	rootCmd.AddCommand(signInCmd)
}
