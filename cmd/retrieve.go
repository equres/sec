package cmd

import (
	"fmt"

	"github.com/equres/sec/util"
	"github.com/spf13/cobra"
)

// retrieveCmd represents the retrieve command
var retrieveCmd = &cobra.Command{
	Use:   "retrieve",
	Short: "Retrieve all the tickers from sec.gov website that are saved in db",
	Long:  `Retrieve all the tickers from sec.gov website that are saved in db`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return util.CheckMigration()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Retrieving all SecTickers
		sec := util.NewSEC("https://www.sec.gov/")

		db, err := util.ConnectDB()
		if err != nil {
			return err
		}

		tickers, err := sec.TickersGetAll(db)
		if err != nil {
			return err
		}
		fmt.Println(tickers)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(retrieveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// retrieveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// retrieveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
