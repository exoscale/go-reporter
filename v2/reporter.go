// v2 implements the v2 version of the go-reporter package.
package v2

import (
	"context"
	goerrors "errors"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/exoscale/go-reporter/v2/errors"
	"github.com/exoscale/go-reporter/v2/logging"
	"github.com/exoscale/go-reporter/v2/metrics"
)

// Reporter represents a reporter instance.
type Reporter struct {
	Errors  *errors.Reporter
	Logging *logging.Reporter
	Metrics *metrics.Reporter

	log log15.Logger

	config *Config
}

// New returns a new reporter instance.
func New(config *Config) (*Reporter, error) {
	var (
		reporter Reporter
		err      error
	)

	reporter.log = log15.New()
	reporter.log.SetHandler(log15.DiscardHandler())

	if config == nil {
		return &reporter, nil
	}

	reporter.config = config

	if config.Debug {
		reporter.log.SetHandler(log15.StderrHandler)
	}

	if config.Errors != nil {
		if reporter.Errors, err = errors.New(config.Errors); err != nil {
			return nil, err
		}
		reporter.Errors.SetLogger(reporter.log)
	}

	if config.Logging != nil {
		if reporter.Logging, err = logging.New(config.Logging); err != nil {
			return nil, err
		}
		reporter.Logging.SetLogger(reporter.log)

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
		if reporter.Metrics, err = metrics.New(config.Metrics); err != nil {
			return nil, err
		}
		reporter.Metrics.SetLogger(reporter.log)
	}

	return &reporter, nil
}

// Start starts the configured reporters.
func (r *Reporter) Start(ctx context.Context) error {
	if r.Errors != nil {
		r.log.Debug("starting errors reporter")
		if err := r.Errors.Start(ctx); err != nil {
			return err
		}
		r.log.Debug("errors reporter started")
	}

	if r.Logging != nil {
		r.log.Debug("starting logging reporter")
		if err := r.Logging.Start(ctx); err != nil {
			return err
		}
		r.log.Debug("logging reporter started")
	}

	if r.Metrics != nil {
		r.log.Debug("starting metrics reporter")
		if err := r.Metrics.Start(ctx); err != nil {
			return err
		}
		r.log.Debug("metrics reporter started")
	}

	return nil
}

// Stop stops the configured reporters.
func (r *Reporter) Stop(ctx context.Context) error {
	if r.Errors != nil {
		r.log.Debug("stopping errors reporter")
		if err := r.Errors.Stop(ctx); err != nil {
			return err
		}
		r.log.Debug("errors reporter stopped")
	}

	if r.Logging != nil {
		r.log.Debug("stopping logging reporter")
		if err := r.Logging.Stop(ctx); err != nil {
			return err
		}
		r.log.Debug("logging reporter stopped")
	}

	if r.Metrics != nil {
		r.log.Debug("stopping metrics reporter")
		if err := r.Metrics.Stop(ctx); err != nil {
			return err
		}
		r.log.Debug("metrics reporter stopped")
	}

	return nil
}

// Config returns the reporter initial configuration.
func (r *Reporter) Config() *Config {
	return r.config
}
