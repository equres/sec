// Copyright (c) 2021 Koszek Systems. All rights reserved.
package cmd

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// findCmd represents the find command
var findCmd = &cobra.Command{
	Use:   "find",
	Short: "find indexes with specific filling date",
	Long:  `find indexes with specific filling date`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			log.Info("please insert a date to search with format YYYY-MM-DD (e.g. 2021-08-23)")
			return nil
		}

		date, err := time.Parse("2006-01-02", args[0])
		if err != nil {
			return err
		}

		secitemfiles, err := S.SearchByFilingDate(DB, date, date)
		if err != nil {
			return err
		}

		log.Info("Title\tCompany Name\tCIK Number\tAccession Number\tXBRLFile Name\n")
		for _, v := range secitemfiles {
			log.Info(v.Title, "\t", v.CompanyName, "\t", v.CIKNumber, "\t", v.AccessionNumber, "\t", v.XbrlFile, "\n")
		}

		if len(secitemfiles) == 0 {
			formattedDate := fmt.Sprintf("%v-%v-%v", date.Year(), int(date.Month()), date.Day())
			log.Info("There are no search results for the date provided -", formattedDate)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(findCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// findCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// findCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
