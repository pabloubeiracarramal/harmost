package main

import (
	"github.com/harmost/agent/internal/daemon"
	"github.com/kardianos/service"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "harmost",
	Short: "Harmost Agent",
	Long:  `Harmost Agent is a component of the Harmost monitoring system.`,
}

func Execute() error {
	return rootCmd.Execute()
}

// getService returns a configured kardianos service.
// This is shared across install, start, stop, etc. commands.
func getService() (service.Service, error) {
	svcConfig := &service.Config{
		Name:        "HarmostAgent",
		DisplayName: "Harmost Agent Service",
		Description: "Maintains a persistent bidirectional gRPC stream with the hub.",
		Arguments:   []string{"run"}, // CRITICAL: Tells the OS to execute "harmost run" when starting the service
	}

	prg := daemon.NewAgentProgram()
	return service.New(prg, svcConfig)
}
