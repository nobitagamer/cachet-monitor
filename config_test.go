package cachet

import (
	"testing"
)

func TestGetMonitorType(t *testing.T) {
	if mt := GetMonitorType("HTTP"); mt != "http" {
		t.Error("does not return correct monitor type")
	}
}
