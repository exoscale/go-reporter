// v2 implements the v2 version of the go-reporter package.
package v2

import (
	"context"
	goerrors "errors"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/exoscale/go-reporter/v2/errors"
	"github.com/exoscale/go-reporter/v2/internal/debug"
	"github.com/exoscale/go-reporter/v2/logging"
	"github.com/exoscale/go-reporter/v2/metrics"
)

// Reporter represents a reporter instance.
type Reporter struct {
	Errors  *errors.Reporter
	Logging *logging.Reporter
	Metrics *metrics.Reporter

	config *Config

	*debug.D
}

// New returns a new reporter instance.
func New(config *Config) (*Reporter, error) {
	var (
		reporter Reporter
		err      error
	)

	if config == nil {
		return &reporter, nil
	}

	reporter.config = config

	// IMPORTANT:
	// To use the debug logger in this package, we have to call its methods via its explicit identifier instead
	// of the methods embedded in the Reporter receiver (i.e. "reporter.D.Debug()" and not "reporter.Debug()"),
	// otherwise those methods will clash with the other logging methods implemented on the Reporter structure
	// (see logging.go file).
	reporter.D = debug.New("reporter")
	if config.Debug {
		reporter.D.On()
	}

	if config.Errors != nil {
		reporter.D.Debug("initializing errors reporter")
		config.Errors.Debug = config.Debug

		if reporter.Errors, err = errors.New(config.Errors); err != nil {
			return nil, err
		}
	}

	if config.Logging != nil {
		reporter.D.Debug("initializing logging reporter")
		config.Logging.Debug = config.Debug

		if reporter.Logging, err = logging.New(config.Logging); err != nil {
			return nil, err
		}

		// Hook the errors reporter's log handler to the logging reporter's logger
		if config.Logging.ReportErrors {
			if reporter.Errors == nil {
				return nil, goerrors.New("logging: errors reporter must be configured to enable error reporting")
			}

			reporter.Logging.SetHandler(log15.MultiHandler(
				reporter.Logging.Handler(),
				reporter.Errors.LogHandler()))
		}
	}

	if config.Metrics != nil {
		reporter.D.Debug("initializing metrics reporter")
		config.Metrics.Debug = config.Debug

		if reporter.Metrics, err = metrics.New(config.Metrics); err != nil {
			return nil, err
		}
	}

	return &reporter, nil
}

// Start starts the configured reporters.
func (r *Reporter) Start(ctx context.Context) error {
	if r.Errors != nil {
		r.D.Debug("starting errors reporter")
		if err := r.Errors.Start(ctx); err != nil {
			return err
		}
		r.D.Debug("errors reporter started")
	}

	if r.Logging != nil {
		r.D.Debug("starting logging reporter")
		if err := r.Logging.Start(ctx); err != nil {
			return err
		}
		r.D.Debug("logging reporter started")
	}

	if r.Metrics != nil {
		r.D.Debug("starting metrics reporter")
		if err := r.Metrics.Start(ctx); err != nil {
			return err
		}
		r.D.Debug("metrics reporter started")
	}

	return nil
}

// Stop stops the configured reporters.
func (r *Reporter) Stop(ctx context.Context) error {
	if r.Errors != nil {
		r.D.Debug("stopping errors reporter")
		if err := r.Errors.Stop(ctx); err != nil {
			return err
		}
		r.D.Debug("errors reporter stopped")
	}

	if r.Logging != nil {
		r.D.Debug("stopping logging reporter")
		if err := r.Logging.Stop(ctx); err != nil {
			return err
		}
		r.D.Debug("logging reporter stopped")
	}

	if r.Metrics != nil {
		r.D.Debug("stopping metrics reporter")
		if err := r.Metrics.Stop(ctx); err != nil {
			return err
		}
		r.D.Debug("metrics reporter stopped")
	}

	return nil
}

// Config returns the reporter initial configuration.
func (r *Reporter) Config() *Config {
	return r.config
}
