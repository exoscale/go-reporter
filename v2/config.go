package v2

import (
	"github.com/exoscale/go-reporter/v2/errors"
	"github.com/exoscale/go-reporter/v2/logging"
	"github.com/exoscale/go-reporter/v2/metrics"
)

type Config struct {
	Metrics *metrics.Config `yaml:"metrics"`
	Logging *logging.Config `yaml:"logging"`
	Errors  *errors.Config  `yaml:"errors"`

	Debug bool `yaml:"debug"`
}
