package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a txt file containing a list of URLs to be downloaded",
	Long:  `Run a txt file containing a list of URLs to be downloaded.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("specify a txt file containing a list of URLs to be downloaded (e.g. sec bench run file.txt)")
		}

		rateLimitString, err := benchCmd.Flags().GetString("rate-limit")
		if err != nil {
			return err
		}

		rateLimit := time.Duration(0)

		if rateLimitString != "" {
			rateLimit, err = time.ParseDuration(fmt.Sprintf("%vms", rateLimitString))
			if err != nil {
				cobra.CheckErr(err)
			}
		}
		filePath := args[0]

		_, err = os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("file %s does not exist", filePath)
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		fileURLs := strings.Split(string(data), "\n")

		startTime := time.Now()

		err = FetchFiles(fileURLs, rateLimit)
		if err != nil {
			return err
		}

		endTime := time.Now()

		if S.Verbose {
			log.Info(fmt.Sprintf("It took %v seconds to download %v files", endTime.Sub(startTime).Seconds(), len(fileURLs)))
		}

		return nil
	},
}

func init() {
	benchCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
