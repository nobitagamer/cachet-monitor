package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"

	cachet "github.com/Soontao/cachet-monitor"
	"github.com/urfave/cli"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Version string, in release version
// This variable will be overwrite by complier
var Version = "SNAPSHOT"

// AppName of this application
var AppName = "Cachet Monitor"

// AppUsage of this application
var AppUsage = "A Command Line Tool for Cachet Monitor"

func main() {

	app := cli.NewApp()

	app.Version = Version
	app.Name = AppName
	app.Usage = AppUsage
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config,c",
			EnvVar: "CONFIG_FILE",
			Value:  "./config.json",
			Usage:  "Path to configuration file",
		},
		cli.StringFlag{
			Name:   "server,s",
			EnvVar: "CACHET_SERVER_API",
			Usage:  "The server api uri, with schema & path",
		},
		cli.StringFlag{
			Name:   "token,t",
			EnvVar: "CACHET_USER_TOKEN",
			Usage:  "The server user token",
		},
		cli.StringFlag{
			Name:   "auto,a",
			EnvVar: "AUTO_CONFIG",
			Usage:  "Auto configuration",
		},
		cli.StringFlag{
			Name:   "log,l",
			EnvVar: "LOG_FILE",
			Usage:  "Path to log file",
		},
		cli.StringFlag{
			Name:   "name,n",
			EnvVar: "SYSTEM_NAME",
			Usage:  "System name",
		},
	}

	app.Action = appAction

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

func appAction(c *cli.Context) (err error) {

	autoConfig := c.GlobalBool("auto")
	configPath := c.GlobalString("config")
	systemName := c.GlobalString("name")
	logFile := c.GlobalString("log")

	var cfg *cachet.CachetMonitor

	if autoConfig {
		logrus.Println("Start auto configuration")
		api := cachet.CachetAPI{
			URL:   c.GlobalString("server"),
			Token: c.GlobalString("token"),
		}
		if cfg, err = api.GetConfigurationFromRemote(); err != nil {
			logrus.Panicf("Unable to start (fetch config): %v", err)
		}
	} else {
		logrus.Println("Start file based configuration")
		if cfg, err = getConfiguration(configPath); err != nil {
			logrus.Panicf("Unable to start (reading config): %v", err)
		}
	}

	if cfg, err = setupMonitors(cfg); err != nil {
		logrus.Panicf("Unable to setup config: %v", err)
	}

	cfg.Immediate = true

	if systemName != "" {
		cfg.SystemName = systemName
	}

	logrus.SetOutput(getLogger(logFile))

	if valid := cfg.Validate(); !valid {
		logrus.Errorf("Invalid configuration")
		os.Exit(1)
	}

	logrus.Debug("Configuration valid")
	logrus.Infof("System: %s", cfg.SystemName)
	logrus.Infof("API: %s", cfg.API.URL)
	logrus.Infof("Monitors: %d\n", len(cfg.Monitors))

	logrus.Infof("Pinging cachet")

	if err := cfg.API.Ping(); err != nil {
		logrus.Errorf("Cannot ping cachet!\n%v", err)
		os.Exit(1)
	}

	logrus.Infof("Ping OK")

	wg := &sync.WaitGroup{}

	for index, monitor := range cfg.Monitors {
		logrus.Infof("Starting Monitor #%d: ", index)
		logrus.Infof("Features: \n - %v", strings.Join(monitor.Describe(), "\n - "))

		go monitor.ClockStart(cfg, monitor, wg)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	<-signals

	logrus.Warnf("Abort: Waiting monitors to finish")

	for _, mon := range cfg.Monitors {
		mon.GetMonitor().ClockStop()
	}

	wg.Wait()

	return err
}

func getLogger(logPath interface{}) *os.File {
	if logPath == nil || len(logPath.(string)) == 0 {
		return os.Stdout
	}

	file, err := os.Create(logPath.(string))
	if err != nil {
		logrus.Errorf("Unable to open file '%v' for logging: \n%v", logPath, err)
		os.Exit(1)
	}

	return file
}

func setupMonitors(cfg *cachet.CachetMonitor) (*cachet.CachetMonitor, error) {
	cfg.Monitors = make([]cachet.MonitorInterface, len(cfg.RawMonitors))

	for index, rawMonitor := range cfg.RawMonitors {
		var t cachet.MonitorInterface
		var err error

		monType := "unknown"

		if t, ok := rawMonitor["type"].(string); ok {
			monType = cachet.GetMonitorType(t)
		}

		switch monType {
		case "http", "https":
			var s cachet.HTTPMonitor
			err = mapstructure.Decode(rawMonitor, &s)
			t = &s
		case "dns":
			var s cachet.DNSMonitor
			err = mapstructure.Decode(rawMonitor, &s)
			t = &s
		case "tcp":
			var s cachet.TCPMonitor
			err = mapstructure.Decode(rawMonitor, &s)
			t = &s
		default:
			logrus.Errorf("Invalid monitor type (index: %d) %v", index, monType)
			continue
		}

		// >> config default incident
		defaultIncident := &cachet.Incident{
			Name:        t.GetMonitor().Name,
			ComponentID: t.GetMonitor().ComponentID,
			Notify:      true,
		}

		componentStatus, err := cfg.API.GetComponentStatus(defaultIncident.ComponentID)

		defaultIncident.ComponentStatus = componentStatus

		// not operational
		if defaultIncident.ComponentStatus > 1 {
			logrus.Infof("%v start with incident.", t.GetMonitor().Name)
			t.GetMonitor().SetDefaultIncident(defaultIncident)
		}

		t.GetMonitor().Type = monType
		// << config default incident

		if err != nil {
			logrus.Errorf("Unable to unmarshal monitor to type (index: %d): %v", index, err)
			continue
		}

		cfg.Monitors[index] = t
	}

	return cfg, nil
}

func getConfiguration(path string) (*cachet.CachetMonitor, error) {
	cfg := &cachet.CachetMonitor{}
	var data []byte

	// test if its a url
	url, err := url.ParseRequestURI(path)
	if err == nil && len(url.Scheme) > 0 {
		// download config
		response, err := http.Get(path)
		if err != nil {
			logrus.Warn("Unable to download network configuration")
			return nil, err
		}

		defer response.Body.Close()
		data, _ = ioutil.ReadAll(response.Body)

		logrus.Info("Downloaded network configuration.")
	} else {
		data, err = ioutil.ReadFile(path)
		if err != nil {
			return nil, errors.New("Unable to open file: '" + path + "'")
		}
	}

	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		err = yaml.Unmarshal(data, cfg)
	} else {
		err = json.Unmarshal(data, cfg)
	}

	if err != nil {
		logrus.Warnf("Unable to parse configuration file")
	}

	return cfg, err
}
