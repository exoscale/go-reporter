package reporter

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"runtime/debug"
	"testing"
	"time"

	"github.com/pkg/errors"
	log "gopkg.in/inconshreveable/log15.v2"

	"github.com/exoscale/go-reporter/logger"
	"github.com/exoscale/go-reporter/sentry"
)

func TestSentryReporting(t *testing.T) {
	// Create a new webserver logging requests
	listener, err := net.Listen("tcp", "127.0.0.1:44444")
	if err != nil {
		t.Fatalf("Listen() error:\n%+v", err)
	}
	requests := make(chan []byte)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/sentry/", func(w http.ResponseWriter, r *http.Request) {
		// We don't really care about the method (POST) and
		// the URL (/api/sentry/store/). Let's send back the
		// body.
		switch r.Header.Get("Content-Type") {
		case "application/octet-stream":
			buf := &bytes.Buffer{}
			b64 := base64.NewDecoder(base64.StdEncoding, r.Body)
			deflate, _ := zlib.NewReader(b64)
			_, _ = io.Copy(buf, deflate)
			deflate.Close()
			requests <- buf.Bytes()
		case "application/json":
			body, _ := ioutil.ReadAll(r.Body)
			requests <- body
		}
		// Send back 200
	})
	go func() {
		_ = http.Serve(listener, mux)
	}()

	// Initialize a reporter
	r, err := New(Configuration{
		Logging: logger.Configuration{
			Console: true,
			Level:   logger.Lvl(log.LvlDebug),
		},
		Sentry: sentry.Configuration{
			DSN: fmt.Sprintf("http://public:secret@%s/sentry",
				listener.Addr().String()),
			Tags: map[string]string{"platform": "test"},
		},
	})
	t.Logf("http://public:secret@%s/sentry\n",
		listener.Addr().String())
	if err != nil {
		t.Fatalf("New() error:\n%+v", err)
	}

	// Trigger an error
	cases := []struct {
		err      string
		message  string
		expected string
	}{
		{"an error", "my first error", "my first error: an error"},
		{"an error", "", "an error"},
	}
	for _, tc := range cases {
		_ = r.Error(errors.New(tc.err), tc.message, "alfred", "batman")
		timeout := time.After(500 * time.Millisecond)
		select {
		case r := <-requests:
			type sentryBody struct {
				Message string
				EventID string `json:"event_id"`
				Project string
				Tags    [][]string
			}
			var got sentryBody
			if err := json.Unmarshal(r, &got); err != nil {
				t.Fatalf("Unmarshal() error:\n%+v", err)
			}
			if got.Message != tc.expected {
				t.Errorf("sentry.message == %q but expected %q", got.Message, tc.expected)
			}
			if got.Project != "sentry" {
				t.Errorf("sentry.project == %q but expected %q", got.Project, "sentry")
			}
			found := false
			for i, tagpair := range got.Tags {
				if len(tagpair) != 2 {
					t.Errorf("sentry.tags contains a non-tuple, at index %d", i+1)
					continue
				}
				key, value := tagpair[0], tagpair[1]
				if key == "platform" && value == "test" {
					found = true
				}
			}
			if !found {
				t.Errorf("sentry.tags doesn't contain platform,test")
			}
		case <-timeout:
			debug.SetTraceback("all")
			panic("timeout!")
		}
		select {
		case r := <-requests:
			t.Fatalf("Error() triggered an additional Sentry request:\n%+v", r)
		case <-timeout:
			// OK
		}
	}
}
