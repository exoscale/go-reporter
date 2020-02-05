// Package metrics handles metrics.
//
// This is a wrapper around Code Hale's Metrics library. It also
// setups the supported exporters.
package metrics

import (
	"errors"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rcrowley/go-metrics"
	"gopkg.in/tomb.v2"
)

const runtimeMetricsInterval = 5 * time.Second

// Metrics represents the internal state of the metric subsystem.
type Metrics struct {
	config   Configuration
	prefix   string
	Registry metrics.Registry

	t tomb.Tomb
}

// New creates a new metric registry and setup the appropriate
// exporters. The provided prefix is used for system-wide metrics.
func New(configuration Configuration, prefix string) (*Metrics, error) {
	reg := metrics.NewRegistry()
	m := Metrics{
		config:   configuration,
		prefix:   prefix,
		Registry: reg,
	}

	return &m, nil
}

// Start starts the metric collection and the exporters.
func (m *Metrics) Start() error {
	// Register runtime metrics
	runtimeRegistry := metrics.NewPrefixedChildRegistry(m.Registry, "go.")
	metrics.RegisterRuntimeMemStats(runtimeRegistry)
	m.t.Go(func() error {
		for {
			timeout := time.After(runtimeMetricsInterval)
			select {
			case <-m.t.Dying():
				return nil
			case <-timeout:
				break
			}
			metrics.CaptureRuntimeMemStatsOnce(runtimeRegistry)
		}
	})

	// Register exporters
	for _, c := range m.config {
		if err := c.initExporter(m); err != nil {
			m.Stop()
			return err
		}
	}
	return nil
}

// MustStart starts the metric collection and panic if there is an
// error.
func (m *Metrics) MustStart() {
	if err := m.Start(); err != nil {
		panic(err)
	}
}

// Push pushes registered metrics to a Prometheus push gateway.
func (m *Metrics) Push() error {
	registry := prometheus.NewRegistry()
	m.Registry.Each(func(name string, metric interface{}) {
		name = strings.Replace(strings.ToLower(name), ".", "_", -1)

		switch metric := metric.(type) {
		case *metrics.StandardGauge:
			g := prometheus.NewGauge(prometheus.GaugeOpts{Name: name})
			g.Set(float64(metric.Value()))
			registry.Register(g)

		case *metrics.StandardGaugeFloat64:
			g := prometheus.NewGauge(prometheus.GaugeOpts{Name: name})
			g.Set(float64(metric.Value()))
			registry.Register(g)

		case *metrics.StandardCounter:
			c := prometheus.NewCounter(prometheus.CounterOpts{Name: name})
			c.Add(float64(metric.Count()))
			registry.Register(c)
		}
	})

	for _, c := range m.config {
		if p, ok := c.(*PromPushGWConfiguration); ok {
			return p.pusher.Gatherer(registry).Push()
		}
	}

	return errors.New("no prompushgw exporter configured")
}

// Stop stops all exporters and wait for them to terminate.
func (m *Metrics) Stop() error {
	m.t.Kill(nil)
	return m.t.Wait()
}
