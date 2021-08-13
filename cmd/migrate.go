// Copyright (c) 2021 Equres LLC. All rights reserved.
package cmd

import (
	"errors"

	"github.com/equres/sec/util"
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "function to migrate the db up or down",
	Long:  `function to migrate the db up or down`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			err := errors.New("please type 'up' to migrate up and 'down' to migrate down (e.g. sec migrate up)")
			return err
		}

		db, err := util.ConnectDB()
		if err != nil {
			return err
		}

		switch args[0] {
		case "up":
			err = util.MigrateUp(db)
			if err != nil {
				return err
			}
		case "down":
			err = util.MigrateDown(db)
			if err != nil {
				return err
			}
		default:
			err := errors.New("please type 'up' to migrate up and 'down' to migrate down (e.g. sec migrate up)")
			return err
		}
		return nil
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
