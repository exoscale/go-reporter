package metrics

import (
	"testing"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/exoscale/go-reporter/config"
	"github.com/exoscale/go-reporter/helpers"
)

func TestUnmarshalConfiguration(t *testing.T) {
	cases := []struct {
		in   string
		want interface{}
	}{
		{
			in: `
- expvar:
    listen: 127.0.0.1:7653
`,
			want: ExpvarConfiguration{
				Listen: config.Addr("127.0.0.1:7653"),
			},
		}, {
			in: `
- file:
    interval: 5m
    path: /var/log/project-metrics.log
`,
			want: FileConfiguration{
				Interval: config.Duration(5 * time.Minute),
				Path:     config.FilePath("/var/log/project-metrics.log"),
			},
		}, {
			in: `
- collectd:
    connect: 127.0.0.2:25675
    interval: 5m
`,
			want: CollectdConfiguration{
				Connect:  config.Addr("127.0.0.2:25675"),
				Interval: config.Duration(5 * time.Minute),
			},
		},
		{
			in: `
- prometheus:
    listen: 127.0.0.1:7653
    interval: 10s
    namespace: foo
    subsystem: bar
`,
			want: PrometheusConfiguration{
				Listen:    config.Addr("127.0.0.1:7653"),
				Interval:  config.Duration(10 * time.Second),
				Namespace: "foo",
				Subsystem: "bar",
			},
		},

		{
			in: `
- prometheus:
    listen: 127.0.0.1:7653
    interval: 10s
    namespace: foo
    subsystem: bar
    certfile: /tmp/foo
    keyfile: /tmp/bar
    cacertfile: /tmp/baz
`,
			want: PrometheusConfiguration{
				Listen:     config.Addr("127.0.0.1:7653"),
				Interval:   config.Duration(10 * time.Second),
				Namespace:  "foo",
				Subsystem:  "bar",
				CertFile:   config.FilePath("/tmp/foo"),
				KeyFile:    config.FilePath("/tmp/bar"),
				CacertFile: config.FilePath("/tmp/baz"),
			},
		},
	}
	for _, c := range cases {
		var got Configuration
		err := yaml.Unmarshal([]byte(c.in), &got)
		if err != nil {
			t.Errorf("Unmarshal(%q) error:\n%+v", c.in, err)
			continue
		}
		if diff := helpers.Diff(got, []interface{}{c.want}); diff != "" {
			t.Errorf("Unmarshal(%q) (-got +want):\n%s", c.in, diff)
		}
	}
}

func TestUnmarshalIncompleteConfiguration(t *testing.T) {
	cases := []string{
		`- expvar: {}`,
		`- file: {}`,
		`- file: {interval: 10m}`,
		`- file: {path: /var/log/project...}`,
		`- collectd: {}`,
		`
- prometheus:
    listen: 127.0.0.1:7653
    interval: 10s
    namespace: foo
    subsystem: bar
    keyfile: /tmp/bar
`,
		`
- prometheus:
    listen: 127.0.0.1:7653
    interval: 10s
    namespace: foo
    subsystem: bar
    certfile: /tmp/foo
`,
	}
	for _, c := range cases {
		var got Configuration
		err := yaml.Unmarshal([]byte(c), &got)
		if err == nil {
			t.Errorf("Unmarshal(%q) == %v but expected error", c, got)
		}
	}
}
