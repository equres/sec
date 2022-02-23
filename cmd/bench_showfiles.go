package cmd

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/equres/sec/pkg/secutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// showfilesCmd represents the showfiles command
var showfilesCmd = &cobra.Command{
	Use:   "showfiles",
	Short: "Prints a specified number of links to files from sec.gov",
	Long:  `Prints a specified number of links to files from sec.gov`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("specify a number of files to be printed out (e.g. sec bench showfiles 1000)")
		}

		countFilesToBeDownloaded, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		if S.Verbose {
			log.Info("Getting a random index file to fetch XBRLFiles...")
		}

		rssFiles, err := secutil.GetAllRSSFiles(S, DB)
		if err != nil {
			return err
		}

		if len(rssFiles) == 0 {
			return fmt.Errorf("please ensure you have enabled downloading of at least 1 index file and run 'sec dow index' in order to download the index files")
		}

		rand.Seed(time.Now().UnixNano())
		rssFileToBeUsed := rssFiles[rand.Intn(len(rssFiles))]

		var fileURLs []string
		for _, item := range rssFileToBeUsed.Channel.Item {
			for _, files := range item.XbrlFiling.XbrlFiles.XbrlFile {
				fileURLs = append(fileURLs, files.URL)
				if len(fileURLs) == countFilesToBeDownloaded {
					break
				}
			}
			if len(fileURLs) == countFilesToBeDownloaded {
				break
			}
		}

		for _, fileURL := range fileURLs {
			fmt.Println(fileURL)
		}

		return nil
	},
}

func init() {
	benchCmd.AddCommand(showfilesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// showfilesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// showfilesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
