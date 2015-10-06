package logger

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/levels"
)

// Level type
type Level uint8

// Convert the Level to a string. E.g. PanicLevel becomes "panic".
func (level Level) String() string {
	switch level {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warning"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	case CritLevel:
		return "cric"
	}

	return "unknown"
}

// ErrNotValidLevel is a type of error returned when the level is not known.
var ErrNotValidLevel = errors.New("log: not a valid Level")

// ParseLevel takes a string level and returns its log level constant.
func ParseLevel(level string) (Level, error) {
	switch level {
	case "crit", "critical":
		return CritLevel, nil
	case "fatal":
		return FatalLevel, nil
	case "error":
		return ErrorLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	}

	var l Level
	return l, ErrNotValidLevel
}

// These are the different logging levels.
const (
	// CritLevel level, highest level of severity. Logs and then calls panic
	// with the message passed to Debug, Info, ...
	CritLevel Level = iota
	// FatalLevel level. Logs and then calls `os.Exit(1)`. It will exit even
	// if the logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be
	// noted. Commonly used for hooks to send errors to an error tracking
	// service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on
	// inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose
	// logging.
	DebugLevel
)

// currentLevel has a default and it's updated later by levelFlag
var currentLevel = InfoLevel

var (
	logger = log.NewLogfmtLogger(os.Stderr)
	ctx    = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)

	// Logger is the global application logger
	Logger = levels.New(ctx).With("caller", log.Caller(5))
)

type levelFlag struct{}

// String implements flag.Value.
func (f levelFlag) String() string {
	return fmt.Sprint(currentLevel)
}

// Set implements flag.Value.
func (f levelFlag) Set(level string) (err error) {
	currentLevel, err = ParseLevel(level)
	return err
}

func init() {
	// In order for this flag to take effect, the user of the package must
	// call flag.Parse() before logging anything.
	flag.Var(levelFlag{}, "log.level", "Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal, crit].")
}

// Debug logs a debug event along with keyvals.
func Debug(keyvals ...interface{}) (err error) {
	if currentLevel >= DebugLevel {
		return Logger.Debug(keyvals...)
	}
	return
}

// Info logs an info event along with keyvals.
func Info(keyvals ...interface{}) (err error) {
	if currentLevel >= InfoLevel {
		return Logger.Info(keyvals...)
	}
	return
}

// Warn logs a warn event along with keyvals.
func Warn(keyvals ...interface{}) (err error) {
	if currentLevel >= WarnLevel {
		return Logger.Warn(keyvals...)
	}
	return
}

// Error logs an error event along with keyvals.
func Error(keyvals ...interface{}) (err error) {
	if currentLevel >= ErrorLevel {
		return Logger.Error(keyvals...)
	}
	return
}

// Crit logs a crit event along with keyvals.
func Crit(keyvals ...interface{}) (err error) {
	if currentLevel >= CritLevel {
		return Logger.Crit(keyvals...)
	}
	return
}
