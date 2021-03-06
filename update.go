// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"flag"
)

type Update struct {
	Common
}

func (cmd *Update) Name() string { return "update" }

func (cmd *Update) Parse(args []string) error {
	set := flag.NewFlagSet("update", flag.ExitOnError)
	return set.Parse(args)
}

func (cmd *Update) Exec() {
	srcdir := cmd.Path("src")
	if !Exists(srcdir) {
		Fatalf("src directory %v missing, run setup", srcdir)
	}

	cmd.DeleteNonRepos()
	cmd.VendorModules()
	cmd.FlattenVendor()
}
