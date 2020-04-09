package prometheus

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

const (
	defaultFlushIntervalSec = 5
)

// Config represents a prometheus metrics export configuration.
type Config struct {
	// Listen represents a net.Dial compatible string indicating the network address to bind the Prometheus metrics
	// scraping endpoint server. If not specified, the server won't be started.
	Listen string `yaml:"listen"`

	// FlushInterval represents the time interval in seconds at which the metrics reporter's registry metrics are
	// flushed to the Prometheus registry.
	FlushInterval int `yaml:"flush_interval"`

	// Namespace represents the namespace to apply to registered Prometheus metrics.
	Namespace string `yaml:"namespace"`

	// Subsystem represents the subsystem to apply to registered Prometheus metrics.
	Subsystem string `yaml:"subsystem"`

	// Debug represents a flags indicating whether to enable internal exporter activity logging.
	// This is mainly for debug purposes.
	Debug bool `yaml:"debug"`
}

func (c *Config) validate() error {
	if c.FlushInterval <= 0 {
		c.FlushInterval = defaultFlushIntervalSec
	}

	return validation.ValidateStruct(c,
		validation.Field(&c.Listen,
			validation.When(c.Listen != "", is.DialString)),
	)
}
