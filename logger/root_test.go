package logger

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	log "gopkg.in/inconshreveable/log15.v2"
)

func TestNewEmpty(t *testing.T) {
	logger, err := New(Configuration{}, nil, "project")
	if err != nil {
		t.Fatalf("New({}) error:\n%+v", err)
	}
	logger.Info("log message", "integer", 15)
}

type logLine struct {
	Level     string
	Module    string
	Message   string
	Timestamp string `json:"@timestamp"`
	Version   int    `json:"@version"`
}

func TestNewFiles(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "logs")
	if err != nil {
		t.Fatalf("Unable to create a temporary directory:\n%+v", err)
	}
	defer os.RemoveAll(tempDir)

	logger, err := New(Configuration{
		Level: Lvl(log.LvlInfo),
		Files: []LogFile{
			LogFile{
				Name:   filepath.Join(tempDir, "out.txt"),
				Format: FormatPlain,
			},
			LogFile{
				Name:   filepath.Join(tempDir, "out.json"),
				Format: FormatJSON,
			},
		},
	}, nil, "project")
	if err != nil {
		t.Fatalf("Unable to initialize new logger:\n%+v", err)
	}

	logger.Info("hello")
	logger.Debug("nothing")
	logger.Crit("important")

	// Json file
	contents, err := ioutil.ReadFile(filepath.Join(tempDir, "out.json"))
	if err != nil {
		t.Fatalf("Unable to read JSON log %q:\n%+v", filepath.Join(tempDir, "out.json"), err)
	}
	lines := strings.Split(string(contents), "\n")
	if len(lines) != 3 {
		t.Fatalf("Got %d line of logs, expected %d:\n%+v",
			len(lines)-1, 2, lines)
	}
	var line1 logLine
	if err := json.Unmarshal([]byte(lines[0]), &line1); err != nil {
		t.Fatalf("Unable to decode JSON log line %q:\n%+v", lines[0], err)
	}
	if line1.Message != "hello" {
		t.Errorf("First line of log message should be %q but got %q",
			"hello", line1.Message)
	}
	if line1.Level != "info" {
		t.Errorf("First line of log message should be level %s but got %s",
			"info", line1.Level)
	}
	if line1.Version != 1 {
		t.Errorf("First line of log message should be version %d but got %d",
			1, line1.Version)
	}

	var line2 logLine
	if err := json.Unmarshal([]byte(lines[1]), &line2); err != nil {
		t.Fatalf("Unable to decode JSON log line %q:\n%+v", lines[1], err)
	}
	if line2.Message != "important" {
		t.Errorf("Second line of log message should be %q but got %q",
			"important", line2.Message)
	}
	if line2.Level != "crit" {
		t.Errorf("Second line of log message should be level %s but got %s",
			"crit", line2.Level)
	}

	// Text file
	contents, err = ioutil.ReadFile(filepath.Join(tempDir, "out.txt"))
	if err != nil {
		t.Fatalf("Unable to read text log %q:\n%+v", filepath.Join(tempDir, "out.txt"), err)
	}
	lines = strings.Split(string(contents), "\n")
	if len(lines) != 3 {
		t.Fatalf("Got %d line of logs, expected %d", len(lines)-1, 2)
	}
	if !strings.Contains(string(lines[0]), " msg=hello") {
		t.Fatalf("First log should be %q, got %q instead",
			"msg=hello", lines[0])
	}
	if !strings.Contains(string(lines[1]), " msg=important") {
		t.Fatalf("Second log should be %q, got %q instead",
			"msg=important", lines[1])
	}

}
