package metrics

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/rcrowley/go-metrics"

	"github.com/exoscale/go-reporter/config"
	"github.com/exoscale/go-reporter/helpers"
)

func TestExpvar(t *testing.T) {
	tcpPort := rand.Intn(1000) + 22000
	var configuration Configuration = make([]ExporterConfiguration, 1, 1)
	configuration[0] = &ExpvarConfiguration{
		Listen: config.Addr(fmt.Sprintf("127.0.0.1:%d", tcpPort)),
	}

	m, err := New(configuration, "project")
	if err != nil {
		t.Fatalf("New(%v) error:\n%+v", configuration, err)
	}
	m.MustStart()
	defer func() {
		m.Stop()
		if !testing.Short() {
			time.Sleep(1 * time.Second) // Slight race condition...
			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", tcpPort))
			if err == nil {
				t.Errorf("Still able to connect to expvar server after stop")
				resp.Body.Close()
			}
		}
	}()

	t.Run("root", func(t *testing.T) {
		c := metrics.NewCounter()
		m.Registry.Register("foo", c)
		c.Inc(47)

		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/", tcpPort))
		if err != nil {
			t.Fatalf("Unable to query expvar server:\n%+v", err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Unable to read body from expvar server:\n%+v", err)
		}
		var got struct {
			Foo int
		}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("Unable to decode JSON body:\n%s\nError:\n%+v", body, err)
		}
		if got.Foo != 47 {
			t.Fatalf("Expected Foo == 47 but got %d instead", got.Foo)
		}
	})

	t.Run("healthz", func(t *testing.T) {
		slurp := func(code int) healthZDetails {
			t.Helper()
			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/healthz", tcpPort))
			if err != nil {
				t.Fatalf("Unable to query healthz endpoint:\n%+v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != code {
				t.Fatalf("Expected status code %d, got %d", code, resp.StatusCode)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Unable to read body from expvar server:\n%+v", err)
			}
			var got healthZDetails
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatalf("Unable to decode JSON body:\n%s\nError:\n%+v", body, err)
			}
			return got
		}

		// No metrics
		got := slurp(200)
		expected := healthZDetails{"ok", map[string]string{}}
		if diff := helpers.Diff(got, expected); diff != "" {
			t.Fatalf("/healthz without healthchecks (-got +want):\n%s", diff)
		}

		// 1 healthy healthcheck
		h := metrics.NewHealthcheck(func(h metrics.Healthcheck) {
			h.Healthy()
		})
		m.Registry.Register("bar", h)
		got = slurp(200)
		expected = healthZDetails{"ok", map[string]string{
			"bar": "+ok",
		}}
		if diff := helpers.Diff(got, expected); diff != "" {
			t.Fatalf("/healthz with 1 ok healthcheck (-got +want):\n%s", diff)
		}

		// 1 unhealthy healthcheck
		h = metrics.NewHealthcheck(func(h metrics.Healthcheck) {
			h.Unhealthy(errors.New("nope"))
		})
		m.Registry.Register("zip", h)
		got = slurp(542)
		expected = healthZDetails{"fail", map[string]string{
			"bar": "+ok",
			"zip": "!nope",
		}}
		if diff := helpers.Diff(got, expected); diff != "" {
			t.Fatalf("/healthz with 1 failing healthcheck (-got +want):\n%s", diff)
		}

	})
}

func TestHealthz(t *testing.T) {
}
