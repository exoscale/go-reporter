package logger

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	log "gopkg.in/inconshreveable/log15.v2"
)

// Lvl is a log level (debug, info, warning, ...)
type Lvl log.Lvl

// LogFile represents a log file (name and output format).
type LogFile struct {
	Name   string
	Format LogFormat
}

// LogFormat represents an output format for LogFile. Currently, only
// plain text and JSON are supported.
type LogFormat int

const (
	// FormatPlain is a plain text log file format.
	FormatPlain LogFormat = iota
	// FormatJSON is a JSONv1 log file format.
	FormatJSON
)

// Configuration if the configuration for logger.
//
// The ability to override log levels for some modules is currently
// missing.
type Configuration struct {
	Level   Lvl
	Console bool
	Syslog  bool
	Files   []LogFile
}

// DefaultConfiguration is the default logging configuration.
var DefaultConfiguration = Configuration{
	Level:   Lvl(log.LvlInfo),
	Console: false,
	Syslog:  true,
}

// String transform a level to string.
func (level Lvl) String() string {
	l := log.Lvl(level)
	return l.String()
}

// UnmarshalText parses a log level from YAML.
func (level *Lvl) UnmarshalText(text []byte) error {
	var l log.Lvl
	var err error
	logLevelString := string(text)
	if l, err = log.LvlFromString(logLevelString); err != nil {
		return errors.Wrap(err,
			fmt.Sprintf("unknown log level %q", logLevelString))
	}
	*level = Lvl(l)

	return nil
}

// UnmarshalText parses a path to a logfile from YAML.
//
// This returns an absolute cleaned path.
func (logFile *LogFile) UnmarshalText(text []byte) error {
	var err error
	logFileString := string(text)

	// Is there a format?
	var supportedFormats = []struct {
		prefix string
		format LogFormat
	}{
		{"plain:", FormatPlain},
		{"json:", FormatJSON},
		{"", FormatPlain}, // default value
	}
	for _, supported := range supportedFormats {
		if strings.HasPrefix(logFileString, supported.prefix) {
			logFile.Format = supported.format
			logFileString = strings.TrimPrefix(logFileString, supported.prefix)
			break
		}
	}

	var absolutePath string
	if absolutePath, err = filepath.Abs(logFileString); err != nil {
		return errors.Wrap(err, "")
	}

	logFile.Name = filepath.Clean(absolutePath)
	return nil
}

// UnmarshalYAML parses a logger configuration from YAML.
func (configuration *Configuration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawConfiguration Configuration
	raw := rawConfiguration(DefaultConfiguration)
	if err := unmarshal(&raw); err != nil {
		return errors.Wrap(err, "unable to decode logging configuration")
	}
	*configuration = Configuration(raw)
	return nil
}
