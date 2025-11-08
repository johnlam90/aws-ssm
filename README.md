# AWS SSM Manager CLI

A native Golang CLI tool for managing AWS SSM (Systems Manager) sessions with EC2 instances. Connect to your instances using instance ID, DNS name, IP address, or tags - **no bastion host or AWS session-manager-plugin required!**

## Features

- 🔍 **Enhanced Interactive Fuzzy Finder**: Multi-select, rich search syntax, color highlighting, and more!
- ⚡ **Remote Command Execution**: Execute commands on instances without starting an interactive shell
- 🌐 **Network Interface Listing**: Display all network interfaces (Multus, EKS) with subnet and security group details
- 🚀 **Multiple Connection Methods**: Connect using instance ID, DNS name, IP address, tags, or instance name
- 📋 **List Instances**: View all your EC2 instances with filtering capabilities
- 🔐 **Secure**: Uses AWS SSM Session Manager for secure, auditable connections
- 💻 **Pure Go Implementation**: Single binary with **NO external dependencies** - no session-manager-plugin needed!
- 🎯 **Smart Instance Discovery**: Automatically detects identifier type and finds matching instances
- 🔌 **Port Forwarding**: Forward local ports to remote services on EC2 instances
- 🎨 **Rich Search Syntax**: Advanced filtering with `name:web`, `state:running`, `tag:Env=prod`, `!Env=dev`
- 📊 **Flexible Display**: Customizable columns, sorting, and color themes
- ⚡ **Performance Optimized**: Caching, streaming pagination, and fuzzy ranking
- 🔖 **Bookmarks**: Favorite your frequently accessed instances

## Why This Tool?

Unlike the official AWS CLI and session-manager-plugin, this tool:

