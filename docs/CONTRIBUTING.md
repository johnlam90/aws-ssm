# Contributing to AWS SSM CLI

Thank you for your interest in contributing to aws-ssm! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Release Process](#release-process)

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please be respectful and constructive in all interactions.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Set up the development environment
4. Create a new branch for your changes
5. Make your changes
6. Test your changes
7. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.24 or later
- Git
- Make (optional but recommended)
- golangci-lint (for linting)

### Clone and Build

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/aws-ssm.git
cd aws-ssm

# Add upstream remote
git remote add upstream https://github.com/johnlam90/aws-ssm.git

# Install dependencies
go mod download

# Build the project
make build

# Or without make
go build -o aws-ssm .
```

### Install Development Tools

```bash
# Install golangci-lint (macOS)
brew install golangci-lint

# Install golangci-lint (Linux)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

## Making Changes

### Branch Naming

Use descriptive branch names:

- `feature/add-xyz` - For new features
- `fix/issue-123` - For bug fixes
- `docs/update-readme` - For documentation changes
- `refactor/improve-xyz` - For refactoring

### Code Style

- Follow standard Go conventions
- Run `go fmt` before committing
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions small and focused

### Commit Messages

Write clear, descriptive commit messages:

```
Add fuzzy finder support for interfaces command

- Integrate go-fuzzyfinder library
- Update interfaces command to use fuzzy finder when no args
- Add documentation for new feature

Fixes #123
```

Format:
- First line: Brief summary (50 chars or less)
- Blank line
- Detailed description (wrap at 72 chars)
- Reference issues/PRs if applicable

## Testing

### Run Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests for specific package
go test -v ./pkg/aws/...
```

### Run Linter

```bash
# Run golangci-lint
make lint

# Or directly
golangci-lint run ./...
```

### Run All Checks

```bash
# Format, vet, lint, and test
make verify
```

### Manual Testing

Before submitting, test your changes manually:

```bash
# Build the binary
make build

# Test the command
./aws-ssm list
./aws-ssm session
./aws-ssm interfaces
./aws-ssm version
```

## Submitting Changes

### Before Submitting

1. **Update your branch** with the latest upstream changes:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all checks**:
   ```bash
   make verify
   ```

3. **Update documentation** if needed:
   - README.md
   - QUICK_REFERENCE.md
   - Relevant documentation files

4. **Add tests** for new features or bug fixes

5. **Update CHANGELOG.md** with your changes

### Pull Request Process

1. Push your changes to your fork:
   ```bash
   git push origin feature/your-feature
   ```

2. Create a pull request on GitHub

3. Fill out the PR template with:
   - Description of changes
   - Related issues
   - Testing performed
   - Screenshots (if UI changes)

4. Wait for review and address feedback

5. Once approved, a maintainer will merge your PR

### Pull Request Guidelines

- Keep PRs focused on a single feature or fix
- Include tests for new functionality
- Update documentation as needed
- Ensure CI checks pass
- Respond to review comments promptly
- Squash commits if requested

## Release Process

Releases are managed by maintainers. The process is:

1. **Update version** in `pkg/version/version.go`

2. **Update CHANGELOG.md** with release notes

3. **Run release preparation**:
   ```bash
   ./scripts/prepare-release.sh 0.1.0
   ```

4. **Create and push tag**:
   ```bash
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin v0.1.0
   ```

5. **GitHub Actions** automatically:
   - Runs tests
   - Builds binaries for all platforms
   - Creates GitHub release
   - Uploads release artifacts
   - Generates Homebrew formula

## Project Structure

```
aws-ssm/
├── cmd/                    # Command implementations
│   ├── root.go            # Root command
│   ├── list.go            # List command
│   ├── session.go         # Session command
│   ├── port_forward.go    # Port forward command
│   ├── interfaces.go      # Interfaces command
│   └── version.go         # Version command
├── pkg/                   # Packages
│   ├── aws/               # AWS client and logic
│   │   ├── client.go      # AWS client
│   │   ├── instance.go    # Instance operations
│   │   ├── session.go     # Session operations
│   │   ├── fuzzy.go       # Fuzzy finder
│   │   ├── command.go     # Command execution
│   │   └── interfaces.go  # Network interfaces
│   └── version/           # Version information
├── .github/               # GitHub Actions workflows
│   └── workflows/
│       ├── ci.yml         # CI workflow
│       └── release.yml    # Release workflow
├── scripts/               # Build and release scripts
├── docs/                  # Documentation
├── Makefile              # Build automation
├── go.mod                # Go module definition
└── README.md             # Main documentation
```

## Getting Help

- **Issues**: Open an issue on GitHub for bugs or feature requests
- **Discussions**: Use GitHub Discussions for questions
- **Documentation**: Check the docs/ directory

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (see LICENSE file).

## Recognition

Contributors will be recognized in:
- GitHub contributors page
- Release notes (for significant contributions)
- CHANGELOG.md (for notable features)

Thank you for contributing to aws-ssm! 🎉

