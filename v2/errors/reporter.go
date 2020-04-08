// errors implements an errors/crashes reporter.
package errors

import (
	"context"
	"errors"
	"time"

	"github.com/getsentry/sentry-go"
	"gopkg.in/inconshreveable/log15.v2"

	"github.com/exoscale/go-reporter/v2/internal/debug"
)

// Reporter represents an errors reporter instance.
type Reporter struct {
	sentry *sentry.Client

	config *Config

	*debug.D
}

// New returns a new errors reporter instance.
func New(config *Config) (*Reporter, error) {
	var (
		reporter Reporter
		err      error
	)

	if config == nil {
		return nil, nil
	}

	if err := config.validate(); err != nil {
		return nil, err
	}
	reporter.config = config

	reporter.D = debug.New("reporter/errors")
	if config.Debug {
		reporter.D.On()
	}

	reporter.Debug("initializing Sentry client")

	reporter.sentry, err = sentry.NewClient(sentry.ClientOptions{
		Dsn:              config.DSN,
		AttachStacktrace: true,
	})
	if err != nil {
		return nil, err
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

// LogHandler is a log15.Handler that sends an event to Sentry if an error-level record message is logged.
func (r *Reporter) LogHandler() log15.Handler {
	return log15.FuncHandler(func(rec *log15.Record) error {
		if rec.Lvl > log15.LvlError {
			return nil
		}

		r.sentry.CaptureException(errors.New(rec.Msg), nil, sentryEventFromLogRecord(rec))
		if r.config.Wait {
			r.sentry.Flush(5 * time.Second)
		}

		return nil
	})
}

// PanicHandler is a function that recovers from a panic and sends an event to Sentry. If a fn function is provided it
// will be executed with the recovered panic value as parameter before returning – typically to exit the program with
// a non-nil return code. This handler should be used with a defer statement at the beginning of the program to watch
// for.
func (r *Reporter) PanicHandler(fn func(interface{})) {
	if re := recover(); re != nil {
		r.sentry.Recover(re, nil, sentryEventFromPanic(re))

		if r.config.Wait {
			r.sentry.Flush(5 * time.Second)
		}

		if fn != nil {
			fn(re)
		}
	}
}

// SetSentryTransport sets the errors reporter's Sentry client transport. This is mainly for testing purposes.
func (r *Reporter) SetSentryTransport(t sentry.Transport) {
	r.sentry.Transport = t
}