- ✅ **No Plugin Required**: Pure Go implementation of the SSM protocol (uses [ssm-session-client](https://github.com/mmmorris1975/ssm-session-client))
- ✅ **Single Binary**: Just download and run - no Python, no Node.js, no external dependencies
- ✅ **Smart Discovery**: Find instances by name, tags, IP, or DNS - not just instance ID
- ✅ **Better UX**: Clean, intuitive command-line interface with helpful error messages

## Prerequisites

- AWS credentials configured (via `~/.aws/credentials` or environment variables)
- SSM Agent version 2.3.68.0 or later on target EC2 instances
- Proper IAM permissions for SSM and EC2

**Note**: The AWS Session Manager Plugin is **NOT required** when using the default native mode (`--native` flag, which is enabled by default). If you want to use the plugin-based mode, you can disable native mode with `--native=false`.

## Installation

> **📖 For detailed installation instructions, see [docs/INSTALLATION.md](docs/INSTALLATION.md)**

### Quick Install

### macOS

#### Homebrew (Recommended)

```bash
# Add the tap
brew tap johnlam90/tap

# Install aws-ssm
brew install aws-ssm

# Verify installation
aws-ssm version
```

#### Manual Installation

**Intel (amd64):**

```bash
curl -L https://github.com/johnlam90/aws-ssm/releases/latest/download/aws-ssm-darwin-amd64.tar.gz | tar xz
chmod +x aws-ssm-darwin-amd64
sudo mv aws-ssm-darwin-amd64 /usr/local/bin/aws-ssm
```

**Apple Silicon (arm64):**

```bash
curl -L https://github.com/johnlam90/aws-ssm/releases/latest/download/aws-ssm-darwin-arm64.tar.gz | tar xz
chmod +x aws-ssm-darwin-arm64
sudo mv aws-ssm-darwin-arm64 /usr/local/bin/aws-ssm
```

### Linux

**amd64:**

```bash
curl -L https://github.com/johnlam90/aws-ssm/releases/latest/download/aws-ssm-linux-amd64.tar.gz | tar xz
chmod +x aws-ssm-linux-amd64
sudo mv aws-ssm-linux-amd64 /usr/local/bin/aws-ssm
```

**arm64:**

```bash
curl -L https://github.com/johnlam90/aws-ssm/releases/latest/download/aws-ssm-linux-arm64.tar.gz | tar xz
chmod +x aws-ssm-linux-arm64
sudo mv aws-ssm-linux-arm64 /usr/local/bin/aws-ssm
```

### Windows

Download the latest release from [GitHub Releases](https://github.com/johnlam90/aws-ssm/releases/latest):

1. Download `aws-ssm-windows-amd64.zip`
2. Extract the archive
3. Add the directory to your PATH

Or using PowerShell:

```powershell
# Download and extract
Invoke-WebRequest -Uri "https://github.com/johnlam90/aws-ssm/releases/latest/download/aws-ssm-windows-amd64.zip" -OutFile "aws-ssm.zip"
Expand-Archive -Path "aws-ssm.zip" -DestinationPath "C:\Program Files\aws-ssm"

# Add to PATH (requires admin)
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\Program Files\aws-ssm", "Machine")
```

### From Source

Requires Go 1.24 or later:

```bash
# Clone the repository
git clone https://github.com/johnlam90/aws-ssm.git
cd aws-ssm

# Build the binary
make build

# Or build for all platforms
make build-all

# Install to GOPATH/bin
make install
```

### Using Go Install

```bash
go install github.com/johnlam90/aws-ssm@latest
```

### Verify Installation

After installation, verify it works:

```bash
aws-ssm version
```

You should see output like:

```sh
aws-ssm version 0.1.0 (commit: abc1234, built: 2025-01-07T12:00:00Z, go: go1.24)
```

## Usage

### List Instances

List all running EC2 instances:

```bash
aws-ssm list
```

List instances with specific tags:

```bash
# Single tag filter
aws-ssm list --tag Name=web-server

# Multiple tag filters
aws-ssm list --tag Environment=production --tag Team=backend

# Show all instances (including stopped)
aws-ssm list --all
```

List instances in a specific region:

```bash
aws-ssm list --region us-west-2 --profile production
```

### Start SSM Session

#### Enhanced Interactive Fuzzy Finder (Recommended)

The easiest way to connect is using the **enhanced interactive fuzzy finder**:

```bash
# No argument - opens interactive selector
aws-ssm session

# Enhanced mode with multi-select and rich search
aws-ssm session --interactive

# Custom columns display
aws-ssm session --interactive --columns name,instance-id,state,type,az

# Show only bookmarked instances
aws-ssm session --interactive --favorites
```

##### Rich Search Syntax

The enhanced fuzzy finder supports powerful search syntax:

```bash
# Filter by name
name:web

# Filter by instance ID  
id:i-123456789

# Filter by state
state:running

# Filter by tags
tag:Environment=production
tag:Team=backend

# Filter by IP pattern
ip:10.0.1.*

# Filter by DNS pattern
dns:*.compute.amazonaws.com

# Negative filters
!state:stopped
!Env=dev

# Tag existence
has:Environment
missing:Team

# Combine multiple filters
name:web state:running tag:Env=prod !Env=dev

# Fuzzy search combined with filters
web state:running has:Backup
```

##### Multi-Select and Batch Operations

When using `--interactive` flag:

- **Space**: Toggle selection of multiple instances
- **Enter**: Connect to selected instances (sequential or batch)
- **c**: Run command on selected instances  
- **p**: Set up port forwarding for selected instances

```bash
# Select multiple instances for batch operations
aws-ssm session --interactive
# Use Space to select multiple, then Enter for actions
```

This will:

1. Fetch all EC2 instances with caching for performance
2. Display interactive interface with customizable columns
3. Support rich search syntax with real-time filtering
4. Allow multi-select for batch operations
5. Show color-coded instance states (green=running, red=stopped)
6. Display detailed preview with metadata and tags

- Full instance details
- All tags
- Public/Private DNS names
- Availability zone
- Instance type

#### Direct Connection

Connect to an instance directly using various identifiers:

**By Instance ID:**

```bash
aws-ssm session i-1234567890abcdef0
```

**By Instance Name (Name tag):**

```bash
aws-ssm session web-server
```

**By Tag (Key:Value format):**

```bash
aws-ssm session Environment:production
aws-ssm session Team:backend
```

**By IP Address:**

```bash
# Private IP
aws-ssm session 10.0.1.100

# Public IP
aws-ssm session 54.123.45.67
```

**By DNS Name:**

```bash
# Public DNS
aws-ssm session ec2-54-123-45-67.us-west-2.compute.amazonaws.com

# Private DNS
aws-ssm session ip-10-0-1-100.us-west-2.compute.internal
```

**With Specific Region and Profile:**

```bash
aws-ssm session web-server --region us-west-2 --profile production
```

**Using Plugin Mode (if you have session-manager-plugin installed):**

```bash
aws-ssm session web-server --native=false
```

### Execute Remote Commands

Execute commands on instances without starting an interactive shell session. The command output is displayed and the CLI exits.

**Basic Command Execution:**

```bash
# Execute a simple command
aws-ssm session web-server "uptime"

# Check disk usage
aws-ssm session i-1234567890abcdef0 "df -h"

# View running processes
aws-ssm session web-server "ps aux"
```

**Multi-word Commands:**

```bash
# Commands with pipes
aws-ssm session web-server "ps aux | grep nginx"

# Commands with multiple arguments
aws-ssm session web-server "systemctl status nginx"

# Complex commands (use quotes)
aws-ssm session web-server "find /var/log -name '*.log' -mtime -1"
```

**With Region and Profile:**

```bash
# Execute command with specific region/profile
aws-ssm session web-server "docker ps" --region us-west-2 --profile production
```

**How It Works:**

1. The command is sent to the instance via AWS SSM `SendCommand` API
2. The CLI waits for the command to complete (up to 2 minutes)
3. Output (stdout and stderr) is displayed
4. The CLI exits with the appropriate status

**Use Cases:**

- Quick health checks: `aws-ssm session web-1 "curl localhost:8080/health"`
- Log inspection: `aws-ssm session app-server "tail -n 100 /var/log/app.log"`
- Service status: `aws-ssm session db-server "systemctl status postgresql"`
- Disk space checks: `aws-ssm session web-server "df -h"`
- Process monitoring: `aws-ssm session api-server "ps aux | grep java"`

### Port Forwarding

Forward a local port to a remote port on an EC2 instance:

```bash
# Forward local port 3306 to remote MySQL port 3306
aws-ssm port-forward db-server --remote-port 3306 --local-port 3306

# Forward local port 8080 to remote port 80
aws-ssm port-forward web-server --remote-port 80 --local-port 8080

# Access RDS through a bastion instance
aws-ssm port-forward bastion --remote-port 5432 --local-port 5432
```

Then connect to `localhost:3306` (or your chosen local port) to access the remote service.

### List Network Interfaces

Display all network interfaces attached to EC2 instances. This is especially useful for instances with multiple network interfaces (Multus, EKS, etc.):

```bash
# Interactive fuzzy finder (no argument)
aws-ssm interfaces

# List interfaces for specific instance by ID
aws-ssm interfaces i-1234567890abcdef0

# List interfaces by instance name
aws-ssm interfaces web-server

# List interfaces by Kubernetes node name
aws-ssm interfaces ip-100-64-149-165.ec2.internal

# List interfaces for multiple nodes
aws-ssm interfaces --node-name ip-100-64-149-165.ec2.internal --node-name ip-100-64-87-43.ec2.internal

# List interfaces with tag filter
aws-ssm interfaces --tag Environment:production

# List interfaces for all instances (including stopped)
aws-ssm interfaces --all
```

**Example Output:**

```sh
Instance: i-07792557b9c1167a4 | DNS Name: ip-100-64-149-165.ec2.internal | Instance Name: nk-rdc-upf-d-mg-worker-node
Interface | Subnet ID                 | CIDR               | SG ID
--------------------------------------------------------------------------------------
ens5      | subnet-06d8b73f0e116b342  | 100.64.128.0/19    | sg-00f82f14e7abe5298
ens6      | subnet-04f50f436a32a474f  | 10.2.9.0/26        | sg-0fd2dbe3c8853e657
ens7      | subnet-033edf3510a4e3f50  | 10.2.11.0/25       | sg-0fd2dbe3c8853e657
ens8      | subnet-03b757845ae511e01  | 10.2.11.128/25     | sg-0fd2dbe3c8853e657

Total instances displayed: 1
```

**Features:**

- Shows interface names (ens5, ens6, etc.) based on device index
- Displays subnet ID and CIDR block for each interface
- Shows security group ID for each interface
- Supports all instance identifier types (ID, name, DNS, IP, tags)
- Can filter by Kubernetes node names
- Works with region and profile flags

### Global Flags

All commands support these flags:

- `--region, -r`: AWS region (defaults to `AWS_REGION` env var or default profile region)
- `--profile, -p`: AWS profile to use (defaults to `AWS_PROFILE` env var or default profile)

## Examples

### Common Workflows

**1. Find and connect to a production web server:**

```bash
# First, list instances to find the right one
aws-ssm list --tag Environment=production --tag Role=web

# Connect to it by name
aws-ssm session prod-web-01 --profile production
```

**2. Connect to an instance in a different region:**

```bash
aws-ssm session my-instance --region eu-west-1
```

**3. Quick connection by IP:**

```bash
aws-ssm session 10.0.1.50
```

## Project Structure

```sh
aws-ssm/
├── main.go              # Application entry point
├── cmd/                 # CLI commands
│   ├── root.go         # Root command and global flags
│   ├── list.go         # List instances command
│   └── session.go      # Start session command
├── pkg/
│   └── aws/            # AWS client wrappers
│       ├── client.go   # AWS client initialization
│       ├── instance.go # EC2 instance queries
│       └── session.go  # SSM session management
└── go.mod              # Go module dependencies
```

## IAM Permissions

Your AWS user/role needs the following permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeInstances",
        "ssm:StartSession",
        "ssm:TerminateSession"
      ],
      "Resource": "*"
    }
  ]
}
```

EC2 instances need an IAM role with:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ssm:UpdateInstanceInformation",
        "ssmmessages:CreateControlChannel",
        "ssmmessages:CreateDataChannel",
        "ssmmessages:OpenControlChannel",
        "ssmmessages:OpenDataChannel"
      ],
      "Resource": "*"
    }
  ]
}
```

