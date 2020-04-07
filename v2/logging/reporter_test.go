package logging

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	gtesting "github.com/exoscale/go-reporter/v2/testing"
)

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
	require.NotNil(t, reporter.log)

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

func TestReporter_SetLogger(t *testing.T) {
	t.Skip()
}
