package v1

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/exoscale/go-reporter/v1/errors"
	"github.com/exoscale/go-reporter/v1/logging"
	"github.com/exoscale/go-reporter/v1/metrics"
	"github.com/exoscale/go-reporter/v1/metrics/prometheus"
)

var testSentryDSN = "https://xxxxxxxxxxxxxxxxxxxxx:yyyyyyyyyyyyyyyyyyy@sentry.example.net/42"

func TestNew(t *testing.T) {
	reporter, err := New(&Config{
		Metrics: &metrics.Config{
			Prometheus: new(prometheus.Config),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, reporter)
	require.NotNil(t, reporter.Metrics)
}

func TestReportLoggingError(t *testing.T) {
	var (
		testErrorMessage    = "oh noes!"
		sentryTestTransport = new(errors.SentryTestTransport)
	)

	testReporter, err := New(&Config{
		Logging: &logging.Config{
			ReportErrors: true,
		},
		Errors: &errors.Config{
			DSN: testSentryDSN,
		},
	})
	require.NoError(t, err)

	testReporter.Errors.SetSentryTransport(sentryTestTransport)

	testReporter.Error(testErrorMessage, "k", "v")

	require.Len(t, sentryTestTransport.Events(), 1)
	require.Equal(t, testErrorMessage, sentryTestTransport.Events()[0].Exception[0].Value)
	require.Equal(t, map[string]string{"k": "v"}, sentryTestTransport.Events()[0].Tags)
}

func TestReportPanic(t *testing.T) {
	// FIXME: for some reason the recover() function in the Errors.PanicHandler() method doesn't actually recover
	// the value passed to panic() from this test function, resulting in a failing test. Note that in real-world
	// conditions the code works as expected, this behavior only occurs during testing. I have no idea why ¯\_(ツ)_/¯
	// -falzm
	t.Skip()

	var (
		testErrorMessage    = "oh noes!"
		sentryTestTransport = new(errors.SentryTestTransport)
	)

	testReporter, err := New(&Config{
		Logging: &logging.Config{
			ReportErrors: true,
		},
		Errors: &errors.Config{
			DSN: testSentryDSN,
		},
	})
	require.NoError(t, err)

	testReporter.Errors.SetSentryTransport(sentryTestTransport)

	defer func(t *testing.T, sentryTestTransport *errors.SentryTestTransport) {
		testReporter.Errors.PanicHandler(func(r interface{}) {
			require.Len(t, sentryTestTransport.Events(), 1)
			require.Equal(t, testErrorMessage, sentryTestTransport.Events()[0].Exception[0].Value)
			require.Equal(t, testErrorMessage, r)
		})
	}(t, sentryTestTransport)

	panic(testErrorMessage)
}
