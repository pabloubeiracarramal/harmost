package main

import (
	"log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installCmd)
}

// TODO: Purpose: register the agent as a persistent OS service so it starts on boot.
//  Implementation is complete via kardianos/service.
//  Consider: warn the user (or block) if credentials are not yet present (i.e. `pair` hasn't been run).
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the agent as a service/daemon",
	Long:  `Installs the agent to run as a background service/daemon on system startup.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getService()
		if err != nil {
			return err
		}

		err = s.Install()
		if err != nil {
			return err
		}

		log.Println("Harmost Agent installed successfully.")
		return nil
	},
}