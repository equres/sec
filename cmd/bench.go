package cmd

import (
	"github.com/spf13/cobra"
)

// benchCmd represents the bench command
var benchCmd = &cobra.Command{
	Use:   "bench",
	Short: "Test how many files can be downloaded from sec.gov together",
	Long:  `Test how many files can be downloaded from sec.gov together`,
}

func NewBenchCMD() *cobra.Command {
	var rateLimit string
	benchCmd.PersistentFlags().StringVarP(&rateLimit, "rate-limit", "w", "", "Time to sleep in between each download")

	return benchCmd
}

func init() {
	rootCmd.AddCommand(NewBenchCMD())

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// benchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// benchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

