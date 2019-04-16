package cachet

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Investigating template
var defaultHTTPInvestigatingTpl = MessageTemplate{
	Subject: `{{ .Monitor.Name }} - {{ .SystemName }}`,
	Message: `{{ .Monitor.Name }} check **failed** (server time: {{ .now }})

{{ .FailReason }}`,
}

// Fixed template
var defaultHTTPFixedTpl = MessageTemplate{
	Subject: `{{ .Monitor.Name }} - {{ .SystemName }}`,
	Message: `**Resolved** - {{ .now }}

Down seconds: {{ .downSeconds }}s`,
}

type HTTPMonitor struct {
	AbstractMonitor `mapstructure:",squash"`

	Method             string
	ExpectedStatusCode int `mapstructure:"expected_status_code"`
	Headers            map[string]string

	// compiled to Regexp
	ExpectedBody string `mapstructure:"expected_body"`
	bodyRegexp   *regexp.Regexp
}

func (monitor *HTTPMonitor) test() bool {

	req, err := http.NewRequest(monitor.Method, monitor.Target, nil)

	for k, v := range monitor.Headers {
		req.Header.Add(k, v)
	}

	transport := http.DefaultTransport.(*http.Transport)

	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: monitor.Strict == false}
	client := &http.Client{
		Timeout:   time.Duration(monitor.Timeout * time.Second),
		Transport: transport,
	}

	resp, err := client.Do(req)
	if err != nil {
		monitor.lastFailReason = err.Error()
		return false
	}

	defer resp.Body.Close()

	// status check
	if monitor.ExpectedStatusCode > 0 && resp.StatusCode != monitor.ExpectedStatusCode {
		monitor.lastFailReason = "Expected HTTP response status: " + strconv.Itoa(monitor.ExpectedStatusCode) + ", got: " + strconv.Itoa(resp.StatusCode)
		return false
	} else if resp.StatusCode > 499 {
		// server internal error also failed
		return false
	}

	// body check
	if monitor.bodyRegexp != nil {
		// check response body
		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			monitor.lastFailReason = err.Error()
			return false
		}

		if !monitor.bodyRegexp.Match(responseBody) {
			monitor.lastFailReason = fmt.Sprintf(
				"Unexpected body: %v.\nExpected to match: %v",
				string(responseBody),
				monitor.ExpectedBody,
			)
			return false
		}
	}

	return true
}

// Validate configuration
func (mon *HTTPMonitor) Validate() []string {

	mon.Template.Investigating.SetDefault(defaultHTTPInvestigatingTpl)
	mon.Template.Fixed.SetDefault(defaultHTTPFixedTpl)

	errs := mon.AbstractMonitor.Validate()

	if len(mon.ExpectedBody) > 0 {
		exp, err := regexp.Compile(mon.ExpectedBody)
		if err != nil {
			errs = append(errs, "Regexp compilation failure: "+err.Error())
		} else {
			mon.bodyRegexp = exp
		}
	}

	mon.Method = strings.ToUpper(mon.Method)
	switch mon.Method {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD":
		break
	case "":
		mon.Method = "GET"
	default:
		errs = append(errs, "Unsupported HTTP method: "+mon.Method)
	}

	return errs
}

func (mon *HTTPMonitor) Describe() []string {
	features := mon.AbstractMonitor.Describe()
	features = append(features, "Method: "+mon.Method)

	return features
}
