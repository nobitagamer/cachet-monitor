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

**Not** support windows platform.

1. Download binary from [release page](https://github.com/Soontao/cachet-monitor/releases)
2. Add the binary to an executable path (/usr/bin, etc.)
3. Create a configuration following provided examples
4. `cachet-monitor -c /etc/cachet-monitor.yaml`

```bash
NAME:
   Cachet Monitor - A Command Line Tool for Cachet Monitor

USAGE:
   cli [global options] command [command options] [arguments...]

VERSION:
   SNAPSHOT

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config value, -c value  Path to configuration file (default: "./config.json") [$CONFIG_FILE]
   --log value, -l value     Path to log file [$LOG_FILE]
   --name value, -n value    System name [$SYSTEM_NAME]
   --help, -h                show help
   --version, -v             print the version
```

## [CHANGELOG](./CHANGELOG.md)

## [LICENSE](./LICENSE)
