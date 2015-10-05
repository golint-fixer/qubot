package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"app"
	"config"
	"logger"
	"qubot"
)

var (
	conf    string
	version bool
)

func init() {
	flag.BoolVar(&version, "version", false, "Show version")
	flag.StringVar(&conf, "conf", "", "Configuration file")
}

func main() {
	flag.Parse()

	if version {
		printVersion()
		os.Exit(0)
	}

	cfg, err := loadConfig()
	if err != nil {
		logger.Error("msg", "The configuration could not be loaded", "error", err, "path", conf)
		os.Exit(1)
	}

	// Start service
	qubot, err := qubot.New(cfg)
	if err != nil {
		logger.Error("msg", "Qubot could not be started", "error", err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	select {
	case <-sigChan:
		qubot.Close()
	case <-qubot.Quit:
		logger.Info("msg", "Qubot stopped unexpectedly")
	}
}

func loadConfig() (*config.Config, error) {
	cfg := config.DefaultConfig

	if conf == "" {
		if cfgfile, err := config.File(); err == nil {
			conf = cfgfile
		}
	}

	if conf != "" {
		var err error
		cfg, err = config.Load(conf)
		if err != nil {
			return nil, err
		}
		logger.Debug("msg", "Using config file", "path", conf)
	}

	return cfg, nil
}

func printVersion() {
	var versionString bytes.Buffer

	fmt.Fprintf(&versionString, "%s v%s", app.Name, app.Version)
	if app.VersionPrerelease != "" {
		fmt.Fprintf(&versionString, "-%s", app.VersionPrerelease)
		if app.Revision != "" {
			fmt.Fprintf(&versionString, " (%s)", app.Revision)
		}
	}
	fmt.Fprintf(&versionString, "\n")

	fmt.Printf(versionString.String())
}
