package cachet

import (
	"fmt"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

// Unit is SECOND

const DefaultInterval = 15
const DefaultTimeout = 10
const DefaultTimeFormat = time.RFC3339
const DefaultThreshold = 3
const HistorySize = 3

type MonitorInterface interface {
	ClockStart(*CachetMonitor, MonitorInterface, *sync.WaitGroup)
	ClockStop()
	tick(MonitorInterface)
	test() bool

	Validate() []string
	GetMonitor() *AbstractMonitor
	Describe() []string
}

// AbstractMonitor data model
type AbstractMonitor struct {
	Name   string
	Target string

	// (default)http / dns
	Type   string
	Strict bool

	Interval time.Duration
	Timeout  time.Duration

	MetricID    int `mapstructure:"metric_id"`
	ComponentID int `mapstructure:"component_id"`

	// Templating stuff
	Template struct {
		Investigating MessageTemplate
		Fixed         MessageTemplate
	}

	// Threshold = percentage / number of down incidents
	Threshold      float32
	ThresholdCount bool `mapstructure:"threshold_count"`

	// lag / average(lagHistory) * 100 = percentage above average lag
	// PerformanceThreshold sets the % limit above which this monitor will trigger degraded-performance
	// PerformanceThreshold float32

	history []bool
	// lagHistory     []float32
	lastFailReason string
	incident       *Incident
	config         *CachetMonitor

	// Closed when mon.Stop() is called
	stopC chan bool
}

// SetDefaultIncident for monitor
func (m *AbstractMonitor) SetDefaultIncident(i *Incident) {
	if m.incident == nil {
		m.incident = i
	}
}

// Validate configuration
func (mon *AbstractMonitor) Validate() []string {
	errs := []string{}

	if len(mon.Name) == 0 {
		errs = append(errs, "Name is required")
	}

	if mon.Interval < 1 {
		mon.Interval = DefaultInterval
	}
	if mon.Timeout < 1 {
		mon.Timeout = DefaultTimeout
	}

	if mon.Timeout > mon.Interval {
		errs = append(errs, "Timeout greater than interval")
	}

	if mon.ComponentID == 0 && mon.MetricID == 0 {
		errs = append(errs, "component_id & metric_id are unset")
	}

	if mon.Threshold <= 0 {
		mon.Threshold = DefaultThreshold
	}

	if err := mon.Template.Fixed.Compile(); err != nil {
		errs = append(errs, "Could not compile \"fixed\" template: "+err.Error())
	}

	if err := mon.Template.Investigating.Compile(); err != nil {
		errs = append(errs, "Could not compile \"investigating\" template: "+err.Error())
	}

	return errs
}

func (mon *AbstractMonitor) GetMonitor() *AbstractMonitor {
	return mon
}

func (mon *AbstractMonitor) Describe() []string {
	features := []string{"Type: " + mon.Type}

	if len(mon.Name) > 0 {
		features = append(features, "Name: "+mon.Name)
	}

	return features
}

func (mon *AbstractMonitor) ClockStart(cfg *CachetMonitor, iface MonitorInterface, wg *sync.WaitGroup) {
	wg.Add(1)
	mon.config = cfg
	mon.stopC = make(chan bool)
	if cfg.Immediate {
		mon.tick(iface)
	}

	ticker := time.NewTicker(mon.Interval * time.Second)

	for {
		select {
		case <-ticker.C:
			mon.tick(iface)
		case <-mon.stopC:
			wg.Done()
			return
		}
	}
}

func (mon *AbstractMonitor) ClockStop() {
	select {
	case <-mon.stopC:
		return
	default:
		close(mon.stopC)
	}
}

func (mon *AbstractMonitor) test() bool { return false }

func (mon *AbstractMonitor) tick(iface MonitorInterface) {

	reqStart := getMs()
	up := iface.test() // do check
	lag := getMs() - reqStart

	historyMaxSize := HistorySize

	currentHistorySize := len(mon.history)

	// remove first history if history queue will be full
	if currentHistorySize == historyMaxSize {
		mon.history = mon.history[1:]
	} else if currentHistorySize > historyMaxSize {
		// why will happened ?
		mon.history = mon.history[currentHistorySize-(historyMaxSize-1):]
	}

	mon.history = append(mon.history, up)

	mon.AnalyseData()

	// report lag
	if mon.MetricID > 0 {
		go mon.config.API.SendMetric(mon.MetricID, lag)
	}

}

// AnalyseData decides if the monitor is statistically up or down and creates / resolves an incident
func (mon *AbstractMonitor) AnalyseData() {

	// look at the past few incidents
	numDown := 0

	for _, up := range mon.history {
		if up == false {
			numDown++
		}
	}

	// the end of history
	currentIsUp := mon.history[len(mon.history)-1]

	l := logrus.WithFields(logrus.Fields{
		"monitor": mon.Name,
		"time":    time.Now().Format(mon.config.DateFormat),
	})

	if currentIsUp {
		l.Printf("monitor is up")
	} else {
		l.Printf("monitor is down")
	}

	if !currentIsUp {
		// create incident
		tplData := getTemplateData(mon)
		tplData["FailReason"] = mon.lastFailReason

		subject, message := mon.Template.Investigating.Exec(tplData)

		if mon.incident == nil {
			mon.incident = &Incident{
				Name:         subject,
				ComponentID:  mon.ComponentID,
				Message:      message,
				Notify:       true,
				incidentTime: time.Now(),
			}
			// is down, create an incident
			l.Infof("creating incident. Monitor is down: %v", mon.lastFailReason)
		} else {
			mon.incident.Message = message
		}

		// set investigating status
		mon.incident.SetInvestigating()
		// create/update incident
		if err, _ := mon.incident.Send(mon.config); err != nil {
			l.Printf("Error sending incident: %v", err)
		}

	} else if mon.incident != nil && currentIsUp {
		// was down, an incident existed
		// its now ok, make it resolved

		// resolve incident
		tplData := getTemplateData(mon)
		tplData["downSeconds"] = fmt.Sprintf("%.2f", time.Since(mon.incident.incidentTime).Seconds())

		subject, message := mon.Template.Fixed.Exec(tplData)
		mon.incident.Name = subject
		mon.incident.Message = message
		mon.incident.SetFixed()

		if err, _ := mon.incident.Send(mon.config); err != nil {
			l.Printf("Error sending incident: %v", err)
		} else {
			// clean history
			mon.lastFailReason = ""
			mon.incident = nil
		}
	}

}
