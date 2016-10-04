package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

	// Use default location if none given
	if conf == "" {
		if cfgfile, err := config.File(); err == nil {
			conf = cfgfile
		}
	}

	cfg := config.DefaultConfig
	if conf != "" {
		var err error
		cfg, err = config.Load(conf)
		if err != nil {
			logger.Error("msg", err)
			os.Exit(1)
		}
		logger.Debug("msg", "Using config file", "path", conf)
	}

	// Wait for reload or termination signals. Start the handler for SIGHUP
	// as early as possible, but ignore it until we are ready to handle
	// reloading our config.
	hup := make(chan os.Signal)
	hupReady := make(chan bool)
	signal.Notify(hup, syscall.SIGHUP)
	go func() {
		<-hupReady
		for {
			select {
			case <-hup:
			}
			// reloadConfig(cfg.configFile, reloadables...)
			reloadConfig()
		}
	}()

	_, err := qubot.New(cfg)
	if err != nil {
		logger.Error("msg", err)
	}

	// Wait for reload or termination signals.
	close(hupReady) // Unblock SIGHUP handler.

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	select {
	case <-term:
		logger.Warn("msg", "Received SIGTERM, exiting gracefully...")
	}

	logger.Info("msg", "See you next time!")
}

func reloadConfig() {
	logger.Error("msg", "reloadConfig() is not implemented yet!")
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
