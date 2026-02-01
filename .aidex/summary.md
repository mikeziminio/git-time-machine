## Architecture
# Git Time Machine Architecture

## Overview
CLI tool for rewriting Git history - modifying author names, emails, and redistributing commits across custom date ranges with randomized timing.

## Core Components

### 1. CLI Parser
- Parses flags: `-i`, `-o`, `--user-name`, `--user-email`, `--date-from`, `--date-to`, `--hour-from`, `--hour-to`, `--min-hour-interval`
- Validates date ranges and time slots
- Feasibility check before processing

### 2. Git Reader
- Reads all branches via `git for-each-ref`
- Parses commit history with DAG structure
- Extracts: SHA, author, date, message, parents, branch info

### 3. Date Distributor
- Calculates new date ranges with `--hour-from`/`--hour-to` constraints
- Applies `--min-hour-interval` between consecutive commits
- Validates feasibility: ensures all commits fit within time window
- Random distribution preserving chronological order

### 4. Git Rewriter
- Creates new repository with `git init`
- Processes commits in topological order
- Uses `git commit --date=` and `--author=` for rewriting
- Preserves merge-parents structure (hybrid approach)
- Recreates all branches with correct commit references

## Processing Strategy
1. Parse input repo structure (all branches + DAG)
2. Calculate new dates with constraints
3. Rewrite commits in topological order
4. Recreate branches pointing to new commits
5. Handle merge commits with preserved parent references
