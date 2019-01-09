package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"

	"github.com/exoscale/go-reporter/config"
)

// ExpvarConfiguration is the configuration for exporting metrics to
// expvar.
type ExpvarConfiguration struct {
	Listen config.Addr
}

// UnmarshalYAML parses a configuration from YAML.
func (c *ExpvarConfiguration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawExpvarConfiguration ExpvarConfiguration
	var raw rawExpvarConfiguration
	if err := unmarshal(&raw); err != nil {
		return errors.Wrap(err, "unable to decode expvar configuration")
	}
	if raw.Listen == "" {
		return errors.Errorf("missing listen value")
	}
	*c = ExpvarConfiguration(raw)
	return nil
}

// initExporter initialize the expvar endpoint.
func (c *ExpvarConfiguration) initExporter(metrics *Metrics) error {
	// Setup the muxer
	mux := http.NewServeMux()
	handler := exp.ExpHandler(metrics.Registry)
	mux.Handle("/", handler)
	mux.HandleFunc("/healthz", healthZ(metrics.Registry))

	// Run the HTTP server
	address := c.Listen
	server := &http.Server{
		Addr:    address.String(),
		Handler: mux,
	}
	listener, err := net.Listen("tcp", address.String())
	if err != nil {
		return errors.Wrapf(err, "unable to listen to %v", address)
	}
	metrics.t.Go(func() error {
		server.Serve(listener)
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

type healthZDetails struct {
	Status  string            `json:"status"`
	Details map[string]string `json:"details"`
}

// healthZ returns a function handler to expose healthchecks in the
// given metrics.
func healthZ(registry metrics.Registry) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		details := healthZDetails{"ok", map[string]string{}}
		registry.Each(func(name string, i interface{}) {
			healthcheck, ok := i.(metrics.Healthcheck)
			if !ok {
				return
			}
			healthcheck.Check()
			if err := healthcheck.Error(); err != nil {
				details.Status = "fail"
				details.Details[name] = fmt.Sprintf("!%s", err.Error())
			} else {
				details.Details[name] = "+ok"
			}
		})
		w.Header().Set("Content-Type", "application/json")
		if output, err := json.Marshal(details); err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "{error: %q}", err)
		} else {
			if details.Status == "fail" {
				w.WriteHeader(542)
			}
			w.Write(output)
		}
	}
}
