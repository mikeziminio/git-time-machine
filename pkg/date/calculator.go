package date

import (
	"fmt"
	"math/rand"
	"time"

	"git-time-machine/pkg/args"
	"git-time-machine/pkg/git"
)

// CalculateNewDates calculates new dates for commits based on configuration
func CalculateNewDates(commits []git.Commit, config *args.Config) ([]time.Time, error) {
	// Determine the date range
	startDate, endDate := determineDateRange(commits, config)

	// Calculate total available time in seconds
	totalDuration := endDate.Sub(startDate).Seconds()

	// Get minimum interval in seconds
	minIntervalSec := float64(config.MinInterval * 3600)

	// Validate feasibility: check if all commits can fit within the time window
	if len(commits) > 1 {
		minRequiredDuration := float64(len(commits)-1) * minIntervalSec
		if totalDuration < minRequiredDuration {
			return nil, fmt.Errorf(
				"impossible to distribute %d commits within %.2f hours with minimum interval of %d hours",
				len(commits),
				totalDuration/3600,
				config.MinInterval,
			)
		}
	}

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Calculate new dates
	newDates := make([]time.Time, len(commits))

	// Distribute commits evenly within the date range
	if len(commits) == 1 {
		// Single commit: random position
		if totalDuration > 0 {
			ratio := rand.Float64()
			offset := time.Duration(ratio * totalDuration)
			newDates[0] = startDate.Add(offset)
		} else {
			newDates[0] = startDate
		}
	} else {
		// Multiple commits: distribute with min-interval respect
		if totalDuration > 0 {
			// Calculate step size considering minimum interval
			// config.MinInterval is in hours
			stepDuration := time.Duration(config.MinInterval) * time.Hour
			
			// Ensure minimum total duration needed
			minTotal := time.Duration(len(commits)-1) * stepDuration
			
			if minTotal > time.Duration(totalDuration)*time.Second {
				// Not enough time, but try to fit
				for i := range commits {
					offset := time.Duration(i) * stepDuration
					newDates[i] = startDate.Add(offset)
				}
			} else {
				// Spread across the range with min-interval as minimum
				// Use a spread calculation that respects the interval
				availableDuration := time.Duration(totalDuration) * time.Second
				spreadDuration := availableDuration - minTotal
				
				// Distribute the remaining time evenly between intervals
				extraPerInterval := time.Duration(0)
				if len(commits) > 1 {
					extraPerInterval = spreadDuration / time.Duration(len(commits)-1)
				}
				
				for i := range commits {
					offset := time.Duration(i) * (stepDuration + extraPerInterval)
					newDates[i] = startDate.Add(offset)
				}
			}
		} else {
			for i := range commits {
				newDates[i] = startDate
			}
		}
	}

	// Apply time slot if specified
	if !config.TimeFrom.IsZero() && !config.TimeTo.IsZero() {
		for i := range commits {
			newDates[i] = applyTimeSlotToRandomTime(newDates[i], config)
		}
	}

	// Sort dates to ensure chronological order
	for i := 0; i < len(newDates)-1; i++ {
		for j := i + 1; j < len(newDates); j++ {
			if newDates[j].Before(newDates[i]) {
				newDates[i], newDates[j] = newDates[j], newDates[i]
			}
		}
	}

	// Ensure minimum interval between consecutive commits
	// If interval is too small, shift later commits
	for i := 1; i < len(newDates); i++ {
		requiredTime := newDates[i-1].Add(time.Duration(minIntervalSec) * time.Second)
		if newDates[i].Before(requiredTime) {
			newDates[i] = requiredTime
		}
	}

	return newDates, nil
}

// determineDateRange determines the effective date range
func determineDateRange(commits []git.Commit, config *args.Config) (time.Time, time.Time) {
	// If both dates are provided, use them
	if config.DateFrom != nil && config.DateTo != nil {
		return *config.DateFrom, *config.DateTo
	}

	// Get actual first and last commit dates
	firstDate := time.Time{}
	lastDate := time.Time{}

	for _, commit := range commits {
		commitDate, err := time.Parse("Mon Jan 2 15:04:05 2006 -0700", commit.Date)
		if err != nil {
			// Try alternate parsing
			if commitDate, err = time.Parse("2006-01-02 15:04:05", commit.Date); err != nil {
				continue
			}
		}

		if firstDate.IsZero() || commitDate.Before(firstDate) {
			firstDate = commitDate
		}
		if lastDate.IsZero() || commitDate.After(lastDate) {
			lastDate = commitDate
		}
	}

	// If no date range provided, use actual commit dates
	if config.DateFrom == nil && config.DateTo == nil {
		return firstDate, lastDate
	}

	// Apply provided date with fallback to commit dates
	if config.DateFrom != nil {
		firstDate = *config.DateFrom
	}
	if config.DateTo != nil {
		lastDate = *config.DateTo
	}

	return firstDate, lastDate
}

// applyTimeSlotToRandomTime applies a random time within the time slot
func applyTimeSlotToRandomTime(date time.Time, config *args.Config) time.Time {
	// Get time slot bounds in minutes from midnight
	timeFrom := config.TimeFrom.Hour*60 + config.TimeFrom.Minute
	timeTo := config.TimeTo.Hour*60 + config.TimeTo.Minute

	// Pick a random minute within the time slot
	randMinute := timeFrom + rand.Intn(timeTo-timeFrom)

	// Add random seconds (0-59)
	randSecond := rand.Intn(60)

	// Apply to the date
	year, month, day := date.Date()
	newDate := time.Date(year, month, day, randMinute/60, randMinute%60, randSecond, 0, date.Location())
	return newDate
}

// clamp restricts a value to be within min and max
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
