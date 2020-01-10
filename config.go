package reporter

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/exoscale/go-reporter/logger"
	"github.com/exoscale/go-reporter/metrics"
	"github.com/exoscale/go-reporter/pushgw"
	"github.com/exoscale/go-reporter/sentry"
)

// Configuration contains the reporter configuration.
type Configuration struct {
	Logging logger.Configuration
	Sentry  sentry.Configuration
	Metrics metrics.Configuration
	Pushgw  pushgw.Configuration
	Prefix  string
}

// DefaultConfiguration is the default reporter configuration.
var DefaultConfiguration = Configuration{
	Logging: logger.DefaultConfiguration,
}

// UnmarshalYAML parses a reporter configuration from YAML.
func (configuration *Configuration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// If a section is entirely absent, it won't get its default
	// value. We work-around that by unmarshalling the empty
	// sections to use its default.
	type rawConfiguration Configuration
	var raw rawConfiguration
	err := yaml.Unmarshal([]byte(`
logging: {}
sentry: {}
`),
		&raw)
	if err != nil {
		return errors.Wrap(err, "unable to decode default reporter configuration")
	}
	if err := unmarshal(&raw); err != nil {
		return errors.Wrap(err, "unable to decode reporter configuration")
	}
	*configuration = Configuration(raw)
	return nil
}
