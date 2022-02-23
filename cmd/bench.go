package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
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

func FetchFiles(fileURLs []string, rateLimit time.Duration) error {
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
