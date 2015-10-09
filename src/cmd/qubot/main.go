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
	"handlers"
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

	v := appVersion()
	if version {
		fmt.Println(v)
		os.Exit(0)
	}
	logger.Info("main", v)

	cfg, err := loadConfig()
	if err != nil {
		logger.Error("main", "The configuration could not be loaded", "error", err, "path", conf)
		os.Exit(1)
	}

	// Start service
	q := qubot.Init(cfg)
	q.Handle(handlers.PingHandler, handlers.TauntHandler)
	err = q.Start()
	if err != nil {
		logger.Error("main", "Qubot could not be started", "error", err)
		q.Close()
		os.Exit(1)
	}

	// More advanced management on signals here: https://goo.gl/fuylKX
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGINT, // aka os.Interrupt
		syscall.SIGTERM,
		syscall.SIGUSR1)

SELECT:
	select {
	case sig := <-sigChan:
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			q.Close()
		case syscall.SIGUSR1:
			q.Report()
			goto SELECT
		}
	case <-q.Done():
		logger.Info("main", "Qubot stopped")
	}

	logger.Info("main", "¡Adiós!")
}

func loadConfig() (*qubot.Config, error) {
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
		logger.Info("main", "Using config file", "path", conf)
	}

	return cfg, config.Validate(cfg)
}

func appVersion() string {
	var versionString bytes.Buffer

	fmt.Fprintf(&versionString, "%s v%s", app.Name, app.Version)
	if app.VersionPrerelease != "" {
		fmt.Fprintf(&versionString, "-%s", app.VersionPrerelease)
		if app.Revision != "" {
			fmt.Fprintf(&versionString, " (%s)", app.Revision)
		}
	}

	return versionString.String()
}
