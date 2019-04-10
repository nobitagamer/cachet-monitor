package cachet

import (
	"net"
	"time"
)

// TCPMonitor struct
type TCPMonitor struct {
	AbstractMonitor `mapstructure:",squash"`
	Port            string
}

// CheckTCPPortAlive func
func CheckTCPPortAlive(ip, port string, timeout int64) bool {

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), time.Duration(timeout)*time.Second)
	if conn != nil {
		defer conn.Close()
	}
	if err != nil {
		return false
	} else {
		return true
	}

}

// test if it available
func (m *TCPMonitor) test() bool {
	return CheckTCPPortAlive(m.Target, m.Port, int64(m.Timeout))
}

// Validate configuration
func (m *TCPMonitor) Validate() []string {
	// super.Validate()
	errs := m.AbstractMonitor.Validate()

	if m.Target == "" {
		errs = append(errs, "Target is required")
	}

	if m.Port == "" {
		errs = append(errs, "Port is required")
	}

	return errs
}
