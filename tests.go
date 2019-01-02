// +build !release

package reporter

import (
	log "gopkg.in/inconshreveable/log15.v2"

	"github.com/exoscale/go-reporter/logger"
)

// NewMock creates a new reporter that does almost nothing. It can be
// used with tests. There is no need to call Close with the returned
// reporter.
func NewMock() *Reporter {
	r, err := New(Configuration{
		Logging: logger.Configuration{
			Console: true,
			Level:   logger.Lvl(log.LvlDebug),
		},
	})
	if err != nil {
		panic(err)
	}
	return r
}

// NewSilentMock creates a new reporter that does nothing.
func NewSilentMock() *Reporter {
	r, err := New(Configuration{
		Logging: logger.Configuration{
			Console: false,
		},
	})
	if err != nil {
		panic(err)
	}
	return r
}