## Troubleshooting

**Connection Issues:**

By default, the CLI uses the **native Go implementation** which doesn't require the session-manager-plugin. If you encounter issues:

1. Make sure SSM Agent is running on the target instance
2. Verify your IAM permissions (see IAM Permissions section)
3. Check that the instance is in the "running" state
4. Try using plugin mode: `aws-ssm session <instance> --native=false` (requires session-manager-plugin)

**No Instances Found:**

```sh
Error: no instances found matching: web-server
```

- Check that the instance exists and is running
- Verify you're using the correct region and profile
- Ensure your AWS credentials have EC2 describe permissions

**Multiple Instances Found:**

```sh
Error: multiple instances found, please use a more specific identifier
```

Use a more specific identifier like the full instance ID or a unique tag combination.

**Instance Not Running:**

```sh
Error: instance i-xxx is not running (current state: stopped)
```

The instance must be in the "running" state to start an SSM session.

## Development

### Building

```bash
go build -o aws-ssm .
```

### Running Tests

```bash
go test ./...
```

### Adding Dependencies

```bash
go get github.com/package/name
go mod tidy
```

## Contributing

Contributions are welcome! Please read our [Contributing Guide](docs/CONTRIBUTING.md) for details on:

- Code of conduct
- Development setup
- Submitting pull requests
- Release process

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Additional Documentation

