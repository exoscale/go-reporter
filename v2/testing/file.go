package testing

import (
	"bytes"
	"io/ioutil"
	"testing"
)

// FileContains returns true if the substring is found in the file located at path, false otherwise.
// If a read error occurs, it makes the associated test fail.
func FileContains(t *testing.T, path, substring string) bool {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("%s", err)
	}

	return bytes.Contains(data, []byte(substring))
}
