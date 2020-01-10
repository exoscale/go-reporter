// Package pushgw wrapper around promethesu/push
package pushgw

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/push"
)

// New creates a pushgateway client from configuration.
func New(config Configuration) (*push.Pusher, error) {
	if config.URL == "" || config.Job == "" {
		return nil, nil
	}

	httpClient, err := NewHTTPClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "can't create http client with given config")
	}

	pwClient := push.New(config.URL, config.Job).Client(httpClient)
	return pwClient, nil
}

// NewHTTPClient creates an http client for a given  config
func NewHTTPClient(config Configuration) (*http.Client, error) {
	var transport = &http.Transport{}
	var tlsConfig *tls.Config

	if config.CacertFile != "" {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, errors.Wrap(err, "can't load certificate pair")
		}

		caCert, err := ioutil.ReadFile(config.CertFile)
		if err != nil {
			return nil, errors.Wrap(err, "can't load certificate")
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		}
		tlsConfig.BuildNameToCertificate()
		transport = &http.Transport{TLSClientConfig: tlsConfig}
	}

	return &http.Client{Transport: transport}, nil
}
