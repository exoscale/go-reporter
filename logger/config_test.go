package logger

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v2"

	log "gopkg.in/inconshreveable/log15.v2"

	"github.com/exoscale/go-reporter/helpers"
)

func TestUnmarshalLogLevel(t *testing.T) {
	cases := []struct {
		in   string
		want Lvl
	}{
		{"info", Lvl(log.LvlInfo)},
		{"warn", Lvl(log.LvlWarn)},
	}
	for _, c := range cases {
		var got Lvl
		err := yaml.Unmarshal([]byte(c.in), &got)
		if err != nil {
			t.Errorf("Unmarshal(%q) error:\n%+v", c.in, err)
			continue
		}
		if diff := helpers.Diff(got, c.want); diff != "" {
			t.Errorf("Unmarshal(%q) (-got +want):\n%s", c.in, diff)
		}
	}

	errorCases := []struct {
		in string
	}{
		{"unknown"},
		{"''"},
	}
	for _, c := range errorCases {
		var got Lvl
		err := yaml.Unmarshal([]byte(c.in), &got)
		if err == nil {
			t.Errorf("Unmarshal(%q) == %s but expected error", c.in, got)
		}
	}
}

func TestUnmarshalLogFile(t *testing.T) {
	var currentDirectory string
	var err error
	if currentDirectory, err = os.Getwd(); err != nil {
		t.Fatalf("Unable to get current directory:\n%+v", err)
	}
	cases := []struct {
		in   string
		want LogFile
	}{
		{"/etc/passwd", LogFile{
			Name:   "/etc/passwd",
			Format: FormatPlain,
		}},
		{"another/file", LogFile{
			Name:   filepath.Join(currentDirectory, "another/file"),
			Format: FormatPlain,
		}},
		{"./another/file", LogFile{
			Name:   filepath.Join(currentDirectory, "another/file"),
			Format: FormatPlain,
		}},
		{filepath.Join("..", filepath.Base(currentDirectory), "another/file"),
			LogFile{
				Name:   filepath.Join(currentDirectory, "another/file"),
				Format: FormatPlain,
			},
		},
		{"json:/var/log/something.json", LogFile{
			Name:   "/var/log/something.json",
			Format: FormatJSON,
		}},
		{"plain:/var/log/something.txt", LogFile{
			Name:   "/var/log/something.txt",
			Format: FormatPlain,
		}},
	}
	for _, c := range cases {
		var got LogFile
		err := yaml.Unmarshal([]byte(c.in), &got)
		if err != nil {
			t.Errorf("Unmarshal(%q) error:\n%+v", c.in, err)
			continue
		}
		if diff := helpers.Diff(got, c.want); diff != "" {
			t.Errorf("Unmarshal(%q) (-got +want):\n%s", c.in, diff)
		}
	}
}

func TestUnmarshalConfiguration(t *testing.T) {
	cases := []struct {
		in   string
		want Configuration
	}{
		{`
level: debug
console: true
syslog: false
files:
  - /var/log/project.log
  - json:/var/log/project.json
`,
			Configuration{
				Level:   Lvl(log.LvlDebug),
				Console: true,
				Syslog:  false,
				Files: []LogFile{
					LogFile{
						Name:   "/var/log/project.log",
						Format: FormatPlain,
					},
					LogFile{
						Name:   "/var/log/project.json",
						Format: FormatJSON,
					},
				}}},
		{"{}", Configuration{
			Level:   Lvl(log.LvlInfo),
			Console: false,
			Syslog:  true}},
	}
	for _, c := range cases {
		var got Configuration
		err := yaml.Unmarshal([]byte(c.in), &got)
		if err != nil {
			t.Errorf("Unmarshal(%q) error:\n%+v", c.in, err)
			continue
		}
		if diff := helpers.Diff(got, c.want); diff != "" {
			t.Errorf("Unmarshal(%q) (-got +want):\n%s", c.in, diff)
		}
	}
}
