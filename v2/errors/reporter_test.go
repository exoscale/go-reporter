package errors

import (
	"errors"
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

func TestReporter_SendError(t *testing.T) {
	var (
		testErrorMessage = errors.New("oh noes!")
		testTags         = map[string]string{
			"k1": "v1",
			"k2": "v2",
		}
		sentryTestTransport = new(SentryTestTransport)
	)

	testReporter, err := New(&Config{DSN: testSentryDSN})
	require.NoError(t, err)

	testReporter.SetSentryTransport(sentryTestTransport)

	testReporter.SendError(testErrorMessage, testTags)
	require.Len(t, sentryTestTransport.Events(), 1)
	require.Equal(t, testErrorMessage.Error(), sentryTestTransport.Events()[0].Exception[0].Value)
	require.Equal(t, testTags, sentryTestTransport.Events()[0].Tags)
}
