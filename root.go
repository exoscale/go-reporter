// Package reporter is a façade for reporting duties.
//
// Such a façade includes logging, error handling and metrics.
package reporter

import (
	"github.com/getsentry/raven-go"
	log "gopkg.in/inconshreveable/log15.v2"

	"github.com/exoscale/go-reporter/logger"
	"github.com/exoscale/go-reporter/metrics"
	"github.com/exoscale/go-reporter/sentry"
)

// Reporter contains the state for a reporter.
type Reporter struct {
	config  *Configuration
	logger  log.Logger
	sentry  *raven.Client
	metrics *metrics.Metrics
	prefix  string
}

// New creates a new reporter from a configuration.
func New(config Configuration) (*Reporter, error) {
	// Initialize sentry
	s, err := sentry.New(config.Sentry)
	if err != nil {
		return nil, err
	}

	// Initialize logger
	l, err := logger.New(config.Logging, sentryHandler(s, config.Sentry.Wait), config.Prefix)
	if err != nil {
		return nil, err
	}

	// Initialize metrics
	m, err := metrics.New(config.Metrics, config.Prefix)
	if err != nil {
		return nil, err
	}

	return &Reporter{
		config:  &config,
		logger:  l,
		sentry:  s,
		metrics: m,
		prefix:  config.Prefix,
	}, nil
}

// Config returns the original configuration provided during reporter initialization.
func (r *Reporter) Config() *Configuration {
	return r.config
}

// Start will start the reporter component
func (r *Reporter) Start() error {
	if r.metrics != nil {
		return r.metrics.Start()
	}
	return nil
}

// Stop will stop reporting and clean the associated resources.
func (r *Reporter) Stop() error {
	if r.sentry != nil {
		r.Debug("shutting down Sentry subsystem")
		r.sentry.Wait()
		r.sentry.Close()
	}
	if r.metrics != nil {
		r.Debug("shutting down metric subsystem")
		err := r.metrics.Stop()
		if err != nil {
			_ = r.Error(err, "fail to stop the metric reporter")
		}
	}
	r.Debug("stop reporting")
	return nil
}
