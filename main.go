// Copyright (c) 2021 Equres LLC. All rights reserved.
package main

import (
	"embed"

	"github.com/equres/sec/cmd"
	_ "github.com/lib/pq"
)

//go:embed migrations
var migrations embed.FS

func main() {

	cmd.GlobalMigrationsFS = migrations

	cmd.Execute()
}
