package reporter

import (
	"fmt"

	"github.com/getsentry/raven-go"
	"github.com/pkg/errors"
	log "gopkg.in/inconshreveable/log15.v2"
)

// Interface Sentry with the logging subsytem. The idea is to register
// an additional handler that will forward errors to Sentry.
func sentryHandler(client *raven.Client) log.Handler {
	if client == nil {
		// No sentry configured
		return nil
	}
	return log.FuncHandler(func(r *log.Record) error {
		if r.Lvl > log.LvlError {
			return nil
		}

		// Convert the context to tags and retrieve the error
		tags := make(map[string]string, len(r.Ctx)/2)
		var err error
		for i := 0; i < len(r.Ctx); i += 2 {
			if v, ok := r.Ctx[i+1].(error); ok && v != nil {
				if v.Error() == r.Msg {
					err = v
				} else {
					err = errors.Wrap(v, r.Msg)
				}
				continue
			}
			k, ok := r.Ctx[i].(string)
			if !ok {
				continue
			}
			tags[k] = fmt.Sprintf("%+v", r.Ctx[i+1])
		}
		if err == nil {
			return nil
		}
		client.CaptureError(err, tags)
		return nil
	})
}
