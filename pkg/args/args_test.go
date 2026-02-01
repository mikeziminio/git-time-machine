package args

import (
	"testing"
	"time"
)

func TestNewTimeOfDay_HourOnly(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		hour    int
		minute  int
		wantErr bool
	}{
		{"Single digit hour", "9", 9, 0, false},
		{"Double digit hour", "12", 12, 0, false},
		{"Zero hour", "0", 0, 0, false},
		{"Two digit zero", "00", 0, 0, false},
		{"Maximum hour", "23", 23, 0, false},
		{"Invalid hour (24)", "24", 0, 0, true},
		{"Invalid hour (negative)", "-1", 0, 0, true},
		{"Invalid input", "abc", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewTimeOfDay(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTimeOfDay() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if result.Hour != tt.hour || result.Minute != tt.minute {
				t.Errorf("NewTimeOfDay() = (%d, %d), want (%d, %d)", result.Hour, result.Minute, tt.hour, tt.minute)
			}
		})
	}
}

func TestNewTimeOfDay_FullFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		hour    int
		minute  int
		wantErr bool
	}{
		{"Standard time", "09:30", 9, 30, false},
		{"Midnight", "00:00", 0, 0, false},
		{"Noon", "12:00", 12, 0, false},
		{"End of day", "23:59", 23, 59, false},
		{"Single digit hour", "9:30", 9, 30, false},
		{"Invalid minute (60)", "09:60", 0, 0, true},
		{"Invalid hour (24)", "24:00", 0, 0, true},
		{"Invalid format", "9:30am", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewTimeOfDay(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTimeOfDay() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if result.Hour != tt.hour || result.Minute != tt.minute {
				t.Errorf("NewTimeOfDay() = (%d, %d), want (%d, %d)", result.Hour, result.Minute, tt.hour, tt.minute)
			}
		})
	}
}

func TestTimeOfDay_ToDuration(t *testing.T) {
	tests := []struct {
		name     string
		hour     int
		minute   int
		expected time.Duration
	}{
		{"Midnight", 0, 0, 0},
		{"One hour", 1, 0, time.Hour},
		{"Ninety minutes", 1, 30, 90 * time.Minute},
		{"End of day", 23, 59, 23*time.Hour + 59*time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tod := &TimeOfDay{Hour: tt.hour, Minute: tt.minute}
			result := tod.ToDuration()
			if result != tt.expected {
				t.Errorf("ToDuration() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTimeOfDay_IsZero(t *testing.T) {
	t.Run("Nil returns true", func(t *testing.T) {
		var tod *TimeOfDay
		if !tod.IsZero() {
			t.Error("IsZero() = false, want true for nil")
		}
	})

	t.Run("Non-nil returns false", func(t *testing.T) {
		tod := &TimeOfDay{Hour: 12, Minute: 0}
		if tod.IsZero() {
			t.Error("IsZero() = true, want false for non-nil")
		}
	})
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Date only", "2023-01-15", false},
		{"Date with time", "2023-01-15T14:30:00", false},
		{"Invalid date", "2023-13-01", true},
		{"Invalid format", "01-15-2023", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if result.IsZero() {
				t.Error("ParseDate() returned zero time")
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{"Valid config", &Config{InputDir: "/input", OutputDir: "/output"}, false},
		{"Missing input", &Config{OutputDir: "/output"}, true},
		{"Missing output", &Config{InputDir: "/input"}, true},
		{"Both missing", &Config{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_ValidateTimeRanges(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{"Valid time range", &Config{TimeFrom: &TimeOfDay{Hour: 9}, TimeTo: &TimeOfDay{Hour: 17}}, false},
		{"Same time", &Config{TimeFrom: &TimeOfDay{Hour: 9}, TimeTo: &TimeOfDay{Hour: 9}}, true},
		{"Invalid range (from > to)", &Config{TimeFrom: &TimeOfDay{Hour: 17}, TimeTo: &TimeOfDay{Hour: 9}}, true},
		{"Nil from, valid to", &Config{TimeTo: &TimeOfDay{Hour: 17}}, false},
		{"Valid from, nil to", &Config{TimeFrom: &TimeOfDay{Hour: 9}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.ValidateTimeRanges()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTimeRanges() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_ValidateInterval(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{"Zero interval", &Config{MinInterval: 0}, false},
		{"Positive interval", &Config{MinInterval: 1}, false},
		{"Negative interval", &Config{MinInterval: -1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.ValidateInterval()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInterval() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
