package errors

import (
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/require"
	"gopkg.in/inconshreveable/log15.v2"
)

var testSentryDSN = "https://xxxxxxxxxxxxxxxxxxxxx:yyyyyyyyyyyyyyyyyyy@sentry.example.net/42"

func TestFilterStackFrames(t *testing.T) {
	var testStacktrace = []sentry.Frame{
		sentry.Frame{Function: "main", Symbol: "", Module: "main", Package: "", Filename: "goreporter.go", AbsPath: "/Users/marc/Documents/Exoscale/git/go-reporter/v1/test/goreporter.go", Lineno: 60, Colno: 0, PreContext: []string(nil), ContextLine: "", PostContext: []string(nil), InApp: true, Vars: map[string]interface{}(nil)},
		sentry.Frame{Function: "doSomething", Symbol: "", Module: "main", Package: "", Filename: "goreporter.go", AbsPath: "/Users/marc/Documents/Exoscale/git/go-reporter/v1/test/goreporter.go", Lineno: 81, Colno: 0, PreContext: []string(nil), ContextLine: "", PostContext: []string(nil), InApp: true, Vars: map[string]interface{}(nil)},
		sentry.Frame{Function: "(*Reporter).Error", Symbol: "", Module: "github.com/exoscale/go-reporter/v1", Package: "", Filename: "logging.go", AbsPath: "/Users/marc/Documents/Exoscale/git/go-reporter/v1/logging.go", Lineno: 15, Colno: 0, PreContext: []string(nil), ContextLine: "", PostContext: []string(nil), InApp: true, Vars: map[string]interface{}(nil)},
		sentry.Frame{Function: "(*Reporter).Error", Symbol: "", Module: "github.com/exoscale/go-reporter/v1/logging", Package: "", Filename: "reporter.go", AbsPath: "/Users/marc/Documents/Exoscale/git/go-reporter/v1/logging/reporter.go", Lineno: 83, Colno: 0, PreContext: []string(nil), ContextLine: "", PostContext: []string(nil), InApp: true, Vars: map[string]interface{}(nil)},
		sentry.Frame{Function: "(*logger).Error", Symbol: "", Module: "gopkg.in/inconshreveable/log15%2ev2", Package: "", Filename: "logger.go", AbsPath: "/Users/marc/.go/pkg/mod/gopkg.in/inconshreveable/log15.v2@v2.0.0-20200109203555-b30bc20e4fd1/logger.go", Lineno: 153, Colno: 0, PreContext: []string(nil), ContextLine: "", PostContext: []string(nil), InApp: true, Vars: map[string]interface{}(nil)},
		sentry.Frame{Function: "(*logger).write", Symbol: "", Module: "gopkg.in/inconshreveable/log15%2ev2", Package: "", Filename: "logger.go", AbsPath: "/Users/marc/.go/pkg/mod/gopkg.in/inconshreveable/log15.v2@v2.0.0-20200109203555-b30bc20e4fd1/logger.go", Lineno: 112, Colno: 0, PreContext: []string(nil), ContextLine: "", PostContext: []string(nil), InApp: true, Vars: map[string]interface{}(nil)},
		sentry.Frame{Function: "(*swapHandler).Log", Symbol: "", Module: "gopkg.in/inconshreveable/log15%2ev2", Package: "", Filename: "handler_go14.go", AbsPath: "/Users/marc/.go/pkg/mod/gopkg.in/inconshreveable/log15.v2@v2.0.0-20200109203555-b30bc20e4fd1/handler_go14.go", Lineno: 14, Colno: 0, PreContext: []string(nil), ContextLine: "", PostContext: []string(nil), InApp: true, Vars: map[string]interface{}(nil)},
		sentry.Frame{Function: "funcHandler.Log", Symbol: "", Module: "gopkg.in/inconshreveable/log15%2ev2", Package: "", Filename: "handler.go", AbsPath: "/Users/marc/.go/pkg/mod/gopkg.in/inconshreveable/log15.v2@v2.0.0-20200109203555-b30bc20e4fd1/handler.go", Lineno: 31, Colno: 0, PreContext: []string(nil), ContextLine: "", PostContext: []string(nil), InApp: true, Vars: map[string]interface{}(nil)},
		sentry.Frame{Function: "MultiHandler.func1", Symbol: "", Module: "gopkg.in/inconshreveable/log15%2ev2", Package: "", Filename: "handler.go", AbsPath: "/Users/marc/.go/pkg/mod/gopkg.in/inconshreveable/log15.v2@v2.0.0-20200109203555-b30bc20e4fd1/handler.go", Lineno: 204, Colno: 0, PreContext: []string(nil), ContextLine: "", PostContext: []string(nil), InApp: true, Vars: map[string]interface{}(nil)},
		sentry.Frame{Function: "funcHandler.Log", Symbol: "", Module: "gopkg.in/inconshreveable/log15%2ev2", Package: "", Filename: "handler.go", AbsPath: "/Users/marc/.go/pkg/mod/gopkg.in/inconshreveable/log15.v2@v2.0.0-20200109203555-b30bc20e4fd1/handler.go", Lineno: 31, Colno: 0, PreContext: []string(nil), ContextLine: "", PostContext: []string(nil), InApp: true, Vars: map[string]interface{}(nil)},
		sentry.Frame{Function: "(*Reporter).LogHandler.func1", Symbol: "", Module: "github.com/exoscale/go-reporter/v1/errors", Package: "", Filename: "reporter.go", AbsPath: "/Users/marc/Documents/Exoscale/git/go-reporter/v1/errors/reporter.go", Lineno: 57, Colno: 0, PreContext: []string(nil), ContextLine: "", PostContext: []string(nil), InApp: true, Vars: map[string]interface{}(nil)},
	}

	filteredFrames := filterStackFrames(testStacktrace)
	require.Len(t, filteredFrames, 2)
	require.Equal(t, "main", filteredFrames[0].Module)
	require.Equal(t, "main", filteredFrames[1].Module)
}

func TestSentryEventFromLogRecord(t *testing.T) {
	var testRec = log15.Record{
		Time: time.Now(),
		Lvl:  log15.LvlError,
		Msg:  "oh noes!",
		Ctx:  []interface{}{"k1", "v1", "k2", "v2"},
	}

	s := sentryEventFromLogRecord(&testRec)
	require.Len(t, s.tags, 2)
	require.Equal(t, map[string]string{
		"k1": "v1",
		"k2": "v2",
	}, s.tags)
}
