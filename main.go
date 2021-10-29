// Copyright (c) 2021 Koszek Systems. All rights reserved.
package main

import (
	"embed"

	"github.com/equres/sec/cmd"
	_ "github.com/lib/pq"
)

//go:embed migrations
var migrations embed.FS

//go:embed templates/*
var templates embed.FS

func main() {
	cmd.GlobalMigrationsFS = migrations
	cmd.GlobalTemplatesFS = templates

	cmd.Execute()
}
