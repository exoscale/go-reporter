// metrics implements a metrics reporter.
package metrics

import (
	"context"
	"time"

	"github.com/rcrowley/go-metrics"
	"gopkg.in/inconshreveable/log15.v2"
	"gopkg.in/tomb.v2"

	"github.com/exoscale/go-reporter/v2/metrics/prometheus"
)

// Reporter represents a metrics reporter instance.
type Reporter struct {
	Prometheus *prometheus.Exporter

	registry        metrics.Registry
	runtimeRegistry metrics.Registry

	log    log15.Logger
	t      *tomb.Tomb
	config *Config
}

// New returns a new metrics reporter instance.
func New(config *Config) (*Reporter, error) {
	var (
		reporter Reporter
		err      error
	)

	reporter.log = log15.New()
	reporter.log.SetHandler(log15.DiscardHandler())

	if config == nil {
		return nil, nil
	}

	if err := config.validate(); err != nil {
		return nil, err
	}
	reporter.config = config

	reporter.registry = metrics.NewRegistry()

	if config.WithRuntimeMetrics {
		reporter.runtimeRegistry = metrics.NewPrefixedChildRegistry(reporter.registry, "go.")
		metrics.RegisterRuntimeMemStats(reporter.runtimeRegistry)
	}

	if config.Prometheus != nil {
		if reporter.Prometheus, err = prometheus.New(config.Prometheus, reporter.registry); err != nil {
			return nil, err
		}
	}

	return &reporter, nil
}

// Register registers a metric in the internal registry.
func (r *Reporter) Register(name string, metric interface{}) error {
	return r.registry.Register(name, metric)
}

// Start starts the metrics reporter.
func (r *Reporter) Start(ctx context.Context) error {
	if r.runtimeRegistry != nil {
		r.t, _ = tomb.WithContext(ctx)

		r.t.Go(func() error {
			ticker := time.NewTicker(time.Duration(r.config.FlushInterval) * time.Second)

			for {
				select {
				case <-ticker.C:
					r.log.Debug("flushing runtime metrics to registry")
					metrics.CaptureRuntimeMemStatsOnce(r.runtimeRegistry)

				case <-r.t.Dying():
					ticker.Stop()
					return nil
				}
			}
		})
	}

	if r.Prometheus != nil {
		r.log.Debug("starting Prometheus metrics exporter")
		if err := r.Prometheus.Start(ctx); err != nil {
			return err
		}
		r.log.Debug("Prometheus metrics exporter started")
	}

	return nil
}

// Stop stops the metrics reporter.
func (r *Reporter) Stop(ctx context.Context) error {
	if r.Prometheus != nil {
		r.log.Debug("stopping metrics Prometheus exporter")
		if err := r.Prometheus.Stop(ctx); err != nil {
			return err
		}
		r.log.Debug("Prometheus metrics exporter stopped")
	}

	// Since tomb activation is conditional, we have to check if it has actually been activated
	// before trying to kill it otherwise we'll get stuck: https://github.com/go-tomb/tomb/issues/21
	if r.t != nil {
		r.t.Kill(nil)
		return r.t.Wait()
	}

	return nil
}

// SetLogger sets the metrics reporter internal logger. This is mainly for debug purposes.
func (r *Reporter) SetLogger(logger log15.Logger) {
	r.log = logger

	if r.Prometheus != nil {
		r.Prometheus.SetLogger(r.log)
	}
}
