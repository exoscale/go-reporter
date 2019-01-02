// Package metrics handles metrics for NIDWALD.
//
// This is a wrapper around Code Hale's Metrics library. It also
// setups the supported exporters.
package metrics

import (
	"time"

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

// Stop stops all exporters and wait for them to terminate.
func (m *Metrics) Stop() error {
	m.t.Kill(nil)
	return m.t.Wait()
}
