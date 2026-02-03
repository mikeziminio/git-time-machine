package args

import (
	"fmt"
	"time"
)

// Config holds all command-line arguments and validated configuration
type Config struct {
	// Required
	InputDir  string
	OutputDir string

	// Optional: Author replacement
	UserName  string
	UserEmail string

	// Optional: Date range
	DateFrom *time.Time
	DateTo   *time.Time

	// Optional: Time slot constraints
	TimeFrom *TimeOfDay
	TimeTo   *TimeOfDay

	// Optional: Minimum interval between commits
	MinInterval int // hours, integer only

	// Output control
	Quiet bool
	Help  bool
}

// TimeOfDay represents a time in hours and minutes (HH:MM format)
type TimeOfDay struct {
	Hour   int
	Minute int
}

// NewTimeOfDay parses time string in format "9", "09", "09:00", "23:50"
// Default values: time-from=00:00, time-to=23:59
func NewTimeOfDay(s string) (*TimeOfDay, error) {
	t := &TimeOfDay{}

	// Check if it's just an hour (e.g., "9", "23")
	if len(s) <= 2 {
		var hour int
		_, err := fmt.Sscanf(s, "%d", &hour)
		if err != nil {
			return nil, fmt.Errorf("invalid hour format: %w", err)
		}
		if hour < 0 || hour > 23 {
			return nil, fmt.Errorf("hour must be between 0 and 23")
		}
		t.Hour = hour
		t.Minute = 0
		return t, nil
	}

	// Check if format is H:MM or HH:MM
	colonPos := -1
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			colonPos = i
			break
		}
	}

	if colonPos > 0 && colonPos < len(s)-1 {
		hourStr := s[:colonPos]
		minuteStr := s[colonPos+1:]

		var hour, minute int
		_, err := fmt.Sscanf(hourStr+" "+minuteStr, "%d %d", &hour, &minute)
		if err != nil {
			return nil, fmt.Errorf("invalid time format: %w", err)
		}
		if hour < 0 || hour > 23 {
			return nil, fmt.Errorf("hour must be between 0 and 23")
		}
		if minute < 0 || minute > 59 {
			return nil, fmt.Errorf("minute must be between 0 and 59")
		}
		t.Hour = hour
		t.Minute = minute
		return t, nil
	}

	return nil, fmt.Errorf("invalid time format: %s (expected HH or HH:MM)", s)
}

// DefaultTimeFrom returns default TimeOfDay (00:00)
func DefaultTimeFrom() *TimeOfDay {
	return &TimeOfDay{Hour: 0, Minute: 0}
}

// DefaultTimeTo returns default TimeOfDay (23:59)
func DefaultTimeTo() *TimeOfDay {
	return &TimeOfDay{Hour: 23, Minute: 59}
}

// IsZero returns true if time is not set
func (t *TimeOfDay) IsZero() bool {
	return t == nil
}

// ToDuration returns the time of day as duration from midnight
func (t *TimeOfDay) ToDuration() time.Duration {
	return time.Duration(t.Hour)*time.Hour + time.Duration(t.Minute)*time.Minute
}

// ParseDate parses a date string in format "2006-01-02" or "2006-01-02T15:04:05"
func ParseDate(s string) (*time.Time, error) {
	t, err := time.Parse("2006-01-02", s)
	if err == nil {
		return &t, nil
	}

	t, err = time.Parse("2006-01-02T15:04:05", s)
	if err == nil {
		return &t, nil
	}

	return nil, fmt.Errorf("invalid date format: %s (expected 2006-01-02 or 2006-01-02T15:04:05)", s)
}

// Validate checks if required flags are provided
func (c *Config) Validate() error {
	if c.InputDir == "" {
		return fmt.Errorf("required flag -i is missing")
	}
	if c.OutputDir == "" {
		return fmt.Errorf("required flag -o is missing")
	}
	return nil
}

// ValidateTimeRanges validates time slot constraints
// Returns default values if not set
func (c *Config) ValidateTimeRanges() error {
	timeFrom := c.TimeFrom
	timeTo := c.TimeTo

	if timeFrom == nil {
		timeFrom = DefaultTimeFrom()
	}
	if timeTo == nil {
		timeTo = DefaultTimeTo()
	}

	if timeFrom.Hour*60+timeFrom.Minute >= timeTo.Hour*60+timeTo.Minute {
		return fmt.Errorf("--time-from must be before --time-to")
	}
	return nil
}

// ValidateInterval validates minimum interval
func (c *Config) ValidateInterval() error {
	if c.MinInterval < 0 {
		return fmt.Errorf("--min-interval must be non-negative")
	}
	return nil
}
