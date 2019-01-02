package reporter

import (
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/exoscale/go-reporter/config"
	"github.com/exoscale/go-reporter/helpers"
	"github.com/exoscale/go-reporter/logger"
	"github.com/exoscale/go-reporter/metrics"
	"github.com/exoscale/go-reporter/sentry"
)

func TestConfiguration(t *testing.T) {
	cases := []struct {
		in   string
		want Configuration
	}{
		{
			in: `
prefix: aargau
logging:
  console: true
  syslog: false
  level: debug
metrics:
  - expvar:
      listen: :8123
sentry:
  dsn: "http://public:secret@errors"
`,
			want: Configuration{
				Logging: logger.Configuration{
					Console: true,
					Syslog:  false,
					Level:   4,
				},
				Metrics: metrics.Configuration([]metrics.ExporterConfiguration{
					&metrics.ExpvarConfiguration{
						Listen: config.Addr(":8123"),
					},
				}),
				Sentry: sentry.Configuration{
					DSN: "http://public:secret@errors",
				},
				Prefix: "aargau",
			},
		},
	}

	for _, tc := range cases {
		var got Configuration
		err := yaml.Unmarshal([]byte(tc.in), &got)
		if err != nil {
			t.Errorf("Unmarshal(%q) error:\n%+v", tc.in, err)
		}
		if diff := helpers.Diff(got, tc.want); diff != "" {
			t.Errorf("Unmarshal(%q) (-got +want):\n%s", tc.in, diff)
		}
	}
}
