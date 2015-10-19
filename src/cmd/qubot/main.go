package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

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
	err = q.Run()
	if err != nil {
		logger.Error("main", "Qubot could not be started", "error", err)
		q.Close()
		os.Exit(1)
	}

	// More advanced management on signals here: https://goo.gl/fuylKX
	var closeTime time.Time
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
	logger.Info("main", "Listening for signals")

	for sig := range sigChan {
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			logger.Info("main", "Initializing clean shutdown...")
			closeTime = time.Now()
			go q.Close()
			goto QUIT
		case syscall.SIGUSR1:
			q.Report()
		}
	}

QUIT:
	select {
	case <-sigChan:
		logger.Info("main", "Second signal received, initializing hard shutdown")
	case <-time.After(time.Second * 10):
		logger.Info("main", "Time limit reached, initializing hard shutdown")
	case <-q.Closed:
		logger.Info("main", "Shutdown completed", "took", time.Since(closeTime))
	}

	logger.Info("main", "Goroutines count", "count", runtime.NumGoroutine())
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
