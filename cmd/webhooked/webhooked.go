package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/42atomys/webhooked"
	"github.com/42atomys/webhooked/cmd/flags"
	"github.com/42atomys/webhooked/internal/config"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

func main() {
	if err := exec(); err != nil {
		slog.Error("an error occurred", "error", err)
		os.Exit(1)
	}

	gracefulShutdown()
	os.Exit(0)
}

func exec() error {
	if err := flags.ValidateFlags(); err != nil {
		return err
	}

	if flags.Version {
		fmt.Printf("Webhooked version: %s\n", "TODO")
		return nil
	}

	if flags.Help {
		pflag.Usage()
		return nil
	}

	if flags.Init {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		// TODO: Initialize a new Webhooked configuration
		fmt.Printf("Initializing a new Webhooked configuration in %s\n", wd)

		return nil
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if flags.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if err := config.Load(flags.Config); err != nil {
		return err
	}

	webhooked.Serve(flags.Port)

	return nil
}

func gracefulShutdown() {
}
