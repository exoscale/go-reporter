package metrics

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/rcrowley/go-metrics"

	"github.com/exoscale/go-reporter/config"
)

// FileConfiguration is the configuration for exporting metrics to
// files.
type FileConfiguration struct {
	Path     config.FilePath
	Interval config.Duration
}

// UnmarshalYAML parses the configuration from YAML.
func (c *FileConfiguration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawFileConfiguration FileConfiguration
	raw := rawFileConfiguration{}
	if err := unmarshal(&raw); err != nil {
		return errors.Wrap(err, "unable to decode file configuration")
	}
	if raw.Interval == config.Duration(0) {
		return errors.Errorf("missing interval value for file configuration")
	}
	if raw.Path == "" {
		return errors.Errorf("missing path value for file configuration")
	}
	*c = FileConfiguration(raw)
	return nil
}

// Initialization
func (c *FileConfiguration) initExporter(m *Metrics) error {
	output, err := os.OpenFile(string(c.Path), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return errors.Wrapf(err, "unable to open file %q", c.Path)
	}

	// Handle stop correctly
	m.t.Go(func() error {
		tick := time.NewTicker(time.Duration(c.Interval))
		defer tick.Stop()
	L:
		for {
			select {
			case <-tick.C:
				metrics.WriteJSONOnce(m.Registry, output)
				output.Sync()
			case <-m.t.Dying():
				break L
			}
		}
		output.Close()
		return nil
	})

	return nil
}
