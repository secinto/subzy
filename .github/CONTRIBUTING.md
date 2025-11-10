# Contributing to Subzy

Thank you for your interest in contributing to Subzy! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Pre-commit Hooks](#pre-commit-hooks)
- [Testing Guidelines](#testing-guidelines)
- [Logging Guidelines](#logging-guidelines)
- [Code Style](#code-style)
- [Pull Request Process](#pull-request-process)

---

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/subzy.git`
3. Add upstream remote: `git remote add upstream https://github.com/LukaSikic/subzy.git`
4. Create a feature branch: `git checkout -b feature/your-feature-name`

---

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git
- (Optional) golangci-lint for linting
- (Optional) pre-commit for automated checks

### Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install pre-commit (Python)
pip install pre-commit
# OR on macOS
brew install pre-commit
```

### Build

```bash
# Build the binary
make build
# OR
go build -o subzy main.go

# Run tests
make test
# OR
go test ./...

# Run with coverage
make test-coverage
```

---

## Pre-commit Hooks

We use pre-commit hooks to ensure code quality before commits.

### Installation

```bash
# Install pre-commit hooks
pre-commit install

# Test hooks on all files
pre-commit run --all-files
```

### What the Hooks Do

1. **go-fmt** - Formats Go code with `gofmt`
2. **go-imports** - Organizes imports with `goimports`
3. **go-vet** - Runs static analysis
4. **go-mod-tidy** - Ensures go.mod/go.sum are clean
5. **go-test-short** - Runs quick tests
6. **golangci-lint** - Comprehensive linting (if installed)
7. **check-added-large-files** - Prevents large files (>1MB)
8. **check-merge-conflict** - Detects merge conflict markers
9. **trailing-whitespace** - Removes trailing whitespace
10. **end-of-file-fixer** - Ensures files end with newline
11. **check-yaml/json** - Validates YAML/JSON files
12. **detect-private-key** - Prevents committing secrets

### Skip Hooks (Emergency Only)

```bash
# Skip pre-commit hooks (not recommended)
git commit --no-verify -m "your message"

# Skip specific hook
SKIP=go-test-short git commit -m "your message"
```

---

## Testing Guidelines

### Test Coverage

- **Target**: 70%+ overall coverage
- **Requirement**: All new code must have tests
- **Run coverage**: `make test-coverage`

### Writing Tests

#### Unit Tests

```go
// runner/example_test.go
package runner

import "testing"

func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "normal case",
            input:    "test",
            expected: "expected",
        },
        // Add more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := FunctionName(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

#### Integration Tests

```go
func TestIntegrationScan(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Create test server
    ts := httptest.NewServer(...)
    defer ts.Close()

    // Run test
    config := &Config{...}
    err := Process(context.Background(), config)

    // Assertions
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

### Running Tests

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# With race detector
go test -race ./...

# Verbose output
go test -v ./...

# Short tests only (skip integration)
go test -short ./...

# Specific package
go test ./runner

# Specific test
go test ./runner -run TestFunctionName
```

### Benchmarks

```bash
# Run benchmarks
go test -bench=. ./runner

# With memory stats
go test -bench=. -benchmem ./runner

# Compare benchmarks
go test -bench=. ./runner > old.txt
# Make changes
go test -bench=. ./runner > new.txt
benchstat old.txt new.txt
```

---

## Logging Guidelines

### Using Structured Logging

We use `zerolog` for structured logging. **Never use `fmt.Println` in production code.**

```go
// Import
import "github.com/rs/zerolog/log"

// Info level
logger.Info().
    Str("subdomain", subdomain).
    Int("status_code", statusCode).
    Msg("Checking subdomain")

// Debug level
logger.Debug().
    Str("url", url).
    Dur("duration", duration).
    Msg("HTTP request completed")

// Warning
logger.Warn().
    Str("subdomain", subdomain).
    Err(err).
    Msg("Retrying after error")

// Error
logger.Error().
    Str("subdomain", subdomain).
    Str("engine", engine).
    Msg("Vulnerability detected")

// With additional context
logger.Info().
    Str("subdomain", subdomain).
    Int("attempt", attempt).
    Dur("backoff", backoff).
    Bool("success", success).
    Msg("Operation result")
```

### Log Levels

- **debug**: Detailed diagnostic information
- **info**: General informational messages
- **warn**: Warning messages (non-critical issues)
- **error**: Error messages (failures)

### When to Log

- **DO**: Log important state changes, errors, warnings
- **DO**: Log with context (subdomain, engine, status)
- **DO**: Use appropriate log levels
- **DON'T**: Log in tight loops without throttling
- **DON'T**: Log sensitive information (passwords, tokens)
- **DON'T**: Use `fmt.Println` in production code

---

## Code Style

### Go Style Guide

Follow the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) and [Effective Go](https://golang.org/doc/effective_go.html).

### Key Points

1. **Formatting**: Use `gofmt` (enforced by pre-commit)
2. **Imports**: Use `goimports` (enforced by pre-commit)
3. **Names**: Use camelCase for private, PascalCase for exported
4. **Comments**: Comment exported functions/types
5. **Errors**: Wrap errors with context
6. **Concurrency**: Use channels and goroutines appropriately

### Example

```go
// Good
func ProcessSubdomain(ctx context.Context, subdomain string) error {
    logger.Debug().Str("subdomain", subdomain).Msg("Processing")

    result, err := checkVulnerability(ctx, subdomain)
    if err != nil {
        return fmt.Errorf("processing %s: %w", subdomain, err)
    }

    return nil
}

// Bad
func process_subdomain(subdomain string) error {
    fmt.Println("Processing:", subdomain)  // Don't use fmt.Println
    result, err := checkVulnerability(subdomain)  // Missing context
    if err != nil {
        return err  // No context wrapping
    }
    return nil
}
```

---

## Pull Request Process

### Before Submitting

1. **Run pre-commit checks**: `pre-commit run --all-files`
2. **Run all tests**: `make test`
3. **Check coverage**: `make test-coverage`
4. **Run linter**: `make lint` or `golangci-lint run`
5. **Update documentation** if adding features
6. **Add/update tests** for your changes

### PR Checklist

- [ ] Tests added/updated
- [ ] Test coverage maintained/improved
- [ ] Documentation updated
- [ ] CHANGELOG.md updated (for user-facing changes)
- [ ] Pre-commit hooks pass
- [ ] All CI checks pass
- [ ] No merge conflicts
- [ ] Commits are meaningful and well-formatted

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add retry logic with exponential backoff
fix: resolve race condition in worker pool
docs: update README with new flags
test: add integration tests for DNS checking
perf: optimize fingerprint matching with Aho-Corasick
refactor: extract logger initialization to separate function
chore: update dependencies
```

### PR Title Format

```
[Type] Brief description

Examples:
[Feature] Add Graylog logging integration
[Fix] Resolve memory leak in worker pool
[Docs] Update installation instructions
[Performance] Optimize fingerprint matching
```

### PR Description Template

```markdown
## Description
Brief description of what this PR does.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
Describe the tests you ran and how to reproduce.

## Checklist
- [ ] Pre-commit hooks pass
- [ ] Tests pass locally
- [ ] Added/updated tests
- [ ] Updated documentation
- [ ] Updated CHANGELOG.md
```

---

## Development Workflow

### Daily Development

```bash
# Update from upstream
git fetch upstream
git rebase upstream/master

# Create feature branch
git checkout -b feature/my-feature

# Make changes and test
go test ./...

# Commit (triggers pre-commit hooks)
git commit -m "feat: my feature"

# Push to your fork
git push origin feature/my-feature

# Create PR on GitHub
```

### Code Review

- Address all review comments
- Keep the PR focused and small
- Respond to feedback promptly
- Update the PR with requested changes

---

## Questions?

- Check existing issues: https://github.com/LukaSikic/subzy/issues
- Read documentation: README.md, IMPLEMENTATION_PLAN.md
- Ask in PR comments

---

## License

By contributing, you agree that your contributions will be licensed under the GPLv2 License.