### User Guides

- [INSTALLATION.md](docs/INSTALLATION.md) - Detailed installation instructions for all platforms
- [QUICKSTART.md](docs/QUICKSTART.md) - Quick start guide
- [QUICK_REFERENCE.md](docs/QUICK_REFERENCE.md) - Quick command reference
- [FUZZY_FINDER.md](docs/FUZZY_FINDER.md) - Interactive instance selection guide
- [COMMAND_EXECUTION.md](docs/COMMAND_EXECUTION.md) - Remote command execution guide
- [NETWORK_INTERFACES.md](docs/NETWORK_INTERFACES.md) - Network interface inspection guide

### Technical Documentation

- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - Technical architecture
- [NATIVE_IMPLEMENTATION.md](docs/NATIVE_IMPLEMENTATION.md) - Pure Go implementation details
- [CHANGELOG.md](docs/CHANGELOG.md) - Version history and release notes

### Development & Release

- [CONTRIBUTING.md](docs/CONTRIBUTING.md) - Contributing guidelines
- [RELEASE_CHECKLIST.md](docs/RELEASE_CHECKLIST.md) - Release checklist
- [HOMEBREW_RELEASE_PROCESS.md](docs/HOMEBREW_RELEASE_PROCESS.md) - Homebrew release and upgrade workflow

## Acknowledgments

- Inspired by [aws-gate](https://github.com/xen0l/aws-gate) Python project
- Built with [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2)
- CLI framework by [Cobra](https://github.com/spf13/cobra)
- SSM protocol implementation by [ssm-session-client](https://github.com/mmmorris1975/ssm-session-client)
- Fuzzy finder by [go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder)
