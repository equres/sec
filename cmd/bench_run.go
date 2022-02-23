package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"
	"net/http"
	"io/ioutil"
	"io"
	"bytes"

	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
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

		err = fetchFiles(fileURLs, rateLimit)
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

func fetchFiles(fileURLs []string, rateLimit time.Duration) error {
	client := &http.Client{}

	var currentDownloadCount int

	for _, fileURL := range fileURLs {
		if S.Verbose {
			log.Info("Downloading file: ", fileURL)
		}

		if fileURL == "" {
			continue
		}

		req, err := http.NewRequest(http.MethodGet, fileURL, nil)
		if err != nil {
			return err
		}

		req.Header.Set("User-Agent", "Firefox, 1234z@asd.aas")

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		size, err := io.Copy(ioutil.Discard, bytes.NewReader(responseBody))
		if err != nil {
			return err
		}

		if S.Verbose {
			log.Info("File size: ", size)
		}
		currentDownloadCount++

		log.Info(fmt.Sprintf("File progress [%d/%d] status_code_%d", currentDownloadCount, len(fileURLs), resp.StatusCode))

		time.Sleep(rateLimit)
	}
	return nil
}
