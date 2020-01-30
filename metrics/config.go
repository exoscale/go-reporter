package metrics

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Configuration is a list list of exporter configurations. However,
// we have several kind of exporters.
type Configuration []ExporterConfiguration

// ExporterConfiguration is an interface for the configuration for an
// exporter.
type ExporterConfiguration interface {
	UnmarshalYAML(f func(interface{}) error) error
	initExporter(m *Metrics) error
}

// UnmarshalYAML parses configuration for the metrics subsystem from YAML.
func (c *Configuration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var rawConfig []yaml.MapSlice
	if err := unmarshal(&rawConfig); err != nil {
		return errors.Wrap(err, "unable to decode list of metric exporters")
	}
	// We now have a list of interfaces. Each interface should be a MapSlice.
	finalConfiguration := make([]ExporterConfiguration, len(rawConfig), len(rawConfig))
	for i, c := range rawConfig {
		if len(c) != 1 {
			return errors.Errorf("metric configuration item %d should contain only one element, not %d",
				i+1, len(c))
		}
		rawExporterName, rawExporterConfiguration := c[0].Key, c[0].Value
		exporterName, ok := rawExporterName.(string)
		if !ok {
			return errors.Errorf("metric configuration item %d exporter name is incorrect",
				i+1)
		}

		// We convert back the exporter configuration to YAML
		// and will unmarshal depending on the exporter name
		strExporterConfiguration, err := yaml.Marshal(rawExporterConfiguration)
		if err != nil {
			return errors.Wrapf(err, "metric configuration item %d", i+1)
		}
		switch exporterName {
		case "prometheus":
			var exporterConfiguration PrometheusConfiguration
			if err := yaml.Unmarshal(strExporterConfiguration, &exporterConfiguration); err != nil {
				return errors.Wrapf(err, "incorrect prometheus configuration for item %d", i+1)
			}
			finalConfiguration[i] = &exporterConfiguration
		case "expvar":
			var exporterConfiguration ExpvarConfiguration
			if err := yaml.Unmarshal(strExporterConfiguration, &exporterConfiguration); err != nil {
				return errors.Wrapf(err, "incorrect expvar configuration for item %d", i+1)
			}
			finalConfiguration[i] = &exporterConfiguration
		case "file":
			var exporterConfiguration FileConfiguration
			if err := yaml.Unmarshal(strExporterConfiguration, &exporterConfiguration); err != nil {
				return errors.Wrapf(err, "incorrect file configuration for item %d", i+1)
			}
			finalConfiguration[i] = &exporterConfiguration
		case "collectd":
			var exporterConfiguration CollectdConfiguration
			if err := yaml.Unmarshal(strExporterConfiguration, &exporterConfiguration); err != nil {
				return errors.Wrapf(err, "incorrect collectd configuration for item %d", i+1)
			}
			finalConfiguration[i] = &exporterConfiguration
		case "prompushgw":
			var exporterConfiguration PromPushGWConfiguration
			if err := yaml.Unmarshal(strExporterConfiguration, &exporterConfiguration); err != nil {
				return errors.Wrapf(err, "incorrect prompushgw configuration for item %d", i+1)
			}
			finalConfiguration[i] = &exporterConfiguration
		default:
			return errors.Errorf("unknown metric system %q for item %d",
				exporterName, i+1)
		}
	}
	*c = finalConfiguration
	return nil
}
