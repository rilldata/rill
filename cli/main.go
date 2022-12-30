package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/rilldata/rill/cli/cmd"
	"github.com/rilldata/rill/cli/pkg/version"
)

// Version details are set using -ldflags
var (
	Version   string
	Commit    string
	BuildDate string
)

func main() {
	ver := version.Version{
		Number:    Version,
		Commit:    Commit,
		Timestamp: BuildDate,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cmd.Execute(ctx, ver)
}
