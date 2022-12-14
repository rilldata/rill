package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/rilldata/rill/cli/cmd"
)

// These are set using -Idflags
var Version string
var Commit string
var BuildDate string

func main() {
	if Version == "" {
		Version = "dev"
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	cmd.Execute(ctx, Version, Commit, BuildDate)
}
