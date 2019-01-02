// Interface for Logrus

package reporter

import (
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

// LogrusHook hooks logrus to reporter
type LogrusHook struct {
	r      *Reporter
	active int32
}

// NewLogrusHook returns a hook suitable for logrus
func (r *Reporter) NewLogrusHook() *LogrusHook {
	return &LogrusHook{r, 1}
}

// Fire handles a logrus entry.
func (hook *LogrusHook) Fire(entry *logrus.Entry) error {
	if atomic.LoadInt32(&hook.active) == 0 {
		return nil
	}

	// Convert to a list
	args := make([]interface{}, len(entry.Data)*2)
	i := 0
	for name, value := range entry.Data {
		args[i] = name
		args[i+1] = value
		i += 2
	}
	switch entry.Level {
	case logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel:
		// Nothing is important enough to get a true error
		hook.r.Warn(entry.Message, args...)
	case logrus.InfoLevel:
		hook.r.Info(entry.Message, args...)
	case logrus.DebugLevel:
		hook.r.Debug(entry.Message, args...)
	}
	return nil
}

// Levels return the levels supported by this hook. We support them
// all.
func (hook *LogrusHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Disable disables the hook. We cannot really remove a hook from logrus. We
// assume during normal run, hook is done only once, so we won't call
// Disable() except on shutdown and the impact on performance is
// negligible.
func (hook *LogrusHook) Disable() {
	atomic.StoreInt32(&hook.active, 0)
}
