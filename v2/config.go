package v2

import (
	"github.com/exoscale/go-reporter/v2/errors"
	"github.com/exoscale/go-reporter/v2/logging"
	"github.com/exoscale/go-reporter/v2/metrics"
)

type Config struct {
	// Errors represents the errors reporter configuration.
	Errors *errors.Config `yaml:"errors"`

	// Logging represents the logging reporter configuration.
	Logging *logging.Config `yaml:"logging"`

	// Metrics represents the metrics reporter configuration.
	Metrics *metrics.Config `yaml:"metrics"`

	// Debug represents a flags indicating whether to enable internal reporter activity logging.
	// This is mainly for debug purposes.
	Debug bool `yaml:"debug"`
}
