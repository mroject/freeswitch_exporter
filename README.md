# FreeSWITCH Exporter for Prometheus

[![Go Report Card](https://goreportcard.com/badge/github.com/mroject/freeswitch_exporter)](https://goreportcard.com/report/github.com/mroject/freeswitch_exporter)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/mroject/freeswitch_exporter/LICENSE)

A [FreeSWITCH](https://freeswitch.org/confluence/display/FREESWITCH/FreeSWITCH+Explained) exporter for Prometheus.

It communicates with FreeSWITCH using [mod_event_socket](https://freeswitch.org/confluence/display/FREESWITCH/mod_event_socket).


Add metrics as below:

1. `sofia gateway status`
2. `module status`
3. `bridged_calls`
4. `current_channels`
5. `detailed_calls`
6. `endpoint`
7. `codec`
8. `api memory`
9. `registration details`

Add feature:

1. `web.config` support tls, authorization and etc. 

configuration exporter web.config visit: https://prometheus.io/docs/guides/basic-auth/

## Getting Started

Pre-built static binaries are available in [releases](https://github.com/mroject/freeswitch_exporter/releases).



To run it:
```bash
./freeswitch_exporter [flags]
```

Help on flags:
```
./freeswitch_exporter --help
usage: freeswitch_exporter [<flags>]

Flags:
      --help                   Show context-sensitive help.
  -l, --web.listen-address=":9282"
                               Address to listen on for web interface and telemetry.
      --web.telemetry-path="/metrics"
                               Path under which to expose metrics.
  -u, --freeswitch.scrape-uri="tcp://localhost:8021"
                               URI on which to scrape freeswitch. E.g.
                               "tcp://localhost:8021"
  -t, --freeswitch.timeout=5s  Timeout for trying to get stats from freeswitch.
  -P, --freeswitch.password="ClueCon"
                               Password for freeswitch event socket.
      --web.config=""          [EXPERIMENTAL] Path to config yaml file that can
                               enable TLS or authentication.
      --version                Show application version.
```

## Usage

Make sure [mod_event_socket](https://freeswitch.org/confluence/display/FREESWITCH/mod_event_socket) is enabled on your FreeSWITCH instance. The default mod_event_socket configuration binds to `::` (i.e., to listen to connections from any host), which will work on IPv4 or IPv6. 

You can specify the scrape URI with the `--freeswitch.scrape-uri` flag. Example:

```
./freeswitch_exporter -u "tcp://localhost:8021"
```

Also, you need to make sure that the exporter will be allowed by the ACL (if any), and that the password matches.

### Multi-Target Exporter

This exporter also supports the multi-target pattern. This means that instead of configuring Prometheus to hit this
exporters `/metrics` endpoint, you would configure it to hit `/probe` instead, like the example Prometheus config below:

```yaml
scrape_configs:
  - job_name: freeswitch
    scrape_interval: 15s
    static_configs:
      - targets:
        - 10.0.0.1:8021 # Note you do need to include the port
        - tcp://10.0.0.2:8021 # tcp:// is injected for you if the value does not start with it
    metrics_path: /probe
    relabel_configs: # This relabeling sets the instance to the target address instead of the address of the exporter
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9282
```

What this does is tell's Prometheus to do the following for each item in `static_configs.targets`:

```shell
curl http://localhost:9282/probe?target=10.0.0.1:8021
```

Then the exporter does what it would normally do and scrapes the target for metrics.

Using this method can help make it a bit easier depending on the size of the platform your monitoring. For example,
this could allow you to utilize Prometheus' discovery plugins.

## Metrics

The exporter will try to fetch values from the following commands:

- `api show calls count`: Calls count
- `api uptime s`: Uptime
- `api strepoch`: Time synced with system
- `status`
- `sofia xmlstatus gateway`: fetch all gateway
- `module`: usage module.conf.xml fetch all module status
- `api show endpoint` all used endpoint
- `api show codec` all used codec
- `registration` all sofia registration details
- `api memory` get freeswitch memory info 

List of exposed metrics:

```bash
# HELP freeswitch_bridged_calls Number of bridged_calls active
# TYPE freeswitch_bridged_calls gauge
# HELP freeswitch_current_calls Number of calls active
# TYPE freeswitch_current_calls gauge
# HELP freeswitch_current_channels Number of channels active
# TYPE freeswitch_current_channels gauge
# HELP freeswitch_current_idle_cpu CPU idle
# TYPE freeswitch_current_idle_cpu gauge
# HELP freeswitch_current_sessions Number of sessions active
# TYPE freeswitch_current_sessions gauge
# HELP freeswitch_current_sessions_peak Peak sessions since startup
# TYPE freeswitch_current_sessions_peak gauge
# HELP freeswitch_current_sessions_peak_last_5min Peak sessions for the last 5 minutes
# TYPE freeswitch_current_sessions_peak_last_5min gauge
# HELP freeswitch_current_sps Number of sessions per second
# TYPE freeswitch_current_sps gauge
# HELP freeswitch_current_sps_peak Peak sessions per second since startup
# TYPE freeswitch_current_sps_peak gauge
# HELP freeswitch_current_sps_peak_last_5min Peak sessions per second for the last 5 minutes
# TYPE freeswitch_current_sps_peak_last_5min gauge
# HELP freeswitch_detailed_bridged_calls Number of detailed_bridged_calls active
# TYPE freeswitch_detailed_bridged_calls gauge
# HELP freeswitch_detailed_calls Number of detailed_calls active
# TYPE freeswitch_detailed_calls gauge
# HELP freeswitch_exporter_failed_scrapes Number of failed freeswitch scrapes.
# TYPE freeswitch_exporter_failed_scrapes counter
# HELP freeswitch_exporter_total_scrapes Current total freeswitch scrapes.
# TYPE freeswitch_exporter_total_scrapes counter
# HELP freeswitch_load_module freeswitch load module status
# TYPE freeswitch_load_module gauge
# HELP freeswitch_max_sessions Max sessions allowed
# TYPE freeswitch_max_sessions gauge
# HELP freeswitch_max_sps Max sessions per second allowed
# TYPE freeswitch_max_sps gauge
# HELP freeswitch_min_idle_cpu Minimum CPU idle
# TYPE freeswitch_min_idle_cpu gauge
# HELP freeswitch_registrations Number of registrations active
# TYPE freeswitch_registrations gauge
# HELP freeswitch_sessions_total Number of sessions since startup
# TYPE freeswitch_sessions_total counter
# HELP freeswitch_sofia_gateway_call_in freeswitch gateway call-in
# TYPE freeswitch_sofia_gateway_call_in gauge
# HELP freeswitch_sofia_gateway_call_out freeswitch gateway call-out
# TYPE freeswitch_sofia_gateway_call_out gauge
# HELP freeswitch_sofia_gateway_failed_call_in freeswitch gateway failed-call-in
# TYPE freeswitch_sofia_gateway_failed_call_in gauge
# HELP freeswitch_sofia_gateway_failed_call_out freeswitch gateway failed-call-out
# TYPE freeswitch_sofia_gateway_failed_call_out gauge
# HELP freeswitch_sofia_gateway_ping freeswitch gateway ping
# TYPE freeswitch_sofia_gateway_ping gauge
# HELP freeswitch_sofia_gateway_pingcount freeswitch gateway pingcount
# TYPE freeswitch_sofia_gateway_pingcount gauge
# HELP freeswitch_sofia_gateway_pingfreq freeswitch gateway pingfreq
# TYPE freeswitch_sofia_gateway_pingfreq gauge
# HELP freeswitch_sofia_gateway_pingmax freeswitch gateway pingmax
# TYPE freeswitch_sofia_gateway_pingmax gauge
# HELP freeswitch_sofia_gateway_pingmin freeswitch gateway pingmin
# TYPE freeswitch_sofia_gateway_pingmin gauge
# HELP freeswitch_sofia_gateway_pingtime freeswitch gateway pingtime
# TYPE freeswitch_sofia_gateway_pingtime gauge
# HELP freeswitch_sofia_gateway_status freeswitch gateways status
# TYPE freeswitch_sofia_gateway_status gauge
# HELP freeswitch_time_synced Is FreeSWITCH time in sync with exporter host time
# TYPE freeswitch_time_synced gauge
# HELP freeswitch_up Was the last scrape successful.
# TYPE freeswitch_up gauge
# HELP freeswitch_uptime_seconds Uptime in seconds
# TYPE freeswitch_uptime_seconds gauge
# HELP freeswitch_endpoint_status freeswitch endpoint status
# TYPE freeswitch_endpoint_status gauge
# HELP freeswitch_codec_status freeswitch endpoint status
# TYPE freeswitch_codec_status gauge
# HELP freeswitch_memory_arena Total non-mmapped bytes
# TYPE freeswitch_memory_arena gauge
# HELP freeswitch_memory_fordblks Total free space
# TYPE freeswitch_memory_fordblks gauge
# HELP freeswitch_memory_fsmblks Free bytes held in fastbins
# TYPE freeswitch_memory_fsmblks gauge
# HELP freeswitch_memory_hblkhd Bytes in mapped regions
# TYPE freeswitch_memory_hblkhd gauge
# HELP freeswitch_memory_hblks # of mapped regions
# TYPE freeswitch_memory_hblks gauge
# HELP freeswitch_memory_keepcost Topmost releasable block
# TYPE freeswitch_memory_keepcost gauge
# HELP freeswitch_memory_ordblks # of free chunks
# TYPE freeswitch_memory_ordblks gauge
# HELP freeswitch_memory_smblks # of free fastbin blocks
# TYPE freeswitch_memory_smblks gauge
# HELP freeswitch_memory_uordblks Total allocated space
# TYPE freeswitch_memory_uordblks gauge
# HELP freeswitch_memory_usmblks Max. total allocated space
# TYPE freeswitch_memory_usmblks gauge
```

## Compiling

With go 1.18+, clone the project and:

```bash
go build
```

Dependencies will be fetched automatically.

## Contributing

Feel free to send pull requests.

Copyright (c) 2022 Zhang Lian Jun <z0413j@outlook.com>

