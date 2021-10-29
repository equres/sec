// Copyright (c) 2021 Koszek Systems. All rights reserved.
package cmd

import (
	"embed"

	log "github.com/sirupsen/logrus"

	"github.com/equres/sec/pkg/database"
	"github.com/spf13/cobra"
)

var GlobalMigrationsFS embed.FS

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "function to migrate the db up or down",
	Long:  `function to migrate the db up or down`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) == 0 {
			log.Info("please type 'up' to migrate up and 'down' to migrate down (e.g. sec migrate up)")
			return nil
		}

		switch args[0] {
		case "up":
			err = database.MigrateUp(DB, GlobalMigrationsFS, RootConfig)
			if err != nil {
				return err
			}
			log.Info("Successfully migrated the DB UP")
		case "down":
			err = database.MigrateDown(DB, GlobalMigrationsFS, RootConfig)
			if err != nil {
				return err
			}
			log.Info("Successfully migrated the DB DOWN")
		default:
			log.Info("please type 'up' to migrate up and 'down' to migrate down (e.g. sec migrate up)")
			return nil
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
