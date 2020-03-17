package v1

import (
	"github.com/exoscale/go-reporter/v1/errors"
	"github.com/exoscale/go-reporter/v1/logging"
	"github.com/exoscale/go-reporter/v1/metrics"
)

type Config struct {
	Metrics *metrics.Config `yaml:"metrics"`
	Logging *logging.Config `yaml:"logging"`
	Errors  *errors.Config  `yaml:"errors"`

	Debug bool `yaml:"debug"`
}
