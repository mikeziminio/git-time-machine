# Git Time Machine

A CLI tool that rewrites Git history by modifying author names/emails and redistributing commits across custom date ranges.

## Features

- **Author rewriting**: Change author name and email for all commits
- **Date redistribution**: Spread commits across custom date ranges
- **Time slot constraints**: Restrict commits to specific hours of the day
- **Minimum interval**: Enforce minimum time between consecutive commits
- **Multiple branches**: Preserve branch structure
- **Compact output**: Optional quiet mode with simple commit mapping

## Installation

```bash
go build -o git-time-machine ./cmd/main.go
```

Or download the binary:

```bash
# Clone the repository
git clone https://github.com/yourusername/git-time-machine.git
cd git-time-machine
go build -o git-time-machine ./cmd/main.go
```

## Usage

```bash
git-time-machine [flags]
```

### Required Flags

| Flag | Description |
|------|-------------|
| `-i, --input` | Input Git repository directory (required) |
| `-o, --output` | Output directory for rewritten repository (required) |

### Optional Flags

| Flag | Description |
|------|-------------|
| `--user-name` | New author name for all commits |
| `--user-email` | New author email for all commits |
| `--date-from` | Start date for rewriting (format: `2006-01-02` or `2006-01-02T15:04:05`) |
| `--date-to` | End date for rewriting (format: `2006-01-02` or `2006-01-02T15:04:05`) |
| `--time-from` | Start time for time slot filtering (format: `9`, `09`, `09:00`, `23:50`) |
| `--time-to` | End time for time slot filtering (format: `9`, `09`, `09:00`, `23:50`) |
| `--min-interval` | Minimum interval between commits in hours (integer, e.g., `1`, `2`, `3`) |
| `-q, --quiet` | Quiet mode (compact output only) |
| `--help` | Display help message |

## Examples

### Basic Author Rewrite

Change the author name for all commits:

```bash
git-time-machine -i /path/to/input/repo -o /path/to/output -u "New Author"
```

### Author and Email Rewrite

Change both author name and email:

```bash
git-time-machine -i /path/to/input/repo -o /path/to/output \
  --user-name "John Doe" \
  --user-email "john@example.com"
```

### Date Range Restriction

Rewrite all commits to a specific date range:

```bash
git-time-machine -i /path/to/input/repo -o /path/to/output \
  --date-from 2023-01-01 \
  --date-to 2023-12-31
```

### Time Slot Constraint

Restrict commits to business hours (9 AM to 5 PM):

```bash
git-time-machine -i /path/to/input/repo -o /path/to/output \
  --date-from 2023-06-01 \
  --date-to 2023-06-10 \
  --time-from 9 \
  --time-to 17
```

### With Minimum Interval

Enforce at least 2 hours between commits:

```bash
git-time-machine -i /path/to/input/repo -o /path/to/output \
  --date-from 2023-01-01 \
  --date-to 2023-01-03 \
  --min-interval 2
```

### Combined Example

All features together:

```bash
git-time-machine -i /path/to/input/repo -o /path/to/output \
  --user-name "John Doe" \
  --user-email "john@example.com" \
  --date-from 2023-01-01 \
  --date-to 2023-06-30 \
  --time-from 9 \
  --time-to 18 \
  --min-interval 1
```

### Quiet Mode

Compact output showing only the commit mapping:

```bash
git-time-machine -i /path/to/input/repo -o /path/to/output -q
```

## Date Format

- **Date-only**: `2006-01-02`
- **Date with time**: `2006-01-02T15:04:05`

## Time Format

- **Hour only**: `9`, `10`, `23` (treated as `09:00`, `10:00`, `23:00`)
- **Full time**: `09:00`, `14:30`, `23:50`

If `--time-to` is not specified, it defaults to `23` (23:59).

## Random Distribution

Commits are distributed across the date range using a semi-random algorithm that:

1. Spreads commits evenly across the available time
2. Respects the minimum interval constraint
3. Applies time slot constraints (if specified)
4. Ensures chronological order

## Output Format

By default, the tool shows:

1. **Input repository summary**: Number of commits and unique days
2. **Original commits**: Full details of each commit
3. **Commit mapping**: `OLD_SHA --> NEW_SHA (author: old_author, date: old_date --> new_date)`
4. **Output repository summary**: Number of commits and unique days

In quiet mode (`-q`), only the summary is shown.

## Error Handling

### Invalid Input Path

```bash
$ git-time-machine -i /nonexistent/path -o /tmp/output
Error: failed to read repository: failed to list branches: chdir /nonexistent/path: no such file or directory
```

### Date Range Validation

```bash
$ git-time-machine -i ./input -o ./output --date-from 2023-01-02 --date-to 2023-01-01
Error: date-from must be before date-to
```

### Impossible Distribution

```bash
$ git-time-machine -i ./input -o ./output \
  --date-from 2023-01-01 --date-to 2023-01-02 \
  --min-interval 24
Error: impossible to distribute 5 commits within 24.00 hours with minimum interval of 24 hours
```

### Time Slot Validation

```bash
$ git-time-machine -i ./input -o ./output --time-from 18 --time-to 9
Error: --time-from must be before --time-to
```

## Feasibility Checking

The tool validates that the specified constraints can accommodate all commits:

- **Date range + min-interval**: Ensures enough time exists for all commits with the minimum interval
- **Time slot**: Ensures the time slot allows valid times for all commits
- **Output**: Errors with clear messages when distribution is impossible

## Branch Preservation

All branches from the input repository are created in the output repository. The commit SHAs change due to rewriting, but the branch structure is preserved.

## Caveats

- This tool rewrites Git history and should only be used on repositories that are not shared with others
- Always backup your repository before using this tool
- Changed commit SHAs affect all dependents (branch references, tags, etc.)
- The tool creates a new repository; the original remains unchanged

## License

MIT License - See LICENSE file for details

## Author

Created by [Your Name]
