package main

import (
	"net/http"
	"os"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/exporter-toolkit/web"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
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
		rtpEnable = kingpin.Flag("rtp.enable", "enable rtp info, default: fasle").Default("false").Bool()
	)
	promlogConfig := &promlog.Config{}
	kingpin.Version("freeswitch_exporter")
	logger := promlog.New(promlogConfig)
	kingpin.Parse()

	c, err := NewCollector(*scrapeURI, *timeout, *password, *rtpEnable)

	if err != nil {
		panic(err)
	}

	prometheus.MustRegister(c)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>FreeSWITCH Exporter</title></head>
			<body>
			<h1>FreeSWITCH Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	level.Info(logger).Log("msg", "Listening on", "address", *listenAddress)
	server := &http.Server{Addr: *listenAddress}
	if err := web.ListenAndServe(server, *configFile, logger); err != nil {
		level.Info(logger).Log("err", err)
		os.Exit(1)
	}
}
