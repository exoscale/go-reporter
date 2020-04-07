package errors

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	var (
		testConfig = &Config{
			DSN: testSentryDSN,
		}
	)

	reporter, err := New(nil)
	require.NoError(t, err)
	require.Nil(t, reporter)

	_, err = New(&Config{})
	require.Error(t, err)

	reporter, err = New(testConfig)
	require.NoError(t, err)
	require.NotNil(t, reporter.sentry)
}
