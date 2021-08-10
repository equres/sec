package cmd

import (
	"fmt"
	"os"

	"github.com/equres/sec/sec"
	"github.com/equres/sec/util"
	"github.com/spf13/cobra"
)

// retrieveCmd represents the retrieve command
var retrieveCmd = &cobra.Command{
	Use:   "retrieve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Retrieving all SecTickers
		sec1 := sec.NewSEC("https://www.sec.gov/")

		db, err := util.ConnectDB()
		if err != nil {
			panic(err)
		}

		tickers, err := sec1.TickersGetAll(db)
		if err != nil {
			panic(err)
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
