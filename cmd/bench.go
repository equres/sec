package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/equres/sec/pkg/secutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// benchCmd represents the bench command
var benchCmd = &cobra.Command{
	Use:   "bench",
	Short: "Test how many files can be downloaded from sec.gov together",
	Long:  `Test how many files can be downloaded from sec.gov together`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please insert the number of files you wish to fetch (e.g. sec bench 100)")
		}

		countArg, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		countFilesToBeDownloaded := countArg

		if S.Verbose {
			log.Info("Number of files to be downloaded: ", countFilesToBeDownloaded)
		}

		// Get random files:
		if S.Verbose {
			log.Info("Getting a random index file to fetch XBRLFiles...")
		}
		rssFiles, err := secutil.GetAllRSSFiles(S, DB)
		if err != nil {
			return err
		}

		rand.Seed(time.Now().Unix())
		rssFileToBeUsed := rssFiles[rand.Intn(len(rssFiles))]

		var fileURLs []string
		for _, item := range rssFileToBeUsed.Channel.Item {
			for _, files := range item.XbrlFiling.XbrlFiles.XbrlFile {
				fileURLs = append(fileURLs, files.URL)
				if len(fileURLs) >= countFilesToBeDownloaded {
					break
				}
			}
			if len(fileURLs) >= countFilesToBeDownloaded {
				break
			}
		}

		if S.Verbose {
			log.Info("Downloading files...")
		}

		rateLimitString, err := cmd.Flags().GetString("rate-limit")
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

		startTime := time.Now()

		err = FetchFiles(fileURLs, rateLimit)
		if err != nil {
			return err
		}

		endTime := time.Now()

		if S.Verbose {
			log.Info(fmt.Sprintf("It took %v seconds to download %v files", endTime.Sub(startTime).Seconds(), countFilesToBeDownloaded))
		}

		return nil
	},
}

func NewBenchCMD() *cobra.Command {
	var rateLimit string
	benchCmd.Flags().StringVarP(&rateLimit, "rate-limit", "w", "", "Time to sleep in between each download")

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
