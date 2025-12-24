package session

import "testing"

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"valid-name", false},
		{"valid_name", false},
		{"ValidName123", false},
		{"a", false},
		{"-", false}, // Special case for previous session
		{"", true},
		{"invalid name", true},
		{"invalid/name", true},
		{"invalid.name", true},
		{"invalid@name", true},
	}

	for _, tt := range tests {
		err := ValidateName(tt.name)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}
