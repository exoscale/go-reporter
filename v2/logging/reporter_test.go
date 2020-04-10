package logging

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/inconshreveable/log15.v2"

	gtesting "github.com/exoscale/go-reporter/v2/testing"
)

type testLogHandler struct {
	records []*log15.Record

	sync.RWMutex
}

func (h *testLogHandler) Log(r *log15.Record) error {
	h.Lock()
	defer h.Unlock()

	h.records = append(h.records, r)

	return nil
}

func newTestLogHandler() *testLogHandler {
	return &testLogHandler{
		records: make([]*log15.Record, 0),
	}
}

func Test_mapToLogContext(t *testing.T) {
	var (
		testContext = map[string]string{
			"k1": "v1",
			"k2": "v2",
		}
	)

	ctx := mapToLogContext(testContext)

	require.Len(t, ctx, 4)
	require.Equal(t, []interface{}{"k1", "v1", "k2", "v2"}, ctx)
}

func Test_withUserContext(t *testing.T) {
	var (
		testErrorMessage = "oh noes!"
	)

	testHandler := newTestLogHandler()

	reporter, err := New(&Config{
		Destinations: []*LogDestinationConfig{
			{Type: "console"},
		},
		Context: map[string]string{
			"k1": "v1",
			"k2": "v2",
		},
	})
	require.NoError(t, err)

	reporter.SetHandler(testHandler)
	reporter.Info(testErrorMessage, "k3", "v3")

	require.Len(t, testHandler.records, 1)
	require.Equal(t, testErrorMessage, testHandler.records[0].Msg)
	require.Equal(t, []interface{}{"k1", "v1", "k2", "v2", "k3", "v3"}, testHandler.records[0].Ctx)
}

func TestNew(t *testing.T) {
	var (
		testDestFilePlain = path.Join(os.TempDir(), "go-reporter.log")
		testDestFileJSON  = path.Join(os.TempDir(), "go-reporter.log.json")
		testErrorMessage  = "oh noes!"
		testConfig        = &Config{
			Destinations: []*LogDestinationConfig{
				{Type: "file", Destination: testDestFilePlain},
				{Type: "file", Destination: testDestFileJSON, Format: "json"},
			},
		}
	)

	defer os.Remove(testDestFilePlain)
	defer os.Remove(testDestFileJSON)

	reporter, err := New(nil)
	require.NoError(t, err)
	require.Nil(t, reporter)

	_, err = New(&Config{})
	require.NoError(t, err)

	reporter, err = New(testConfig)
	require.NoError(t, err)
	require.NotNil(t, reporter)
	require.NotNil(t, reporter.config)

	reporter.Error(testErrorMessage)
	require.FileExists(t, testDestFilePlain)
	require.True(t, gtesting.FileContains(t, testDestFilePlain, testErrorMessage))

	record := make(map[string]interface{})
	data, err := ioutil.ReadFile(testDestFileJSON)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &record))
	require.Equal(t, testErrorMessage, record["msg"])
}

func TestReporter_Start(t *testing.T) {
	t.Skip()
}

func TestReporter_Stop(t *testing.T) {
	t.Skip()
}

func TestReporter_SetHandler(t *testing.T) {
	var (
		testErrorMessage = "oh noes!"
	)

	testHandler := newTestLogHandler()

	reporter, err := New(&Config{Destinations: []*LogDestinationConfig{
		{Type: "console"},
	}})
	require.NoError(t, err)

	reporter.SetHandler(testHandler)
	reporter.Info(testErrorMessage)

	require.Len(t, testHandler.records, 1)
	require.Equal(t, testErrorMessage, testHandler.records[0].Msg)
}

func TestReporter_Handler(t *testing.T) {
	testHandler := newTestLogHandler()

	reporter, err := New(&Config{Destinations: []*LogDestinationConfig{
		{Type: "console"},
	}})
	require.NoError(t, err)

	reporter.logger.SetHandler(testHandler)

	require.Equal(t, testHandler, reporter.Handler())
}
