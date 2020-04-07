package errors

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	var config *Config

	config = &Config{}
	require.Error(t, config.validate())

	config = &Config{DSN: testSentryDSN}
	require.NoError(t, config.validate())
}
