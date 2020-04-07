package v2

// Crit is a convenience wrapper around Reporter.Logging.Crit().
// It is effective only if the reporter has its logging reporter configured.
func (r *Reporter) Crit(msg string, ctx ...interface{}) {
	if r.Logging != nil {
		r.Logging.Crit(msg, ctx...)
	}
}

// Error is a convenience wrapper around Reporter.Logging.Error().
// It is effective only if the reporter has its logging reporter configured.
func (r *Reporter) Error(msg string, ctx ...interface{}) {
	if r.Logging != nil {
		r.Logging.Error(msg, ctx...)
	}
}

// Warn is a convenience wrapper around Reporter.Logging.Warn().
// It is effective only if the reporter has its logging reporter configured.
func (r *Reporter) Warn(msg string, ctx ...interface{}) {
	if r.Logging != nil {
		r.Logging.Warn(msg, ctx...)
	}
}

// Info is a convenience wrapper around Reporter.Logging.Info().
// It is effective only if the reporter has its logging reporter configured.
func (r *Reporter) Info(msg string, ctx ...interface{}) {
	if r.Logging != nil {
		r.Logging.Info(msg, ctx...)
	}
}

// Debug is a convenience wrapper around Reporter.Logging.Debug().
// It is effective only if the reporter has its logging reporter configured.
func (r *Reporter) Debug(msg string, ctx ...interface{}) {
	if r.Logging != nil {
		r.Logging.Debug(msg, ctx...)
	}
}
