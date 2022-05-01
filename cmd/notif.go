package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Destination string
var Count uint

// notifCmd represents the notif command
var notifCmd = &cobra.Command{
	Use:   "notif",
	Short: "Sends a notification to a chosen Social Media account",
	Long:  `Sends a notification to a chosen Social Media account`,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch Destination {
		case "twitter":
			// Send a Tweet

		default:
			return fmt.Errorf("Invalid destination: %s", Destination)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(notifCmd)

	notifCmd.PersistentFlags().StringVarP(&Destination, "dest", "d", "", "Social Media to send notification to (e.g. twitter)")
	notifCmd.PersistentFlags().UintVarP(&Count, "count", "c", 5, "Number of filings to include in the notification (e.g. 5)")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// notifCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// notifCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
