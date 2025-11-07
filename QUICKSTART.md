# Quick Start Guide

This guide will help you get started with the AWS SSM Manager CLI in just a few minutes.

## Step 1: Prerequisites

Before you begin, ensure you have:

1. **AWS Session Manager Plugin** installed
2. **AWS credentials** configured
3. **Go 1.21+** (if building from source)

### Install AWS Session Manager Plugin

**macOS:**
```bash
brew install --cask session-manager-plugin
```

**Linux (Ubuntu/Debian):**
```bash
curl "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_64bit/session-manager-plugin.deb" -o "session-manager-plugin.deb"
sudo dpkg -i session-manager-plugin.deb
```

### Configure AWS Credentials

If you haven't already, configure your AWS credentials:

```bash
aws configure
```

Or set environment variables:
```bash
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key
export AWS_REGION=us-east-1
```

## Step 2: Build the CLI

```bash
# Clone or navigate to the project directory
cd aws-ssm

# Download dependencies
go mod download

# Build the binary
go build -o aws-ssm .

# (Optional) Move to your PATH
sudo mv aws-ssm /usr/local/bin/
```

Or use the Makefile:
```bash
make build
```

## Step 3: Verify Installation

```bash
./aws-ssm --help
```

You should see the help output with available commands.

## Step 4: List Your Instances

See all running EC2 instances:

```bash
./aws-ssm list
```

Filter by tags:
```bash
./aws-ssm list --tag Environment=production
```

Use a specific AWS profile and region:
```bash
./aws-ssm list --region us-west-2 --profile myprofile
```

## Step 5: Connect to an Instance

### By Instance Name (easiest)

```bash
./aws-ssm session my-web-server
```

### By Instance ID

```bash
./aws-ssm session i-1234567890abcdef0
```

### By Tag

```bash
./aws-ssm session Environment:production
```

### By IP Address

```bash
./aws-ssm session 10.0.1.100
```

## Common Use Cases

### 1. Connect to a Production Database Server

```bash
# List production instances
./aws-ssm list --tag Environment=production --tag Role=database

# Connect to the database server
./aws-ssm session prod-db-01 --profile production
```

### 2. Quick Debug Session

```bash
# Find the instance by IP
./aws-ssm session 10.0.5.42
```

### 3. Multi-Region Management

```bash
# List instances in different regions
./aws-ssm list --region us-east-1
./aws-ssm list --region eu-west-1

# Connect to instance in specific region
./aws-ssm session web-server --region eu-west-1
```

## Troubleshooting

### "session-manager-plugin not found"

Install the AWS Session Manager Plugin (see Step 1).

### "No instances found"

- Verify you're using the correct region: `--region us-west-2`
- Check your AWS profile: `--profile myprofile`
- Ensure the instance is running: `./aws-ssm list --all`

### "Permission denied"

Ensure your AWS user/role has the required IAM permissions. See `examples/iam-policy.json` for the required permissions.

### "Instance not running"

The instance must be in the "running" state. Check with:
```bash
./aws-ssm list --all
```

## Next Steps

- Read the full [README.md](README.md) for detailed documentation
- Check [examples/iam-policy.json](examples/iam-policy.json) for IAM permission requirements
- Review [examples/instance-role-policy.json](examples/instance-role-policy.json) for EC2 instance role setup

## Tips

1. **Use tab completion**: Most shells support command completion for better UX
2. **Set default region**: Export `AWS_REGION` to avoid typing `--region` every time
3. **Use profiles**: Organize multiple AWS accounts with named profiles
4. **Tag your instances**: Use consistent tagging for easier instance discovery

## Getting Help

```bash
# General help
./aws-ssm --help

# Command-specific help
./aws-ssm list --help
./aws-ssm session --help
```

Happy SSM-ing! 🚀

