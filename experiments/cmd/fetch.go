package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var RateLimit string
var ProxiesFilePath string
var ProxyChangeOccurence int

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch URLs from a text file",
	Long:  `Fetch URLs from a text file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("Please provide a path to the file to read URLs from (e.g. fetch fileURLs.txt)")
		}

		var err error
		rateLimit := time.Duration(0)

		if RateLimit != "" {
			rateLimit, err = time.ParseDuration(fmt.Sprintf("%vms", RateLimit))
			if err != nil {
				return err
			}
		}

		var proxies []string
		if ProxiesFilePath != "" {
			proxies, err = LoadProxies(ProxiesFilePath)
			if err != nil {
				return err
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

		err = fetchFiles(fileURLs, rateLimit, proxies, ProxyChangeOccurence)
		if err != nil {
			return err
		}

		endTime := time.Now()

		fmt.Printf("It took %v seconds to download %v files", endTime.Sub(startTime).Seconds(), len(fileURLs))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	fetchCmd.PersistentFlags().StringVarP(&RateLimit, "rate-limit", "w", "", "Time to sleep in between each download")
	fetchCmd.PersistentFlags().StringVarP(&ProxiesFilePath, "proxies-file", "p", "", "Path to proxies file (e.g. ../proxies.txt)")
	fetchCmd.PersistentFlags().IntVarP(&ProxyChangeOccurence, "proxy-change", "o", 0, "Number of files to download before changing proxy")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fetchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fetchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func fetchFiles(fileURLs []string, rateLimit time.Duration, proxies []string, proxyChangeOccurence int) error {
	client := &http.Client{}

	var currentDownloadCount int

	if len(proxies) > 0 {
		rand.Seed(time.Now().Unix())
		proxyNum := rand.Intn(len(proxies))
		proxyUrl, err := url.Parse(fmt.Sprintf("http://%v", proxies[proxyNum]))
		if err != nil {
			return err
		}

		client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
	}

	for _, fileURL := range fileURLs {
		if len(proxies) > 0 && proxyChangeOccurence > 0 && currentDownloadCount%proxyChangeOccurence == 0 {
			proxyNum := rand.Intn(len(proxies))
			proxyUrl, err := url.Parse(fmt.Sprintf("http://%v", proxies[proxyNum]))
			if err != nil {
				return err
			}

			if proxies[proxyNum] == "" {
				client = &http.Client{}
			} else {
				client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
			}
			fmt.Println("Proxy changed to:", proxyUrl)
		}

		fmt.Println("Downloading file: ", fileURL)

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

		fmt.Println("File size: ", size)

		currentDownloadCount++

		fmt.Println(fmt.Sprintf("File progress [%d/%d] status_code_%d", currentDownloadCount, len(fileURLs), resp.StatusCode))

		time.Sleep(rateLimit)
	}
	return nil
}

// Create function LoadProxies
func LoadProxies(filePath string) ([]string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	proxies := strings.Split(string(data), "\n")

	return proxies, nil
}
