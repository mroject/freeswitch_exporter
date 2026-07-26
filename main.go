package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/exporter-toolkit/web"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		logLevel = kingpin.Flag(
			"log.level",
			"Log level for structured logging (debug, info, warn, error).",
		).Default("debug").Enum("debug", "info", "warn", "error")
		listenAddress = kingpin.Flag(
			"web.listen-address",
			"Address to listen on for web interface and telemetry.").Short('l').Default(":9282").String()
		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.").Default("/metrics").String()
		scrapeURI = kingpin.Flag(
			"freeswitch.scrape-uri",
			`URI on which to scrape freeswitch. E.g. "tcp://localhost:8021"`).Short('u').Default("tcp://localhost:8021").String()
		timeout = kingpin.Flag(
			"freeswitch.timeout",
			"Timeout for trying to get stats from freeswitch.").Short('t').Default("5s").Duration()
		password = kingpin.Flag(
			"freeswitch.password",
			"Password for freeswitch event socket.").Short('P').Default("ClueCon").String()
		configFile = kingpin.Flag(
			"web.config",
			"[EXPERIMENTAL] Path to config yaml file that can enable TLS or authentication.",
		).Default("").String()
		rtpEnable              = kingpin.Flag("rtp.enable", "enable rtp info(feature:todo!), default: fasle").Default("false").Bool()
		channelDurationDisable = kingpin.Flag(
			"freeswitch.channel-duration.disable",
			"Disable the freeswitch_current_channels_by_duration gauge of active channel counts by age threshold.").Default("false").Bool()
		channelDurationThresholdsFlag = kingpin.Flag(
			"freeswitch.channel-duration.thresholds",
			"Comma-separated, strictly increasing channel-age thresholds in seconds for freeswitch_current_channels_by_duration.").Default("30,60,120,300,600,900,1800,3600,7200,14400,21600,43200,86400,172800").String()
	)
	kingpin.Version("freeswitch_exporter\nversion: 1.0.6")
	kingpin.Parse()

	promlogConfig := &promlog.Config{}
	promlogConfig.Level = &promlog.AllowedLevel{}   // ensure level flag has a target
	promlogConfig.Format = &promlog.AllowedFormat{} // keep format configurable
	if err := promlogConfig.Level.Set(*logLevel); err != nil {
		panic(fmt.Sprintf("invalid log level %q: %v", *logLevel, err))
	}
	logger := promlog.New(promlogConfig)

	channelDurationThresholds, err := parseThresholds(*channelDurationThresholdsFlag)

	if err != nil {
		panic(fmt.Sprintf("invalid --freeswitch.channel-duration.thresholds: %v", err))
	}

	c, err := NewCollector(*scrapeURI, *timeout, *password, *rtpEnable, !*channelDurationDisable, channelDurationThresholds, logger)

	if err != nil {
		panic(err)
	}

	prometheus.MustRegister(c)

	http.Handle(*metricsPath, promhttp.Handler())

	// This implements Prometheus' multi-target exporter support
	// Example project: Official Blackbox Exporter
	// https://github.com/prometheus/blackbox_exporter#prometheus-configuration
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "'target' query param not provided, but required.", http.StatusBadRequest)

		}

		// Not checking for the port to allow the port to be configured in
		// the Prometheus scrape target config.
		if !strings.HasPrefix(target, "tcp://") {
			target = fmt.Sprintf("tcp://%s", target)
		}

		c, colErr := NewCollector(target, *timeout, *password, *rtpEnable, !*channelDurationDisable, channelDurationThresholds, logger)
		if colErr != nil {
			http.Error(w, fmt.Sprintf("failed to create collector for %s: %s", target, colErr), http.StatusInternalServerError)
		}

		registry := prometheus.NewRegistry()
		registry.MustRegister(c)

		promhttp.HandlerFor(registry, promhttp.HandlerOpts{}).ServeHTTP(w, r)

	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>FreeSWITCH Exporter</title></head>
			<body>
			<h1>FreeSWITCH Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	server := &http.Server{Addr: *listenAddress}
	flags := &web.FlagConfig{
		WebListenAddresses: &[]string{*listenAddress},
		WebSystemdSocket:   new(bool),
		WebConfigFile:      configFile,
	}
	if err := web.ListenAndServe(server, flags, logger); err != nil {
		level.Info(logger).Log("err", err)
		os.Exit(1)
	}
}
