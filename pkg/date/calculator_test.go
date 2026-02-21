package date

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

	assert.Equal(t, expectedStart.Unix(), start.Unix())
	assert.Equal(t, expectedEnd.Unix(), end.Unix())
}

func TestDetermineDateRange_NoDatesProvided(t *testing.T) {
	config := &args.Config{}

	commits := []git.Commit{
		{Date: "Mon Jan 1 10:00:00 2023 +0000"},
		{Date: "Fri Dec 31 23:00:00 2023 +0000"},
	}

	start, end := determineDateRange(commits, config)

	expectedStart := mustParseTime("2023-01-01")
	*expectedStart = time.Date(expectedStart.Year(), expectedStart.Month(), expectedStart.Day(), 10, 0, 0, 0, expectedStart.Location())
	expectedEnd := mustParseTime("2023-12-31")
	*expectedEnd = time.Date(expectedEnd.Year(), expectedEnd.Month(), expectedEnd.Day(), 23, 0, 0, 0, expectedEnd.Location())

	assert.Equal(t, expectedStart.Unix(), start.Unix())
	assert.Equal(t, expectedEnd.Unix(), end.Unix())
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

	assert.Equal(t, expectedStart.Unix(), start.Unix())
	assert.False(t, end.IsZero())
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

	assert.Equal(t, expectedEnd.Unix(), end.Unix())
	assert.False(t, start.IsZero())
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

	require.NoError(t, err)
	assert.Len(t, newDates, 1)
	assert.False(t, newDates[0].IsZero())
}

func TestCalculateNewDates_MinInterval(t *testing.T) {
	config := &args.Config{
		DateFrom:    mustParseTime("2023-01-01"),
		DateTo:      mustParseTime("2023-01-03"),
		MinInterval: 24,
	}

	commits := []git.Commit{
		{Date: "Mon Jan 1 10:00:00 2023 +0000"},
		{Date: "Tue Jan 2 10:00:00 2023 +0000"},
		{Date: "Wed Jan 3 10:00:00 2023 +0000"},
	}

	newDates, err := CalculateNewDates(commits, config)

	require.NoError(t, err)
	assert.Len(t, newDates, 3)

	for i := 1; i < len(newDates); i++ {
		diff := newDates[i].Sub(newDates[i-1]).Hours()
		assert.GreaterOrEqual(t, diff, float64(24), "Interval should be at least 24 hours")
	}
}

func TestCalculateNewDates_ImpossibleDistribution(t *testing.T) {
	config := &args.Config{
		DateFrom:    mustParseTime("2023-01-01"),
		DateTo:      mustParseTime("2023-01-02"),
		MinInterval: 48,
	}

	commits := []git.Commit{
		{Date: "Mon Jan 1 10:00:00 2023 +0000"},
		{Date: "Tue Jan 2 10:00:00 2023 +0000"},
	}

	_, err := CalculateNewDates(commits, config)

	assert.Error(t, err)
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

	require.NoError(t, err)

	for _, date := range newDates {
		hour := date.Hour()
		assert.GreaterOrEqual(t, hour, 10, "Hour should be >= 10")
		assert.Less(t, hour, 12, "Hour should be < 12")
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

	require.NoError(t, err)

	for i := 1; i < len(newDates); i++ {
		assert.False(t, newDates[i].Before(newDates[i-1]), "Dates should be in chronological order")
	}
}

func mustParseTime(s string) *time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return &t
}
