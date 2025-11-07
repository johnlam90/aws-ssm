# Release Checklist for v0.1.0

This document outlines the steps to create and publish a release.

## Pre-Release Checklist

### Code Quality

- [ ] All tests pass (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Code is formatted (`make fmt`)
- [ ] Go vet passes (`make vet`)
- [ ] All verification checks pass (`make verify`)

### Documentation

- [ ] README.md is up to date
- [ ] CHANGELOG.md has release notes for v0.1.0
- [ ] INSTALLATION.md has correct installation instructions
- [ ] QUICK_REFERENCE.md reflects all commands
- [ ] All feature documentation is complete:
  - [ ] FUZZY_FINDER.md
  - [ ] COMMAND_EXECUTION.md
  - [ ] NETWORK_INTERFACES.md
  - [ ] NATIVE_IMPLEMENTATION.md
  - [ ] ARCHITECTURE.md

### Version Information

- [ ] Version updated in `pkg/version/version.go` to `0.1.0`
- [ ] CHANGELOG.md has entry for `[0.1.0]`
- [ ] Release date is set in CHANGELOG.md

### Repository State

- [ ] All changes are committed
- [ ] Working directory is clean (`git status`)
- [ ] On main branch
- [ ] Synced with remote (`git pull`)
- [ ] No uncommitted changes

### Build Verification

- [ ] Local build succeeds (`make build`)
- [ ] Multi-platform builds succeed (`make build-all`)
- [ ] Binary works (`./aws-ssm version`)
- [ ] All platforms build successfully:
  - [ ] darwin/amd64
  - [ ] darwin/arm64
  - [ ] linux/amd64
  - [ ] linux/arm64
  - [ ] windows/amd64

## Release Process

### 1. Run Release Preparation Script

```bash
./scripts/prepare-release.sh 0.1.0
```

This script will:
- Verify working directory is clean
- Check if on main branch
- Verify tag doesn't exist
- Run all tests
- Run linter
- Build for all platforms
- Test the binary

### 2. Create Git Tag

```bash
# Create annotated tag
git tag -a v0.1.0 -m "Release v0.1.0"

# Or use Makefile
make tag VERSION=0.1.0
```

### 3. Push Tag to GitHub

```bash
# Push the tag
git push origin v0.1.0
```

### 4. GitHub Actions Workflow

Once the tag is pushed, GitHub Actions will automatically:

1. **Run CI checks**:
   - Run tests on multiple platforms
   - Run linter
   - Build binaries

2. **Create release**:
   - Build binaries for all platforms
   - Create compressed archives (tar.gz for Unix, zip for Windows)
   - Generate SHA256 checksums
   - Create GitHub release with release notes
   - Upload all artifacts

3. **Generate Homebrew formula**:
   - Calculate checksums for Homebrew
   - Create Homebrew formula
   - Upload as artifact

### 5. Verify GitHub Release

After the workflow completes:

- [ ] Go to https://github.com/johnlam90/aws-ssm/releases
- [ ] Verify release v0.1.0 exists
- [ ] Check all artifacts are uploaded:
  - [ ] aws-ssm-darwin-amd64.tar.gz
  - [ ] aws-ssm-darwin-arm64.tar.gz
  - [ ] aws-ssm-linux-amd64.tar.gz
  - [ ] aws-ssm-linux-arm64.tar.gz
  - [ ] aws-ssm-windows-amd64.zip
  - [ ] checksums.txt
- [ ] Verify release notes are correct
- [ ] Test download links work

### 6. Test Installation

Test installation on different platforms:

**macOS (Intel):**
```bash
curl -L https://github.com/johnlam90/aws-ssm/releases/download/v0.1.0/aws-ssm-darwin-amd64.tar.gz | tar xz
chmod +x aws-ssm-darwin-amd64
./aws-ssm-darwin-amd64 version
```

**macOS (Apple Silicon):**
```bash
curl -L https://github.com/johnlam90/aws-ssm/releases/download/v0.1.0/aws-ssm-darwin-arm64.tar.gz | tar xz
chmod +x aws-ssm-darwin-arm64
./aws-ssm-darwin-arm64 version
```

**Linux (amd64):**
```bash
curl -L https://github.com/johnlam90/aws-ssm/releases/download/v0.1.0/aws-ssm-linux-amd64.tar.gz | tar xz
chmod +x aws-ssm-linux-amd64
./aws-ssm-linux-amd64 version
```

**Windows:**
- Download aws-ssm-windows-amd64.zip
- Extract and run `aws-ssm-windows-amd64.exe version`

### 7. Verify Checksums

```bash
# Download checksums file
curl -L https://github.com/johnlam90/aws-ssm/releases/download/v0.1.0/checksums.txt -o checksums.txt

# Download a binary
curl -L https://github.com/johnlam90/aws-ssm/releases/download/v0.1.0/aws-ssm-darwin-arm64.tar.gz -o aws-ssm-darwin-arm64.tar.gz

# Verify checksum
shasum -a 256 -c checksums.txt --ignore-missing
```

## Post-Release Tasks

### 1. Update Homebrew Tap (Future)

When creating a Homebrew tap:

```bash
# Clone homebrew-aws-ssm repository
git clone https://github.com/johnlam90/homebrew-aws-ssm.git

# Copy the generated formula
cp homebrew-formula/aws-ssm.rb homebrew-aws-ssm/Formula/

# Update checksums for all platforms
# Commit and push
```

### 2. Announce Release

- [ ] Create announcement on GitHub Discussions
- [ ] Update project website (if applicable)
- [ ] Share on social media (if applicable)
- [ ] Notify users/contributors

### 3. Monitor for Issues

- [ ] Watch for bug reports
- [ ] Monitor GitHub issues
- [ ] Check download statistics
- [ ] Gather user feedback

### 4. Plan Next Release

- [ ] Create milestone for next version
- [ ] Review and prioritize issues
- [ ] Update roadmap

## Rollback Procedure

If issues are discovered after release:

### Option 1: Patch Release

1. Fix the issue
2. Create patch release (v0.1.1)
3. Follow release process

### Option 2: Delete Release

```bash
# Delete the tag locally
git tag -d v0.1.0

# Delete the tag remotely
git push origin :refs/tags/v0.1.0

# Delete the GitHub release manually
# Go to https://github.com/johnlam90/aws-ssm/releases
# Click on the release
# Click "Delete"
```

## Release Artifacts Checklist

After release, verify all artifacts exist:

- [ ] Source code (zip)
- [ ] Source code (tar.gz)
- [ ] aws-ssm-darwin-amd64.tar.gz
- [ ] aws-ssm-darwin-arm64.tar.gz
- [ ] aws-ssm-linux-amd64.tar.gz
- [ ] aws-ssm-linux-arm64.tar.gz
- [ ] aws-ssm-windows-amd64.zip
- [ ] checksums.txt
- [ ] Homebrew formula (as artifact)

## Success Criteria

Release is successful when:

- [ ] All artifacts are available
- [ ] Installation works on all platforms
- [ ] Checksums are correct
- [ ] Version command shows correct version
- [ ] No critical bugs reported within 24 hours
- [ ] Documentation is accurate
- [ ] GitHub release is published

## Notes

- Always test on a clean machine if possible
- Keep release notes concise but informative
- Include migration notes if breaking changes
- Thank contributors in release notes
- Update documentation immediately after release

## Contact

For questions about the release process:
- Open an issue on GitHub
- Contact maintainers

---

**Release Manager**: John Lam (@johnlam90)
**Release Date**: 2025-01-07
**Version**: 0.1.0

