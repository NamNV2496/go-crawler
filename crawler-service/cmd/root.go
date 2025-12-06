package cmd

import "github.com/spf13/cobra"

func Execute() error {
	rootCmd := &cobra.Command{
		Short: "A simple web crawler",
	}
	rootCmd.AddCommand(CrawlerWorkerCmd)
	rootCmd.AddCommand(CrawlerWorkerRetryCmd)
	return rootCmd.Execute()
}
