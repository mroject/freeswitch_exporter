package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// parseThresholds parses a comma-separated list of channel-age thresholds (seconds).
// Values must be > 0 and strictly increasing.
func parseThresholds(s string) ([]float64, error) {
	parts := strings.Split(s, ",")

	thresholds := make([]float64, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)

		value, err := strconv.ParseFloat(trimmed, 64)

		if err != nil {
			return nil, fmt.Errorf("invalid threshold value %q: not a number", trimmed)
		}

		if value <= 0 {
			return nil, fmt.Errorf("invalid threshold value %q: must be > 0", trimmed)
		}

		if len(thresholds) > 0 && value <= thresholds[len(thresholds)-1] {
			return nil, fmt.Errorf("invalid threshold value %q: thresholds must be strictly increasing", trimmed)
		}

		thresholds = append(thresholds, value)
	}

	return thresholds, nil
}

// parseCreatedEpoch parses a FreeSWITCH created_epoch string (epoch seconds).
// Returns an error if the value is not a positive decimal integer.
func parseCreatedEpoch(s string) (time.Time, error) {
	v, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)

	if err != nil {
		return time.Time{}, fmt.Errorf("invalid created_epoch %q: not a decimal integer", s)
	}

	if v <= 0 {
		return time.Time{}, fmt.Errorf("invalid created_epoch %q: must be > 0", s)
	}

	return time.Unix(v, 0), nil
}

// countChannelsByDuration counts, for each threshold, the number of active channels
// whose age (now - created_epoch) is >= that threshold. Counts are independent per
// threshold (a channel older than multiple thresholds increments each). Skips rows
// with missing/invalid created_epoch and clamps negative ages to 0.
func countChannelsByDuration(rows []map[string]any, now time.Time, thresholds []float64) ([]float64, int) {
	counts := make([]float64, len(thresholds))
	skipped := 0

	for _, row := range rows {
		raw, ok := row["created_epoch"].(string)

		if !ok {
			skipped++
			continue
		}

		start, err := parseCreatedEpoch(raw)

		if err != nil {
			skipped++
			continue
		}

		age := now.Sub(start).Seconds()

		if age < 0 {
			age = 0
		}

		for i, threshold := range thresholds {
			if age >= threshold {
				counts[i]++
			}
		}
	}

	return counts, skipped
}

// channelRowsPayload mirrors the JSON shape of "api show channels as json".
type channelRowsPayload struct {
	RowCount int              `json:"row_count"`
	Rows     []map[string]any `json:"rows"`
}

// decodeChannelRows decodes the channel rows JSON payload into a slice of maps.
func decodeChannelRows(payload []byte) ([]map[string]any, error) {
	var p channelRowsPayload

	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("cannot decode channel rows JSON: %w", err)
	}

	return p.Rows, nil
}

var channelDurationDesc = prometheus.NewDesc(
	namespace+"_current_channels_by_duration",
	"Number of currently active FreeSWITCH channels whose age is >= threshold_seconds, observed at scrape time.",
	[]string{"threshold_seconds"},
	nil,
)

// channelDurationMetrics fetches channels, counts them per age threshold, and emits
// one gauge sample per threshold.
func (c *Collector) channelDurationMetrics(ch chan<- prometheus.Metric) error {
	response, err := c.fsCommand("api show channels as json")

	if err != nil {
		return err
	}

	rows, err := decodeChannelRows(response)

	if err != nil {
		return err
	}

	counts, skipped := countChannelsByDuration(rows, time.Now(), c.channelDurationThresholds)

	if skipped > 0 {
		level.Debug(c.logger).Log("msg", "skipped unparseable channel row", "count", skipped)
	}

	for i, threshold := range c.channelDurationThresholds {
		metric, err := prometheus.NewConstMetric(
			channelDurationDesc,
			prometheus.GaugeValue,
			counts[i],
			strconv.FormatInt(int64(threshold), 10),
		)

		if err != nil {
			return err
		}

		ch <- metric
	}

	return nil
}
