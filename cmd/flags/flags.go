package flags

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/spf13/pflag"
)

var (
	Config   string
	Help     bool
	Init     bool
	Port     int
	Validate bool
	Version  bool
	Debug    bool
)

const usage = `Usage: webhooked [options]

Options:
	-h, --help       Show this help message and exit
	--version        Show the version and exit
	
	-c, --config     The path to the configuration file
	-i, --init       Initialize the webhooked configuration
	-p, --port       The port to listen on
	-v, --validate   Validate the webhooked configuration
`

func init() {
	slog.SetDefault(
		slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}),
		),
	)

	pflag.Usage = usageFn
	pflag.StringVarP(&Config, "config", "c", "webhooked.yaml", "The path to the configuration file.")
	pflag.BoolVarP(&Init, "init", "i", false, "Initialize a new Webhooked configuration.")
	pflag.BoolVarP(&Help, "help", "h", false, "Show Webhooked usage.")
	pflag.IntVarP(&Port, "port", "p", 8080, "The port to listen on.")
	pflag.BoolVar(&Version, "version", false, "Show Webhooked version.")
	pflag.BoolVarP(&Validate, "validate", "v", false, "Validate the Webhooked configuration.")
	pflag.BoolVarP(&Debug, "debug", "d", false, "Enable debug logging.")

	pflag.Parse()
}

func ValidateFlags() error {
	if Port < 1 || Port > 65535 {
		return fmt.Errorf("invalid port number: %d (must be between 1 and 65535)", Port)
	}

	if Config == "" {
		return fmt.Errorf("config file path is required")
	}

	return nil
}

func usageFn() {
	log.Print(usage)
	pflag.PrintDefaults()
}
