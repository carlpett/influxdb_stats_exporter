# Influxdb stats Exporter

[![CircleCI](https://circleci.com/gh/carlpett/influxdb_stats_exporter.svg?style=shield)](https://circleci.com/gh/carlpett/influxdb_stats_exporter)
[![DockerHub](https://img.shields.io/docker/build/carlpett/influxdb_stats_exporter.svg?style=shield)](https://hub.docker.com/r/carlpett/influxdb_stats_exporter/)

Prometheus exporter for InfluxDB stats, as reported by a `SHOW STATS` query. Tested to work with InfluxDB 1.4 and 1.5.

Not to be confused with [influxdb_exporter](https://github.com/prometheus/influxdb_exporter/), which accepts the InfluxDB line protocol and converts it to Prometheus format.

# Metrics
The exporter will make a `SHOW STATS` query to InfluxDB, and report all the returned statistics as metrics. The metrics are therefore dependent on the underlying InfluxDB installation. All metrics are prefixed with `influxdb`.

Two meta-metrics are added by the exporter, `influxdb_exporter_stats_query_duration_seconds` and `influxdb_exporter_stats_query_success`. `influxdb_exporter_stats_query_duration_seconds` is a gauge showing the number of seconds it took to get a query response from InfluxDB. `influxdb_exporter_stats_query_success` is `1` if a valid response was received, and `0` if there was an error.

As of InfluxDB 1.5, there is a Prometheus `/metrics` endpoint on InfluxDB itself. However, this does not yet return any of the statistics about InfluxDB, only the Golang process-level metrics. In future developments of InfluxDB, this exporter may (hopefully) become obsolete.

# Usage
`./influxdb_stats_exporter` launches the exporter with all default options: Querying a local Influx server, without any authentication, and serving metrics on port 9424. If you need to alter these options, see the section on flags and environment variables below.

## Flags and environment variables
`influxdb_stats_exporter` has a number of flags to set different options, some of which can also be set with environment variables:

Name     | Description | Default value | Environment variable name
---------|-------------|---------------|--------------------------
`--influx.url` | Url to InfluxDB | `http://localhost:8086` | `INFLUX_URL`
`--influx.user` | Username for InfluxDB | _(Not set)_ | `INFLUX_USER`
`--influx.password` | Password for InfluxDB | _(Not set)_ | `INFLUX_PASSWORD`
`--ssl.skip-verify` | Skip HTTPS certificate verification | `false` | -
`--log.level` | Log level for console output | `info` | -
`--web.listen-address` | Address on which to expose metrics | `:9424` | -
`--web.metrics-path` | Path under which the metrics are available | `/metrics` | -

