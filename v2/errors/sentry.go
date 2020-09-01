package errors

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"gopkg.in/inconshreveable/log15.v2"

	"github.com/exoscale/go-reporter/v2/internal/debug"
)

const sentryFlushTimeout = 5 * time.Second

var (
	// internalPackages represents a list of packages to be excluded from the errors stack trace sent to Sentry.
	// Note: the Sentry SDK encodes dots in packages' import path into "%2e".
	internalPackages = []string{
		"github.com/exoscale/go-reporter",
		"gopkg.in/inconshreveable/log15%2ev2",
	}
)

// sentryEventModifier implements the sentry.EventModifier interface.
type sentryEventModifier struct {
	tags  map[string]string
	panic bool
}

func (m *sentryEventModifier) ApplyToEvent(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
	// Clean stack traces by filtering non-relevant frames
	for i := range event.Exception {
		event.Exception[i].Stacktrace.Frames = filterStackFrames(event.Exception[i].Stacktrace.Frames)
	}
	for i := range event.Threads {
		event.Threads[i].Stacktrace.Frames = filterStackFrames(event.Threads[i].Stacktrace.Frames)
	}

	// Add tags extracted from the original log record context
	event.Tags = m.tags

	// If the event has been created following a panic, flag it as crashed
	if m.panic && len(event.Threads) == 1 {
		event.Threads[0].Crashed = true
	}

	return event
}

// sentryEventFromLogRecord returns a sentryEventModifier instance containing tags extracted from a log record's
// context.
func sentryEventFromLogRecord(rec *log15.Record) *sentryEventModifier {
	var s = sentryEventModifier{tags: make(map[string]string)}

	// Extract context key/value pairs from the log record's context,
	// which is stored as a string slice of key/value pairs.
	for i := 0; i < len(rec.Ctx); i += 2 {
		// Guard against bogus records containing an odd number of context values
		if len(rec.Ctx[i:]) == 1 {
			break
		}

		k, v := rec.Ctx[i], rec.Ctx[i+1]
		s.tags[fmt.Sprint(k)] = fmt.Sprint(v)
	}

	return &s
}

// sentryEventFromPanic returns a sentryEventModifier instance indicating a crash.
func sentryEventFromPanic(_ interface{}) *sentryEventModifier {
	return &sentryEventModifier{
		tags:  make(map[string]string),
		panic: true,
	}
}

// sentryEventWithTags returns a sentryEventModifier instance adding user-specifed tags.
func sentryEventWithTags(tags map[string]string) *sentryEventModifier {
	return &sentryEventModifier{tags: tags}
}

// filterStackFrames filters out all frames related to internal packages.
func filterStackFrames(frames []sentry.Frame) []sentry.Frame {
	var filteredFrames = make([]sentry.Frame, 0)

nextFrame:
	for _, frame := range frames {
		for _, pkg := range internalPackages {
			if strings.HasPrefix(frame.Module, pkg) {
				continue nextFrame
			}
		}
		filteredFrames = append(filteredFrames, frame)
	}

	return filteredFrames
}

// SentryTestTransport is an implementation of the sentry.Transport interface for testing purposes.
type SentryTestTransport struct {
	mu        sync.Mutex
	events    []*sentry.Event
	lastEvent *sentry.Event
}

// Configure is a no-op for SentryTestTransport.
func (t *SentryTestTransport) Configure(_ sentry.ClientOptions) {}

// SendEvent assembles a new packet out of `Event` and sends it to remote server.
func (t *SentryTestTransport) SendEvent(event *sentry.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, event)
	t.lastEvent = event
}

// Flush is a no-op for SentryTestTransport. It always returns true immediately.
func (t *SentryTestTransport) Flush(_ time.Duration) bool {
	return true
}

// Events returns a list of events received by the Transport.
func (t *SentryTestTransport) Events() []*sentry.Event {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.events
}

type sentryDebugWriter struct {
	d *debug.D
}

func (w *sentryDebugWriter) Write(p []byte) (n int, err error) {
	w.d.Debug(string(p))

	return len(p), nil
}
