package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/Freedom-Club-Sec/Coldwire-server/internal/authenticate"
	"github.com/Freedom-Club-Sec/Coldwire-server/internal/config"
	"github.com/Freedom-Club-Sec/Coldwire-server/internal/data"
	"github.com/Freedom-Club-Sec/Coldwire-server/internal/httpserver"
)

type CLIFlags struct {
	ConfigPath string
	Host       string
	Port       int
}

func parseFlags() (*CLIFlags, error) {
	const (
		defaultConfig = "configs/config.json"
		configUsage   = "Path to JSON configuration file"

		defaultHost = "127.0.0.1"
		hostUsage   = "Server address to listen on"

		defaultPort = 8000
		portUsage   = "Server port to listen on"
	)

	f := &CLIFlags{}

	flag.StringVar(&f.ConfigPath, "config", defaultConfig, configUsage)
	flag.StringVar(&f.ConfigPath, "c", defaultConfig, configUsage+" (shorthand)")

	flag.StringVar(&f.Host, "host", defaultHost, hostUsage)
	flag.StringVar(&f.Host, "h", defaultHost, hostUsage+" (shorthand)")

	flag.IntVar(&f.Port, "port", defaultPort, portUsage)
	flag.IntVar(&f.Port, "p", defaultPort, portUsage+" (shorthand)")

	flag.Parse()

	// The reason we don't just use uint16 for f.Port, is because we
	// would still need to convert it back to int to be accepted by
	// other functions
	if f.Port < 0 || f.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", f.Port)
	}

	return f, nil
}

func main() {
	flags, err := parseFlags()
	if err != nil {
		slog.Error("Invalid CLI flags", "error", err)
		os.Exit(1)
	}

	cfg, err := config.Load(flags.ConfigPath)
	if err != nil {
		slog.Error("Failed to parse config file", "path", flags.ConfigPath, "error", err)
		os.Exit(1)
	}

	slog.Info("Initializing storage services", "UserStorage", cfg.UserStorage, "DataStorage", cfg.DataStorage)
	userSvc, err := authenticate.NewUserService(cfg)
	if err != nil {
		slog.Error("Error while initializing UserStorage service", "error", err)
		os.Exit(1)
	}

	dataSvc, err := data.NewDataService(cfg, userSvc.Store)
	if err != nil {
		slog.Error("Error while initializing DataStorage service", "error", err)
		os.Exit(1)
	}

	dbSvcs := httpserver.DBServices{
		UserService: userSvc,
		DataService: dataSvc,
	}

	// Ensures the database services closes properly before the server exits.
	defer dbSvcs.UserService.Store.ExitCleanup()
	defer dbSvcs.DataService.Store.ExitCleanup()

	// Clears up all previous challenges, clean slate basically.
	dbSvcs.UserService.Store.CleanupChallenges()

	slog.Info("Server starting",
		"host", flags.Host,
		"port", flags.Port,
		"configPath", flags.ConfigPath,
	)

	srv := httpserver.New(flags.Host, flags.Port, cfg, &dbSvcs)
	if err := srv.Start(); err != nil {
		slog.Error("server crashed", "error", err)
		os.Exit(1)
	}

}
