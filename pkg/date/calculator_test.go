package date

import (
	"testing"
	"time"

	"git-time-machine/pkg/args"
	"git-time-machine/pkg/git"
)

func TestDetermineDateRange_BothDatesProvided(t *testing.T) {
	config := &args.Config{
		DateFrom: mustParseTime("2023-01-01"),
		DateTo:   mustParseTime("2023-12-31"),
	}

	commits := []git.Commit{
		{Date: "Mon Jan 15 10:00:00 2022 +0000"},
	}

	start, end := determineDateRange(commits, config)

	expectedStart := mustParseTime("2023-01-01")
	expectedEnd := mustParseTime("2023-12-31")

	if start.Unix() != expectedStart.Unix() {
		t.Errorf("start = %v, want %v", start, expectedStart)
	}
	if end.Unix() != expectedEnd.Unix() {
		t.Errorf("end = %v, want %v", end, expectedEnd)
	}
}

func TestDetermineDateRange_NoDatesProvided(t *testing.T) {
	config := &args.Config{}

	commits := []git.Commit{
		{Date: "Mon Jan 1 10:00:00 2023 +0000"},
		{Date: "Fri Dec 31 23:00:00 2023 +0000"},
	}

	start, end := determineDateRange(commits, config)

	expectedStart := mustParseTime("2023-01-01")
	*expectedStart = time.Date(expectedStart.Year(), expectedStart.Month(), expectedStart.Day(), 10, 0, 0, 0, expectedStart.Location()) // 10:00
	expectedEnd := mustParseTime("2023-12-31")
	*expectedEnd = time.Date(expectedEnd.Year(), expectedEnd.Month(), expectedEnd.Day(), 23, 0, 0, 0, expectedEnd.Location()) // 23:00

	if start.Unix() != expectedStart.Unix() {
		t.Errorf("start = %v, want %v", start, expectedStart)
	}
	if end.Unix() != expectedEnd.Unix() {
		t.Errorf("end = %v, want %v", end, expectedEnd)
	}
}

func TestDetermineDateRange_OnlyDateFrom(t *testing.T) {
	config := &args.Config{
		DateFrom: mustParseTime("2023-01-01"),
	}

	commits := []git.Commit{
		{Date: "Mon Jan 15 10:00:00 2023 +0000"},
		{Date: "Fri Dec 31 23:00:00 2023 +0000"},
	}

	start, end := determineDateRange(commits, config)

	expectedStart := mustParseTime("2023-01-01")
	// end should be the last commit date

	if start.Unix() != expectedStart.Unix() {
		t.Errorf("start = %v, want %v", start, expectedStart)
	}
	if end.IsZero() {
		t.Error("end should not be zero")
	}
}

func TestDetermineDateRange_OnlyDateTo(t *testing.T) {
	config := &args.Config{
		DateTo: mustParseTime("2023-12-31"),
	}

	commits := []git.Commit{
		{Date: "Mon Jan 1 10:00:00 2023 +0000"},
		{Date: "Fri Dec 31 23:00:00 2023 +0000"},
	}

	start, end := determineDateRange(commits, config)

	expectedEnd := mustParseTime("2023-12-31")
	// start should be the first commit date

	if end.Unix() != expectedEnd.Unix() {
		t.Errorf("end = %v, want %v", end, expectedEnd)
	}
	if start.IsZero() {
		t.Error("start should not be zero")
	}
}

func TestCalculateNewDates_SingleCommit(t *testing.T) {
	config := &args.Config{
		DateFrom: mustParseTime("2023-01-01"),
		DateTo:   mustParseTime("2023-01-02"),
	}

	commits := []git.Commit{
		{Date: "Mon Jan 15 10:00:00 2023 +0000"},
	}

	newDates, err := CalculateNewDates(commits, config)

	if err != nil {
		t.Fatalf("CalculateNewDates() error = %v", err)
	}

	if len(newDates) != 1 {
		t.Errorf("Expected 1 date, got %d", len(newDates))
	}

	date := newDates[0]
	if date.IsZero() {
		t.Error("Date should not be zero")
	}
}

func TestCalculateNewDates_MinInterval(t *testing.T) {
	config := &args.Config{
		DateFrom:    mustParseTime("2023-01-01"),
		DateTo:      mustParseTime("2023-01-03"),
		MinInterval: 24, // 24 hours
	}

	commits := []git.Commit{
		{Date: "Mon Jan 1 10:00:00 2023 +0000"},
		{Date: "Tue Jan 2 10:00:00 2023 +0000"},
		{Date: "Wed Jan 3 10:00:00 2023 +0000"},
	}

	newDates, err := CalculateNewDates(commits, config)

	if err != nil {
		t.Fatalf("CalculateNewDates() error = %v", err)
	}

	if len(newDates) != 3 {
		t.Fatalf("Expected 3 dates, got %d", len(newDates))
	}

	// Check that dates are at least 24 hours apart
	for i := 1; i < len(newDates); i++ {
		diff := newDates[i].Sub(newDates[i-1]).Hours()
		if diff < 24 {
			t.Errorf("Interval of %.1f hours is less than required 24 hours", diff)
		}
	}
}

func TestCalculateNewDates_ImpossibleDistribution(t *testing.T) {
	config := &args.Config{
		DateFrom:    mustParseTime("2023-01-01"),
		DateTo:      mustParseTime("2023-01-02"),
		MinInterval: 48, // 48 hours
	}

	commits := []git.Commit{
		{Date: "Mon Jan 1 10:00:00 2023 +0000"},
		{Date: "Tue Jan 2 10:00:00 2023 +0000"},
	}

	_, err := CalculateNewDates(commits, config)

	if err == nil {
		t.Error("Expected error for impossible distribution")
	}
}

func TestCalculateNewDates_TimeSlot(t *testing.T) {
	config := &args.Config{
		DateFrom: mustParseTime("2023-06-01"),
		DateTo:   mustParseTime("2023-06-02"),
		TimeFrom: &args.TimeOfDay{Hour: 10},
		TimeTo:   &args.TimeOfDay{Hour: 12},
	}

	commits := []git.Commit{
		{Date: "Thu Jun 1 08:00:00 2023 +0000"},
		{Date: "Thu Jun 1 09:00:00 2023 +0000"},
	}

	newDates, err := CalculateNewDates(commits, config)

	if err != nil {
		t.Fatalf("CalculateNewDates() error = %v", err)
	}

	// Check that all times are within 10:00-12:00
	for _, date := range newDates {
		hour := date.Hour()
		if hour < 10 || hour >= 12 {
			t.Errorf("Time %d:00 is outside 10:00-12:00 slot", hour)
		}
	}
}

func TestCalculateNewDates_ChronologicalOrder(t *testing.T) {
	config := &args.Config{
		DateFrom: mustParseTime("2023-01-01"),
		DateTo:   mustParseTime("2023-01-10"),
	}

	commits := []git.Commit{
		{Date: "Mon Jan 1 10:00:00 2023 +0000"},
		{Date: "Tue Jan 2 10:00:00 2023 +0000"},
		{Date: "Wed Jan 3 10:00:00 2023 +0000"},
	}

	newDates, err := CalculateNewDates(commits, config)

	if err != nil {
		t.Fatalf("CalculateNewDates() error = %v", err)
	}

	// Check that dates are in chronological order
	for i := 1; i < len(newDates); i++ {
		if newDates[i].Before(newDates[i-1]) {
			t.Errorf("Date %v is before previous date %v", newDates[i], newDates[i-1])
		}
	}
}

// Helper function to parse time
func mustParseTime(s string) *time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	// Set the time to midnight for consistent testing
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return &t
}
