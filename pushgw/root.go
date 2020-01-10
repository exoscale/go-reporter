// Package pushgw wrapper around promethesu/push
package pushgw

import (
	"github.com/prometheus/client_golang/prometheus/push"
)

// New creates a pushgateway client from configuration.
// If URL or Job is not provided, return nil,error
//TODO: this should return error as well and handle tls
func New(config Configuration) *push.Pusher {
	if config.URL == "" {
		return nil
	}
	client := push.New(config.URL, config.Job)
	return client

}
