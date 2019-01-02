package sentry

import (
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/exoscale/go-reporter/helpers"
)

func TestUnmarshalConfiguration(t *testing.T) {
	cases := []struct {
		in   string
		want Configuration
	}{
		{"dsn: http://public:secret@sentry.errors",
			Configuration{
				DSN:  "http://public:secret@sentry.errors",
				Tags: nil,
			},
		},
		{"{}",
			Configuration{
				DSN:  "",
				Tags: nil,
			},
		},
		{`
dsn: http://public:secret@sentry.errors
tags:
  environment: prod
  dc: us-east-4
version: "foo"
`,
			Configuration{
				DSN: "http://public:secret@sentry.errors",
				Tags: map[string]string{
					"environment": "prod",
					"dc":          "us-east-4",
				},
				Version: "foo",
			},
		},
	}
	for _, c := range cases {
		var got Configuration
		err := yaml.Unmarshal([]byte(c.in), &got)
		if err != nil {
			t.Errorf("Unmarshal(%q) error:\n%+v", c.in, err)
		}
		if diff := helpers.Diff(got, c.want); diff != "" {
			t.Errorf("Unmarshal(%q) (-got +want):\n%s", c.in, diff)
		}
	}
}
