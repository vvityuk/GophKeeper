package models

import "testing"

func TestDataType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		dataType DataType
		want     bool
	}{
		{"CREDENTIALS", DataTypeCredentials, true},
		{"TEXT", DataTypeText, true},
		{"BINARY", DataTypeBinary, true},
		{"CARD", DataTypeCard, true},
		{"INVALID", DataType("INVALID"), false},
		{"empty", DataType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dataType.IsValid(); got != tt.want {
				t.Errorf("DataType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

