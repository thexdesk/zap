package main

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/hinshun/zap/command"
	"github.com/rs/zerolog"
)

func init() {
	// UNIX Time is faster and smaller than most timestamps. If you set
	// zerolog.TimeFieldFormat to an empty string, logs will write with UNIX
	// time.
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	ih := NewInterruptHandler(cancel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	defer ih.Close()

	app := command.App(ctx)
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "zap: %s\n", err)
		os.Exit(1)
	}
}
