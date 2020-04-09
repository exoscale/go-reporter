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
	"gopkg.in/tomb.v2"

	"github.com/exoscale/go-reporter/v2/internal/debug"
)

// Exporter represents a metrics exporter to a Prometheus server.
type Exporter struct {
	registry *prom.Registry
	pm       *prometheusmetrics.PrometheusConfig

	t      *tomb.Tomb // Goroutines manager
	config *Config

	*debug.D
}

// New returns a new Prometheus metrics exporter based on provided configuration.
// If a go-metrics registry is provided, it is hooked into the exporter's internal Prometheus registry and its
// metrics will be flushed periodically to the Prometheus endpoint (flush interval to be specified in the config).
func New(config *Config, registry metrics.Registry) (*Exporter, error) {
	var exporter Exporter

	if err := config.validate(); err != nil {
		return nil, err
	}
	exporter.config = config

	exporter.D = debug.New("reporter/metrics/prometheus")
	if config.Debug {
		exporter.D.On()
	}

	exporter.Debug("enabling exporter",
		"namespace", config.Namespace,
		"subsystem", config.Subsystem,
		"flush_interval", config.FlushInterval)

	exporter.registry = prom.NewRegistry()

	if registry != nil {
		exporter.Debug("enabling go-metrics registry export to Prometheus")

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
	// Before initializing the goroutines management tomb we have to check that we actually have goroutines to
	// handle with it, otherwise it'll get stuck during shutdown (see Stop() method).
	if e.pm == nil && e.config.Listen == "" {
		return nil
	}

	e.t, _ = tomb.WithContext(ctx)

	if e.pm != nil {
		e.t.Go(e.registryFlushLoop)
	}

	if e.config.Listen != "" {
		e.t.Go(func() error {
			return e.serveHTTP(nil)
		})
	}

	return nil
}

// Stop stops the metrics exporter.
func (e *Exporter) Stop(_ context.Context) error {
	// Since tomb activation is conditional, we have to check if it has actually been activated
	// before trying to kill it otherwise we'll get stuck: https://github.com/go-tomb/tomb/issues/21
	if e.t == nil {
		return nil
	}

	e.t.Kill(nil)

	return e.t.Wait()
}

// registryFlushLoop periodically flushes the go-metrics registry provided during exporter initialization to the
// Prometheus registry. This method blocks the caller until the exporter's tomb dies.
func (e *Exporter) registryFlushLoop() error {
	tick := time.NewTicker(time.Duration(e.config.FlushInterval) * time.Second)

	e.Debug("starting go-metrics flush loop")

	for {
		select {
		case <-tick.C:
			if err := e.pm.UpdatePrometheusMetricsOnce(); err != nil {
				e.Error("unable to flush go-metrics registry", "err", err)
				return err
			}

		case <-e.t.Dying():
			tick.Stop()
			e.Debug("terminating go-metrics flush loop")
			return nil
		}
	}
}

// serveHTTP runs an HTTP server to serve the Prometheus metrics scraping endpoint. This method blocks the caller
// until the exporter's tomb dies.
func (e *Exporter) serveHTTP(server *http.Server) error {
	if server == nil {
		server = &http.Server{
			Addr:    e.config.Listen,
			Handler: e.HTTPHandler(),
		}
	}

	e.Debug("starting scraping endpoint server")

	go server.ListenAndServe()

	_ = <-e.t.Dying()
	e.Debug("terminating scraping endpoint server")
	return server.Shutdown(context.Background())
}
