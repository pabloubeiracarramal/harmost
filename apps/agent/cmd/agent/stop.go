package main

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(stopCmd)
}

// TODO: Purpose: tell the OS service manager to stop the running agent service.
//  Implementation is complete via kardianos/service. Nothing further needed here.
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the agent",
	Long:  `Stop the agent`,
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getService()
		if err != nil {
			return err
		}

		err = s.Stop()
		if err != nil {
			return err
		}

		log.Println("Agent stopped successfully")
		return nil
	},
}
