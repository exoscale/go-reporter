package debug

import (
	"gopkg.in/inconshreveable/log15.v2"
)

// D represents a debug-specific logger that writes to stderr.
type D struct {
	prefix string
	logger log15.Logger
}

// New returns a new debug instance.
func New(prefix string) *D {
	debug := &D{
		prefix: prefix,
		logger: log15.New(),
	}
	debug.Off()

	return debug
}

// On enables the debug logger.
func (d *D) On() {
	d.logger.SetHandler(log15.StderrHandler)
}

// Off disables the debug logger.
func (d *D) Off() {
	d.logger.SetHandler(log15.DiscardHandler())
}

// Debug logs a debug-level message.
func (d *D) Debug(msg string, ctx ...interface{}) {
	if d.prefix != "" {
		msg = d.prefix + ": " + msg
	}

	d.logger.Debug(msg, ctx...)
}

// Error logs an error-level message.
func (d *D) Error(msg string, ctx ...interface{}) {
	if d.prefix != "" {
		msg = d.prefix + ": " + msg
	}

	d.logger.Error(msg, ctx...)
}
