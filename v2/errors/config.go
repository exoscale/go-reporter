package errors

import validation "github.com/go-ozzo/ozzo-validation/v4"

// Config represents an errors reporter configuration.
type Config struct {
	// DSN represents the Sentry DSN.
	DSN string `yaml:"dsn"`

	// Wait represents a flag indicating if the calls to Sentry should be done synchronously
	// (effectively blocking the caller).
	Wait bool `yaml:"wait"`

	// Debug represents a flags indicating whether to enable internal reporter activity logging.
	// This is mainly for debug purposes.
	Debug bool `yaml:"debug"`
}

func (c *Config) validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.DSN, validation.Required))
}
