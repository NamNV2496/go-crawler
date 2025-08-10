package cmd

import "github.com/spf13/cobra"

func Execute() error {
	rootCmd := &cobra.Command{
		Short: "A simple web crawler",
	}
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(schedulerWorkerCmd)
	return rootCmd.Execute()
}
