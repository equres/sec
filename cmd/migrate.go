// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/equres/sec/util"
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "function to migrate the db up or down",
	Long:  `function to migrate the db up or down`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			err := errors.New("please type 'up' to migrate up and 'down' to migrate down (e.g. sec migrate up)")
			fmt.Println(err)
			os.Exit(1)
		}

		db, err := util.ConnectDB()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		switch args[0] {
		case "up":
			err = util.MigrateUp(db)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		case "down":
			err = util.MigrateDown(db)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		default:
			err := errors.New("please type 'up' to migrate up and 'down' to migrate down (e.g. sec migrate up)")
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// migrateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// migrateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
