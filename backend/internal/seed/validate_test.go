package seed

import (
	"strings"
	"testing"
	"time"
)

func reading(meter string, hour int, kwh float64) Reading {
	return Reading{MeterID: meter, Timestamp: time.Date(2026, 9, 1, hour, 0, 0, 0, time.UTC), ConsumptionKWh: kwh, VoltageV: 220, CurrentA: 100, PowerFactor: 0.95}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		in        []Reading
		wantKept  int
		wantIssue string // substring, empty = no issues
	}{
		{"clean", []Reading{reading("M-1", 0, 1), reading("M-1", 1, 1), reading("M-1", 2, 1)}, 3, ""},
		{"duplicate", []Reading{reading("M-1", 0, 1), reading("M-1", 0, 2)}, 1, "duplicate"},
		{"gap", []Reading{reading("M-1", 0, 1), reading("M-1", 3, 1)}, 2, "gap of 3h"},
		{"negative", []Reading{reading("M-1", 0, -5)}, 1, "negative"},
		{"meters are independent", []Reading{reading("M-1", 0, 1), reading("M-2", 5, 1)}, 2, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kept, issues := Validate(tt.in)
			if len(kept) != tt.wantKept {
				t.Errorf("kept %d readings, want %d", len(kept), tt.wantKept)
			}
			if tt.wantIssue == "" && len(issues) != 0 {
				t.Errorf("unexpected issues: %v", issues)
			}
			if tt.wantIssue != "" && (len(issues) == 0 || !strings.Contains(issues[0].String(), tt.wantIssue)) {
				t.Errorf("issues %v, want one containing %q", issues, tt.wantIssue)
			}
		})
	}
}
