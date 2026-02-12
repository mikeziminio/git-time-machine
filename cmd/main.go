package main

import (
	"fmt"
	"os"
	"time"

	"git-time-machine/pkg/args"
	"git-time-machine/pkg/git"
	"git-time-machine/pkg/date"
	"git-time-machine/pkg/writer"
)

// Global config for the run
var config *args.Config

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		printHelp()
		os.Exit(1)
	}
}

func run() error {
	// Define flags
	var input, output, userName, userEmail, dateFromStr, dateToStr, timeFromStr, timeToStr string
	var minInterval int
	var quiet, help bool

	// Parse manually
	cmdArgs := os.Args[1:]
	for i := 0; i < len(cmdArgs); i++ {
		arg := cmdArgs[i]
		if arg == "-i" || arg == "--input" {
			if i+1 < len(cmdArgs) {
				input = cmdArgs[i+1]
				i++
			}
		} else if arg == "-o" || arg == "--output" {
			if i+1 < len(cmdArgs) {
				output = cmdArgs[i+1]
				i++
			}
		} else if arg == "--user-name" {
			if i+1 < len(cmdArgs) {
				userName = cmdArgs[i+1]
				i++
			}
		} else if arg == "--user-email" {
			if i+1 < len(cmdArgs) {
				userEmail = cmdArgs[i+1]
				i++
			}
		} else if arg == "--date-from" {
			if i+1 < len(cmdArgs) {
				dateFromStr = cmdArgs[i+1]
				i++
			}
		} else if arg == "--date-to" {
			if i+1 < len(cmdArgs) {
				dateToStr = cmdArgs[i+1]
				i++
			}
		} else if arg == "--time-from" {
			if i+1 < len(cmdArgs) {
				timeFromStr = cmdArgs[i+1]
				i++
			}
		} else if arg == "--time-to" {
			if i+1 < len(cmdArgs) {
				timeToStr = cmdArgs[i+1]
				i++
			}
		} else if arg == "--min-interval" {
			if i+1 < len(cmdArgs) {
				_, err := fmt.Sscanf(cmdArgs[i+1], "%d", &minInterval)
				if err != nil {
					return fmt.Errorf("Invalid min-interval: %v", err)
				}
				i++
			}
		} else if arg == "-q" || arg == "--quiet" {
			quiet = true
		} else if arg == "--help" || arg == "-help" {
			help = true
			break
		}
	}

	if help {
		printHelp()
		return nil
	}

	// Check if no flags were provided at all
	if input == "" && output == "" {
		printHelp()
		return nil
	}

	// Check for -i only mode (information mode)
	if input != "" && output == "" {
		return printRepoInfo(input)
	}

	// Check required flags
	if input == "" {
		return fmt.Errorf("required flag -i is missing")
	}
	if output == "" {
		return fmt.Errorf("required flag -o is missing")
	}

	// Parse optional flags
	var DateFrom, DateTo *time.Time

	if dateFromStr != "" {
		t, err := args.ParseDate(dateFromStr)
		if err != nil {
			return fmt.Errorf("Error parsing date-from: %v", err)
		}
		DateFrom = t
	}

	if dateToStr != "" {
		t, err := args.ParseDate(dateToStr)
		if err != nil {
			return fmt.Errorf("Error parsing date-to: %v", err)
		}
		DateTo = t
	}

	// Create config
	config = &args.Config{
		InputDir:    input,
		OutputDir:   output,
		Quiet:       quiet,
		UserName:    userName,
		UserEmail:   userEmail,
		MinInterval: minInterval,
		DateFrom:    DateFrom,
		DateTo:      DateTo,
	}

	// Set time slots
	if timeFromStr != "" {
		t, err := args.NewTimeOfDay(timeFromStr)
		if err != nil {
			return fmt.Errorf("Error parsing time-from: %v", err)
		}
		config.TimeFrom = t
	}

	if timeToStr != "" {
		t, err := args.NewTimeOfDay(timeToStr)
		if err != nil {
			return fmt.Errorf("Error parsing time-to: %v", err)
		}
		config.TimeTo = t
	}

	// Run processor
	p := &Processor{config: config}
	return p.Run()
}

func printHelp() {
	fmt.Println("Git Time Machine - Rewrite Git history with custom dates and authors")
	fmt.Println()
	fmt.Println("Usage: git-time-machine [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -i, --input string           Input Git repository directory (required)")
	fmt.Println("  -o, --output string          Output directory for rewritten repository (required)")
	fmt.Println("      --user-name string       New author name for all commits")
	fmt.Println("      --user-email string      New author email for all commits")
	fmt.Println("      --date-from string       Start date for rewriting (format: 2006-01-02 or 2006-01-02T15:04:05)")
	fmt.Println("      --date-to string         End date for rewriting (format: 2006-01-02 or 2006-01-02T15:04:05)")
	fmt.Println("      --time-from string       Start time for time slot filtering (format: 9, 09, 09:00, 23:50)")
	fmt.Println("      --time-to string         End time for time slot filtering (format: 9, 09, 09:00, 23:50)")
	fmt.Println("      --min-interval int       Minimum interval between commits in hours (integer)")
	fmt.Println("  -q, --quiet                  Quiet mode (compact output)")
	fmt.Println("      --help                   Display help message")
}

// Processor manages the git history rewriting process
type Processor struct {
	config       *args.Config
	reader       *git.Reader
	outputWriter *writer.Writer
	commits      []git.Commit
}

func (p *Processor) Run() error {
	// 1. Read input repo
	if err := p.readInputRepo(); err != nil {
		return err
	}

	// 2. Calculate new dates
	if err := p.calculateNewDates(); err != nil {
		return err
	}

	// 3. Create output repo
	if err := p.createOutputRepo(); err != nil {
		return err
	}

	// 4. Write commits
	if err := p.writeRewrittenCommits(); err != nil {
		return err
	}

	// 5. Output summary
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

	// Collect all commits (git log is newest to oldest, we need oldest to newest)
	var totalCommits int
	daysMap := make(map[string]bool)

	for _, branch := range branches {
		commits := branch.Commits
		// Reverse commits to get oldest first
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

	// Show commits if not quiet
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

	// Store new dates with commits
	for i := range p.commits {
		p.commits[i].NewDate = newDates[i]
	}
	return nil
}

func (p *Processor) createOutputRepo() error {
	writer := writer.NewWriter(p.config.OutputDir, p.config.Quiet)
	if err := writer.InitRepository(); err != nil {
		return err
	}
	p.outputWriter = writer
	return nil
}

func (p *Processor) writeRewrittenCommits() error {
	for i, commit := range p.commits {
		// Apply author changes if specified
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

		// Store mapping
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

func (p *Processor) printOutputSummary() error {
	// Count output commits and days
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
