package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"git-time-machine/pkg/args"
	"github.com/spf13/cobra"
)

func TestCobraCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "unknown flag",
			args:    []string{"--unknown", "value"},
			wantErr: true,
		},
		{
			name:    "unknown flag short",
			args:    []string{"-x"},
			wantErr: true,
		},
		{
			name:    "missing required flags",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "help flag",
			args:    []string{"--help"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createRootCommand()
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func createRootCommand() *cobra.Command {
	config := &args.Config{}

	cmd := &cobra.Command{
		Use:   "git-time-machine",
		Short: "Git Time Machine - Rewrite Git history with custom dates and authors",
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.Help {
				cmd.Usage()
				return nil
			}

			if config.InputDir == "" && config.OutputDir == "" {
				cmd.Usage()
				return nil
			}

			if config.InputDir != "" && config.OutputDir == "" {
				return printRepoInfo(config.InputDir)
			}

			if err := config.Validate(); err != nil {
				return err
			}

			if err := config.ValidateTimeRanges(); err != nil {
				return err
			}

			if err := config.ValidateInterval(); err != nil {
				return err
			}

			p := &Processor{config: config}
			return p.Run()
		},
	}

	cmd.Flags().StringVarP(&config.InputDir, "input", "i", "", "Input Git repository directory (required)")
	cmd.Flags().StringVarP(&config.OutputDir, "output", "o", "", "Output directory for rewritten repository (required)")
	cmd.Flags().StringVar(&config.UserName, "user-name", "", "New author name for all commits")
	cmd.Flags().StringVar(&config.UserEmail, "user-email", "", "New author email for all commits")
	cmd.Flags().Func("date-from", "Start date for rewriting (format: 2006-01-02 or 2006-01-02T15:04:05)", func(s string) error {
		t, err := args.ParseDate(s)
		if err != nil {
			return err
		}
		config.DateFrom = t
		return nil
	})
	cmd.Flags().Func("date-to", "End date for rewriting (format: 2006-01-02 or 2006-01-02T15:04:05)", func(s string) error {
		t, err := args.ParseDate(s)
		if err != nil {
			return err
		}
		config.DateTo = t
		return nil
	})
	cmd.Flags().Func("time-from", "Start time for time slot filtering (format: 9, 09, 09:00, 23:50)", func(s string) error {
		t, err := args.NewTimeOfDay(s)
		if err != nil {
			return err
		}
		config.TimeFrom = t
		return nil
	})
	cmd.Flags().Func("time-to", "End time for time slot filtering (format: 9, 09, 09:00, 23:50, default: 23)", func(s string) error {
		t, err := args.NewTimeOfDay(s)
		if err != nil {
			return err
		}
		config.TimeTo = t
		return nil
	})
	cmd.Flags().IntVar(&config.MinInterval, "min-interval", 0, "Minimum interval between commits in hours (integer)")
	cmd.Flags().BoolVarP(&config.Quiet, "quiet", "q", false, "Quiet mode (compact output)")
	cmd.Flags().BoolVar(&config.Help, "help", false, "Display help message")

	return cmd
}
