package cmd

import "github.com/spf13/cobra"

func Execute() error {
	rootCmd := &cobra.Command{
		Use:   "crawler",
		Short: "A simple web crawler",
	}
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(CrawlerWorkerCmd)
	return rootCmd.Execute()
}
