module github.com/exoscale/go-reporter

require (
	collectd.org v0.3.0
	github.com/certifi/gocertifi v0.0.0-20190105021004-abcd57078448 // indirect
	github.com/deathowl/go-metrics-prometheus v0.0.0-20190530215645-35bace25558f
	github.com/getsentry/raven-go v0.2.0
	github.com/go-stack/stack v1.8.0
	github.com/kylelemons/godebug v0.0.0-20170820004349-d65d576e9348
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.1
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a
	gopkg.in/inconshreveable/log15.v2 v2.0.0-20180818164646-67afb5ed74ec
	gopkg.in/tomb.v2 v2.0.0-20161208151619-d5d1b5820637
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/rcrowley/go-metrics => github.com/exoscale/go-metrics v0.0.0-20180729161012-6a0b1c6c28ec

go 1.13
