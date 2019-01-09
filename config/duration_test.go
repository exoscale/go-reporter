package config

import (
	"testing"
	"time"

	"gopkg.in/yaml.v2"
)

var cases = []struct {
	in  string
	out time.Duration
}{
	{
		in:  "324ms",
		out: 324 * time.Millisecond,
	}, {
		in:  "3s",
		out: 3 * time.Second,
	}, {
		in:  "5m",
		out: 5 * time.Minute,
	}, {
		in:  "1h",
		out: time.Hour,
	}, {
		in:  "1h4m",
		out: 1*time.Hour + 4*time.Minute,
	}, {
		in:  "48h18s",
		out: 48*time.Hour + 18*time.Second,
	}, {
		in:  "35ns",
		out: 35 * time.Nanosecond,
	},
}

func TestUnmarshalDuration(t *testing.T) {
	for _, c := range cases {
		var got Duration
		err := yaml.Unmarshal([]byte(c.in), &got)
		if err != nil {
			t.Errorf("Unmarshal(%q) error:\n%+v", c.in, err)
			continue
		}
		if got != Duration(c.out) {
			t.Errorf("Unmarshal(%q) == %v but expected %v", c.in, got, c.out)
		}
	}
}
