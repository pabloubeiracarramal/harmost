package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/harmost/agent/internal/docker"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(containersCmd)
}

var containersCmd = &cobra.Command{
	Use:   "containers",
	Short: "List all containers on this host (debugging aid)",
	Long:  `Lists all containers (running and stopped) visible to the agent's Docker daemon, including the harmost job that created them, if any.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		d, err := docker.New()
		if err != nil {
			return err
		}

		list, err := d.ListAllContainers(cmd.Context())
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "CONTAINER ID\tIMAGE\tSTATE\tSTATUS\tJOB ID\tNAMES")
		for _, c := range list {
			id := c.ID
			if len(id) > 12 {
				id = id[:12]
			}
			names := make([]string, len(c.Names))
			for i, n := range c.Names {
				names[i] = strings.TrimPrefix(n, "/")
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				id, c.Image, c.State, c.Status, c.Labels[docker.JobIDLabel], strings.Join(names, ","))
		}
		return w.Flush()
	},
}
