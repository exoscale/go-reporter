// logging implements a logs reporter.
package logging

import (
	"context"

	"gopkg.in/inconshreveable/log15.v2"
)

// Reporter represents a logging reporter instance.
type Reporter struct {
	log log15.Logger

	config *Config
}

// New returns a new logging reporter instance.
func New(config *Config) (*Reporter, error) {
	var reporter Reporter

	if config == nil {
		return nil, nil
	}

	if err := config.validate(); err != nil {
		return nil, err
	}
	reporter.config = config

	reporter.log = log15.New()
	reporter.log.SetHandler(log15.DiscardHandler())

	handlers := make([]log15.Handler, 0)
	for _, d := range reporter.config.Destinations {
		var (
			h   log15.Handler
			err error
		)

		switch d.Type {
		case "file":
			h, err = newFileHandler(d)

		case "syslog":
			h, err = newSyslogHandler(d)

		case "console":
			h, err = newConsoleHandler(d)
		}
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, h)
	}
	if len(handlers) > 0 {
		reporter.log.SetHandler(log15.MultiHandler(handlers...))
	}

	return &reporter, nil
}

// Start is a no-op operation.
func (r *Reporter) Start(_ context.Context) error {
	return nil
}

// Stop is a no-op operation.
func (r *Reporter) Stop(_ context.Context) error {
	return nil
}

// Handler returns the logging reporter's log15.Handler.
func (r *Reporter) Handler() log15.Handler {
	return r.log.GetHandler()
}

// SetHandler sets the logging reporter's log15.Handler.
func (r *Reporter) SetHandler(h log15.Handler) {
	r.log.SetHandler(h)
}

// SetLogger is a no-op operation.
func (r *Reporter) SetLogger(_ log15.Logger) {}

// Crit logs a message with a "critical" severity level.
func (r *Reporter) Crit(msg string, ctx ...interface{}) {
	r.log.Crit(msg, ctx...)
}

// Error logs a message with an "error" severity level.
func (r *Reporter) Error(msg string, ctx ...interface{}) {
	r.log.Error(msg, ctx...)
}

// Warn logs a message with a "warning" severity level.
func (r *Reporter) Warn(msg string, ctx ...interface{}) {
	r.log.Warn(msg, ctx...)
}

// Info logs a message with an "info" severity level.
func (r *Reporter) Info(msg string, ctx ...interface{}) {
	r.log.Info(msg, ctx...)
}

// Debug logs a message with a "debug" severity level.
func (r *Reporter) Debug(msg string, ctx ...interface{}) {
	r.log.Debug(msg, ctx...)
}
