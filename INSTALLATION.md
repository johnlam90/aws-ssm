# Installation Guide

This guide provides detailed installation instructions for aws-ssm on different platforms.

## Table of Contents

- [macOS](#macos)
- [Linux](#linux)
- [Windows](#windows)
- [From Source](#from-source)
- [Docker](#docker)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)

## macOS

### Homebrew (Recommended)

The easiest way to install on macOS is using Homebrew:

```bash
# Add the tap
brew tap johnlam90/tap

# Install aws-ssm
brew install aws-ssm

# Verify installation
aws-ssm version
```

**Updating:**

```bash
# Update the tap
brew update

# Upgrade aws-ssm
brew upgrade aws-ssm
```

**Uninstalling:**

```bash
# Uninstall aws-ssm
brew uninstall aws-ssm

# Optionally remove the tap
brew untap johnlam90/tap
```

### Manual Installation

#### Intel Macs (amd64)

```bash
# Download the latest release
curl -L https://github.com/johnlam90/aws-ssm/releases/latest/download/aws-ssm-darwin-amd64.tar.gz -o aws-ssm.tar.gz

# Extract
tar -xzf aws-ssm.tar.gz

# Make executable
chmod +x aws-ssm-darwin-amd64

# Move to PATH
sudo mv aws-ssm-darwin-amd64 /usr/local/bin/aws-ssm

# Clean up
rm aws-ssm.tar.gz

# Verify
aws-ssm version
```

#### Apple Silicon Macs (arm64)

```bash
# Download the latest release
curl -L https://github.com/johnlam90/aws-ssm/releases/latest/download/aws-ssm-darwin-arm64.tar.gz -o aws-ssm.tar.gz

# Extract
tar -xzf aws-ssm.tar.gz

# Make executable
chmod +x aws-ssm-darwin-arm64

# Move to PATH
sudo mv aws-ssm-darwin-arm64 /usr/local/bin/aws-ssm

# Clean up
rm aws-ssm.tar.gz

# Verify
aws-ssm version
```

### Specific Version

To install a specific version, replace `latest` with the version number:

```bash
VERSION=0.1.0
curl -L https://github.com/johnlam90/aws-ssm/releases/download/v${VERSION}/aws-ssm-darwin-arm64.tar.gz -o aws-ssm.tar.gz
```

## Linux

### Debian/Ubuntu

```bash
# Download the latest release
curl -L https://github.com/johnlam90/aws-ssm/releases/latest/download/aws-ssm-linux-amd64.tar.gz -o aws-ssm.tar.gz

# Extract
tar -xzf aws-ssm.tar.gz

# Make executable
chmod +x aws-ssm-linux-amd64

# Move to PATH
sudo mv aws-ssm-linux-amd64 /usr/local/bin/aws-ssm

# Clean up
rm aws-ssm.tar.gz

# Verify
aws-ssm version
```

### RHEL/CentOS/Fedora

Same as Debian/Ubuntu above.

### ARM64 (Raspberry Pi, AWS Graviton, etc.)

```bash
# Download the latest release
curl -L https://github.com/johnlam90/aws-ssm/releases/latest/download/aws-ssm-linux-arm64.tar.gz -o aws-ssm.tar.gz

# Extract
tar -xzf aws-ssm.tar.gz

# Make executable
chmod +x aws-ssm-linux-arm64

# Move to PATH
sudo mv aws-ssm-linux-arm64 /usr/local/bin/aws-ssm

# Clean up
rm aws-ssm.tar.gz

# Verify
aws-ssm version
```

### Install to User Directory (No sudo)

If you don't have sudo access:

```bash
# Create bin directory in home
mkdir -p ~/bin

# Download and extract
curl -L https://github.com/johnlam90/aws-ssm/releases/latest/download/aws-ssm-linux-amd64.tar.gz | tar -xz

# Move to ~/bin
mv aws-ssm-linux-amd64 ~/bin/aws-ssm

# Add to PATH (add this to ~/.bashrc or ~/.zshrc)
export PATH="$HOME/bin:$PATH"

# Reload shell
source ~/.bashrc  # or source ~/.zshrc

# Verify
aws-ssm version
```

## Windows

### PowerShell (Recommended)

```powershell
# Create installation directory
New-Item -ItemType Directory -Force -Path "C:\Program Files\aws-ssm"

# Download the latest release
Invoke-WebRequest -Uri "https://github.com/johnlam90/aws-ssm/releases/latest/download/aws-ssm-windows-amd64.zip" -OutFile "aws-ssm.zip"

# Extract
Expand-Archive -Path "aws-ssm.zip" -DestinationPath "C:\Program Files\aws-ssm" -Force

# Add to PATH (requires admin)
[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable("Path", "Machine") + ";C:\Program Files\aws-ssm",
    "Machine"
)

# Clean up
Remove-Item "aws-ssm.zip"

# Restart PowerShell and verify
aws-ssm version
```

### Manual Installation

1. Download `aws-ssm-windows-amd64.zip` from [GitHub Releases](https://github.com/johnlam90/aws-ssm/releases/latest)
2. Extract the ZIP file
3. Move `aws-ssm-windows-amd64.exe` to a directory in your PATH
4. Rename to `aws-ssm.exe` (optional)
5. Open a new Command Prompt or PowerShell window
6. Run `aws-ssm version` to verify

### User Directory Installation (No Admin)

```powershell
# Create bin directory in user profile
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\bin"

# Download
Invoke-WebRequest -Uri "https://github.com/johnlam90/aws-ssm/releases/latest/download/aws-ssm-windows-amd64.zip" -OutFile "aws-ssm.zip"

# Extract
Expand-Archive -Path "aws-ssm.zip" -DestinationPath "$env:USERPROFILE\bin" -Force

# Add to user PATH
[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable("Path", "User") + ";$env:USERPROFILE\bin",
    "User"
)

# Clean up
Remove-Item "aws-ssm.zip"

# Restart PowerShell and verify
aws-ssm version
```

## From Source

### Prerequisites

- Go 1.24 or later
- Git
- Make (optional)

### Build and Install

```bash
# Clone the repository
git clone https://github.com/johnlam90/aws-ssm.git
cd aws-ssm

# Build using Make
make build

# Or build directly with go
go build -o aws-ssm .

# Install to GOPATH/bin
make install

# Or install directly
go install .

# Verify
aws-ssm version
```

### Build for Specific Platform

```bash
# Build for Linux amd64
GOOS=linux GOARCH=amd64 go build -o aws-ssm-linux-amd64 .

# Build for macOS arm64
GOOS=darwin GOARCH=arm64 go build -o aws-ssm-darwin-arm64 .

# Build for Windows amd64
GOOS=windows GOARCH=amd64 go build -o aws-ssm-windows-amd64.exe .

# Build for all platforms
make build-all
```

## Docker

### Using Docker

```bash
# Pull the image (coming soon)
docker pull johnlam90/aws-ssm:latest

# Run with AWS credentials
docker run --rm -it \
  -v ~/.aws:/root/.aws:ro \
  johnlam90/aws-ssm:latest list

# Create an alias for convenience
alias aws-ssm='docker run --rm -it -v ~/.aws:/root/.aws:ro johnlam90/aws-ssm:latest'

# Use the alias
aws-ssm list
aws-ssm session
```

### Build Docker Image

```bash
# Clone the repository
git clone https://github.com/johnlam90/aws-ssm.git
cd aws-ssm

# Build the image
docker build -t aws-ssm:local .

# Run
docker run --rm -it -v ~/.aws:/root/.aws:ro aws-ssm:local version
```

## Verification

After installation, verify that aws-ssm is working:

```bash
# Check version
aws-ssm version

# Expected output:
# aws-ssm version 0.1.0 (commit: abc1234, built: 2025-01-07T12:00:00Z, go: go1.24)

# Check help
aws-ssm --help

# List instances (requires AWS credentials)
aws-ssm list
```

## Troubleshooting

### Homebrew Installation Issues

#### Tap Not Found

**Problem**: `Error: No available formula with the name "aws-ssm"`

**Solution**: Make sure you've added the tap first:

```bash
brew tap johnlam90/tap
brew install aws-ssm
```

#### Old Version Installed

**Problem**: Running `aws-ssm version` shows an old version

**Solution**: Update Homebrew and upgrade aws-ssm:

```bash
brew update
brew upgrade aws-ssm
```

#### Conflicting Binary

**Problem**: `aws-ssm` command uses a different binary than the Homebrew installation

**Solution**: Check which binary is being used and remove conflicts:

```bash
# Check which binary is being used
which aws-ssm

# If it's not /opt/homebrew/bin/aws-ssm (Apple Silicon) or /usr/local/bin/aws-ssm (Intel)
# Remove the conflicting binary
sudo rm /path/to/conflicting/aws-ssm

# Verify Homebrew installation
brew info aws-ssm
```

### Command Not Found

**Problem**: `aws-ssm: command not found`

**Solutions**:

1. **Check if binary is in PATH**:

   ```bash
   which aws-ssm
   ```

2. **Add directory to PATH**:

   ```bash
   # For bash
   echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
   source ~/.bashrc

   # For zsh
   echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.zshrc
   source ~/.zshrc
   ```

3. **Use full path**:

   ```bash
   /usr/local/bin/aws-ssm version
   # or for Apple Silicon
   /opt/homebrew/bin/aws-ssm version
   ```

### Permission Denied

**Problem**: `Permission denied` when running aws-ssm

**Solution**: Make the binary executable:

```bash
chmod +x /usr/local/bin/aws-ssm
```

### macOS Security Warning

**Problem**: "aws-ssm cannot be opened because the developer cannot be verified"

**Solution**:
1. Go to System Preferences → Security & Privacy
2. Click "Allow Anyway" next to the aws-ssm message
3. Or run: `xattr -d com.apple.quarantine /usr/local/bin/aws-ssm`

### AWS Credentials Not Found

**Problem**: `NoCredentialProviders: no valid providers in chain`

**Solutions**:

1. **Configure AWS credentials**:
   ```bash
   aws configure
   ```

2. **Set environment variables**:
   ```bash
   export AWS_ACCESS_KEY_ID=your_access_key
   export AWS_SECRET_ACCESS_KEY=your_secret_key
   export AWS_REGION=us-east-1
   ```

3. **Use AWS profile**:
   ```bash
   aws-ssm list --profile production
   ```

### Version Mismatch

**Problem**: Old version showing after update

**Solution**:
1. Check which binary is being used: `which aws-ssm`
2. Remove old versions: `sudo rm /usr/local/bin/aws-ssm`
3. Reinstall the latest version
4. Clear shell hash: `hash -r` or restart terminal

## Uninstallation

### macOS/Linux

```bash
# Remove binary
sudo rm /usr/local/bin/aws-ssm

# If installed via Homebrew
brew uninstall aws-ssm
```

### Windows

```powershell
# Remove binary
Remove-Item "C:\Program Files\aws-ssm\aws-ssm-windows-amd64.exe"

# Remove from PATH (requires admin)
# Manually edit System Environment Variables
```

## Next Steps

After installation:

1. **Configure AWS credentials** (if not already done):
   ```bash
   aws configure
   ```

2. **Test the installation**:
   ```bash
   aws-ssm list
   ```

3. **Read the documentation**:
   - [README.md](README.md) - Main documentation
   - [QUICK_REFERENCE.md](QUICK_REFERENCE.md) - Command reference
   - [FUZZY_FINDER.md](FUZZY_FINDER.md) - Interactive selection guide

4. **Try the interactive fuzzy finder**:
   ```bash
   aws-ssm session
   ```

Enjoy using aws-ssm! 🚀

