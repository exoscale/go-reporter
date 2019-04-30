package metrics

import (
	"context"
	"path"
	"strings"
	"time"

	"collectd.org/api"
	"collectd.org/network"
	"github.com/pkg/errors"
	"github.com/rcrowley/go-metrics"

        "github.com/exoscale/go-reporter/config"
)

var separator = "."

// CollectdConfiguration represents the configuration for exporting
// metrics to collectd
type CollectdConfiguration struct {
	Connect  config.Addr
	Interval config.Duration
	Exclude  []string
}

// UnmarshalYAML parses a configuration for collectd from YAML.
func (c *CollectdConfiguration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawCollectdConfiguration CollectdConfiguration
	raw := rawCollectdConfiguration{
		Connect: config.Addr("127.0.0.1:25826"),
	}
	if err := unmarshal(&raw); err != nil {
		return errors.Wrap(err, "unable to decode collectd configuration")
	}
	if raw.Interval == config.Duration(0) {
		return errors.Errorf("missing interval value for collectd configuration")
	}
	*c = CollectdConfiguration(raw)
	return nil
}

// initExporter initializes collectd reporter
func (c *CollectdConfiguration) initExporter(m *Metrics) error {
	// Client is using UDP. No need to reconnect.
	address := c.Connect.String()
	client, err := network.Dial(address, network.ClientOptions{})
	if err != nil {
		return errors.Wrapf(err, "unable to connect to collectd (%v)", address)
	}

	m.t.Go(func() error {
		tick := time.NewTicker(time.Duration(c.Interval))
		defer tick.Stop()
	L:
		for {
			select {
			case <-tick.C:
				collectdReportOnce(m.Registry, client, time.Duration(c.Interval), m.prefix, c.Exclude)
			case <-m.t.Dying():
				break L
			}
		}
		client.Flush()
		client.Close()
		return nil
	})

	return nil
}

// collectdReportOnce will export the current metrics to collectd
func collectdReportOnce(r metrics.Registry, client *network.Client, interval time.Duration,
	prefix string, excluded []string) {
	ctx := context.Background()
        hostname, err := config.GetFQDN()
	if err != nil {
		return
	}
	now := time.Now()
	vls := make([]*api.ValueList, 0, 10)
	r.Each(func(name string, i interface{}) {
		// Filter metrics matching any configured pattern
		// Any error parsing an exclusion pattern is silently ignored.
		for _, pattern := range excluded {
			if matched, _ := path.Match(pattern, name); matched {
				return
			}
		}

		var identifierType string
		var values []api.Value
		switch metric := i.(type) {
		case metrics.Gauge:
			identifierType = "gauge"
			values = []api.Value{api.Gauge(metric.Value())}
		case metrics.GaugeFloat64:
			identifierType = "gauge"
			values = []api.Value{api.Gauge(metric.Value())}
		case metrics.Counter:
			identifierType = "counter"
			values = []api.Value{api.Counter(metric.Count())}
		case metrics.Meter:
			identifierType = "meter"
			values = []api.Value{
				api.Counter(metric.Count()),
				api.Gauge(metric.Rate1()),
				api.Gauge(metric.Rate5()),
				api.Gauge(metric.Rate15()),
				api.Gauge(metric.RateMean()),
			}
		case metrics.Histogram:
			identifierType = "histogram"
			ps := metric.Percentiles([]float64{0.5, 0.75, 0.95, 0.98, 0.99, 0.999})
			values = []api.Value{
				api.Counter(metric.Count()),
				api.Gauge(metric.Max()),
				api.Gauge(metric.Mean()),
				api.Gauge(metric.Min()),
				api.Gauge(metric.StdDev()),
				api.Gauge(ps[0]),
				api.Gauge(ps[1]),
				api.Gauge(ps[2]),
				api.Gauge(ps[3]),
				api.Gauge(ps[4]),
				api.Gauge(ps[5]),
			}
		case metrics.Timer:
			identifierType = "timer"
			ps := metric.Percentiles([]float64{0.5, 0.75, 0.95, 0.98, 0.99, 0.999})
			values = []api.Value{
				api.Gauge(metric.Max()),
				api.Gauge(metric.Mean()),
				api.Gauge(metric.Min()),
				api.Gauge(metric.StdDev()),
				api.Gauge(ps[0]),
				api.Gauge(ps[1]),
				api.Gauge(ps[2]),
				api.Gauge(ps[3]),
				api.Gauge(ps[4]),
				api.Gauge(ps[5]),
			}
		case metrics.Healthcheck:
			// Don't do anything with them.
		default:
			metrics.GetOrRegisterMeter(
				"github.com/exoscale/go-reporter.metrics.collectd.unknown-metrics",
				r).Mark(1)
			return
		}
		plugin, pluginInstance := collectdGetPluginName(name)
		plugin = strings.Join([]string{prefix, plugin}, ".")
		identifier := api.Identifier{
			Host:           hostname,
			Plugin:         plugin,
			PluginInstance: pluginInstance,
			Type:           identifierType,
		}
		vl := api.ValueList{
			Identifier: identifier,
			Time:       now,
			Interval:   interval,
			Values:     values,
		}
		vls = append(vls, &vl)
	})
	sent := 0
	failed := 0
	for _, vl := range vls {
		if err := client.Write(ctx, vl); err != nil {
			failed++
		} else {
			sent++
		}
	}
	if sent > 0 {
		metrics.GetOrRegisterMeter(
			"github.com/exoscale/go-reporter.metrics.collectd.sent-metrics",
			r).Mark(int64(sent))
	}
	if failed > 0 {
		metrics.GetOrRegisterMeter(
			"github.com/exoscale/go-reporter.metrics.collectd.failed-writes",
			r).Mark(int64(failed))
	}
}

// Compute plugin and plugin instance from metric name.
func collectdGetPluginName(metricName string) (plugin string, pluginInstance string) {
	index := strings.LastIndex(metricName, separator)
	if index == -1 {
		plugin = metricName
		pluginInstance = ""
	} else {
		plugin = metricName[:index]
		pluginInstance = metricName[index+1:]
	}
	return
}
