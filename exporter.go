package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/serenize/snaker"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type collector struct {
	client influx.Client
	query  influx.Query
}

var (
	queryDuration = prometheus.NewDesc(
		prometheus.BuildFQName("influxdb", "exporter", "stats_query_duration_seconds"),
		"Duration of SHOW STATS query",
		nil,
		nil,
	)
	querySuccess = prometheus.NewDesc(
		prometheus.BuildFQName("influxdb", "exporter", "stats_query_success"),
		"1 if SHOW STATS query succeeded",
		nil,
		nil,
	)
)

var (
	influxUrl      = kingpin.Flag("influx.url", "Url to InfluxDB").Default("http://localhost:8086").Envar("INFLUX_URL").URL()
	influxUser     = kingpin.Flag("influx.user", "InfluxDB username").Default("").Envar("INFLUX_USER").String()
	influxPassword = kingpin.Flag("influx.password", "InfluxDB password").Default("").Envar("INFLUX_PASSWORD").String()
	sslSkipVerify  = kingpin.Flag("ssl.skip-verify", "Skip HTTPS certificate verification").Default("false").String()
	bindAddr       = kingpin.Flag("web.listen-address", "Address to serve metrics on").Default(":9424").String()
	metricsPath    = kingpin.Flag("web.metrics-path", "Path to serve metrics on").Default("/metrics").String()
	logLevel       = kingpin.Flag("log.level", "Log level").Default(levelString(logrus.InfoLevel)).Enum(levelStrings(logrus.AllLevels)...)
)

func levelString(l logrus.Level) string {
	return l.String()
}
func levelStrings(l []logrus.Level) []string {
	ls := make([]string, len(l))
	for i, level := range l {
		ls[i] = level.String()
	}
	return ls
}

var versionMap = logrus.Fields{
	"version":   version.Version,
	"revision":  version.Revision,
	"branch":    version.Branch,
	"buildUser": version.BuildUser,
	"buildDate": version.BuildDate,
	"goVersion": version.GoVersion,
}

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Version(version.Print("influxdb_stats_exporter"))
	kingpin.Parse()

	// Validity is checked in kingpin
	level, _ := logrus.ParseLevel(*logLevel)
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	logrus.WithFields(versionMap).Info("Starting influxdb_stats_exporter")

	config := buildConfig()
	c := newCollector(config)
	defer func() {
		err := c.client.Close()
		if err != nil {
			logrus.WithError(err).Error("Error closing influx client")
		}
	}()

	prometheus.MustRegister(c)
	prometheus.MustRegister(version.NewCollector("influxdb_stats_exporter"))

	http.Handle(*metricsPath, withLogging(promhttp.Handler()))
	logrus.Infof("Serving Influx metrics on %v%v", *bindAddr, *metricsPath)
	err := http.ListenAndServe(*bindAddr, nil)
	if err != nil {
		logrus.WithError(err).Fatalf("Error serving metrics endpoint")
	}
}

func withLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{
			"userAgent": r.UserAgent(),
			"client":    r.RemoteAddr,
		}).Debugf("Serving metrics request")
		h.ServeHTTP(w, r)
	})
}

func buildConfig() influx.HTTPConfig {
	config := influx.HTTPConfig{
		Addr: (*influxUrl).String(),
	}
	if *influxUser != "" {
		config.Username = *influxUser
	}
	if *influxPassword != "" {
		config.Password = *influxPassword
	}
	if strings.ToLower(*sslSkipVerify) == "true" {
		config.InsecureSkipVerify = true
	}

	return config
}
func newCollector(config influx.HTTPConfig) collector {
	logrus.Infof("Using InfluxDB at %v", *influxUrl)
	client, err := influx.NewHTTPClient(config)
	if err != nil {
		logrus.WithError(err).Panic("Failed to set up influx client")
	}

	return collector{
		client: client,
		query:  influx.NewQuery("SHOW STATS", "", ""),
	}
}
func (c collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("influxdb_exporter", "Bogus desc", []string{}, prometheus.Labels{})
}
func (c collector) Collect(ch chan<- prometheus.Metric) {
	t := time.Now()
	r, err := c.client.Query(c.query)
	ch <- prometheus.MustNewConstMetric(queryDuration, prometheus.GaugeValue, time.Since(t).Seconds())

	if err != nil {
		logrus.WithError(err).Error("SHOW STATS query failed")
		ch <- prometheus.MustNewConstMetric(querySuccess, prometheus.GaugeValue, 0)
		return
	} else if r.Error() != nil {
		logrus.WithError(r.Error()).Error("SHOW STATS query failed")
		ch <- prometheus.MustNewConstMetric(querySuccess, prometheus.GaugeValue, 0)
		return
	}
	ch <- prometheus.MustNewConstMetric(querySuccess, prometheus.GaugeValue, 1)

	for _, res := range r.Results {
		for _, s := range res.Series {
			for idx := 0; idx < len(s.Columns); idx++ {
				seriesName := strings.ToLower(snaker.CamelToSnake(s.Name))
				colName := strings.ToLower(snaker.CamelToSnake(s.Columns[idx]))
				fqName := fmt.Sprintf("influxdb_%s_%s", seriesName, colName)

				desc := prometheus.NewDesc(fqName, colName, []string{}, s.Tags)

				asNum, ok := s.Values[0][idx].(json.Number)
				if !ok {
					logrus.
						WithFields(logrus.Fields{"series": s.Name, "column": colName, "value": s.Values[0][idx]}).
						Warn("Failed to convert value to number")
				}
				val, err := asNum.Float64()
				if err != nil {
					logrus.WithFields(logrus.Fields{"series": s.Name, "column": colName, "value": s.Values[0][idx]}).
						Warn("Failed to convert value to float")
				} else {
					m := prometheus.MustNewConstMetric(desc, prometheus.UntypedValue, val)
					ch <- m
				}
			}
		}
	}
}
