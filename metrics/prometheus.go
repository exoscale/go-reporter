package metrics

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/deathowl/go-metrics-prometheus"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/exoscale/go-reporter/config"
)

// PrometheusConfiguration is the configuration for exporting metrics to
// files.
type PrometheusConfiguration struct {
	Listen     config.Addr
	Interval   config.Duration
	Namespace  string
	Subsystem  string
	CertFile   config.FilePath
	KeyFile    config.FilePath
	CacertFile config.FilePath
}

// UnmarshalYAML parses the configuration from YAML.
func (c *PrometheusConfiguration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawPrometheusConfiguration PrometheusConfiguration
	raw := rawPrometheusConfiguration{}
	if err := unmarshal(&raw); err != nil {
		return errors.Wrap(err, "unable to decode prometheus configuration")
	}
	if raw.Listen == "" {
		return errors.Errorf("missing listen value")
	}
	if raw.Namespace == "" {
		return errors.Errorf("missing namespace value")
	}
	if raw.Subsystem == "" {
		return errors.Errorf("missing subsystem value")
	}
	if raw.Interval == config.Duration(0) {
		return errors.Errorf("missing interval value for prometheus configuration")
	}

	if (raw.CertFile != "" || raw.KeyFile != "" || raw.CacertFile != "") &&
		(raw.CertFile == "" || raw.KeyFile == "" || raw.CacertFile == "") {
		return errors.Errorf("certfile, keyfile and cacertfile should be configured")
	}
	*c = PrometheusConfiguration(raw)
	return nil
}

// initExporter initialize the Prometheus exporter
func (c *PrometheusConfiguration) initExporter(metrics *Metrics) error {
	prometheusRegistry := prometheus.DefaultRegisterer
	pClient := prometheusmetrics.NewPrometheusProvider(
		metrics.Registry,
		c.Namespace,
		c.Subsystem,
		prometheusRegistry,
		time.Duration(c.Interval))
	metrics.t.Go(func() error {
		tick := time.NewTicker(time.Duration(c.Interval))
		for {
			select {
			case <-tick.C:
				pClient.UpdatePrometheusMetricsOnce()
			case <-metrics.t.Dying():
				return nil
			}
		}
	})

	address := c.Listen
	listener, err := net.Listen("tcp", address.String())
	if err != nil {
		return errors.Wrapf(err, "unable to listen to %v", address)
	}
	server := &http.Server{
		Addr:    address.String(),
		Handler: promhttp.Handler(),
	}

	metrics.t.Go(func() error {
		if c.CertFile == "" && c.KeyFile == "" && c.CacertFile == "" {
			server.Serve(listener)
		} else {
			cacert, err := ioutil.ReadFile(string(c.CacertFile))
			if err != nil {
				return err
			}
			certpool := x509.NewCertPool()
			certpool.AppendCertsFromPEM(cacert)

			tlsConfig := &tls.Config{
				ClientAuth: tls.RequireAndVerifyClientCert,
				ClientCAs:  certpool,
			}
			server.TLSConfig = tlsConfig
			server.ServeTLS(listener, string(c.CertFile), string(c.KeyFile))
		}
		return nil
	})

	// Handle stop correctly
	metrics.t.Go(func() error {
		<-metrics.t.Dying()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		return nil
	})
	return nil
}
