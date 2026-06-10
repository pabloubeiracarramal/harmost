package main

import (
	"log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

// TODO: Purpose: tell the OS service manager to start the already-installed agent service.
//  Implementation is complete via kardianos/service.
//  Consider: validate credentials exist before starting and print a helpful error if `pair` hasn't been run.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the installed background service",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getService()
		if err != nil {
			return err
		}
		
		err = s.Start()
		if err != nil {
			return err
		}
		
		log.Println("Service started.")
		return nil
	},
}