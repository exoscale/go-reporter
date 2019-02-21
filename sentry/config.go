package sentry

// Configuration is the configuration for Sentry.
type Configuration struct {
	DSN     string
	Tags    map[string]string
	Version string
	Wait    bool
}
