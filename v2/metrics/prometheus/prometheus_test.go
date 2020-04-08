package prometheus

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	gometrics "github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	var (
		testConfig = new(Config)
		metrics    = gometrics.NewRegistry()
	)

	exporter, err := New(testConfig, metrics)
	require.NoError(t, err)
	require.NotNil(t, exporter)
	require.NotNil(t, exporter.registry)
	require.NotNil(t, exporter.pm)
	require.Equal(t, testConfig, exporter.config)
}

func TestExporter_Register(t *testing.T) {
	var (
		testConfig     = new(Config)
		testMetricName = "test"
		testMetricHelp = "Test Prometheus metric"
		testMetric     = prom.NewGauge(prom.GaugeOpts{
			Name: testMetricName,
			Help: testMetricHelp,
		})
	)

	exporter, err := New(testConfig, nil)
	require.NoError(t, err)

	require.NoError(t, exporter.Register(testMetric))

	registeredMetrics, err := exporter.registry.Gather()
	require.NoError(t, err)
	require.Len(t, registeredMetrics, 1)
	require.Equal(t, testMetricName, registeredMetrics[0].GetName())
	require.Equal(t, testMetricHelp, registeredMetrics[0].GetHelp())
}

func TestExporter_HTTPHandler(t *testing.T) {
	var (
		testConfig      = new(Config)
		testMetricName  = "test"
		testMetricHelp  = "Test Prometheus metric"
		testMetricValue = 42.0
		testMetric      = prom.NewGauge(prom.GaugeOpts{
			Name: testMetricName,
			Help: testMetricHelp,
		})
	)

	exporter, err := New(testConfig, nil)
	require.NoError(t, err)

	require.NoError(t, exporter.Register(testMetric))

	testMetric.Set(testMetricValue)
	ts := httptest.NewServer(exporter.HTTPHandler())
	defer ts.Close()

	res, err := ts.Client().Get(ts.URL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf(`# HELP %[1]s %[2]s
# TYPE %[1]s gauge
%[1]s %[3]d
`,
		testMetricName,
		testMetricHelp,
		int64(testMetricValue),
	), string(body))
}

func TestExporter_Start(t *testing.T) {
	var (
		testCtx        = context.Background()
		testConfig     = &Config{FlushInterval: 1}
		testMetricName = "test"
	)

	registry := gometrics.NewRegistry()
	metric := gometrics.NewGauge()
	require.NoError(t, registry.Register(testMetricName, metric))

	exporter, err := New(testConfig, registry)
	require.NoError(t, err)

	require.NoError(t, exporter.Start(testCtx))
	require.NotNil(t, exporter.t)
	require.True(t, exporter.t.Alive())

	time.Sleep(time.Duration(testConfig.FlushInterval) * time.Second * 2)
	registeredMetrics, err := exporter.registry.Gather()
	require.NoError(t, err)
	require.Len(t, registeredMetrics, 1)
	require.Equal(t, testMetricName, registeredMetrics[0].GetName())

	exporter.t.Kill(nil)
	require.Eventually(t,
		func() bool { return exporter.t.Wait() == nil },
		time.Second*3,
		500*time.Millisecond,
		"exporter tomb failed to be killed")
}

func TestExporter_Stop(t *testing.T) {
	var (
		testCtx    = context.Background()
		testConfig = new(Config)
		metrics    = gometrics.NewRegistry()
	)

	exporter, err := New(testConfig, metrics)
	require.NoError(t, err)

	require.NoError(t, exporter.Start(testCtx))
	require.NotNil(t, exporter.t)
	require.True(t, exporter.t.Alive())

	require.Eventually(t,
		func() bool { return exporter.Stop(testCtx) == nil },
		time.Second*3,
		500*time.Millisecond,
	)
}
