# go-reporter

## Configuration

*go-reporter* is configured through a variable of type `Configuration`.

```yaml
config := reporter.Configuration{...}
r, err := reporter.New(config)
// check err
log.SetOutput(r)
```

It can also be configured through a YAML file.

Currently, due to a technical limitation, if a key is unknown, it will
just be ignored. Be extra-careful that all keys are correctly spelled.

## Yaml

The reporter encompasses the following aspects, each of them having its
own subsection:

 - ``logging``: `Logging`_,
 - ``metrics``: `Metrics`_,
 - ``sentry``: `Crash reporting`_.
 - ``prefix``: `Prefix`_.

### Logging

Here is an example configuration:

```yaml
reporting:
  logging:
    level: info
    console: true
    syslog: false
    format: json
    files:
      - /var/log/project/project.log
      - json:/var/log/project/project.log
```

``level`` specify the maximum log level to use. It can be one of:

 - ``crit``
 - ``error``
 - ``warn``
 - ``info``
 - ``debug``

``console`` enables logging to console while ``syslog`` enables
logging to the local syslog daemon. No configuration knobs are
available for those targets.

`format` allows you to set the log format for `console` and `syslog`. Currently, `json` and `plain` are supported (default to `plain`).

``files`` allows one to set a list of files to log to. It's possible
to prefix the file with the expected format. Currently, only ``json:``
is allowed to get JSON event format.

### Metrics

Metrics can be exported using various output plugins. Here is an example:

```yaml
reporting:
  metrics:
    - expvar:
        listen: 127.0.0.1:8123
```

The support outputs are:

 - ``expvar``
 - ``file``
 - ``collectd``

The ``expvar`` output supports the following key:

 - ``listen`` to specify an HTTP endpoint to listen to (mandatory)

The metrics are then available at the root of the provided HTTP
endpoint. Moreover, when querying ``/healthz``, one can get the status
of all healthcheckers. If one of them is unhealthy, it will return a
542 error, otherwise it will return a 200 status code. The returned
JSON object contains a ``status`` key which can either be ``ok`` or
``fail``. It also contains a ``detail`` key which contains another
object which maps healthcheck names with either ``+ok`` if they are
healthy or the error string prefixed by ``!`` if not.

The ``file`` output supports the following keys:

 - ``path`` to specify a file path (mandatory)
 - ``interval`` to specify an interval (mandatory)

At each tick, the current metric values will be written to the
specified file as a one-line JSON object. For debug purpose, it's
possible to filter the metrics concerning only *project*
by using the following command::

    tailf /var/log/project/metrics \
      | jq 'with_entries(select(.key | startswith("project.")))'

The ``collectd`` output supports the following keys:

 - ``connect`` to specify the target (default to ``127.0.0.1:25826``)
 - ``interval`` to specify an interval (mandatory)

For collectd output to work correctly, you need to append the
following to ``types.db`` file::

     histogram count:COUNTER:0:U, max:GAUGE:U:U, mean:GAUGE:U:U, min:GAUGE:U:U, stddev:GAUGE:0:U, p50:GAUGE:U:U, p75:GAUGE:U:U, p95:GAUGE:U:U, p98:GAUGE:U:U, p99:GAUGE:U:U, p999:GAUGE:U:U
     meter     count:COUNTER:0:U, m1_rate:GAUGE:0:U, m5_rate:GAUGE:0:U, m15_rate:GAUGE:0:U, mean_rate:GAUGE:0:U
     timer     max:GAUGE:U:U, mean:GAUGE:U:U, min:GAUGE:U:U, stddev:GAUGE:0:U, p50:GAUGE:U:U, p75:GAUGE:U:U, p95:GAUGE:U:U, p98:GAUGE:U:U, p99:GAUGE:U:U, p999:GAUGE:U:U

Note that the configuration should be a list of output plugins. An
output plugin is a map from plugin type to its configuration. Only one
item per map is allowed.

Intervals are specified with a number and a unit. For example:

 - ``5s``
 - ``1m``
 - ``30m``

### Crash reporting

Crash reporting is done with Sentry. Here is an example configuration:

```yaml
reporting:
  sentry:
    dsn: https://public:secret@sentry.example.com/project
    tags:
      environment: production
    wait: false
```

Sentry calls are performed asynchronously by default. To wait for the Sentry
API to acknowledge the event, set `wait: true`.

### Prefix

The ``prefix`` of the logs and metrics.

```yaml
reporting:
  prefix: myproject
```
