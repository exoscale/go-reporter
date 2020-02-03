package metrics

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"

	"github.com/exoscale/go-reporter/config"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/push"
)

type PromPushGWConfiguration struct {
	URL        string
	Job        string
	CertFile   config.FilePath
	KeyFile    config.FilePath
	CacertFile config.FilePath

	pusher *push.Pusher
}

func (c *PromPushGWConfiguration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawPromPushGWConfiguration PromPushGWConfiguration
	raw := rawPromPushGWConfiguration{}
	if err := unmarshal(&raw); err != nil {
		return errors.Wrap(err, "unable to decode prompushgw configuration")
	}
	if raw.URL == "" {
		return errors.Errorf("missing URL value")
	}
	if raw.Job == "" {
		return errors.Errorf("missing job value")
	}
	if (raw.CertFile != "" || raw.KeyFile != "" || raw.CacertFile != "") &&
		(raw.CertFile == "" || raw.KeyFile == "" || raw.CacertFile == "") {
		return errors.Errorf("certfile, keyfile and cacertfile should be configured")
	}
	*c = PromPushGWConfiguration(raw)
	return nil
}

func (c *PromPushGWConfiguration) initExporter(m *Metrics) error {
	httpClient, err := NewHTTPClient(*c)
	if err != nil {
		return errors.Wrap(err, "can't create http client with given prompushgw config")
	}

	c.pusher = push.New(c.URL, c.Job).Client(httpClient)
	return nil
}

// NewHTTPClient creates an HTTP client for a given config.
func NewHTTPClient(config PromPushGWConfiguration) (*http.Client, error) {
	var transport = &http.Transport{}
	var tlsConfig *tls.Config

	if config.CacertFile != "" {
		cert, err := tls.LoadX509KeyPair(string(config.CertFile), string(config.KeyFile))
		if err != nil {
			return nil, errors.Wrap(err, "can't load certificate pair")
		}

		caCert, err := ioutil.ReadFile(string(config.CertFile))
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
