// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

// findCmd represents the find command
var findCmd = &cobra.Command{
	Use:   "find",
	Short: "find indexes with specific filling date",
	Long:  `find indexes with specific filling date`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please insert a date to search with format YYYY-MM-DD (e.g. 2021-08-23)")
		}

		date, err := time.Parse("2006-01-02", args[0])
		if err != nil {
			return err
		}

		formattedDate := fmt.Sprintf("%v-%v-%v", date.Year(), int(date.Month()), date.Day())
		secitemfiles, err := S.SearchByFillingDate(DB, formattedDate)
		if err != nil {
			return err
		}

		tabWriter := tabwriter.NewWriter(os.Stdout, 12, 0, 2, ' ', 0)
		fmt.Fprint(tabWriter, "Title", "\t", "Company Name", "\t", "CIK Number", "\t", "Accession Number", "\t", "XBRLFile Name", "\n")
		for _, v := range secitemfiles {
			fmt.Fprint(tabWriter, v.Title, "\t", v.CompanyName, "\t", v.CIKNumber, "\t", v.AccessionNumber, "\t", v.XbrlFile, "\n")
		}
		err = tabWriter.Flush()
		if err != nil {
			return err
		}

		if len(secitemfiles) == 0 {
			fmt.Println("There are no search results for the date provided -", formattedDate)
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