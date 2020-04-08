// metrics implements a metrics reporter.
package metrics

import (
	"context"
	"time"

	"github.com/rcrowley/go-metrics"
	"gopkg.in/tomb.v2"

	"github.com/exoscale/go-reporter/v2/internal/debug"
	"github.com/exoscale/go-reporter/v2/metrics/prometheus"
)

// Reporter represents a metrics reporter instance.
type Reporter struct {
	Prometheus *prometheus.Exporter

	registry        metrics.Registry
	runtimeRegistry metrics.Registry

	t      *tomb.Tomb // Goroutines manager
	config *Config

	*debug.D
}

// New returns a new metrics reporter instance.
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

	reporter.D = debug.New("reporter/metrics")
	if config.Debug {
		reporter.D.On()
	}

	reporter.registry = metrics.NewRegistry()

	if config.WithRuntimeMetrics {
		reporter.Debug("enabling Go runtime metrics collection")
		reporter.runtimeRegistry = metrics.NewPrefixedChildRegistry(reporter.registry, "go.")
		metrics.RegisterRuntimeMemStats(reporter.runtimeRegistry)
	}

	if config.Prometheus != nil {
		config.Prometheus.Debug = config.Debug
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
					r.Debug("flushing runtime metrics to registry")
					metrics.CaptureRuntimeMemStatsOnce(r.runtimeRegistry)

				case <-r.t.Dying():
					ticker.Stop()
					return nil
				}
			}
		})
	}

	if r.Prometheus != nil {
		r.Debug("starting Prometheus exporter")
		if err := r.Prometheus.Start(ctx); err != nil {
			return err
		}
		r.Debug("Prometheus exporter started")
	}

	return nil
}

// Stop stops the metrics reporter.
func (r *Reporter) Stop(ctx context.Context) error {
	if r.Prometheus != nil {
		r.Debug("stopping Prometheus exporter")
		if err := r.Prometheus.Stop(ctx); err != nil {
			return err
		}
		r.Debug("Prometheus exporter stopped")
	}

	// Since tomb activation is conditional, we have to check if it has actually been activated
	// before trying to kill it otherwise we'll get stuck: https://github.com/go-tomb/tomb/issues/21
	if r.t != nil {
		r.t.Kill(nil)
		return r.t.Wait()
	}

	return nil
}
