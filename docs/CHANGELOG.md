# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-01-07

### Added

#### Core Features
- **Interactive Fuzzy Finder**: Select EC2 instances interactively with real-time filtering
  - Works with `session` and `interfaces` commands
  - Arrow key navigation and type-to-filter
  - Rich preview panel with instance details and tags
  - Powered by `go-fuzzyfinder`

- **Remote Command Execution**: Execute commands on EC2 instances without interactive sessions
  - Syntax: `aws-ssm session <instance> "<command>"`
  - Captures stdout and stderr
  - Automatic timeout handling (2 minutes)
  - Uses AWS SSM SendCommand API

- **Network Interface Listing**: Display all network interfaces for EC2 instances
  - Shows interface names (ens5, ens6, etc.), subnet IDs, CIDR blocks, and security groups
  - Perfect for Kubernetes/EKS nodes with Multus CNI
  - Supports filtering by node names, instance IDs, and tags
  - Interactive fuzzy finder when no arguments provided

- **Native Go SSM Implementation**: Pure Go implementation without external AWS Session Manager Plugin
  - Uses `mmmorris1975/ssm-session-client` library
  - No external dependencies required
  - `--native` flag (default: true) to toggle between native and plugin modes

- **Port Forwarding**: Forward local ports to remote EC2 instance ports
  - Syntax: `aws-ssm port-forward <instance> --remote-port <port> --local-port <port>`
  - Useful for accessing databases, web servers, and other services
  - Native Go implementation

#### Instance Discovery
- Multiple instance identifier types supported:
  - Instance ID (e.g., `i-1234567890abcdef0`)
  - Instance name (Name tag, e.g., `web-server`)
  - DNS name (e.g., `ip-100-64-149-165.ec2.internal`)
  - IP address (public or private)
  - Tag filters (format: `Key:Value`)

#### Commands
- `aws-ssm list` - List EC2 instances with filtering options
- `aws-ssm session [instance] [command]` - Start interactive session or execute command
- `aws-ssm port-forward <instance>` - Forward ports to remote instance
- `aws-ssm interfaces [instance]` - List network interfaces
- `aws-ssm version` - Display version information

#### Documentation
- Comprehensive README.md with examples
- QUICK_REFERENCE.md for command reference
- FUZZY_FINDER.md for interactive selection guide
- COMMAND_EXECUTION.md for remote command execution
- NETWORK_INTERFACES.md for network interface inspection
- NATIVE_IMPLEMENTATION.md for pure Go implementation details
- ARCHITECTURE.md for technical architecture

#### Build & Release
- Multi-platform builds (macOS amd64/arm64, Linux amd64/arm64, Windows amd64)
- GitHub Actions CI/CD workflows
- Automated releases with binaries and checksums
- Homebrew formula generation
- Version information embedded in binary

### Technical Details

#### Dependencies
- AWS SDK for Go v2 (github.com/aws/aws-sdk-go-v2)
- Cobra CLI framework (github.com/spf13/cobra)
- SSM Session Client (github.com/mmmorris1975/ssm-session-client v0.200.0)
- Go Fuzzyfinder (github.com/ktr0731/go-fuzzyfinder v0.9.0)

#### Requirements
- Go 1.24 or later
- AWS credentials configured
- IAM permissions for EC2 and SSM

#### Supported Platforms
- macOS (Intel and Apple Silicon)
- Linux (amd64 and arm64)
- Windows (amd64)

### Notes

This is the initial release of aws-ssm, a modern AWS Systems Manager CLI tool built in pure Go. It provides an intuitive interface for managing EC2 instances without requiring SSH access or bastion hosts.

The tool is inspired by [aws-gate](https://github.com/xen0l/aws-gate) but reimplemented in Go with additional features like fuzzy finding, remote command execution, and network interface inspection.

[0.1.0]: https://github.com/johnlam90/aws-ssm/releases/tag/v0.1.0

