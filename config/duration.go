package config

import (
	"time"

	"github.com/pkg/errors"
)

// Duration is an "alias" for time.Duration. It just gets an
// UnmarshalText helper function.
type Duration time.Duration

func (d Duration) String() string {
	return time.Duration(d).String()
}

// UnmarshalText parses a duration string.
func (d *Duration) UnmarshalText(text []byte) error {
	dur, err := time.ParseDuration(string(text))
	if err != nil {
		return errors.Wrapf(err, "invalid duration")
	}
	*d = Duration(dur)
	return nil
}
