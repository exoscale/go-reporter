package logging

import (
	"log/syslog"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"gopkg.in/inconshreveable/log15.v2"
)

var (
	defaultLogLevel  = "error"
	defaultLogFormat = "plain"
)

// LogDestinationConfig represents a logging reporter destination.
type LogDestinationConfig struct {
	// Type represents the destination type (file|console|syslog).
	Type string `yaml:"type"`

	// Destination represents the log destination depending on the type:
	// - For type "file", is must be a filesystem path
	// - For "console", it is ignored
	// - For type "syslog", it can be either empty (local syslog) or a net.Dial format string  for remote syslog
	Destination string `yaml:"destination"`

	// Level represents the highest message severity level to report (crit..debug).
	Level string `yaml:"level"`

	// Format represents the format to apply to logged messages (plain|json).
	Format string `yaml:"format"`
}

func (c *LogDestinationConfig) logFormat() log15.Format {
	switch c.Format {
	case "json":
		return log15.JsonFormat()

	default:
		return log15.LogfmtFormat()
	}
}

func (c *LogDestinationConfig) validate() error {
	if c.Level == "" {
		c.Level = defaultLogLevel
	}

	if c.Format == "" {
		c.Format = defaultLogFormat
	}

	return validation.ValidateStruct(c,
		validation.Field(&c.Type,
			validation.Required,
			validation.In(
				"file",
				"console",
				"syslog",
			)),

		validation.Field(&c.Destination,
			validation.When(c.Type == "file", validation.Required),
			validation.When(c.Type == "syslog" && c.Destination != "", is.DialString)),

		validation.Field(&c.Level,
			validation.By(func(v interface{}) error {
				_, err := log15.LvlFromString(v.(string))
				return err
			})),

		validation.Field(&c.Format,
			validation.In(
				"plain",
				"json",
			)),
	)
}

// Config represents a logging reporter configuration.
type Config struct {
	// Destinations represents the list of logging destinations.
	Destinations []*LogDestinationConfig `yaml:"destination"`

	// ReportErrors represents a flag indicating whether to automatically send error-level and higher
	// log messages to the errors reporter (the errors reporter has to be configured).
	ReportErrors bool `yaml:"report_errors"`
}

func (c *Config) validate() error {
	for _, d := range c.Destinations {
		if err := d.validate(); err != nil {
			return err
		}
	}

	return nil
}

func newFileHandler(d *LogDestinationConfig) (log15.Handler, error) {
	fileHandler, err := log15.FileHandler(d.Destination, d.logFormat())
	if err != nil {
		return nil, err
	}

	logLevel, err := log15.LvlFromString(d.Level)
	if err != nil {
		return nil, err
	}

	return log15.LvlFilterHandler(logLevel, fileHandler), nil
}

func newSyslogHandler(d *LogDestinationConfig) (log15.Handler, error) {
	var (
		syslogHandler log15.Handler
		err           error
	)

	syslogHandler, err = log15.SyslogHandler(syslog.LOG_INFO, "", d.logFormat())
	if d.Destination != "" {
		syslogHandler, err = log15.SyslogNetHandler("tcp", d.Destination, syslog.LOG_INFO, "", d.logFormat())
	}
	if err != nil {
		return nil, err
	}

	logLevel, err := log15.LvlFromString(d.Level)
	if err != nil {
		return nil, err
	}

	return log15.LvlFilterHandler(logLevel, syslogHandler), nil
}

func newConsoleHandler(d *LogDestinationConfig) (log15.Handler, error) {
	logLevel, err := log15.LvlFromString(d.Level)
	if err != nil {
		return nil, err
	}

	return log15.LvlFilterHandler(logLevel, log15.StderrHandler), nil
}
