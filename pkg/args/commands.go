package args

import (
	"fmt"

	"github.com/spf13/cobra"
)

// GetCommand returns the root Cobra command
func GetCommand() *cobra.Command {
	config := &Config{}

	cmd := &cobra.Command{
		Use:   "git-time-machine",
		Short: "Git Time Machine - Rewrite Git history with custom dates and authors",
		Long: `Git Time Machine is a CLI tool that rewrites Git history by:
  - Changing author names and emails
  - Redistributing commits across custom date ranges
  - Applying time slot constraints`,
		Example: `  git-time-machine -i ./my-repo -o ./rewritten-repo --user-name "John Doe" --user-email "john@example.com"
  git-time-machine -i ./my-repo -o ./rewritten-repo --date-from 2023-01-01 --date-to 2023-12-31
  git-time-machine -i ./my-repo -o ./rewritten-repo --time-from 9 --time-to 18 --min-interval 2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.Help {
				cmd.Usage()
				return nil
			}

			// Validate required flags
			if err := config.Validate(); err != nil {
				return err
			}

			// Validate time ranges
			if err := config.ValidateTimeRanges(); err != nil {
				return err
			}

			// Validate interval
			if err := config.ValidateInterval(); err != nil {
				return err
			}

			fmt.Printf("Input: %s\n", config.InputDir)
			fmt.Printf("Output: %s\n", config.OutputDir)
			fmt.Printf("Quiet: %v\n", config.Quiet)
			return nil
		},
	}

	// Required flags
	cmd.Flags().StringVarP(&config.InputDir, "input", "i", "", "Input Git repository directory (required)")
	cmd.Flags().StringVarP(&config.OutputDir, "output", "o", "", "Output directory for rewritten repository (required)")

	// Optional: Author replacement
	cmd.Flags().StringVar(&config.UserName, "user-name", "", "New author name for all commits")
	cmd.Flags().StringVar(&config.UserEmail, "user-email", "", "New author email for all commits")

	// Optional: Date range
	cmd.Flags().Func("date-from", "Start date for rewriting (format: 2006-01-02 or 2006-01-02T15:04:05)", func(s string) error {
		t, err := ParseDate(s)
		if err != nil {
			return err
		}
		config.DateFrom = t
		return nil
	})
	cmd.Flags().Func("date-to", "End date for rewriting (format: 2006-01-02 or 2006-01-02T15:04:05)", func(s string) error {
		t, err := ParseDate(s)
		if err != nil {
			return err
		}
		config.DateTo = t
		return nil
	})

	// Optional: Time slot constraints
	cmd.Flags().Func("time-from", "Start time for time slot filtering (format: 9, 09, 09:00, 23:50)", func(s string) error {
		t, err := NewTimeOfDay(s)
		if err != nil {
			return err
		}
		config.TimeFrom = t
		return nil
	})
	cmd.Flags().Func("time-to", "End time for time slot filtering (format: 9, 09, 09:00, 23:50, default: 23)", func(s string) error {
		t, err := NewTimeOfDay(s)
		if err != nil {
			return err
		}
		config.TimeTo = t
		return nil
	})

	// Optional: Minimum interval
	cmd.Flags().IntVar(&config.MinInterval, "min-interval", 0, "Minimum interval between commits in hours (integer)")

	// Output control
	cmd.Flags().BoolVarP(&config.Quiet, "quiet", "q", false, "Quiet mode (compact output)")
	cmd.Flags().BoolVar(&config.Help, "help", false, "Display help message")

	return cmd
}
