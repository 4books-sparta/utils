package utils

import (
	"os"

	"github.com/go-kit/kit/log"
)

type LoggingReporter struct {
	logger log.Logger
}

type ErrorReporter interface {
	Report(error, ...string)
	Message(string, ...string)
}

func NewLoggingReporter(logger log.Logger) ErrorReporter {
	return &LoggingReporter{logger}
}

func (l *LoggingReporter) Report(err error, args ...string) {
	in := []interface{}{"error", err.Error()}
	for _, a := range args {
		in = append(in, a)
	}

	_ = l.logger.Log(in...)
}

func (l *LoggingReporter) Message(msg string, args ...string) {
	in := []interface{}{msg, "::"}
	for _, a := range args {
		in = append(in, a)
	}

	_ = l.logger.Log(in...)
}

func NewLogger() log.Logger {
	logger := log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	return logger
}
