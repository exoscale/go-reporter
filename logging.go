// Logging façade for reporter.

package reporter

import (
	"strings"

	"github.com/pkg/errors"
	log "gopkg.in/inconshreveable/log15.v2"
)

// Debug logs a debug message with additional context.
func (r *Reporter) Debug(msg string, ctx ...interface{}) {
	r.logger.Debug(msg, ctx...)
}

// Info logs an info message with additional context.
func (r *Reporter) Info(msg string, ctx ...interface{}) {
	r.logger.Info(msg, ctx...)
}

// Warn logs a warning message with additional context.
func (r *Reporter) Warn(msg string, ctx ...interface{}) {
	r.logger.Warn(msg, ctx...)
}

// Error logs an error message with additional context. It takes the
// error as a first argument, a descriptive message (or "") as a
// second. It will add "err" with the error as a context. It will
// return a wrapped error with the message, except if message was
// empty. In this case, it returns the original error.
func (r *Reporter) Error(err error, msg string, ctx ...interface{}) error {
	ctx = append(ctx, "err", err)
	if msg == "" {
		r.logger.Error(err.Error(), ctx...)
		return err
	}
	r.logger.Error(msg, ctx...)
	return errors.Wrap(err, msg)
}

// Write will take some bytes and log them.
func (r *Reporter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	r.logger.Debug(msg)
	return len(p), nil
}

// Logger returns the reporter logger instance.
func (r *Reporter) Logger() log.Logger {
	return r.logger
}
