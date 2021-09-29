package cmd

import (
	"github.com/sirupsen/logrus"
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
		if S.Verbose {
			logrus.Info("Checking/Downloading index files...")
		}

		err := S.DownloadIndex(DB)
		if err != nil {
			return err
		}

		err = S.ForEachWorklist(DB, S.DownloadZIPFiles, "")
		if err != nil {
			return err
		}
		return nil
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
