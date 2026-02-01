# Initial Project Tasks

## Core Features

### 1. Parse Command-Line Arguments
- **Description**: Implement CLI argument parsing with flags `-i`, `-o`, `--user-name`, `--user-email`, `--date-from`, `--date-to`, and additional time flags
- **Subtasks**:
  - [x] Define argument structure for all flags:
    - [x] Required: `-i`, `-o`
    - [x] Optional: `--user-name`, `--user-email`, `--date-from`, `--date-to`, `--time-from`, `--time-to`
    - [x] Optional: `--min-interval` (integer, hours only)
    - [x] Optional: `-q` (quiet mode for compact output)
    - [x] Optional: `--help` (display formatted help message)
  - [x] Implement validation for required flags (`-i`, `-o`)
  - [x] Add date range validation: `--date-from` < `--date-to`
  - [x] Add time slot validation: `--time-from` < `--time-to` (if both provided)
  - [x] Implement feasibility check: verify if date range + time slots + min interval can accommodate all commits
  - [x] Add pretty help message generation using a table format
  - [x] Test argument parsing with various combinations

**Time slot format**: `--time-from` and `--time-to` support:
- Full format: `09:00`, `23:50`
- Hour only: `9` (treated as `09:00`)
- `--time-to` default: `23` (treated as `23:59`)
- `--min-interval`: integer in hours (e.g., `1`, `2`, `3`)

**Output format**: If `-q` is NOT set, show compact comparison: `OLD_SHA --> NEW_SHA (author: old@new, date: old_date --> new_date)`

**Implementation Note**: Use Go's `flag` package for CLI argument parsing. Validate date ranges and time slots before processing. Calculate if the distribution is possible given constraints.

### 2. Read Git Repository
- **Description**: Read the source Git repository from the input directory
- **Subtasks**:
  - [x] Validate that the input directory is a Git repository
  - [x] Read all commits from the repository using `git log` CLI
  - [x] Store commit history in intermediate format
  - [x] Handle repository errors gracefully
  - [x] Calculate and display summary: `Commits: X   Days: Y` for input repository

**Implementation Note**: Use `git for-each-ref` to get all branches and their commit hashes for hybrid DAG preservation.

### 3. Filter and Collect Commits
- **Description**: Collect all commits that need to be rewritten using Git CLI commands
- **Subtasks**:
  - [x] Implement commit iteration from `git log` output
  - [x] Parse commit metadata (SHA, author, date, message) from Git CLI output
  - [x] Handle edge cases (empty repository, single commit)

### 4. Rewrite Author Information
- **Description**: Change author name and email for all commits using Git CLI
- **Subtasks**:
  - [x] Apply `--user-name` changes to all commits using `git commit --amend --author=`
  - [x] Apply `--user-email` changes to all commits using `git commit --amend --author=`
  - [x] Handle cases where flags are not provided (keep original values)

**Implementation Note**: For rewritten commits, store original metadata for display in output.

### 5. Calculate New Dates
- **Description**: Redistribute commits across the new date range with time slot constraints
- **Subtasks**:
  - [x] Determine the new date range:
    - [x] `--date-from` to `--date-to` if both provided
    - [x] `--date-from` to last commit date if only `--date-from` provided
    - [x] First commit date to `--date-to` if only `--date-to` provided
    - [x] First commit date to last commit date if neither provided
  - [x] Parse `--time-from` and `--time-to` flags (format: `9`, `09`, `09:00`, `23:50`) with default `23` for `--time-to`
  - [x] Apply time slot constraints: only allow commits within `HOURS:MINUTES` range
  - [x] Apply `--min-interval` constraint between consecutive commits
  - [x] Implement random time distribution algorithm within constraints
  - [x] Validate feasibility: check if all commits can fit within the time window with minimum interval
  - [x] Error with clear message if distribution is impossible

