package pushgw

// Configuretation is config for push gateway
// for grouping see https://github.com/prometheus/client_golang/blob/master/prometheus/push/push.go#L152
type Configuration struct {
	URL string
	Job string
	// CertFile   config.FilePath
	// KeyFile    config.FilePath
	// CacertFile config.FilePath
}
