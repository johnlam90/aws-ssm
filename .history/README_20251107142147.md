# AWS SSM Manager CLI

A native Golang CLI tool for managing AWS SSM (Systems Manager) sessions with EC2 instances. Connect to your instances using instance ID, DNS name, IP address, or tags - no bastion host required!

## Features

- 🚀 **Multiple Connection Methods**: Connect using instance ID, DNS name, IP address, tags, or instance name
- 📋 **List Instances**: View all your EC2 instances with filtering capabilities
- 🔐 **Secure**: Uses AWS SSM Session Manager for secure, auditable connections
- ⚡ **Native Go**: Fast, single binary with no dependencies (except AWS session-manager-plugin)
- 🎯 **Smart Instance Discovery**: Automatically detects identifier type and finds matching instances

## Prerequisites

- Go 1.21 or later (for building from source)
- [AWS Session Manager Plugin](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html)
- AWS credentials configured (via `~/.aws/credentials` or environment variables)
- SSM Agent version 2.3.68.0 or later on target EC2 instances
- Proper IAM permissions for SSM and EC2

### Installing AWS Session Manager Plugin

**macOS:**
```bash
brew install --cask session-manager-plugin
```

**Linux:**
```bash
curl "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_64bit/session-manager-plugin.deb" -o "session-manager-plugin.deb"
sudo dpkg -i session-manager-plugin.deb
```

**Windows:**
Download from [AWS Documentation](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/aws-ssm.git
cd aws-ssm

# Install dependencies
go mod download

# Build the binary
go build -o aws-ssm .

# Move to your PATH (optional)
sudo mv aws-ssm /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/yourusername/aws-ssm@latest
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

Connect to an instance using various identifiers:

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

```
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

**Session Manager Plugin Not Found:**
```
Error: session-manager-plugin not found in PATH
```
Install the AWS Session Manager Plugin (see Prerequisites section).

**No Instances Found:**
```
Error: no instances found matching: web-server
```
- Check that the instance exists and is running
- Verify you're using the correct region and profile
- Ensure your AWS credentials have EC2 describe permissions

**Multiple Instances Found:**
```
Error: multiple instances found, please use a more specific identifier
```
Use a more specific identifier like the full instance ID or a unique tag combination.

**Instance Not Running:**
```
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

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License.

## Acknowledgments

- Inspired by [aws-gate](https://github.com/xen0l/aws-gate) Python project
- Built with [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2)
- CLI framework by [Cobra](https://github.com/spf13/cobra)
