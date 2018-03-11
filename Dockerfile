FROM golang:1.9 AS builder
WORKDIR /go/src/github.com/carlpett/influxdb_stats_exporter/
COPY . .
RUN make build

FROM busybox:glibc
EXPOSE 9424
USER nobody

COPY --from=builder /go/src/github.com/carlpett/influxdb_stats_exporter/influxdb_stats_exporter /influxdb_stats_exporter

ENTRYPOINT ["/influxdb_stats_exporter"]
