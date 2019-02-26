package metrics

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"

	"collectd.org/api"
	"collectd.org/network"
	"github.com/rcrowley/go-metrics"

	"github.com/exoscale/go-reporter/config"
	"github.com/exoscale/go-reporter/helpers"
)

func TestCollectd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip collectd exporter test in short mode")
	}
	port := rand.Intn(1000) + 23000
	address := config.Addr(fmt.Sprintf("127.0.0.1:%d", port))

	// Create a collectd server
	udpAddress, err := net.ResolveUDPAddr("udp", address.String())
	if err != nil {
		t.Fatalf("Unable to resolve %v:\n%+v", address, err)
	}
	sock, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		t.Fatalf("Unable to listen to %v:\n%+v", udpAddress, err)
	}
	defer sock.Close()

	// Initialize collectd exporter
	var configuration Configuration = make([]ExporterConfiguration, 1, 1)
	configuration[0] = &CollectdConfiguration{
		Connect:  address,
		Interval: config.Duration(500 * time.Millisecond),
		Exclude:  []string{"metrics.test.donot*"},
	}
	m, err := New(configuration, "project")
	if err != nil {
		t.Fatalf("New(%v) error:\n%+v", configuration, err)
	}
	m.MustStart()

	// Create a metric of each type
	metrics.NewRegisteredCounter("foo", m.Registry).Inc(47)
	metrics.NewRegisteredCounter("metrics.test.counter", m.Registry).Inc(41)
	metrics.NewRegisteredGauge("metrics.test.gauge", m.Registry).Update(18)
	metrics.NewRegisteredGaugeFloat64("metrics.test.gauge64", m.Registry).Update(787.14)
	metrics.NewRegisteredHistogram("metrics.test.histogram", m.Registry,
		metrics.NewUniformSample(10)).Update(1871)
	metrics.NewRegisteredMeter("metrics.test.meter", m.Registry).Mark(19)

	// Create a metric that we'll exclude from reporting
	metrics.NewRegisteredGauge("metrics.test.donotwant", m.Registry).Update(13)

	// For timer, we would like to test percentiles too. 50 and 75
	// percentile will be 16, but others will be 18.
	tt := metrics.NewRegisteredTimer("metrics.test.timer", m.Registry)
	tt.Update(18 * time.Second)
	for i := 0; i < 9; i++ {
		tt.Update(16 * time.Second)
	}

	// Read all received values
	deadline := time.After(750 * time.Millisecond)
	deadline2 := time.After(1000 * time.Millisecond)
	socketData := make(chan []byte)
	socketErr := make(chan error)
	go func() {
		for {
			buf := make([]byte, network.DefaultBufferSize)
			n, err := sock.Read(buf)
			if err != nil {
				socketErr <- err
				return
			}
			socketData <- buf[:n]
		}
	}()
	receivedValues := make([][]*api.ValueList, 0, 100)
L:
	for i := 0; true; i++ {
		select {
		case <-deadline:
			go m.Stop()
		case <-deadline2:
			break L
		case err := <-socketErr:
			t.Fatalf("Received an unexpected error (after %d iterations) while reading:\n%+v",
				i, err)
		case buf := <-socketData:
			valueLists, err := network.Parse(buf, network.ParseOpts{})
			if err != nil {
				t.Fatalf("Received an unexpected error (after %d iterations) while parsing %v:\n%+v",
					i, buf, err)
			}
			receivedValues = append(receivedValues, valueLists)
		}
	}

	// Wait a bit to check we don't have anything left
	deadline = time.After(750 * time.Millisecond)
	select {
	case <-deadline:
		break
	case err := <-socketErr:
		t.Fatalf("Received an unexpected error while reading:\n%+v", err)
	case buf := <-socketData:
		t.Errorf("Received additional late data (%d bytes)", len(buf))
	}

	// Flatten received values
	if len(receivedValues) == 0 {
		t.Fatal("Nothing was received")
	}
	values := receivedValues[0]
	for _, v := range receivedValues[1:] {
		values = append(values, v...)
	}

	// Check them
	cases := []struct {
		plugin         string
		pluginInstance string
		kind           string
		compare        int
		values         []api.Value
	}{
		{
			plugin: "project.foo",
			kind:   "counter",
			values: []api.Value{api.Counter(47)},
		}, {
			plugin:         "project.metrics.test",
			pluginInstance: "counter",
			kind:           "counter",
			values:         []api.Value{api.Counter(41)},
		}, {
			plugin:         "project.metrics.test",
			pluginInstance: "gauge",
			kind:           "gauge",
			values:         []api.Value{api.Gauge(18)},
		}, {
			plugin:         "project.metrics.test",
			pluginInstance: "gauge64",
			kind:           "gauge",
			values:         []api.Value{api.Gauge(787.14)},
		}, {
			plugin:         "project.metrics.test",
			pluginInstance: "histogram",
			kind:           "histogram",
			values: []api.Value{
				api.Counter(1),
				api.Gauge(1871),
				api.Gauge(1871),
				api.Gauge(1871),
				api.Gauge(0),
				api.Gauge(1871),
				api.Gauge(1871),
				api.Gauge(1871),
				api.Gauge(1871),
				api.Gauge(1871),
				api.Gauge(1871),
			},
		}, {
			plugin:         "project.metrics.test",
			pluginInstance: "meter",
			kind:           "meter",
			compare:        4,
			values: []api.Value{
				api.Counter(19),
				api.Gauge(0),
				api.Gauge(0),
				api.Gauge(0),
			},
		}, {
			plugin:         "project.metrics.test",
			pluginInstance: "timer",
			kind:           "timer",
			values: []api.Value{
				api.Gauge(1.8e+10),
				api.Gauge(1.62e+10),
				api.Gauge(1.6e+10),
				api.Gauge(6e+8),
				api.Gauge(1.6e+10),
				api.Gauge(1.6e+10),
				api.Gauge(1.8e+10),
				api.Gauge(1.8e+10),
				api.Gauge(1.8e+10),
				api.Gauge(1.8e+10),
			},
		},
	}

	for _, c := range cases {
		found := false
		var v *api.ValueList
		for _, v = range values {
			if v.Plugin == c.plugin && v.PluginInstance == c.pluginInstance {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Unable to find %s/%s among %d received metrics",
				c.plugin, c.pluginInstance, len(values))
			continue
		}

		hostname, _ := os.Hostname()
		if v.Host != hostname {
			t.Errorf("Received metric %+v has hostname == %v, expected %v",
				v, v.Host, hostname)
		}
		if v.Type != c.kind {
			t.Errorf("Received metric %+v has type == %v, expected %v",
				v, v.Type, c.kind)
		}
		if v.Interval != 500*time.Millisecond {
			t.Errorf("Received metric %+v has interval == %v, expected %v",
				v, v.Interval, 500*time.Millisecond)
		}
		if c.compare == 0 {
			c.compare = len(c.values)
		}
		if diff := helpers.Diff(v.Values[:c.compare], c.values[:c.compare]); diff != "" {
			t.Errorf("Received metric %+v (-got +want):\n%s", v, diff)
		}
	}

	for _, v := range values {
		if v.PluginInstance == "donotwant" {
			t.Error("Excluded metric found in received metrics")
		}
	}
}
