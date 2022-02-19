package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/equres/sec/pkg/download"
	"github.com/equres/sec/pkg/secreq"
	"github.com/equres/sec/pkg/secutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var GlobalRateLimit string
var GlobalGenerateTxt bool

// benchCmd represents the bench command
var benchCmd = &cobra.Command{
	Use:   "bench",
	Short: "Test how many files can be downloaded from sec.gov together",
	Long:  `Test how many files can be downloaded from sec.gov together`,
	RunE: func(cmd *cobra.Command, args []string) error {
		countFilesToBeDownloaded := 20
		if len(args) > 0 {
			countArg, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			countFilesToBeDownloaded = countArg
		}

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

		if GlobalGenerateTxt {
			if S.Verbose {
				log.Info("Generating a text file with the URLs...")
			}
			err = os.WriteFile(filepath.Join(S.Config.Main.CacheDir, "file_urls.txt"), []byte(strings.Join(fileURLs, "\n")), 0644)
			if err != nil {
				return err
			}
		}

		if S.Verbose {
			log.Info("Downloading files...")
		}

		downloader := download.NewDownloader(S.Config)
		downloader.IsEtag = true
		downloader.Verbose = S.Verbose
		downloader.Debug = S.Debug
		downloader.TotalDownloadsCount = countFilesToBeDownloaded
		downloader.CurrentDownloadCount = 1

		rateLimit, err := time.ParseDuration(fmt.Sprintf("%vms", S.Config.Main.RateLimitMs))
		if err != nil {
			return err
		}

		if GlobalRateLimit != "" {
			rateLimit, err = time.ParseDuration(fmt.Sprintf("%vms", GlobalRateLimit))
			if err != nil {
				return err
			}
		}

		if S.Verbose {
			log.Info("Rate limit used between downloads in ms: ", rateLimit.Milliseconds())
		}

		if GlobalGenerateTxt {
			data, err := os.ReadFile(filepath.Join(S.Config.Main.CacheDir, "file_urls.txt"))
			if err != nil {
				return err
			}

			fileURLs = strings.Split(string(data), "\n")
		}

		startTime := time.Now()

		for _, v := range fileURLs {
			if S.Verbose {
				log.Info(fmt.Sprintf("Download progress [%d/%d/%f%%]", downloader.CurrentDownloadCount, downloader.TotalDownloadsCount, downloader.GetDownloadPercentage()))
			}

			req := secreq.NewSECReqGET(S.Config)
			req.IsEtag = true

			if S.Verbose {
				log.Info("Downloading File: ", v)
			}

			resp, err := req.SendRequest(10, rateLimit, v)
			if err != nil {
				return err
			}

			responseBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			log.Info(fmt.Sprintf("File %v progress [%d/%d/%f%%] status_code_%d", v, downloader.CurrentDownloadCount, downloader.TotalDownloadsCount, downloader.GetDownloadPercentage(), resp.StatusCode))
			if download.IsErrorPage(string(responseBody)) {
				return fmt.Errorf("requested file but received an error instead")
			}

			size, err := io.Copy(ioutil.Discard, bytes.NewReader(responseBody))
			if err != nil {
				return err
			}
			time.Sleep(rateLimit)

			if S.Verbose {
				log.Info("Size of file downloaded: ", size)
			}

			downloader.CurrentDownloadCount += 1
		}

		endTime := time.Now()

		if S.Verbose {
			log.Info(fmt.Sprintf("It took %v seconds to download %v files", endTime.Sub(startTime).Seconds(), countFilesToBeDownloaded))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(benchCmd)

	benchCmd.Flags().StringVarP(&GlobalRateLimit, "rate limit in ms", "w", "", "Time to sleep in between each download")
	benchCmd.Flags().BoolVar(&GlobalGenerateTxt, "txt", false, "Generate and read from a text file")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// benchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// benchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
