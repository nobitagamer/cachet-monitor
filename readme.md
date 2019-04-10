# Enhanced Cachet Monitor

[![CircleCI](https://circleci.com/gh/Soontao/cachet-monitor.svg?style=shield)](https://circleci.com/gh/Soontao/cachet-monitor)
[![codecov](https://codecov.io/gh/Soontao/cachet-monitor/branch/master/graph/badge.svg)](https://codecov.io/gh/Soontao/cachet-monitor)

## Features

- [x] Creates & Resolves Incidents
- [x] Posts monitor lag to cachet graphs
- [x] HTTP Checks (body/status code)
- [x] DNS Checks
- [x] Updates Component to Partial Outage
- [x] Updates Component to Major Outage if already in Partial Outage (works with distributed monitors)
- [x] Can be run on multiple servers and geo regions
- [x] **NEW** TCP Checks
- [ ] **NEW** SAP Cloud Application Status Checks
- [x] **NEW** Configuration schema file

## Example Configuration

**Note:** configuration can be in json or yaml format. [`example.config.json`](./example.config.json), [`example.config.yaml`](./example.config.yml) files.

## Installation

1. Download binary from [release page](https://github.com/CastawayLabs/cachet-monitor/releases)
2. Add the binary to an executable path (/usr/bin, etc.)
3. Create a configuration following provided examples
4. `cachet-monitor -c /etc/cachet-monitor.yaml`

pro tip: run in background using `nohup cachet-monitor 2>&1 > /var/log/cachet-monitor.log &`, or use a tmux/screen session

```
Usage:
  cachet-monitor (-c PATH | --config PATH) [--log=LOGPATH] [--name=NAME] [--immediate]
  cachet-monitor -h | --help | --version

Arguments:
  PATH     path to config.json
  LOGPATH  path to log output (defaults to STDOUT)
  NAME     name of this logger

Examples:
  cachet-monitor -c /root/cachet-monitor.json
  cachet-monitor -c /root/cachet-monitor.json --log=/var/log/cachet-monitor.log --name="development machine"

Options:
  -c PATH.json --config PATH     Path to configuration file
  -h --help                      Show this screen.
  --version                      Show version
  --immediate                    Tick immediately (by default waits for first defined interval)
  
Environment varaibles:
  CACHET_API      override API url from configuration
  CACHET_TOKEN    override API token from configuration
  CACHET_DEV      set to enable dev logging
```

## Init script

If your system is running systemd (like Debian, Ubuntu 16.04, Fedora, RHEL7, or Archlinux) you can use the provided example file: [example.cachet-monitor.service](https://github.com/CastawayLabs/cachet-monitor/blob/master/example.cachet-monitor.service).

1. Simply put it in the right place with `cp example.cachet-monitor.service /etc/systemd/system/cachet-monitor.service`
2. Then do a `systemctl daemon-reload` in your terminal to update Systemd configuration
3. Finally you can start cachet-monitor on every startup with `systemctl enable cachet-monitor.service`! 👍

## Templates

This package makes use of [`text/template`](https://godoc.org/text/template). [Default HTTP template](https://github.com/CastawayLabs/cachet-monitor/blob/master/http.go#L14)

The following variables are available:

| Root objects  | Description                         |
| ------------- | ------------------------------------|
| `.SystemName` | system name                         | 
| `.API`        | `api` object from configuration     |
| `.Monitor`    | `monitor` object from configuration |
| `.now`        | formatted date string               |

| Monitor variables  |
| ------------------ |
| `.Name`            |
| `.Target`          |
| `.Type`            |
| `.Strict`          |
| `.MetricID`        |
| ...                |

All monitor variables are available from `monitor.go`

## Vision and goals

We made this tool because we felt the need to have our own monitoring software (leveraging on Cachet).
The idea is a stateless program which collects data and pushes it to a central cachet instance.

This gives us power to have an army of geographically distributed loggers and reveal issues in both latency & downtime on client websites.

## Package usage

When using `cachet-monitor` as a package in another program, you should follow what `cli/main.go` does. It is important to call `Validate` on `CachetMonitor` and all the monitors inside.

[API Documentation](https://godoc.org/github.com/Soontao/cachet-monitor)
