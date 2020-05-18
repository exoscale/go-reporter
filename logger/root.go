// Package logger handles logging.
//
// This is a thing wrapper around inconshreveable's log15. It
// additionally brings a configuration format (from YAML) with the
// ability to log to several destinations.
//
// It also brings some conventions like the presence of "module" in
// each context to be able to filter logs more easily. However, this
// convention is not really enforced. Once you have a root logger,
// create sublogger with New and provide a new value for "module".
package logger

import (
	"encoding/json"
	"fmt"
	"log/syslog"
	"os"
	"strings"
	"time"

	"github.com/go-stack/stack"
	"github.com/pkg/errors"
	log "gopkg.in/inconshreveable/log15.v2"
)

const (
	timeFormat = "2006-01-02T15:04:05-0700"
)

// // Find our own name to be able to extract caller function and package reliably.
// var (
// 	ownPackageCall    = stack.Caller(0)                                               // reporter/logger.init
// 	ownPackageName    = strings.SplitN(fmt.Sprintf("%+n", ownPackageCall), ".", 2)[0] // reporter/logger
// 	parentPackageName = ownPackageName[0:strings.LastIndex(ownPackageName, "/")]      // reporter
// )

func logFormat(logFormat LogFormat) (log.Format, error) {
	switch logFormat {
	case FormatPlain:
		return log.LogfmtFormat(), nil
	case FormatJSON:
		return JSONv1Format(), nil
	default:
		return nil, fmt.Errorf("unknown format provided: %v", logFormat)
	}
}

// New creates a new logger from a configuration.
func New(config Configuration, additionalHandler log.Handler, prefix string) (log.Logger, error) {
	handlers := make([]log.Handler, 0, 10)
	defaultFormatter, err := logFormat(config.Format)
	if err != nil {
		return nil, err
	}
	// We need to build the appropriate handler.
	if config.Console {
		if config.Format == FormatPlain {
			handlers = append(handlers, log.StdoutHandler)

		} else {
			handlers = append(handlers, log.StreamHandler(os.Stdout, defaultFormatter))
		}
	}
	if config.Syslog {
		handler, err := log.SyslogHandler(syslog.LOG_INFO, prefix, defaultFormatter)
		if err != nil {
			return nil, errors.Wrap(err, "unable to open syslog connection")
		}
		handlers = append(handlers, handler)
	}
	for _, logFile := range config.Files {
		formatter, err := logFormat(logFile.Format)
		if err != nil {
			return nil, err
		}
		handler, err := log.FileHandler(logFile.Name, formatter)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to open log file %q", logFile.Name)
		}
		handlers = append(handlers, handler)
	}

	// Initialize the logger
	var logger = log.New()

	funcHandler := log.FuncHandler(func(r *log.Record) error {
		return log.MultiHandler(handlers...).Log(r)
	})

	if config.IncludeCaller {
		funcHandler = contextHandler(log.MultiHandler(handlers...), prefix)
	}

	logHandler := log.LvlFilterHandler(
		log.Lvl(config.Level),
		funcHandler,
	)

	if additionalHandler != nil {
		logHandler = log.MultiHandler(logHandler, additionalHandler)
	}

	logger.SetHandler(logHandler)

	return logger, nil
}

// Add more context to log entry. This is similar to
// log.CallerFileHandler and log.CallerFuncHandler but it's a bit
// smarter on how the stack trace is inspected to avoid logging
// modules. It adds a "caller" (qualified function + line number) and
// a "module" (PROJECT package name).
func contextHandler(h log.Handler, prefix string) log.Handler {
	skipPrefixes := []string{
		"github.com/exoscale/go-reporter",
		"github.com/sirupsen/logrus",
		fmt.Sprintf("%s/vendor/github.com/sirupsen/logrus", prefix),
	}
	return log.FuncHandler(func(r *log.Record) error {
		callStack := stack.Trace().TrimBelow(r.Call)
		callerFound := false
	outer:
		for _, call := range callStack {
			if !callerFound {
				// Searching for the first caller.
				caller := fmt.Sprintf("%+v", call)
				for _, prefix := range skipPrefixes {
					if strings.HasPrefix(caller, prefix) {
						continue outer
					}
				}
				r.Ctx = append(r.Ctx, "caller", caller)
				callerFound = true
			}
			if callerFound {
				// Searching for the package name. We
				// want the first inside our specified
				// prefix.
				module := fmt.Sprintf("%+n", call)
				if !strings.HasPrefix(module, prefix) {
					continue
				}
				if strings.HasPrefix(module, fmt.Sprintf("%s/vendor/", prefix)) {
					continue
				}
				module = strings.SplitN(module, ".", 2)[0]
				r.Ctx = append(r.Ctx, "module", module)
				return h.Log(r)
			}
		}
		return h.Log(r)
	})
}

// JSONv1Format formats log records as JSONv1 objects separated by newlines.
func JSONv1Format() log.Format {
	return log.FormatFunc(func(r *log.Record) []byte {
		props := make(map[string]interface{})

		var level string
		switch r.Lvl {
		case log.LvlDebug:
			level = "debug"
		case log.LvlInfo:
			level = "info"
		case log.LvlWarn:
			level = "warn"
		case log.LvlError:
			level = "error"
		case log.LvlCrit:
			level = "crit"
		default:
			panic("bad level")
		}

		props["@version"] = 1
		props["@timestamp"] = r.Time
		props["level"] = level
		props["message"] = r.Msg

		for i := 0; i < len(r.Ctx); i += 2 {
			k, ok := r.Ctx[i].(string)
			if !ok {
				props["LOG_ERROR"] = fmt.Sprintf("%+v is not a string key", r.Ctx[i])
			}
			props[k] = formatJSONValue(r.Ctx[i+1])
		}

		b, err := json.Marshal(props)
		if err != nil {
			b, _ = json.Marshal(map[string]string{
				"LOG_ERROR": err.Error(),
			})
		}

		b = append(b, '\n')
		return b
	})

}

func formatJSONValue(value interface{}) interface{} {
	switch v := value.(type) {
	case time.Time:
		return v.Format(timeFormat)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	case int, int8, int16, int32, int64, float32, float64, uint, uint8, uint16, uint32, uint64, string:
		return value
	default:
		return fmt.Sprintf("%+v", value)
	}
}
