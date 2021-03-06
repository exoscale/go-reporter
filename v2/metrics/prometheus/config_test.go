package prometheus

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	testConfig := new(Config)
	require.NoError(t, testConfig.validate())
	require.Equal(t, defaultFlushIntervalSec, testConfig.FlushInterval,
		"should have been set to default value")
}
