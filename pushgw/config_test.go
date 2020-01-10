package pushgw

import (
	"testing"

	"github.com/exoscale/go-reporter/helpers"
	"gopkg.in/yaml.v2"
)

func TestUnmarshalConfiguration(t *testing.T) {
	cases := []struct {
		in   string
		want Configuration
	}{
		{
			in: `
url: http://pushgateway.net
job: bar
certfile: /tmp/foo
keyfile: /tmp/bar
cacertfile: /tmp/baz
`,
			want: Configuration{
				URL:        "http://pushgateway.net",
				Job:        "bar",
				CertFile:   "/tmp/foo",
				KeyFile:    "/tmp/bar",
				CacertFile: "/tmp/baz",
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
