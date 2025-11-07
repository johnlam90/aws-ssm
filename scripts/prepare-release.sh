#!/bin/bash

# Release preparation script for aws-ssm
# This script verifies that everything is ready for a release

set -e

VERSION=${1:-"0.1.0"}
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "AWS SSM Release Preparation"
echo "Version: $VERSION"
echo "========================================="
echo ""

# Function to print success
success() {
    echo -e "${GREEN}✓${NC} $1"
}

# Function to print error
error() {
    echo -e "${RED}✗${NC} $1"
}

# Function to print warning
warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    error "Not in the project root directory"
    exit 1
fi
success "In project root directory"

# Check if git is clean
if ! git diff-index --quiet HEAD --; then
    error "Working directory has uncommitted changes"
    echo "  Please commit or stash your changes before releasing"
    exit 1
fi
success "Working directory is clean"

# Check if on main branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$CURRENT_BRANCH" != "main" ]; then
    warning "Not on main branch (current: $CURRENT_BRANCH)"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    success "On main branch"
fi

# Check if tag already exists
if git tag | grep -q "^v${VERSION}$"; then
    error "Tag v${VERSION} already exists"
    exit 1
fi
success "Tag v${VERSION} is available"

# Check Go version
GO_VERSION=$(go version | awk '{print $3}')
if [[ ! "$GO_VERSION" =~ go1\.2[4-9] ]]; then
    warning "Go version is $GO_VERSION (recommended: go1.24+)"
else
    success "Go version is $GO_VERSION"
fi

# Check if dependencies are up to date
echo ""
echo "Checking dependencies..."
go mod download
go mod verify
success "Dependencies verified"

# Run tests
echo ""
echo "Running tests..."
if go test -v -race ./... > /tmp/test-output.log 2>&1; then
    success "All tests passed"
else
    error "Tests failed"
    cat /tmp/test-output.log
    exit 1
fi

# Run go vet
echo ""
echo "Running go vet..."
if go vet ./... > /tmp/vet-output.log 2>&1; then
    success "go vet passed"
else
    error "go vet failed"
    cat /tmp/vet-output.log
    exit 1
fi

# Run golangci-lint if available
echo ""
if command -v golangci-lint &> /dev/null; then
    echo "Running golangci-lint..."
    if golangci-lint run ./... > /tmp/lint-output.log 2>&1; then
        success "golangci-lint passed"
    else
        warning "golangci-lint found issues"
        cat /tmp/lint-output.log
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
else
    warning "golangci-lint not installed (recommended: brew install golangci-lint)"
fi

# Check if version is updated in version.go
echo ""
echo "Checking version in code..."
if grep -q "Version = \"$VERSION\"" pkg/version/version.go; then
    success "Version $VERSION found in pkg/version/version.go"
else
    error "Version $VERSION not found in pkg/version/version.go"
    echo "  Please update the Version variable in pkg/version/version.go"
    exit 1
fi

# Check if CHANGELOG.md is updated
echo ""
echo "Checking CHANGELOG.md..."
if grep -q "\[$VERSION\]" CHANGELOG.md; then
    success "Version $VERSION found in CHANGELOG.md"
else
    warning "Version $VERSION not found in CHANGELOG.md"
    echo "  Please update CHANGELOG.md with release notes"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Build for all platforms
echo ""
echo "Building for all platforms..."
if make build-all > /tmp/build-output.log 2>&1; then
    success "All platform builds succeeded"
    echo ""
    echo "Build artifacts:"
    ls -lh dist/
else
    error "Build failed"
    cat /tmp/build-output.log
    exit 1
fi

# Test the binary
echo ""
echo "Testing binary..."
if ./aws-ssm version > /tmp/version-output.log 2>&1; then
    success "Binary works"
    echo ""
    cat /tmp/version-output.log
else
    error "Binary test failed"
    cat /tmp/version-output.log
    exit 1
fi

# Summary
echo ""
echo "========================================="
echo "Release Preparation Summary"
echo "========================================="
echo ""
success "All checks passed!"
echo ""
echo "Next steps:"
echo "  1. Review the changes one more time"
echo "  2. Create and push the tag:"
echo "     git tag -a v${VERSION} -m \"Release v${VERSION}\""
echo "     git push origin v${VERSION}"
echo "  3. GitHub Actions will automatically:"
echo "     - Run tests"
echo "     - Build binaries for all platforms"
echo "     - Create a GitHub release"
echo "     - Upload release artifacts"
echo "     - Generate Homebrew formula"
echo ""
echo "Or use the Makefile:"
echo "  make tag VERSION=${VERSION}"
echo "  git push origin v${VERSION}"
echo ""

