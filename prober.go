package main

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func probeHandler(w http.ResponseWriter, r *http.Request, logger log.Logger, timeout time.Duration, params url.Values) {
	if params == nil {
		params = r.URL.Query()
	}
	target, err := getTarget(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	timeoutSeconds, err := getTimeout(r, float64(timeout))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	password := params.Get("password")
	rtpEnable, _ := strconv.ParseBool(params.Get("rtp_enable"))

	scrapeLogger := log.With(logger, "target", target)
	col, err := NewCollector(target, time.Duration(float64(time.Second)*(timeoutSeconds)), password, rtpEnable, scrapeLogger)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	registry := prometheus.NewRegistry()
	registry.MustRegister(col)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func getTarget(q url.Values) (string, error) {
	if v := q.Get("target"); v != "" {
		return v, nil
	}
	if host, port := q.Get("host"), q.Get("port"); host != "" && port != "" {
		return "tcp://" + net.JoinHostPort(host, port), nil
	}
	return "", errors.New("target or host/port parameter are missing")
}

func getTimeout(r *http.Request, timeout float64) (timeoutSeconds float64, err error) {
	// If a timeout is configured via the Prometheus header, add it to the request.
	if v := r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"); v != "" {
		var err error
		timeoutSeconds, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, err
		}
	}
	if timeoutSeconds == 0 {
		timeoutSeconds = timeout
	}
	return timeoutSeconds, nil
}
