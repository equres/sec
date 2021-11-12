package cmd

import (
	"github.com/spf13/cobra"
)

// DID NOT FIX IMPORT:
// ERROR EXPECTED

// dowzCmd represents the dowz command
var dowzCmd = &cobra.Command{
	Use:   "dowz",
	Short: "download all of the referenced file from XBRL index as ZIP files",
	Long:  `download all of the referenced file from XBRL index as ZIP files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return S.ForEachWorklist(DB, S.DownloadZIPFiles, "")
	},
}

func init() {
	rootCmd.AddCommand(dowzCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dowzCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dowzCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
