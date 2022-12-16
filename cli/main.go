package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/rilldata/rill/cli/cmd"
)

// Version info is set using -Idflags.
var (
	Version   string
	Commit    string
	BuildDate string
)

func main() {
	if Version == "" {
		Version = "dev"
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	cmd.Execute(ctx, Version, Commit, BuildDate)
}
