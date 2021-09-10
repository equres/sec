// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"embed"

	"github.com/equres/sec/pkg/server"
	"github.com/spf13/cobra"
)

var GlobalTemplatesFS embed.FS

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server to serve files",
	Long:  `Start the HTTP server to serve files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		server, err := server.NewServer(DB, RootConfig, GlobalTemplatesFS)
		if err != nil {
			return err
		}

		err = server.StartServer()
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
