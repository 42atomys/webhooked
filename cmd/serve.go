/*
Package cmd : cobra package

# Copyright © 2022 42Stellar

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"atomys.codes/webhooked/internal/config"
	"atomys.codes/webhooked/internal/server"
)

var (
	flagPort *int
	// serveCmd represents the serve command
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "serve the http server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := config.Load(configFilePath); err != nil {
				log.Fatal().Err(err).Msg("invalid configuration")
			}

			srv, err := server.NewServer(*flagPort)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to create server")
			}

			log.Fatal().Err(srv.Serve()).Msg("Error during server start")
		},
	}
)

func init() {
	rootCmd.AddCommand(serveCmd)

	flagPort = serveCmd.Flags().IntP("port", "p", 8080, "port to listen on")
}
