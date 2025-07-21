package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/42atomys/webhooked"
	"github.com/42atomys/webhooked/cmd/flags"
	"github.com/42atomys/webhooked/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
)

type app struct {
	config *config.Config
	server *webhooked.Server
}

func main() {
	// Create context that will be cancelled on interrupt signals
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	if err := exec(ctx); err != nil {
		log.Error().Err(err).Msg("application failed to start")
		os.Exit(1)
	}

}

func exec(ctx context.Context) error {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	log.Logger = log.Logger.Level(zerolog.InfoLevel)
	if flags.Debug {
		log.Logger = log.Logger.Level(zerolog.DebugLevel)
	}

	if err := flags.ValidateFlags(); err != nil {
		return err
	}

	if flags.Version {
		fmt.Printf("Webhooked version: %s\n", webhooked.Version)
		return nil
	}

	if flags.Help {
		pflag.Usage()
		return nil
	}

	if flags.Init {
		return initializeConfig()
	}

	if flags.Validate {
		if _, err := config.Load(flags.Config); err != nil {
			return fmt.Errorf("configuration validation failed: %w", err)
		}
		fmt.Println("✅ Configuration is valid")
		return nil
	}

	cfg, err := config.Load(flags.Config)
	if err != nil {
		return err
	}

	// Create server instance
	server, err := webhooked.NewServer(cfg, flags.Port)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	app := &app{
		config: cfg,
		server: server,
	}

	// Start server in goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		log.Info().Int("port", flags.Port).Msg("starting webhooked server")
		if err := app.server.Start(); err != nil {
			serverErrChan <- fmt.Errorf("server failed to start: %w", err)
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		log.Info().Msg("shutdown signal received, gracefully shutting down...")
		return app.gracefulShutdown()
	case err := <-serverErrChan:
		return err
	}
}

func (a *app) gracefulShutdown() error {
	if a.server == nil {
		log.Info().Msg("no server to shutdown")
		return nil
	}

	// Give the server 30 seconds to gracefully shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Info().Msg("gracefully shutting down server...")
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server shutdown failed")
		return err
	}

	log.Info().Msg("server shutdown completed")
	return nil
}

func initializeConfig() error {
	var configPath string
	if filepath.IsAbs(flags.Config) {
		configPath = flags.Config
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		configPath = fmt.Sprintf("%s/%s", wd, flags.Config)
	}
	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("configuration file already exists at %s", configPath)
	}

	exampleConfig := `apiVersion: v1alpha2
kind: Configuration
metadata:
  name: example-webhooked-config
specs:
- metricsEnabled: true
  webhooks:
  - name: example-webhook
    entrypointUrl: /example
    security:
      type: noop
    storage:
    - type: noop
    response:
      statusCode: 200
      contentType: application/json
      formatting:
        templateString: |
          {
            "message": "Webhook received successfully",
            "timestamp": "{{ now }}"
          }
`

	if err := os.WriteFile(configPath, []byte(exampleConfig), 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	fmt.Println("✅ Webhooked configuration initialized at ", configPath)
	fmt.Println("📝 Edit the configuration file to customize your webhook endpoints")
	fmt.Println("🚀 Start the server with: webhooked serve --config \n", configPath)

	return nil
}
