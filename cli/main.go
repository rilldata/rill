package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/rilldata/rill/cli/cmd"
)

// These are set using -Idflags
var Version string
var Commit string
var BuildDate = time.Now().String()

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	cmd.Execute(ctx, Version, Commit, BuildDate)
}
