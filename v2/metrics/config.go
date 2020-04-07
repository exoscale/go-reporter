package metrics

import (
	"github.com/exoscale/go-reporter/v2/metrics/prometheus"
)

const (
	defaultFlushIntervalSec = 5
)

// Config represents a metrics reporter configuration.
type Config struct {
	// Prometheus represents a Prometheus metrics exporter configuration.
	Prometheus *prometheus.Config `yaml:"prometheus"`

	// FlushInterval represents the time interval in seconds at which to flush metrics to the internal registry.
	FlushInterval int `yaml:"flush_interval"`

	// WithRuntimeMetrics represents a flag indicating whether Go runtime metrics should be included to the registered
	// metrics.
	WithRuntimeMetrics bool `yaml:"runtime_metrics"`
}

func (c *Config) validate() error {
	if c.FlushInterval <= 0 {
		c.FlushInterval = defaultFlushIntervalSec
	}

	return nil
}
