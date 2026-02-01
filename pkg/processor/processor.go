package processor

import (
	"fmt"

	"git-time-machine/pkg/args"
	"git-time-machine/pkg/date"
	"git-time-machine/pkg/git"
	"git-time-machine/pkg/writer"
)

// Processor manages the git history rewriting process
type Processor struct {
	config  *args.Config
	reader  *git.Reader
	outputWriter *writer.Writer
	commits  []git.Commit
}

// New creates a new processor
func New(config *args.Config) *Processor {
	return &Processor{
		config:  config,
	}
}

// Run executes the full rewriting process
func (p *Processor) Run() error {
	if err := p.runWithValidation(); err != nil {
		return err
	}
	return nil
}

func (p *Processor) runWithValidation() error {
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
