package domain

import "testing"

func TestGetRoleName(t *testing.T) {
	tests := []struct {
		role     int
		expected string
	}{
		{0, "client"},
		{2, "router"},
		{6, "sensor"},
		{11, "router_late"},
		{999, "unknown"},
	}

	for _, tt := range tests {
		result := GetRoleName(tt.role)
		if result != tt.expected {
			t.Errorf("GetRoleName(%d) = %s, expected %s", tt.role, result, tt.expected)
		}
	}
}
