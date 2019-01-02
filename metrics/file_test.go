package metrics

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rcrowley/go-metrics"

	"github.com/exoscale/go-reporter/config"
)

func TestFile(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "logs")
	if err != nil {
		t.Fatalf("Unable to create a temporary directory:\n%+v", err)
	}
	defer os.RemoveAll(tempDir)

	outputFile := filepath.Join(tempDir, "out.txt")
	var configuration Configuration = make([]ExporterConfiguration, 1, 1)
	configuration[0] = &FileConfiguration{
		Interval: config.Duration(1 * time.Second),
		Path:     config.FilePath(outputFile),
	}

	m, err := New(configuration, "project")
	if err != nil {
		t.Fatalf("New(%v) error:\n%+v", configuration, err)
	}
	m.MustStart()
	defer func() {
		m.Stop()
	}()

	// Increase some counter
	c := metrics.NewCounter()
	m.Registry.Register("foo", c)
	c.Inc(47)

	if testing.Short() {
		t.Skip("Skip logfile test in short mode")
	}

	// Check we can get the value from file
	time.Sleep(1500 * time.Millisecond)
	contents, err := ioutil.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Unable to read JSON metrics %q:\n%+v", outputFile, err)
	}
	lines := strings.Split(string(contents), "\n")
	if len(lines) != 2 {
		t.Fatalf("Got %d line of logs, expected %d:\n%+v",
			len(lines)-1, 1, lines)
	}
	var got struct {
		Foo struct {
			Count int
		}
	}
	if err := json.Unmarshal([]byte(lines[0]), &got); err != nil {
		t.Fatalf("Unable to decode JSON body:\n%s\nError:\n%+v", lines[0], err)
	}
	if got.Foo.Count != 47 {
		t.Fatalf("Expected Foo == 47 but got %d instead", got.Foo)
	}
}
