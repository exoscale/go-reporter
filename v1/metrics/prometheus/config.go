package prometheus

const (
	defaultFlushInterval = 5
)

// Config represents a prometheus metrics export configuration.
type Config struct {
	// FlushInterval represents the time interval at which the metrics reporter's registry metrics are flushed to the
	// Prometheus registry.
	FlushInterval int `yaml:"flush_interval"`

	// Namespace represents the namespace to apply to registered Prometheus metrics.
	Namespace string `yaml:"namespace"`

	// Subsystem represents the subsystem to apply to registered Prometheus metrics.
	Subsystem string `yaml:"subsystem"`
}

func (c *Config) validate() error {
	if c.FlushInterval <= 0 {
		c.FlushInterval = defaultFlushInterval
	}

	return nil
}