### 6. Generate New Git History
- **Description**: Create new Git repository with rewritten history using Git CLI commands
- **Subtasks**:
  - [x] Create new Git repository in output directory using `git init`
  - [x] Apply commits with new author info and dates using `git commit --date=` and `--author=`
  - [x] Preserve commit messages and content
  - [x] Handle branch structure (default branch or all branches via `git branch`)
  - [x] Output compact comparison: `OLD_SHA --> NEW_SHA (author: old@new, date: old_date --> new_date)` when `-q` is NOT set
  - [x] At the end, display summary: `Commits: X   Days: Y`

**Implementation Note**: Create branches after all commits are processed, mapping old SHA to new SHA.

### 7. Error Handling and Validation
- **Description**: Add comprehensive error handling using Git CLI exit codes and output parsing
- **Subtasks**:
  - [x] Validate input directory exists and is a Git repo (check `.git/` directory or `git status`)
  - [x] Validate output directory (create if needed, clear if not empty)
  - [x] Handle date range validation (`--date-from` < `--date-to`)
  - [x] Parse Git CLI error messages for meaningful error reports
  - [x] Display formatted error messages for all failure cases
  - [x] Display summary for output repository after processing

**Implementation Note**: Use `github.com/fatih/color` package for colored output when possible.

## Testing

### 8. Write Unit Tests
- **Description**: Test individual components (exclude Git CLI interaction for E2E)
- **Subtasks**:
  - [x] Test argument parsing
  - [x] Test date calculation algorithm
  - [x] Test commit metadata parsing (from Git CLI output format)
  - [x] Test random distribution function

### 9. Write Integration Tests
- **Description**: Test full workflow using Git CLI commands for setup and verification
- **Subtasks**:
  - [x] Create test repository programmatically using Git CLI
  - [x] Test full rewriting pipeline
  - [x] Verify author changes via `git log --format=`
  - [x] Verify date redistribution via `git log --format=`
  - [x] Test with various flag combinations

### 10. Write End-to-End Tests
- **Description**: Test the complete CLI tool with real Git repositories
- **Subtasks**:
  - [x] Create test repository structure in `testdata/` folder:
    - [x] `project-simple/` - 3-5 commits on single branch
    - [x] `project-multibranch/` - multiple branches
    - [x] `project-older-first/` - commits in non-chronological order
    - [x] `project-large/` - 50+ commits
    - [x] `project-time-constrained/` - test for feasibility checking with tight time slots
  - [x] Each test case should have pre-generated Git history with varied commit dates
  - [x] E2E tests should run CLI tool on real repositories (no Docker)
  - [x] Tests should verify output via Git CLI commands (`git log`, `git show`)
  - [x] Test with various flag combinations:
    - [x] Basic author/date rewriting
    - [x] With `--time-from` and `--time-to` time slots
    - [x] With `--min-interval` constraint
    - [x] Impossible distribution scenarios (should error)
    - [x] With `-q` flag for compact output
  - [x] Test edge cases (empty repo, single commit, all dates equal)
  - [x] Test error handling scenarios

**Implementation Note**: Use Go's `testing` package with `os/exec` for running CLI tool in tests.

## Documentation

### 11. Write Usage Documentation
- **Description**: Document how to use the CLI tool
- **Subtasks**:
  - [ ] Create README.md with installation instructions
  - [ ] Document all command-line flags with examples:
    - [ ] Basic flags (`-i`, `-o`, `--user-name`, `--user-email`)
    - [ ] Date range flags (`--date-from`, `--date-to`)
    - [ ] Time slot flags (`--time-from`, `--time-to`, `--min-interval`)
    - [ ] Output control flags (`-q`)
    - [ ] `--help` for formatted help message
  - [ ] Include usage examples for each flag combination
  - [ ] Explain the random date distribution behavior
  - [ ] Explain time slot and minimum interval constraints
  - [ ] Add troubleshooting guide for common Git errors
  - [ ] Document feasibility checking and error messages
  - [ ] Document compact output format when `-q` is NOT used
  - [ ] Document help message format with table layout

##_future_todo