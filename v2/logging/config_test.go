package logging

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogDestinationConfig_Validate(t *testing.T) {
	var config *LogDestinationConfig

	config = &LogDestinationConfig{}
	require.Error(t, config.validate())

	config = &LogDestinationConfig{Type: "lolnope"}
	require.Error(t, config.validate())

	config = &LogDestinationConfig{
		Type:  "file",
		Level: "lolnope",
	}
	require.Error(t, config.validate())

	config = &LogDestinationConfig{
		Type:   "file",
		Level:  "debug",
		Format: "lolnope",
	}
	require.Error(t, config.validate())

	config = &LogDestinationConfig{Type: "file"}
	require.Error(t, config.validate())

	config = &LogDestinationConfig{Type: "file", Destination: "/tmp/test.log"}
	require.NoError(t, config.validate())
	require.Equal(t, defaultLogLevel, config.Level, "should have been set to default value")
	require.Equal(t, defaultLogFormat, config.Format, "should have been set to default value")

	config = &LogDestinationConfig{Type: "syslog", Destination: "lolnope"}
	require.Error(t, config.validate())

	config = &LogDestinationConfig{Type: "syslog", Destination: "server:1514"}
	require.NoError(t, config.validate())
}

func TestLogDestinationConfig_LogFormat(t *testing.T) {
	// Cannot be tested because the logFormat() method returns a function, and testify
	// cannot compare functions equality: https://godoc.org/github.com/stretchr/testify/require#Equal
	t.Skip()
}

func TestConfig_Validate(t *testing.T) {
	var config *Config

	config = &Config{}
	require.NoError(t, config.validate())

	config = &Config{Destinations: []*LogDestinationConfig{
		{Type: "file"},
	}}
	require.Error(t, config.validate())

	config = &Config{Destinations: []*LogDestinationConfig{
		{Type: "file", Destination: "/tmp/test.log"},
	}}
	require.NoError(t, config.validate())
}
