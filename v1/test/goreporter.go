package main

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/thecodeteam/goodbye"

	reporter "github.com/exoscale/go-reporter/v1"
	"github.com/exoscale/go-reporter/v1/errors"
	"github.com/exoscale/go-reporter/v1/logging"
)

var r *reporter.Reporter

func init() {
	var err error

	if r, err = reporter.New(&reporter.Config{
		// Metrics: &metrics.Config{
		// 	WithRuntimeMetrics: true,
		// 	Prometheus: &prometheus.Config{
		// 		Listen:        ":9090",
		// 		FlushInterval: 2,
		// 	},
		// },
		Logging: &logging.Config{
			Destinations: []*logging.LogDestinationConfig{
				{Type: "console"},
			},
			ReportErrors: true,
		},
		Errors: &errors.Config{
			DSN:  "https://3b2a58e948de46a1a30671560b66dc7f:72097d3692c449bc8e32a6724f230704@errors.exoscale.ch/56",
			Wait: true,
		},
		Debug: true,
	}); err != nil {
		panic(err)
	}
}

func main() {
	ctx := context.Background()
	defer goodbye.Exit(ctx, 0)
	goodbye.Notify(ctx, syscall.SIGINT, syscall.SIGTERM)

	goodbye.Register(func(ctx context.Context, _ os.Signal) {
		_ = r.Stop(ctx)
	})

	if err := r.Start(context.Background()); err != nil {
		panic(err)
	}

	defer r.Errors.PanicHandler(func(r interface{}) {
		fmt.Fprintf(os.Stderr, "process crashed! %v\n", r)
		os.Exit(1)
	})

	doSomething()

	// gm := gometrics.NewGauge()
	// if err := r.Metrics.Register("gometric", gm); err != nil {
	// 	panic(err)
	// }
	// gm.Update(42)
	//
	// pm := prom.NewGauge(prom.GaugeOpts{
	// 	Name: "prommetric",
	// 	Help: "Native Prometheus metric",
	// })
	// r.Metrics.Prometheus.MustRegister(pm)
	// pm.Set(42)
	//
	// http.Handle("/metrics", r.Metrics.Prometheus.HTTPHandler())
	// log.Fatal(http.ListenAndServe(":9090", nil))
}

func doSomething() {
	panic("panic! at the Disco")
	// r.Error("o noes!",
	// 	"why", "because!",
	// 	"no_seriously", 42,
	// 	"err", goerrors.New("fffffuuuu"))
}
