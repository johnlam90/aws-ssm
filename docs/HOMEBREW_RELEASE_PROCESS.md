# Homebrew Release Process

This document explains the complete workflow for releasing new versions of aws-ssm and how Homebrew users upgrade to those new versions.

## Table of Contents

- [Release Process Overview](#release-process-overview)
- [Homebrew Formula Update Process](#homebrew-formula-update-process)
- [User Upgrade Experience](#user-upgrade-experience)
- [Automation Opportunities](#automation-opportunities)
- [Testing and Verification](#testing-and-verification)
- [Troubleshooting](#troubleshooting)

---

## Release Process Overview

### What Happens When You Push a New Version Tag

When you push a new version tag (e.g., `v0.2.0`) to the `johnlam90/aws-ssm` repository, the following automated process occurs:

#### 1. Tag Creation and Push

```bash
# Create an annotated tag
git tag -a v0.2.0 -m "Release version 0.2.0

- Feature 1
- Feature 2
- Bug fixes"

# Push the tag to GitHub
git push origin v0.2.0
```

#### 2. GitHub Actions Release Workflow Triggered

The `.github/workflows/release.yml` workflow is automatically triggered when a tag matching `v*` is pushed.

**Workflow Steps:**

1. **Checkout Code** - Clones the repository at the tagged commit
2. **Set Up Go** - Installs Go 1.24 or later
3. **Build Binaries** - Compiles binaries for all platforms:
   - macOS Intel (darwin-amd64)
   - macOS Apple Silicon (darwin-arm64)
   - Linux amd64
   - Linux arm64
   - Windows amd64

4. **Embed Version Information** - Uses LDFLAGS to embed:
   - Version number (from git tag)
   - Git commit hash
   - Build timestamp
   - Go version

5. **Create Archives** - Packages binaries:
   - `.tar.gz` for Unix-like systems (macOS, Linux)
   - `.zip` for Windows

6. **Generate Checksums** - Creates `checksums.txt` with SHA256 hashes for all artifacts

7. **Create GitHub Release** - Publishes a new release with:
   - Auto-generated release notes
   - All binary artifacts attached
   - SHA256 checksums file

#### 3. Artifacts Published

After the workflow completes (~5-6 minutes), the following artifacts are available:

```
https://github.com/johnlam90/aws-ssm/releases/download/v0.2.0/
├── aws-ssm-darwin-amd64.tar.gz
├── aws-ssm-darwin-arm64.tar.gz
├── aws-ssm-linux-amd64.tar.gz
├── aws-ssm-linux-arm64.tar.gz
├── aws-ssm-windows-amd64.zip
└── checksums.txt
```

**Example checksums.txt:**
```
11ed1020a7b98106a1c0702f486dbbcd119214063253bd9dac99b33f422fbe86  aws-ssm-darwin-amd64.tar.gz
2abb8deefb8c5d5086afb2d5d8994f06757e89fe08ddab6b9d3d6ab792c1a64e  aws-ssm-darwin-arm64.tar.gz
a46920a62e90c94849ffcb7cc176fcc58e2f9ffdd2d2813846a31658f8aa031d  aws-ssm-linux-amd64.tar.gz
5e0a7a6f72fd5679bc124821244a9650b0b742feb2f30b5271a90cfb978c7d75  aws-ssm-linux-arm64.tar.gz
9c8e5d2f1a3b4c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d  aws-ssm-windows-amd64.zip
```

---

## Homebrew Formula Update Process

### Current State: Manual Update Required

**Important:** The Homebrew formula in the `johnlam90/homebrew-tap` repository must be manually updated after each release. This is a **required step** for users to be able to upgrade via Homebrew.

### Step-by-Step Formula Update

#### 1. Download the New Checksums

After the GitHub release is published, download the checksums:

```bash
# Download checksums from the new release
curl -sL https://github.com/johnlam90/aws-ssm/releases/download/v0.2.0/checksums.txt
```

#### 2. Clone the Homebrew Tap Repository

```bash
# Clone the tap repository
git clone https://github.com/johnlam90/homebrew-tap.git
cd homebrew-tap
```

#### 3. Update the Formula File

Edit `aws-ssm.rb` and update the following fields:

**Fields to Update:**

1. **Version number** (line 5)
2. **Download URLs** (lines 9, 17, 26, 34)
3. **SHA256 checksums** (lines 10, 18, 27, 35)

**Example Update for v0.2.0:**

```ruby
class AwsSsm < Formula
  desc "Native Golang CLI tool for managing AWS SSM sessions"
  homepage "https://github.com/johnlam90/aws-ssm"
  version "0.2.0"  # ← UPDATE THIS
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/johnlam90/aws-ssm/releases/download/v0.2.0/aws-ssm-darwin-amd64.tar.gz"  # ← UPDATE VERSION
      sha256 "NEW_SHA256_HASH_FOR_DARWIN_AMD64"  # ← UPDATE THIS

      def install
        bin.install "aws-ssm-darwin-amd64" => "aws-ssm"
      end
    end

    if Hardware::CPU.arm?
      url "https://github.com/johnlam90/aws-ssm/releases/download/v0.2.0/aws-ssm-darwin-arm64.tar.gz"  # ← UPDATE VERSION
      sha256 "NEW_SHA256_HASH_FOR_DARWIN_ARM64"  # ← UPDATE THIS

      def install
        bin.install "aws-ssm-darwin-arm64" => "aws-ssm"
      end
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://github.com/johnlam90/aws-ssm/releases/download/v0.2.0/aws-ssm-linux-amd64.tar.gz"  # ← UPDATE VERSION
      sha256 "NEW_SHA256_HASH_FOR_LINUX_AMD64"  # ← UPDATE THIS

      def install
        bin.install "aws-ssm-linux-amd64" => "aws-ssm"
      end
    end

    if Hardware::CPU.arm?
      url "https://github.com/johnlam90/aws-ssm/releases/download/v0.2.0/aws-ssm-linux-arm64.tar.gz"  # ← UPDATE VERSION
      sha256 "NEW_SHA256_HASH_FOR_LINUX_ARM64"  # ← UPDATE THIS

      def install
        bin.install "aws-ssm-linux-arm64" => "aws-ssm"
      end
    end
  end

  test do
    system "#{bin}/aws-ssm", "version"
  end
end
```

#### 4. Commit and Push the Updated Formula

```bash
# Add the updated formula
git add aws-ssm.rb

# Commit with a descriptive message
git commit -m "Update aws-ssm formula to v0.2.0"

# Push to GitHub
git push origin main
```

#### 5. Verify the Update

```bash
# Update Homebrew
brew update

# Check the formula version
brew info johnlam90/tap/aws-ssm
```

---

## User Upgrade Experience

### How Users Upgrade to New Versions

Once the Homebrew formula is updated in the tap repository, users can upgrade using standard Homebrew commands.

#### Step 1: Update Homebrew

Users first update Homebrew to fetch the latest formula definitions:

```bash
brew update
```

**What happens:**
- Homebrew fetches the latest changes from all tapped repositories
- The updated `aws-ssm.rb` formula is downloaded
- Homebrew compares the installed version with the formula version

#### Step 2: Check for Updates

Users can check if an update is available:

```bash
# Check outdated packages
brew outdated

# Check specific package info
brew info aws-ssm
```

**Example output:**
```
==> johnlam90/tap/aws-ssm: stable 0.2.0
Native Golang CLI tool for managing AWS SSM sessions
https://github.com/johnlam90/aws-ssm
Installed
/opt/homebrew/Cellar/aws-ssm/0.1.0 (4 files, 16.7MB)
  Built from source on 2025-11-07 at 16:37:16
From: https://github.com/johnlam90/homebrew-tap/blob/HEAD/aws-ssm.rb
License: MIT
==> Caveats
aws-ssm 0.2.0 is available and can be upgraded with:
  brew upgrade aws-ssm
```

#### Step 3: Upgrade aws-ssm

```bash
# Upgrade aws-ssm specifically
brew upgrade aws-ssm

# Or upgrade all outdated packages
brew upgrade
```

**What happens during upgrade:**

1. **Download** - Homebrew downloads the new version from GitHub releases
2. **Verify** - SHA256 checksum is verified against the formula
3. **Extract** - Archive is extracted to a temporary location
4. **Install** - Binary is installed to `/opt/homebrew/Cellar/aws-ssm/0.2.0/bin/`
5. **Link** - Symlink is updated: `/opt/homebrew/bin/aws-ssm` → `../Cellar/aws-ssm/0.2.0/bin/aws-ssm`
6. **Cleanup** - Old version is removed (unless `HOMEBREW_NO_INSTALL_CLEANUP=1`)

#### Step 4: Verify the Upgrade

```bash
# Check the installed version
aws-ssm version

# Expected output:
# aws-ssm version 0.2.0 (commit: abc1234, built: 2025-11-08T10:00:00-06:00, go: go1.24.10)
```

### How Homebrew Detects New Versions

Homebrew detects new versions by comparing:

1. **Installed Version** - Stored in `/opt/homebrew/Cellar/aws-ssm/<version>/`
2. **Formula Version** - Defined in the `version` field of `aws-ssm.rb`

When `brew update` is run:
- Homebrew pulls the latest formula from `johnlam90/homebrew-tap`
- Compares the formula version with installed version
- Marks the package as "outdated" if formula version > installed version

---

## Automation Opportunities

### Option 1: Automated Formula Update with GitHub Actions (Recommended)

Create a GitHub Actions workflow in the `johnlam90/aws-ssm` repository that automatically updates the Homebrew formula when a new release is published.

**Create `.github/workflows/update-homebrew-formula.yml`:**

```yaml
name: Update Homebrew Formula

on:
  release:
    types: [published]

jobs:
  update-formula:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout aws-ssm repository
        uses: actions/checkout@v4

      - name: Get release version and checksums
        id: release
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          echo "version=$VERSION" >> $GITHUB_OUTPUT
          
          # Download checksums
          curl -sL "https://github.com/johnlam90/aws-ssm/releases/download/v${VERSION}/checksums.txt" -o checksums.txt
          
          # Extract SHA256 hashes
          DARWIN_AMD64_SHA=$(grep "darwin-amd64" checksums.txt | awk '{print $1}')
          DARWIN_ARM64_SHA=$(grep "darwin-arm64" checksums.txt | awk '{print $1}')
          LINUX_AMD64_SHA=$(grep "linux-amd64" checksums.txt | awk '{print $1}')
          LINUX_ARM64_SHA=$(grep "linux-arm64" checksums.txt | awk '{print $1}')
          
          echo "darwin_amd64_sha=$DARWIN_AMD64_SHA" >> $GITHUB_OUTPUT
          echo "darwin_arm64_sha=$DARWIN_ARM64_SHA" >> $GITHUB_OUTPUT
          echo "linux_amd64_sha=$LINUX_AMD64_SHA" >> $GITHUB_OUTPUT
          echo "linux_arm64_sha=$LINUX_ARM64_SHA" >> $GITHUB_OUTPUT

      - name: Checkout homebrew-tap repository
        uses: actions/checkout@v4
        with:
          repository: johnlam90/homebrew-tap
          token: ${{ secrets.HOMEBREW_TAP_TOKEN }}
          path: homebrew-tap

      - name: Update formula
        run: |
          cd homebrew-tap
          
          # Update version
          sed -i "s/version \".*\"/version \"${{ steps.release.outputs.version }}\"/" aws-ssm.rb
          
          # Update URLs and SHA256 hashes
          sed -i "s|download/v[0-9.]\+/|download/v${{ steps.release.outputs.version }}/|g" aws-ssm.rb
          sed -i "0,/sha256 \".*\"/s//sha256 \"${{ steps.release.outputs.darwin_amd64_sha }}\"/" aws-ssm.rb
          sed -i "0,/sha256 \".*\"/s//sha256 \"${{ steps.release.outputs.darwin_arm64_sha }}\"/" aws-ssm.rb
          sed -i "0,/sha256 \".*\"/s//sha256 \"${{ steps.release.outputs.linux_amd64_sha }}\"/" aws-ssm.rb
          sed -i "0,/sha256 \".*\"/s//sha256 \"${{ steps.release.outputs.linux_arm64_sha }}\"/" aws-ssm.rb

      - name: Commit and push changes
        run: |
          cd homebrew-tap
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add aws-ssm.rb
          git commit -m "Update aws-ssm formula to v${{ steps.release.outputs.version }}"
          git push
```

**Setup Requirements:**

1. Create a Personal Access Token (PAT) with `repo` scope
2. Add it as a secret named `HOMEBREW_TAP_TOKEN` in the `johnlam90/aws-ssm` repository settings
3. The workflow will automatically update the formula when a release is published

### Option 2: Using Homebrew's `brew bump-formula-pr` (Alternative)

Homebrew provides a built-in command to create pull requests for formula updates:

```bash
# Bump the formula to a new version
brew bump-formula-pr \
  --url="https://github.com/johnlam90/aws-ssm/releases/download/v0.2.0/aws-ssm-darwin-arm64.tar.gz" \
  --sha256="NEW_SHA256_HASH" \
  johnlam90/tap/aws-ssm
```

**Limitations:**
- Only works for formulas in the official Homebrew repositories or well-known taps
- Requires manual execution for each release
- Not suitable for automated workflows

### Option 3: Manual Script for Formula Updates

Create a helper script to automate the manual update process:

**Create `scripts/update-homebrew-formula.sh`:**

```bash
#!/bin/bash
set -e

# Check if version is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 0.2.0"
  exit 1
fi

VERSION=$1
TAG="v${VERSION}"

echo "Updating Homebrew formula to version ${VERSION}..."

# Download checksums
echo "Downloading checksums..."
curl -sL "https://github.com/johnlam90/aws-ssm/releases/download/${TAG}/checksums.txt" -o /tmp/checksums.txt

# Extract SHA256 hashes
DARWIN_AMD64_SHA=$(grep "darwin-amd64" /tmp/checksums.txt | awk '{print $1}')
DARWIN_ARM64_SHA=$(grep "darwin-arm64" /tmp/checksums.txt | awk '{print $1}')
LINUX_AMD64_SHA=$(grep "linux-amd64" /tmp/checksums.txt | awk '{print $1}')
LINUX_ARM64_SHA=$(grep "linux-arm64" /tmp/checksums.txt | awk '{print $1}')

echo "SHA256 Hashes:"
echo "  darwin-amd64: $DARWIN_AMD64_SHA"
echo "  darwin-arm64: $DARWIN_ARM64_SHA"
echo "  linux-amd64:  $LINUX_AMD64_SHA"
echo "  linux-arm64:  $LINUX_ARM64_SHA"

# Clone or update homebrew-tap
if [ -d "/tmp/homebrew-tap" ]; then
  echo "Updating existing homebrew-tap clone..."
  cd /tmp/homebrew-tap
  git pull
else
  echo "Cloning homebrew-tap..."
  git clone https://github.com/johnlam90/homebrew-tap.git /tmp/homebrew-tap
  cd /tmp/homebrew-tap
fi

# Create updated formula
echo "Updating formula..."
cat > aws-ssm.rb << EOF
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
    system "#{bin}/aws-ssm", "version"
  end
end
EOF

# Show diff
echo ""
echo "Changes to be committed:"
git diff aws-ssm.rb

# Commit and push
echo ""
read -p "Commit and push changes? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  git add aws-ssm.rb
  git commit -m "Update aws-ssm formula to v${VERSION}"
  git push origin main
  echo "Formula updated successfully!"
else
  echo "Changes not committed."
fi

# Cleanup
rm /tmp/checksums.txt
```

**Usage:**

```bash
# Make the script executable
chmod +x scripts/update-homebrew-formula.sh

# Run the script with the new version
./scripts/update-homebrew-formula.sh 0.2.0
```

### Best Practices for Maintaining the Homebrew Tap

1. **Keep Formula in Sync** - Update the formula immediately after each release
2. **Test Before Pushing** - Always test the formula locally before pushing
3. **Use Semantic Versioning** - Follow semver for version numbers
4. **Document Changes** - Include release notes in git tags
5. **Automate When Possible** - Use GitHub Actions for automatic updates
6. **Monitor Issues** - Watch for user-reported installation issues
7. **Keep Checksums Secure** - Always verify SHA256 checksums match the release

---

## Testing and Verification

### Testing the Updated Formula Locally

Before pushing formula updates, test them locally to ensure they work correctly.

#### 1. Test Formula Syntax

```bash
# Audit the formula for issues
brew audit --strict johnlam90/tap/aws-ssm

# Check for style issues
brew style johnlam90/tap/aws-ssm
```

#### 2. Test Installation from Updated Formula

```bash
# Uninstall current version
brew uninstall aws-ssm

# Install from the updated formula
brew install johnlam90/tap/aws-ssm

# Verify version
aws-ssm version
```

#### 3. Test the Formula's Test Block

```bash
# Run the formula's test
brew test johnlam90/tap/aws-ssm
```

**Expected output:**
```
Testing johnlam90/tap/aws-ssm
==> /opt/homebrew/Cellar/aws-ssm/0.2.0/bin/aws-ssm version
```

#### 4. Test Upgrade Path

```bash
# Install old version first
brew install johnlam90/tap/aws-ssm

# Update Homebrew
brew update

# Upgrade to new version
brew upgrade aws-ssm

# Verify upgrade
aws-ssm version
```

### Verifying User Upgrades

After pushing the formula update, verify that users can upgrade successfully:

#### 1. Test on a Clean System

Use a Docker container or VM to test the installation:

```bash
# macOS (using Docker with Homebrew)
docker run -it homebrew/brew:latest bash

# Inside container
brew tap johnlam90/tap
brew install aws-ssm
aws-ssm version
```

#### 2. Test Upgrade from Previous Version

```bash
# Install old version
brew install johnlam90/tap/aws-ssm@0.1.0

# Update and upgrade
brew update
brew upgrade aws-ssm

# Verify
aws-ssm version
```

#### 3. Monitor GitHub Issues

Watch for user-reported issues:
- Installation failures
- Checksum mismatches
- Architecture-specific problems
- Version detection issues

### Automated Testing with GitHub Actions

Add a test workflow to the `homebrew-tap` repository:

**Create `.github/workflows/test-formula.yml`:**

```yaml
name: Test Formula

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    strategy:
      matrix:
        os: [macos-latest, ubuntu-latest]
    runs-on: ${{ matrix.os }}

    steps:
      - name: Set up Homebrew
        uses: Homebrew/actions/setup-homebrew@master

      - name: Tap repository
        run: brew tap johnlam90/tap

      - name: Install aws-ssm
        run: brew install aws-ssm

      - name: Test installation
        run: |
          which aws-ssm
          aws-ssm version

      - name: Run formula test
        run: brew test aws-ssm

      - name: Audit formula
        run: brew audit --strict johnlam90/tap/aws-ssm
```

---

## Troubleshooting

### Common Issues During Formula Updates

#### Issue 1: SHA256 Checksum Mismatch

**Error:**
```
Error: SHA256 mismatch
Expected: abc123...
  Actual: def456...
```

**Solution:**
1. Verify you downloaded the correct checksums.txt from the release
2. Ensure the version number in URLs matches the release tag
3. Re-download checksums and update the formula

#### Issue 2: Formula Not Updating

**Error:** Users report `brew upgrade` doesn't show new version

**Solution:**
1. Verify the formula was pushed to the tap repository
2. Check that the version number was updated in the formula
3. Ask users to run `brew update` first
4. Clear Homebrew cache: `rm -rf $(brew --cache)`

#### Issue 3: Download URL 404

**Error:**
```
Error: Failed to download resource "aws-ssm"
Download failed: https://github.com/johnlam90/aws-ssm/releases/download/v0.2.0/aws-ssm-darwin-arm64.tar.gz
```

**Solution:**
1. Verify the GitHub release was published successfully
2. Check that the release tag matches the URL (e.g., `v0.2.0` not `0.2.0`)
3. Ensure the repository is public or the formula includes authentication

#### Issue 4: Binary Not Executable

**Error:** `Permission denied` when running aws-ssm

**Solution:**
Update the formula to ensure the binary is executable:

```ruby
def install
  bin.install "aws-ssm-darwin-arm64" => "aws-ssm"
  # Ensure binary is executable
  chmod 0755, bin/"aws-ssm"
end
```

### Getting Help

If you encounter issues:

1. **Check GitHub Actions Logs** - Review the release workflow logs
2. **Test Locally** - Install the formula on your machine first
3. **Review Homebrew Docs** - https://docs.brew.sh/Formula-Cookbook
4. **Ask for Help** - Open an issue in the homebrew-tap repository

---

## Complete Release Checklist

Use this checklist for each new release:

### Pre-Release

- [ ] Update CHANGELOG.md with new version and changes
- [ ] Update version in code if needed
- [ ] Run tests locally: `make test`
- [ ] Run linter: `make lint`
- [ ] Build locally: `make build`
- [ ] Commit all changes

### Release

- [ ] Create annotated git tag: `git tag -a v0.2.0 -m "Release v0.2.0"`
- [ ] Push tag: `git push origin v0.2.0`
- [ ] Wait for GitHub Actions release workflow to complete (~5-6 minutes)
- [ ] Verify release on GitHub: https://github.com/johnlam90/aws-ssm/releases
- [ ] Verify all artifacts are present (5 binaries + checksums.txt)

### Post-Release (Manual Formula Update)

- [ ] Download checksums.txt from the release
- [ ] Clone homebrew-tap repository
- [ ] Update formula version number
- [ ] Update download URLs with new version
- [ ] Update SHA256 checksums for all platforms
- [ ] Test formula locally: `brew audit --strict johnlam90/tap/aws-ssm`
- [ ] Commit and push formula changes
- [ ] Test installation: `brew upgrade aws-ssm`
- [ ] Verify version: `aws-ssm version`

### Post-Release (Automated Formula Update)

- [ ] Verify GitHub Actions workflow triggered in homebrew-tap repository
- [ ] Check that formula was updated automatically
- [ ] Test installation: `brew upgrade aws-ssm`
- [ ] Verify version: `aws-ssm version`

### Verification

- [ ] Test on macOS Intel (if available)
- [ ] Test on macOS Apple Silicon
- [ ] Test on Linux (optional)
- [ ] Update documentation if needed
- [ ] Announce release (optional)

---

## Summary

The complete release workflow is:

1. **Developer** pushes a version tag → GitHub Actions builds and publishes release
2. **Maintainer** updates Homebrew formula (manually or automatically)
3. **Users** run `brew update && brew upgrade aws-ssm` to get the new version

**Key Points:**

- ✅ GitHub Actions automates binary building and release publishing
- ⚠️ Homebrew formula updates are currently manual (can be automated)
- ✅ Users upgrade with standard Homebrew commands
- ✅ SHA256 checksums ensure download integrity
- ✅ Multi-architecture support for macOS and Linux

**Recommended:** Implement automated formula updates using GitHub Actions (Option 1) to streamline the release process and reduce manual work.

