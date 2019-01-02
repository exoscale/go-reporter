package reporter

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestMetricsExpandName(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"test.name", "test.name"},
		{"name", "name"},
		{".test.name", "test.name"},
		{".name", "name"},
	}
	for _, c := range cases {
		got := expandName(c.in, "project")
		if got != c.want {
			t.Errorf("expandName(%q) == %q but expected %q",
				c.in, got, c.want)
		}
	}
}

func BenchmarkMetricsExpandName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		expandName("name", "project")
	}
}

func BenchmarkMetricsNonExpandName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		expandName(".name", "project")
	}
}

func TestMetricsCounter(t *testing.T) {
	r := NewMock()
	r.Counter("counter1").Inc(10)
	snap := r.Counter("counter1").Snapshot()
	got := snap.Count()
	if got != 10 {
		t.Errorf("Expected counter %v value == 10, got %d",
			snap, got)
	}
}

func TestMetricsConcurrentCounter(t *testing.T) {
	r := NewMock()
	wg := sync.WaitGroup{}
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			r.Counter("counter1").Inc(1)
			wg.Done()
		}()
	}
	wg.Wait()
	v := r.Counter("counter1").Count()
	if v != 200 {
		t.Errorf("Expected counter value == 100, got %d",
			v)
	}
}

func TestMetricsGauge(t *testing.T) {
	r := NewMock()
	r.Gauge("gauge1").Update(719)
	snap := r.Gauge("gauge1").Snapshot()
	got := snap.Value()
	if got != 719 {
		t.Errorf("Expected gauge %v value == 719, got %d",
			snap, got)
	}
}

func TestMetricsGaugeFloat64(t *testing.T) {
	r := NewMock()
	r.GaugeFloat64("gauge2").Update(71.9)
	snap := r.GaugeFloat64("gauge2").Snapshot()
	got := snap.Value()
	if got != 71.9 {
		t.Errorf("Expected gauge %v value == 71.9, got %f",
			snap, got)
	}
}

func TestMetricsHistogram(t *testing.T) {
	r := NewMock()
	r.Histogram("histogram2").Update(72)
	snap := r.Histogram("histogram2").Snapshot()
	got := snap.Max()
	if got != 72 {
		t.Errorf("Expected histogram %v value == 72, got %d",
			snap, got)
	}
}

func TestMetricsMeter(t *testing.T) {
	r := NewMock()
	r.Meter("meter2").Mark(72)
	snap := r.Meter("meter2").Snapshot()
	got := snap.Count()
	if got != 72 {
		t.Errorf("Expected meter %v value == 72, got %d",
			snap, got)
	}
}

func TestMetricsTimer(t *testing.T) {
	r := NewMock()
	r.Timer("timer2").Update(time.Second)
	snap := r.Timer("timer2").Snapshot()
	got := snap.Max()
	if got != int64(time.Second) {
		t.Errorf("Expected timer %v value == %v, got %d",
			snap, time.Second, got)
	}
}

func TestMetricsHealthcheck(t *testing.T) {
	r := NewMock()
	r.Healthcheck("pipo", func(h Healthcheck) {
		h.Unhealthy(errors.New("nope"))
	}).Check()
	if h := r.Healthcheck("pipo", nil); h.Error().Error() != "nope" {
		t.Errorf("Expected healthcheck error %q, got %q", "nope", h.Error())
	}
}
