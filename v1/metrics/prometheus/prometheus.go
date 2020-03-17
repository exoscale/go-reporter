// prometheus implements a Prometheus-compatible metrics exporter.
package prometheus

import (
	"context"
	"net/http"
	"time"

	prometheusmetrics "github.com/deathowl/go-metrics-prometheus"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rcrowley/go-metrics"
	"gopkg.in/inconshreveable/log15.v2"
	"gopkg.in/tomb.v2"
)

// Exporter represents a metrics exporter to a Prometheus server.
type Exporter struct {
	config   *Config
	registry *prom.Registry
	pm       *prometheusmetrics.PrometheusConfig

	log log15.Logger
	t   *tomb.Tomb
}

// New returns a new Prometheus metrics exporter based on provided configuration.
// If a go-metrics registry is provided, it is hooked into the exporter's internal Prometheus registry and its
// metrics will be flushed periodically to the Prometheus endpoint (flush interval to be specified in the config).
func New(config *Config, registry metrics.Registry) (*Exporter, error) {
	var exporter Exporter

	exporter.log = log15.New()
	exporter.log.SetHandler(log15.DiscardHandler())

	if err := config.validate(); err != nil {
		return nil, err
	}
	exporter.config = config

	exporter.registry = prom.NewRegistry()

	if registry != nil {
		exporter.pm = prometheusmetrics.NewPrometheusProvider(
			registry,
			exporter.config.Namespace,
			exporter.config.Subsystem,
			exporter.registry,
			time.Duration(exporter.config.FlushInterval)*time.Second)
	}

	return &exporter, nil
}

// HTTPHandler returns an HTTP.Handler exposing the Prometheus exporter registry.
func (e *Exporter) HTTPHandler() http.Handler {
	return promhttp.HandlerFor(e.registry, promhttp.HandlerOpts{})
}

// Register registers the provided metric to the Prometheus exporter registry.
func (e *Exporter) Register(m prom.Collector) error {
	return e.registry.Register(m)
}

// MustRegister registers the provided metric to the Prometheus exporter registry, and panics if an error occur.
func (e *Exporter) MustRegister(m prom.Collector) {
	e.registry.MustRegister(m)
}

// Start starts the metrics exporter.
func (e *Exporter) Start(ctx context.Context) error {
	if e.pm != nil {
		e.t, _ = tomb.WithContext(ctx)

		e.t.Go(func() error {
			tick := time.NewTicker(time.Duration(e.config.FlushInterval) * time.Second)

			for {
				select {
				case <-tick.C:
					e.log.Debug("flushing go-metrics registry to Prometheus registry")
					if err := e.pm.UpdatePrometheusMetricsOnce(); err != nil {
						e.log.Error("unable to flush go-metrics registry to Prometheus registry",
							"err", err)
						return err
					}

				case <-e.t.Dying():
					tick.Stop()
					return nil
				}
			}
		})
	}

	return nil
}

// Stop stops the metrics exporter.
func (e *Exporter) Stop(ctx context.Context) error {
	// Since tomb activation is conditional, we have to check if it has actually been activated
	// before trying to kill it otherwise we'll get stuck: https://github.com/go-tomb/tomb/issues/21
	if e.t != nil {
		e.t.Kill(nil)
		return e.t.Wait()
	}

	return nil
}

// SetLogger sets the metrics exporter internal logger.
func (e *Exporter) SetLogger(logger log15.Logger) {
	e.log = logger
}
