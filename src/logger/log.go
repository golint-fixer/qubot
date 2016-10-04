package logger

import (
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/levels"
)

// Logger is the global application logger
var Logger levels.Levels

func init() {
	logger := log.NewLogfmtLogger(os.Stderr)
	ctx := log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
	Logger = levels.New(ctx).With("caller", log.Caller(5))
}

// Debug logs a debug event along with keyvals.
func Debug(keyvals ...interface{}) error {
	return Logger.Debug(keyvals...)
}

// Info logs an info event along with keyvals.
func Info(keyvals ...interface{}) error {
	return Logger.Info(keyvals...)
}

// Warn logs a warn event along with keyvals.
func Warn(keyvals ...interface{}) error {
	return Logger.Warn(keyvals...)
}

// Error logs an error event along with keyvals.
func Error(keyvals ...interface{}) error {
	return Logger.Error(keyvals...)
}

// Crit logs a crit event along with keyvals.
func Crit(keyvals ...interface{}) error {
	return Logger.Crit(keyvals...)
}
