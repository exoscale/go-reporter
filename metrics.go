// Metrics façade for reporter.
//
// All the functions take a name and return an appropriate
// metric. This metric is created if it doesn't exist. Unless the name
// starts with ".", it will be prepended with the module name (guessed
// from the stack trace).

package reporter

import (
	"fmt"
	"strings"

	"github.com/go-stack/stack"
	"github.com/rcrowley/go-metrics"
)

// Counter returns a counter with the given name.
func (r *Reporter) Counter(name string) metrics.Counter {
	return metrics.GetOrRegisterCounter(expandName(name, r.prefix), r.metrics.Registry)
}

// Gauge returns a gauge with the given name.
func (r *Reporter) Gauge(name string) metrics.Gauge {
	return metrics.GetOrRegisterGauge(expandName(name, r.prefix), r.metrics.Registry)
}

// GaugeFloat64 returns a 64-bit float gauge with the given name.
func (r *Reporter) GaugeFloat64(name string) metrics.GaugeFloat64 {
	return metrics.GetOrRegisterGaugeFloat64(expandName(name, r.prefix), r.metrics.Registry)
}

// Histogram returns an histogram with the given name. This uses an
// exponentially-decating sample with a forward-decaying priority
// reservoir. It uses fixed values for the reservoir size and
// alpha. Those values were scientifically obtained through random
// Google searches.
func (r *Reporter) Histogram(name string) metrics.Histogram {
	reservoir := 100
	alpha := 0.015
	return metrics.GetOrRegisterHistogram(expandName(name, r.prefix),
		r.metrics.Registry,
		metrics.NewExpDecaySample(reservoir, alpha))
}

// Meter returns a meter with the given name.
func (r *Reporter) Meter(name string) metrics.Meter {
	return metrics.GetOrRegisterMeter(expandName(name, r.prefix), r.metrics.Registry)
}

// Timer returns a timer with the given name.
func (r *Reporter) Timer(name string) metrics.Timer {
	return metrics.GetOrRegisterTimer(expandName(name, r.prefix), r.metrics.Registry)
}

// Healthcheck holds healthcheck state.
type Healthcheck struct {
	metrics.Healthcheck
}

// Healthcheck returns an healthcheck with the given name.
func (r *Reporter) Healthcheck(name string, f func(h Healthcheck)) Healthcheck {
	check := func(h metrics.Healthcheck) {
		f(Healthcheck{h})
	}
	return Healthcheck{metrics.GetOrRegisterHealthcheck(
		expandName(name, r.prefix), r.metrics.Registry, check)}
}

const separator = "."

// ExpandName a metric name to include the module name, unless there is
// already a dot.
func expandName(name, prefix string) string {
	if name[:1] == "." {
		return name[1:]
	}

	callStack := stack.Trace()

	// Find the first module (but try to skip project.reporter)
	var best string
	for _, call := range callStack {
		//nolint
		module := fmt.Sprintf("%+n", call)
		if !strings.HasPrefix(module, prefix) {
			continue
		}
		module = module[len(prefix):]
		if len(module) == 0 || module[0] != '/' {
			continue
		}
		module = module[1:]
		if strings.HasPrefix(module, "vendor/") {
			continue
		}
		module = strings.SplitN(module, separator, 2)[0]
		module = strings.Replace(module, "/", separator, -1)
		best = strings.Join([]string{module, name}, separator)
		if module != "reporter" {
			return best
		}
		continue
	}

	if best != "" {
		return best
	}
	return name
}
