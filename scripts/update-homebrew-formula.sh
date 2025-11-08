#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# Script to update the Homebrew formula for aws-ssm.
# Usage:
#   ./scripts/update-homebrew-formula.sh [--auto|-y] [--force] [--tap <tap_repo>] <version>
# Example:
#   ./scripts/update-homebrew-formula.sh 0.2.0
#   ./scripts/update-homebrew-formula.sh --auto --force --tap johnlam90/homebrew-tap 0.2.1

AUTO_CONFIRM="${UPDATE_HOMEBREW_AUTO_CONFIRM:-}"  # env var shortcut
FORCE_RESET="false"
TAP_REPO="johnlam90/homebrew-tap"
VERSION=""

# Colors (disabled if NO_COLOR set or non-TTY)
if [[ -n "${NO_COLOR:-}" || ! -t 1 ]]; then
  RED=""; GREEN=""; YELLOW=""; NC=""
else
  RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
fi

die() { echo -e "${RED}Error: $*${NC}" >&2; exit 1; }
info() { echo -e "${YELLOW}$*${NC}"; }
success() { echo -e "${GREEN}$*${NC}"; }
warn() { echo -e "${YELLOW}Warning: $*${NC}"; }

# Dependency check
need() { command -v "$1" >/dev/null 2>&1 || die "Missing required dependency: $1"; }
for dep in curl git grep awk; do need "$dep"; done

# Parse arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    -y|--auto) AUTO_CONFIRM="1"; shift ;;
    --force) FORCE_RESET="true"; shift ;;
    --tap) TAP_REPO="${2:-}"; [[ -z "$TAP_REPO" ]] && die "--tap requires a value"; shift 2 ;;
    -h|--help)
      cat <<EOF
Usage: $0 [options] <version>
Options:
  -y, --auto        Auto confirm commit/push (non-interactive)
  --force           Force reset local tap clone (git reset --hard origin/main)
  --tap <repo>      Override tap repo (default: johnlam90/homebrew-tap)
  -h, --help        Show this help
Environment:
  UPDATE_HOMEBREW_AUTO_CONFIRM=1  Auto confirm
  NO_COLOR=1                      Disable colored output
  GITHUB_TOKEN=...                Use token for GitHub API (higher rate limits)
EOF
      exit 0
      ;;
    --*) die "Unknown flag: $1" ;;
    -*) die "Unknown short flag: $1" ;;
    *) VERSION="$1"; shift ;;
  esac
done

