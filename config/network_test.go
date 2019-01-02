package config

import (
	"fmt"
	"net"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestUnmarshalAddr(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{":80", ":80"},
		{"0.0.0.0:80", "0.0.0.0:80"},
		{"127.0.0.1:80", "127.0.0.1:80"},
		{"localhost:80", "127.0.0.1:80"},
		//		{"ip6-localhost:80", "[::1]:80"},
		{"[2001:db8:cafe::1]:80", "[2001:db8:cafe::1]:80"},
		{"[::]:80", "[::]:80"},
		{"127.0.0.1", ""},
		{"127.0.0.1:65536", ""},
		{"127.0.0.1:-1", ""},
		{"127.0.0.1.14.5:80", ""},
		{"~~hello!!:80", ""},
		{"::15::16:80", ""},
		{"i.shoud.get.nxdomain.invalid:80", ""},
	}
	for _, tc := range cases {
		var got Addr
		input := fmt.Sprintf("%q", tc.in)
		err := yaml.Unmarshal([]byte(input), &got)
		switch {
		case err != nil && tc.want != "":
			t.Errorf("Unmarshal(%q) error\n%+v", tc.in, err)
		case err == nil && tc.want == "":
			t.Errorf("Unmarshal(%q) == %q but expected error", tc.in, got.String())
		case err == nil && tc.want != got.String():
			t.Errorf("Unmarshal(%q) == %q but expected %q", tc.in, got.String(), tc.want)
		}
	}
}

func TestUnmarshalHostname(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"127.0.0.1", ""},
		{"::1", ""},
		{"192.0.2.16", ""},
		{"2001:db8:cafe:cafe::1", ""},
		{"0.0.0.0", ""},
	}
	for _, c := range cases {
		var got Hostname
		err := yaml.Unmarshal([]byte(c.in), &got)
		if err != nil {
			t.Errorf("Unmarshal(%q) error:\n%+v", c.in, err)
		}
		expected := c.want
		if len(expected) == 0 {
			expected = c.in
		}
		expectedHost := net.ParseIP(expected)
		if err == nil && !net.IP(got).Equal(expectedHost) {
			t.Errorf("Unmarshal(%q) == %v but expected %v", c, got, expectedHost)
		}
	}
}

func TestUnmarshalInvalidHostname(t *testing.T) {
	cases := []string{
		"127.0.0.1.14.5",
		"~~hello!!",
		"::15::16",
	}
	for _, c := range cases {
		var got Hostname
		err := yaml.Unmarshal([]byte(c), &got)
		if err == nil {
			t.Errorf("Unmarshal(%q) == %v but expected error", c, got)
		}
	}
}

func TestUnmarshalUnresolvableHostname(t *testing.T) {
	var got Hostname
	c := "i.cant.exist.invalid"
	err := yaml.Unmarshal([]byte(c), &got)
	if err == nil {
		t.Errorf("Unmarshal(%q) == %v but expected error", c, got)
	}
}

func TestUnmarshalMACAddr(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"0a:00:27:00:00:00", ""},
		{"0b:00:27:00:00:00", ""},
	}
	for _, c := range cases {
		var got MACAddr
		err := yaml.Unmarshal([]byte(c.in), &got)
		if err != nil {
			t.Errorf("Unmarshal(%q) error:\n%+v", c.in, err)
		}
		expected := c.want
		if len(expected) == 0 {
			expected = c.in
		}
		if net.HardwareAddr(got).String() != expected {
			t.Errorf("Unmarshal(%q) == %v but expected %v", c, got, expected)
		}
	}
}

func TestUnmarshalInvalidMac(t *testing.T) {
	cases := []string{
		"0a:00:27:00:00:00",
		"~~hello!!",
		"0a:00:27:00:00",
	}
	for _, c := range cases {
		var got Hostname
		err := yaml.Unmarshal([]byte(c), &got)
		if err == nil {
			t.Errorf("Unmarshal(%q) == %v but expected error", c, got)
		}
	}
}
