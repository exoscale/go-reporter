package reporter

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/exoscale/go-reporter/helpers"
	"github.com/exoscale/go-reporter/logger"
)

func TestLogrusIntegration(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "logs")
	if err != nil {
		t.Fatalf("Unable to create a temporary directory:\n%+v", err)
	}
	defer os.RemoveAll(tempDir)

	r, err := New(Configuration{
		Prefix: "reporter",
		Logging: logger.Configuration{
			Level: logger.DefaultConfiguration.Level,
			Files: []logger.LogFile{
				logger.LogFile{
					Name:   filepath.Join(tempDir, "out.json"),
					Format: logger.FormatJSON,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Unable to initialize new logger:\n%+v", err)
	}

	hook := r.NewLogrusHook()
	log := logrus.New()
	log.AddHook(hook)
	log.Out = ioutil.Discard
	log.Info("hello")
	log.WithFields(logrus.Fields{
		"foo": "baz",
		"bar": 45,
	}).Warn("great!")
	hook.Disable()
	log.Info("bye!")

	contents, err := ioutil.ReadFile(filepath.Join(tempDir, "out.json"))
	if err != nil {
		t.Fatalf("Unable to read JSON log %q:\n%+v", filepath.Join(tempDir, "out.json"), err)
	}
	lines := make([](map[string]interface{}), 0)
	for _, line := range strings.Split(string(contents), "\n") {
		if len(line) == 0 {
			continue
		}
		var decodedLine map[string]interface{}
		if err := json.Unmarshal([]byte(line), &decodedLine); err != nil {
			t.Fatalf("Unable to decode JSON log line %q:\n%+v", line, err)
		}
		delete(decodedLine, "@timestamp")
		delete(decodedLine, "caller")
		lines = append(lines, decodedLine)
	}
	expected := [](map[string]interface{}){
		{
			"@version": 1,
			"level":    "info",
			"message":  "hello",
		}, {
			"@version": 1,
			"level":    "warn",
			"message":  "great!",
			"foo":      "baz",
			"bar":      45,
		},
	}
	if diff := helpers.Diff(lines, expected); diff != "" {
		t.Errorf("Not the logs we were expecting (-got +want):\n%s", diff)
	}
}
