package cmd

import (
	"github.com/equres/sec/pkg/secutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// compareCmd represents the compare command
var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compares files in a directory to ones in index file",
	Long:  `Compares files in a directory to ones in index file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			log.Info("Please type 'zip' to compare ZIP files, 'unzipped' to compare unzipped files, and 'raw' to compare the raw files")
			return nil
		}

		switch args[0] {
		case "zip":
			return secutil.CompareZipFiles(S, DB)
		case "unzipped":
			return secutil.CompareUnzippedFiles(S, DB)
		case "raw":
			return secutil.CompareRawFiles(S, DB)
		default:
			log.Info("Please type 'zip' to compare ZIP files, 'unzipped' to compare unzipped files, and 'raw' to compare the raw files")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(compareCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// compareCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// compareCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
