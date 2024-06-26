package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	versioncollector "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

const app = "freeswitch_exporter"

var totalScrapes = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Name:      "exporter_total_scrapes",
	Help:      "Current total freeswitch scrapes.",
}, []string{"status"})

func init() {
	prometheus.MustRegister(versioncollector.NewCollector(app))
	prometheus.MustRegister(totalScrapes)
}

func main() {
	os.Exit(run())
}

func newLandingPage(metricsPath, healthzPath, probePath string, probeEnable bool) (http.Handler, error) {
	landingConfig := web.LandingConfig{
		Name:        app,
		Description: "exporter for freeswitch",
		Version:     version.Info(),
		Links: []web.LandingLinks{
			{
				Address:     metricsPath,
				Text:        "Metrics",
				Description: "for self-metrics or running in single target mode",
			},
			{
				Address:     healthzPath,
				Text:        "Healthz",
				Description: "for liveness or readiness probe",
			},
		},
	}
	if probeEnable {
		landingConfig.Links = append(landingConfig.Links, web.LandingLinks{
			Address:     probePath,
			Text:        "Probe",
			Description: "for probe handler, currently supported parameters are [target or (host and port), password]",
		})
	}
	return web.NewLandingPage(landingConfig)
}

func run() int {
	var (
		toolkitFlags = webflag.AddFlags(kingpin.CommandLine, ":9282")

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
		rtpEnable   = kingpin.Flag("rtp.enable", "enable rtp info(feature:todo!), default: false").Default("false").Bool()
		probeEnable = kingpin.Flag("probe.enable", "enable probe handler /probe").Default("false").Bool()
	)
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print(app))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	if *probeEnable {
		http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
			probeHandler(w, r, logger, *timeout, nil)
		})
	} else {
		c, err := NewCollector(*scrapeURI, *timeout, *password, *rtpEnable, logger)
		if err != nil {
			level.Error(logger).Log("msg", "error creating collector", "err", err)
			return 1
		}
		prometheus.MustRegister(c)
	}

	http.Handle(*metricsPath, promhttp.Handler())

	healthzPath := "/-/healthy"
	http.HandleFunc(healthzPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Healthy"))
	})

	landingPage, err := newLandingPage(*metricsPath, healthzPath, "/probe", *probeEnable)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	http.Handle("/", landingPage)

	srv := &http.Server{}
	srvc := make(chan struct{})
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := web.ListenAndServe(srv, toolkitFlags, logger); err != nil {
			level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
			close(srvc)
		}
	}()

	for {
		select {
		case <-term:
			level.Info(logger).Log("msg", "Received SIGTERM, exiting gracefully...")
			return 0
		case <-srvc:
			return 1
		}
	}
}
