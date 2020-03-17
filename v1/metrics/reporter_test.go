package metrics

import (
	"context"
	"testing"
	"time"

	gometrics "github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/require"

	"github.com/exoscale/go-reporter/v1/metrics/prometheus"
	gtesting "github.com/exoscale/go-reporter/v1/testing"
)

/*
	/!\ /!\ /!\ IMPORTANT /!\ /!\ /!\

	The rcrowley/go-metrics package has a limitation related to the runtime
	metrics registration into a registry (see [1] for details), meaning that
	a runtime registry can only be initialized once per package: in the
	following tests, the `Config.WithRuntimeMetrics` flag should be used only
	once across all tests in the package or else tests that rely on runtime
	metrics will fail.

	[1]: https://github.com/rcrowley/go-metrics/issues/221

	/!\ /!\ /!\ IMPORTANT /!\ /!\ /!\
*/

func TestNew(t *testing.T) {
	var (
		testConfig = &Config{
			Prometheus: new(prometheus.Config),
		}
	)

	reporter, err := New(nil)
	require.NoError(t, err)
	require.Nil(t, reporter)

	reporter, err = New(testConfig)
	require.NoError(t, err)
	require.NotNil(t, reporter)
	require.NotNil(t, reporter.log)
	require.NotNil(t, reporter.registry)
	require.NotNil(t, reporter.Prometheus)
}

func TestReporter_Register(t *testing.T) {
	var testMetricName = "test"

	reporter, err := New(&Config{})
	require.NoError(t, err)
	reporter.log.SetHandler(gtesting.NewLogHandler(t))

	metric := gometrics.NewGauge()
	require.NoError(t, reporter.Register(testMetricName, metric))
	require.Equal(t, metric, reporter.registry.Get(testMetricName))
}

func TestReporter_Start(t *testing.T) {
	var testCtx = context.Background()

	reporter, err := New(&Config{
		FlushInterval:      1,
		WithRuntimeMetrics: true,
	})
	require.NoError(t, err)
	reporter.log.SetHandler(gtesting.NewLogHandler(t))

	require.NoError(t, reporter.Start(testCtx))
	require.NotNil(t, reporter.t)

	time.Sleep(time.Duration(reporter.config.FlushInterval) * time.Second * 2)
	require.NotNil(t, reporter.registry.Get("go.runtime.NumGoroutine").(gometrics.Gauge).Value())

	reporter.t.Kill(nil)
	require.Eventually(t,
		func() bool { return reporter.t.Wait() == nil },
		time.Second*3,
		500*time.Millisecond,
		"reporter tomb failed to be killed")
}

func TestReporter_Stop(t *testing.T) {
	var testCtx = context.Background()

	reporter, err := New(&Config{
		FlushInterval:      1,
		WithRuntimeMetrics: true,
	})
	require.NoError(t, err)
	reporter.log.SetHandler(gtesting.NewLogHandler(t))

	require.NoError(t, reporter.Start(testCtx))
	require.NotNil(t, reporter.t)
	require.True(t, reporter.t.Alive())

	require.Eventually(t,
		func() bool { return reporter.Stop(testCtx) == nil },
		time.Second*3,
		500*time.Millisecond,
	)
}
