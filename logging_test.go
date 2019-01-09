package reporter

import (
	"errors"
	"testing"
)

func TestLogging(t *testing.T) {
	r := NewMock()
	r.Info("hello")
	r.Debug("help")
	r.Debug("help", "context", 76, "another context", 8.9)
}

func TestLoggingError(t *testing.T) {
	r := NewMock()
	err := errors.New("hello")
	got := r.Error(err, "", "context", "nothing")
	if got != err {
		t.Errorf("Error(%q, \"\") == %q but expected %q",
			err, got, err)
	}
	got = r.Error(err, "hey", "context", "nothing")
	expected := "hey: hello"
	if got.Error() != expected {
		t.Errorf("Error(%q, \"hey\") == %q but expected %q",
			err, got, expected)
	}
}

func BenchmarkNonLogging(b *testing.B) {
	r := NewSilentMock()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Debug("help", "some", 76, "context", 89)
	}
}

func BenchmarkLogging(b *testing.B) {
	r := NewSilentMock()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Error(errors.New("batman"), "help", "some", 76, "context", 89)
	}
}
