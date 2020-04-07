package testing

import (
	"testing"

	"gopkg.in/inconshreveable/log15.v2"
)

// LogHandler represents a testing logger implementing the inconshreveable/log15.v2 Handler interface.
type LogHandler struct {
	t *testing.T
}

// NewLogHandler returns a new LogHandler instance.
func NewLogHandler(t *testing.T) *LogHandler {
	return &LogHandler{t}
}

// Log logs a log15.Record to the testing logger if the tests are run in verbose mode.
func (h *LogHandler) Log(r *log15.Record) error {
	if testing.Verbose() {
		h.t.Logf("%s: lvl=%q msg=%q ctx=%v", h.t.Name(), r.Lvl.String(), r.Msg, r.Ctx)
	}

	return nil
}
