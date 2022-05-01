package cmd

import (
	"net/url"

	"github.com/equres/sec/pkg/seccache"
	"github.com/equres/sec/pkg/secutil"
	"github.com/spf13/cobra"
)

var NumberToAdd int

// queueCmd represents the queue command
var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "Add stock information to the queue",
	Long:  `Add stock information to the queue`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filings, err := secutil.GetMostRecentFilings(S, DB, NumberToAdd)
		if err != nil {
			return err
		}

		S.Log("Adding filings to the queue...")

		sc := seccache.NewSECCache(DB, S)
		for _, filing := range filings {
			xbrlurl, err := url.Parse(filing.XbrlURL)
			if err != nil {
				return err
			}

			xbrlurl.Host = "equres.com"
			filing.XbrlURL = xbrlurl.String()

			err = sc.AddToFilingsNotificationQueue(filing)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(queueCmd)

	queueCmd.PersistentFlags().IntVarP(&NumberToAdd, "number", "n", 25, "Number of stocks to add to the queue (e.g. 25)")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// queueCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// queueCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
