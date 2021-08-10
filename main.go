// Copyright (c) 2021 Equres LLC. All rights reserved.
package main

import (
	"github.com/equres/sec/cmd"
	_ "github.com/lib/pq"
)

func main() {
	cmd.Execute()
}
