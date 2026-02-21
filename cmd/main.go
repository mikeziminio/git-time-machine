package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"git-time-machine/pkg/args"
	"git-time-machine/pkg/date"
	"git-time-machine/pkg/git"
	"git-time-machine/pkg/writer"
)

type Processor struct {
	config       *args.Config
	reader       *git.Reader
	outputWriter *writer.Writer
	commits      []git.Commit
}

func (p *Processor) Run() error {
	if err := p.readInputRepo(); err != nil {
		return err
	}

	if err := p.calculateNewDates(); err != nil {
		return err
	}

	if err := p.createOutputRepo(); err != nil {
		return err
	}

	if err := p.writeRewrittenCommits(); err != nil {
		return err
	}

	if err := p.copyFiles(); err != nil {
		return err
	}

	if err := p.printOutputSummary(); err != nil {
		return err
	}

	return nil
}

func (p *Processor) readInputRepo() error {
	reader := git.NewReader(p.config.InputDir)
	branches, err := reader.GetBranches()
	if err != nil {
		return fmt.Errorf("failed to read repository: %w", err)
	}

	var totalCommits int
	daysMap := make(map[string]bool)

	for _, branch := range branches {
		commits := branch.Commits
		for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
			commits[i], commits[j] = commits[j], commits[i]
		}

		for _, commit := range commits {
			totalCommits++
			p.commits = append(p.commits, commit)
			if !commit.DateParsed.IsZero() {
				day := commit.DateParsed.Format("2006-01-02")
				daysMap[day] = true
			}
		}
	}

	fmt.Printf("Input repository summary:\n")
	fmt.Printf("  Commits: %d   Days: %d\n", totalCommits, len(daysMap))

	if !p.config.Quiet {
		fmt.Println("\nOriginal commits:")
		for _, branch := range branches {
			fmt.Printf("\nBranch: %s\n", branch.Name)
			for i, commit := range branch.Commits {
				fmt.Printf("  %d. SHA: %s\n", i+1, commit.SHA)
				fmt.Printf("     Author: %s <%s>\n", commit.Author, commit.Email)
				fmt.Printf("     Date: %s\n", commit.Date)
				fmt.Printf("     Message: %s\n", commit.Message)
			}
		}
	}

	return nil
}

func (p *Processor) calculateNewDates() error {
	if len(p.commits) == 0 {
		return nil
	}

	newDates, err := date.CalculateNewDates(p.commits, p.config)
	if err != nil {
		return err
	}

	for i := range p.commits {
		p.commits[i].NewDate = newDates[i]
	}
	return nil
}

func (p *Processor) createOutputRepo() error {
	w := writer.NewWriter(p.config.OutputDir, p.config.Quiet)
	if err := w.InitRepository(); err != nil {
		return err
	}
	p.outputWriter = w
	return nil
}

func (p *Processor) writeRewrittenCommits() error {
	for i, commit := range p.commits {
		author := commit.Author
		email := commit.Email
		if p.config.UserName != "" {
			author = p.config.UserName
		}
		if p.config.UserEmail != "" {
			email = p.config.UserEmail
		}

		sha, err := p.outputWriter.CreateCommit(author, email, commit.NewDate, commit.Message)
		if err != nil {
			return fmt.Errorf("failed to create commit %d: %w", i+1, err)
		}

		if !p.config.Quiet {
			fmt.Printf("%s --> %s (author: %s / %s, date: %s --> %s)\n",
				commit.SHA[:12],
				sha[:12],
				commit.Author,
				author,
				commit.Date[:25],
				commit.NewDate.Format("2006-01-02"),
			)
		}
	}

	return nil
}

func (p *Processor) copyFiles() error {
	if err := p.outputWriter.CopyFiles(p.config.InputDir); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}
	return nil
}

func (p *Processor) printOutputSummary() error {
	daysMap := make(map[string]bool)
	for _, commit := range p.commits {
		if commit.NewDate.IsZero() {
			continue
		}
		day := commit.NewDate.Format("2006-01-02")
		daysMap[day] = true
	}

	fmt.Printf("\nOutput repository summary:\n")
	fmt.Printf("  Commits: %d   Days: %d\n", len(p.commits), len(daysMap))

	return nil
}

func printRepoInfo(inputDir string) error {
	reader := git.NewReader(inputDir)
	branches, err := reader.GetBranches()
	if err != nil {
		return fmt.Errorf("failed to read repository: %w", err)
	}

	var totalCommits int
	daysMap := make(map[string]bool)

	for _, branch := range branches {
		commits := branch.Commits
		for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
			commits[i], commits[j] = commits[j], commits[i]
		}

		for _, commit := range commits {
			totalCommits++
			if !commit.DateParsed.IsZero() {
				day := commit.DateParsed.Format("2006-01-02")
				daysMap[day] = true
			}
		}
	}

	fmt.Println("Repository Information:")
	fmt.Printf("  Commits: %d   Days: %d\n", totalCommits, len(daysMap))

	fmt.Println("\nOriginal commits:")
	for _, branch := range branches {
		fmt.Printf("\nBranch: %s\n", branch.Name)
		for i, commit := range branch.Commits {
			fmt.Printf("  %d. SHA: %s\n", i+1, commit.SHA)
			fmt.Printf("     Author: %s <%s>\n", commit.Author, commit.Email)
			fmt.Printf("     Date: %s\n", commit.Date)
			fmt.Printf("     Message: %s\n", commit.Message)
		}
	}

	fmt.Println("\nFlags will be ignored, as -o is missing")

	return nil
}

func main() {
	config := &args.Config{}

	// Set default time range
	defaultTimeFrom, _ := args.NewTimeOfDay("19:00")
	defaultTimeTo, _ := args.NewTimeOfDay("23:59")
	config.TimeFrom = defaultTimeFrom
	config.TimeTo = defaultTimeTo

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

			if config.InputDir == "" && config.OutputDir == "" {
				return cmd.Help()
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
	cmd.Flags().StringVar(&config.UserName, "user-name", "Mike Zimin", "New author name for all commits")
	cmd.Flags().StringVar(&config.UserEmail, "user-email", "mikeziminio@gmail.com", "New author email for all commits")
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

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
