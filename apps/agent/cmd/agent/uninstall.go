package main

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

// TODO: Purpose: remove the agent from the OS service manager.
//  Implementation is complete via kardianos/service.
//  Consider: stop the service first if it is running, and optionally delete stored credentials.
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the agent",
	Long:  `Uninstall the agent from the system.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getService()
		if err != nil {
			return err
		}

		err = s.Uninstall()
		if err != nil {
			return err
		}

		log.Println("Agent uninstalled successfully.")
		return nil
	},
}
