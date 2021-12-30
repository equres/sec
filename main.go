// Copyright (c) 2021 Koszek Systems. All rights reserved.
package main

import (
	"embed"

	"github.com/equres/sec/cmd"
	"github.com/equres/sec/pkg/server"
	_ "github.com/lib/pq"
)

//go:embed migrations
var migrations embed.FS

//go:embed templates/*
var templates embed.FS

//go:embed _assets/*
var assets embed.FS

func main() {
	cmd.GlobalMigrationsFS = migrations
	cmd.GlobalTemplatesFS = templates
	server.GlobalAssetsFS = assets

	// Start CLI
	cmd.Execute()
}
