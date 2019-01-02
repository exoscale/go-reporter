package helpers

import (
	"net"
	"testing"
)

func TestIncrementIPv4(t *testing.T) {
	cases := []struct {
		in        string
		increment int
		want      string
	}{
		{"10.0.0.0", 10, "10.0.0.10"},
		{"10.0.0.240", -10, "10.0.0.230"},
		{"10.0.0.240", 20, "10.0.1.4"},
		{"10.0.0.240", -300, "9.255.255.196"},
		{"0.0.0.5", -6, "255.255.255.255"},
		{"255.255.255.254", 6, "0.0.0.4"},
	}
	for _, tc := range cases {
		got := IncrementIPv4(net.ParseIP(tc.in), tc.increment)
		if diff := Diff(got, tc.want); diff != "" {
			t.Errorf("IncrementIPv4(%q, %d) (-got +want):\n%s",
				tc.in, tc.increment, diff)
		}
	}
}
