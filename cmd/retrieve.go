package cmd

import (
	"fmt"
	"os"

	"github.com/equres/sec/util"
	"github.com/spf13/cobra"
)

// retrieveCmd represents the retrieve command
var retrieveCmd = &cobra.Command{
	Use:   "retrieve",
	Short: "Retrieve all the tickers from sec.gov website that are saved in db",
	Long:  `Retrieve all the tickers from sec.gov website that are saved in db`,
	Run: func(cmd *cobra.Command, args []string) {
		// Retrieving all SecTickers
		sec := util.NewSEC("https://www.sec.gov/")

		db, err := util.ConnectDB()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		tickers, err := sec.TickersGetAll(db)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(tickers)
		os.Exit(0)
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
