package main

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

// TODO: Purpose: entry point called by the OS service manager (via Arguments: []string{"run"} in root.go).
//  This command itself is complete — it hands control to kardianos/service which calls AgentProgram.Start().
//  All remaining work lives in internal/daemon/program.go.
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the agent in daemon mode",
	Long:  `Starts the agent as a background service/daemon.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getService()
		if err != nil {
			return err
		}

		// If run interactively, this blocks until Ctrl+C.
		// If run by the OS, this blocks until the OS stops the service.
		log.Println("Booting Harmost Agent...")
		return s.Run()
	},
}
