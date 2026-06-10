package main

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pairCmd)
}

// TODO: Implement OAuth2 device flow pairing:
//  Purpose: authenticate this agent with the hub and persist credentials for use by the daemon.
//  1. Accept the hub address as a flag or argument.
//  2. POST to the hub's device authorization endpoint to get a device_code and user_code.
//  3. Print the verification URL and user_code so the user can authorize in a browser.
//  4. Poll the hub's token endpoint until the user completes authorization or it times out.
//  5. On success, persist the access token + hub address to the OS config dir (os.UserConfigDir).
var pairCmd = &cobra.Command{
	Use:   "pair",
	Short: "Pair the agent with the server",
	Long:  `Pair the agent with the server using a pairing code.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("Placehorder command")
		return nil
	},
}
