package cmd

import (
	"fmt"
	"time"

	"github.com/equres/sec/database"
	"github.com/equres/sec/sec"
	"github.com/spf13/cobra"
)

// dowzCmd represents the dowz command
var dowzCmd = &cobra.Command{
	Use:   "dowz",
	Short: "download all of the referenced file from XBRL index as ZIP files",
	Long:  `download all of the referenced file from XBRL index as ZIP files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.ConnectDB(RootConfig)
		if err != nil {
			return err
		}

		s, err := sec.NewSEC(RootConfig)
		if err != nil {
			return err
		}

		s.Verbose, err = cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}

		worklist, err := sec.WorklistWillDownloadGet(db)
		if err != nil {
			return err
		}

		err = s.DownloadIndex()
		if err != nil {
			return err
		}

		for _, v := range worklist {
			date, err := time.Parse("2006-1", fmt.Sprintf("%d-%d", v.Year, v.Month))
			if err != nil {
				return err
			}
			formatted := date.Format("2006-01")

			fileURL := fmt.Sprintf("%v/Archives/edgar/monthly/xbrlrss-%v.xml", s.Config.Main.CacheDir, formatted)

			rssFile, err := s.ParseRSSGoXML(fileURL)
			if err != nil {
				return err
			}

			total_count := len(rssFile.Channel.Item)
			var current_count int
			for _, v1 := range rssFile.Channel.Item {
				err = s.DownloadFile(v1.Enclosure.URL, s.Config)
				if err != nil {
					return err
				}
				time.Sleep(1 * time.Second)
				current_count++
				if !s.Verbose {
					fmt.Printf("\r[%d/%d files already downloaded]. Will download %d remaining files. Pass --verbose to see progress report", current_count, total_count, (total_count - current_count))
				}

				if s.Verbose {
					fmt.Printf("[%d/%d] %s downloaded...\n", current_count, total_count, time.Now().Format("2006-01-02 03:04:05"))
				}
			}
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
