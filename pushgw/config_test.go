package pushgw

import (
	"github.com/exoscale/go-reporter/helpers"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestUnmarshalConfiguration(t *testing.T) {
	cases := []struct {
		in   string
		want Configuration
	}{
		{
			in: `
url: http://pushgateway.exoscale.net
job: bar
`,
			want: Configuration{
				URL: "http://pushgateway.exoscale.net",
				Job: "bar",
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