[[ -z "$VERSION" ]] && die "Version number required (e.g. 0.2.0)"
[[ $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "Invalid version format '$VERSION' (expected semver: X.Y.Z)"
TAG="v${VERSION}"

success "==> Updating Homebrew formula to version ${VERSION} (tag ${TAG})"
echo

API_URL="https://api.github.com/repos/johnlam90/aws-ssm/releases/tags/${TAG}"
AUTH_HEADER=""
if [[ -n "${GITHUB_TOKEN:-}" ]]; then
  AUTH_HEADER="Authorization: token ${GITHUB_TOKEN}"
fi

info "Checking if release ${TAG} exists..."
if ! curl -sf -H "$AUTH_HEADER" "$API_URL" > /dev/null; then
  die "Release ${TAG} not found. Create and push the tag first (git tag -a ${TAG} -m 'Release ${VERSION}' && git push origin ${TAG})"
fi
success "✓ Release ${TAG} found"
echo

TMP_CHECKSUM=$(mktemp /tmp/aws-ssm-checksums.XXXXXX)
trap 'rm -f "$TMP_CHECKSUM"' EXIT

info "Downloading checksums..."
CHECKSUM_URL="https://github.com/johnlam90/aws-ssm/releases/download/${TAG}/checksums.txt"
curl -sLf "$CHECKSUM_URL" -o "$TMP_CHECKSUM" || die "Failed to download checksums.txt (${CHECKSUM_URL})"
success "✓ Checksums downloaded"
echo

info "Extracting SHA256 hashes..."
DARWIN_AMD64_SHA=$(grep "darwin-amd64" "$TMP_CHECKSUM" | awk '{print $1}')
DARWIN_ARM64_SHA=$(grep "darwin-arm64" "$TMP_CHECKSUM" | awk '{print $1}')
LINUX_AMD64_SHA=$(grep "linux-amd64" "$TMP_CHECKSUM" | awk '{print $1}')
LINUX_ARM64_SHA=$(grep "linux-arm64" "$TMP_CHECKSUM" | awk '{print $1}')

validate_sha() { [[ $1 =~ ^[a-f0-9]{64}$ ]] || die "Invalid SHA256 hash: '$1'"; }
for h in "$DARWIN_AMD64_SHA" "$DARWIN_ARM64_SHA" "$LINUX_AMD64_SHA" "$LINUX_ARM64_SHA"; do
  [[ -z "$h" ]] && die "Missing one or more required hashes"
  validate_sha "$h"
done
success "✓ Hashes validated"
echo

echo "SHA256 Hashes:";
echo "  darwin-amd64: $DARWIN_AMD64_SHA"
echo "  darwin-arm64: $DARWIN_ARM64_SHA"
echo "  linux-amd64:  $LINUX_AMD64_SHA"
echo "  linux-arm64:  $LINUX_ARM64_SHA"
echo

WORKDIR=$(mktemp -d /tmp/homebrew-tap.XXXXXX || true)
[[ -z "$WORKDIR" ]] && WORKDIR="/tmp/homebrew-tap.$$" && mkdir -p "$WORKDIR"
info "Cloning tap repo '$TAP_REPO' into $WORKDIR ..."
git clone "https://github.com/${TAP_REPO}.git" "$WORKDIR" >/dev/null 2>&1 || die "Failed to clone tap repo ${TAP_REPO}"
cd "$WORKDIR"

info "Fetching latest changes..."
git fetch origin >/dev/null 2>&1
if [[ "$FORCE_RESET" == "true" ]]; then
  warn "Force reset enabled: discarding local changes"
  git reset --hard origin/main >/dev/null 2>&1 || die "git reset failed"
else
  git checkout main >/dev/null 2>&1 || die "Failed to checkout main"
  git pull --ff-only origin main >/dev/null 2>&1 || die "Failed to pull latest main"
fi
success "✓ Repository ready"
echo

EXISTING_VERSION=""
if [[ -f aws-ssm.rb ]]; then
  EXISTING_VERSION=$(grep -E 'version "[0-9]+\.[0-9]+\.[0-9]+"' aws-ssm.rb | head -1 | sed -E 's/.*version "([0-9]+\.[0-9]+\.[0-9]+)".*/\1/') || true
fi
if [[ "$EXISTING_VERSION" == "$VERSION" ]]; then
  warn "Formula already at version ${VERSION}. Nothing to do."
  echo "Location: $WORKDIR/aws-ssm.rb"
  exit 0
fi

info "Updating formula..."
cat > aws-ssm.rb <<EOF
# typed: false
# frozen_string_literal: true

class AwsSsm < Formula
  desc "Native Golang CLI tool for managing AWS SSM sessions"
  homepage "https://github.com/johnlam90/aws-ssm"
  version "${VERSION}"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/johnlam90/aws-ssm/releases/download/${TAG}/aws-ssm-darwin-amd64.tar.gz"
      sha256 "${DARWIN_AMD64_SHA}"
      def install
        bin.install "aws-ssm-darwin-amd64" => "aws-ssm"
      end
    end
    if Hardware::CPU.arm?
      url "https://github.com/johnlam90/aws-ssm/releases/download/${TAG}/aws-ssm-darwin-arm64.tar.gz"
      sha256 "${DARWIN_ARM64_SHA}"
      def install
        bin.install "aws-ssm-darwin-arm64" => "aws-ssm"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/johnlam90/aws-ssm/releases/download/${TAG}/aws-ssm-linux-amd64.tar.gz"
      sha256 "${LINUX_AMD64_SHA}"
      def install
        bin.install "aws-ssm-linux-amd64" => "aws-ssm"
      end
    end
    if Hardware::CPU.arm?
      url "https://github.com/johnlam90/aws-ssm/releases/download/${TAG}/aws-ssm-linux-arm64.tar.gz"
      sha256 "${LINUX_ARM64_SHA}"
      def install
        bin.install "aws-ssm-linux-arm64" => "aws-ssm"
      end
    end
  end

  test do
    system "
      \\#{bin}/aws-ssm", "version"
  end
end
EOF
success "✓ Formula updated"
echo

info "Changes to be committed:"
git diff aws-ssm.rb || true
echo

COMMIT_MSG="Update aws-ssm formula to v${VERSION}"
if [[ -n "$AUTO_CONFIRM" ]]; then
  info "Auto confirm enabled; committing and pushing..."
  git add aws-ssm.rb
  git commit -m "$COMMIT_MSG" >/dev/null 2>&1 || warn "Commit failed (possibly no changes)"
  git push origin main || die "Push failed"
  success "✓ Formula updated successfully (auto mode)"
else
  read -p "Commit and push changes? (y/n) " -n 1 -r; echo
  if [[ $REPLY =~ ^[Yy]$ ]]; then
    git add aws-ssm.rb
    git commit -m "$COMMIT_MSG" >/dev/null 2>&1 || warn "Commit failed (possibly no changes)"
    info "Pushing to GitHub..."
    git push origin main || die "Push failed"
    success "✓ Formula updated successfully!"
    echo "Next steps:"; echo "  1. brew update"; echo "  2. brew upgrade aws-ssm"; echo "  3. aws-ssm version"
  else
    warn "Changes not committed."
    echo "Formula file remains at: $WORKDIR/aws-ssm.rb"
  fi
fi

echo
success "Done."
