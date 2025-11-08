# Quick Reference Guide

## Installation

```bash
# Build from source
git clone <repository>
cd aws-ssm
go build -o aws-ssm .

# Or use make
make build

# Install to system
sudo mv aws-ssm /usr/local/bin/
```

## Basic Commands

### List Instances

```bash
# List all instances
aws-ssm list

# List all instances (including stopped)
aws-ssm list --all

# Filter by tag
aws-ssm list --tag Environment=production
aws-ssm list --tag Name=web-server

# Multiple tags
aws-ssm list --tag Environment=production --tag Role=web

# Specific region
aws-ssm list --region us-west-2

# Specific profile
aws-ssm list --profile production
```

### Start Session

```bash
# Interactive fuzzy finder (recommended - no argument)
aws-ssm session

# By instance ID
aws-ssm session i-1234567890abcdef0

# By instance name (Name tag)
aws-ssm session web-server

# By tag (Key:Value)
aws-ssm session Environment:production

# By private IP
aws-ssm session 10.0.1.100

# By public IP
aws-ssm session 54.123.45.67

# By DNS name
aws-ssm session ec2-54-123-45-67.us-west-2.compute.amazonaws.com

# With region and profile
aws-ssm session web-server --region us-west-2 --profile production

# Use plugin mode (requires session-manager-plugin)
aws-ssm session web-server --native=false
```

### Execute Remote Commands

```bash
# Execute a command on an instance
aws-ssm session <instance> "<command>"

# Examples:
aws-ssm session web-server "uptime"
aws-ssm session i-1234567890abcdef0 "df -h"
aws-ssm session web-server "ps aux | grep nginx"
aws-ssm session web-server "systemctl status nginx"
aws-ssm session web-server "docker ps" --region us-west-2 --profile production
```

### Port Forwarding

```bash
# Forward local port to remote port
aws-ssm port-forward <instance> --remote-port <remote> --local-port <local>

# Examples:
# MySQL
aws-ssm port-forward db-server --remote-port 3306 --local-port 3306

# PostgreSQL
aws-ssm port-forward db-server --remote-port 5432 --local-port 5432

# HTTP
aws-ssm port-forward web-server --remote-port 80 --local-port 8080

# Redis
aws-ssm port-forward cache-server --remote-port 6379 --local-port 6379

# Then connect to localhost:<local-port>
mysql -h 127.0.0.1 -P 3306 -u user -p
```

### List Network Interfaces

```bash
# Interactive fuzzy finder (recommended - no argument)
aws-ssm interfaces

# List interfaces for specific instance
aws-ssm interfaces i-1234567890abcdef0
aws-ssm interfaces web-server

# List interfaces by Kubernetes node name
aws-ssm interfaces ip-100-64-149-165.ec2.internal
aws-ssm interfaces -n ip-100-64-149-165.ec2.internal

# List interfaces for multiple nodes
aws-ssm interfaces -n ip-100-64-149-165.ec2.internal -n ip-100-64-87-43.ec2.internal

# List interfaces with tag filter
aws-ssm interfaces --tag Environment:production

# List interfaces for all instances (including stopped)
aws-ssm interfaces --all

# With region and profile
aws-ssm interfaces web-server --region us-west-2 --profile production
```

## Global Flags

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--region` | `-r` | AWS region | `--region us-west-2` |
| `--profile` | `-p` | AWS profile | `--profile production` |
| `--help` | `-h` | Show help | `--help` |

## Session Command Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--native` | `-n` | `true` | Use native Go implementation (no plugin) |

## Environment Variables

```bash
# Set default region
export AWS_REGION=us-west-2

# Set default profile
export AWS_PROFILE=production

# AWS credentials (if not using profile)
export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_SESSION_TOKEN=<token>  # Optional, for temporary credentials
```

## Instance Identifier Types

| Type | Format | Example |
|------|--------|---------|
| Instance ID | `i-*` | `i-1234567890abcdef0` |
| Name | `<name>` | `web-server` |
| Tag | `Key:Value` | `Environment:production` |
| Private IP | `10.*` or `172.*` or `192.168.*` | `10.0.1.100` |
| Public IP | Any other IP | `54.123.45.67` |
| Public DNS | `ec2-*` | `ec2-54-123-45-67.us-west-2.compute.amazonaws.com` |
| Private DNS | `ip-*` | `ip-10-0-1-100.ec2.internal` |

## Common Workflows

### 1. Find and Connect to Instance

**Using Interactive Fuzzy Finder (Easiest):**
```bash
# Just run session without arguments
aws-ssm session

# Then:
# - Type to filter instances
# - Use arrow keys to navigate
# - Press Enter to connect
```

**Using List and Connect:**
```bash
# Step 1: List instances to find the right one
aws-ssm list --tag Environment=production

# Step 2: Connect using instance ID or name
aws-ssm session i-1234567890abcdef0
# or
aws-ssm session web-server-1
```

### 2. Execute Quick Health Checks

```bash
# Check if web server is responding
aws-ssm session web-server "curl -s localhost:8080/health"

# Check disk space
aws-ssm session web-server "df -h"

# Check service status
aws-ssm session web-server "systemctl status nginx"

# View recent logs
aws-ssm session app-server "tail -n 50 /var/log/app.log"
```

### 3. Inspect Network Interfaces (Multus/EKS)

```bash
# List all network interfaces for a Kubernetes node
aws-ssm interfaces ip-100-64-149-165.ec2.internal

# Check network configuration for all worker nodes
aws-ssm interfaces --tag Role:worker

# Verify network setup for specific instance
aws-ssm interfaces web-server
```

### 4. Access Database Through Bastion

```bash
# Start port forwarding
aws-ssm port-forward bastion --remote-port 5432 --local-port 5432

# In another terminal, connect to database
psql -h localhost -p 5432 -U dbuser -d mydb
```

### 5. Multi-Region Management

```bash
# List instances in different regions
aws-ssm list --region us-east-1
aws-ssm list --region eu-west-1
aws-ssm list --region ap-southeast-1

# Connect to instance in specific region
aws-ssm session web-server --region eu-west-1
```

### 6. Multi-Account Management

```bash
# List instances in different accounts
aws-ssm list --profile account-dev
aws-ssm list --profile account-staging
aws-ssm list --profile account-prod

# Connect to instance in specific account
aws-ssm session web-server --profile account-prod
```

## Troubleshooting

### Check Instance SSM Status

```bash
# List instances and check if they're running
aws-ssm list

# If instance is not listed, check:
# 1. SSM Agent is running on the instance
# 2. Instance has proper IAM role
# 3. You're using the correct region
```

### Connection Issues

```bash
# Try plugin mode to isolate issue
aws-ssm session <instance> --native=false

# Check AWS credentials
aws sts get-caller-identity

# Check SSM permissions
aws ssm describe-instance-information --region <region>
```

### Permission Issues

```bash
# Verify your IAM permissions
aws iam get-user
aws iam list-attached-user-policies --user-name <your-username>

# Check if you can describe instances
aws ec2 describe-instances --region <region>

# Check if you can start SSM session
aws ssm start-session --target <instance-id> --region <region>
```

## Required IAM Permissions

### For Users/Roles

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ssm:StartSession",
        "ssm:TerminateSession",
        "ssm:ResumeSession",
        "ssm:DescribeSessions",
        "ssm:GetConnectionStatus"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeInstances",
        "ec2:DescribeInstanceStatus"
      ],
      "Resource": "*"
    }
  ]
}
```

### For EC2 Instances

Attach the `AmazonSSMManagedInstanceCore` managed policy to the instance role.

## Tips and Tricks

### Bash Aliases

Add to your `~/.bashrc` or `~/.zshrc`:

```bash
# Quick aliases
alias ssm='aws-ssm session'
alias ssm-list='aws-ssm list'
alias ssm-pf='aws-ssm port-forward'

# Region-specific aliases
alias ssm-us='aws-ssm session --region us-east-1'
alias ssm-eu='aws-ssm session --region eu-west-1'

# Profile-specific aliases
alias ssm-prod='aws-ssm session --profile production'
alias ssm-dev='aws-ssm session --profile development'
```

### Shell Completion

```bash
# Generate completion script
aws-ssm completion bash > /etc/bash_completion.d/aws-ssm
aws-ssm completion zsh > /usr/local/share/zsh/site-functions/_aws-ssm

# Or for current user
aws-ssm completion bash > ~/.aws-ssm-completion.bash
echo 'source ~/.aws-ssm-completion.bash' >> ~/.bashrc
```

### Quick Instance Access

```bash
# Create a function for quick access
ssm-quick() {
  aws-ssm session "$1" --region us-west-2 --profile production
}

# Usage
ssm-quick web-server
```

## Performance Tips

1. **Use instance ID when possible** - Fastest lookup
2. **Use Name tag** - Second fastest (single tag lookup)
3. **Avoid IP/DNS lookups** - Requires full instance scan
4. **Cache instance IDs** - Store frequently used instance IDs

## Security Best Practices

1. **Use IAM roles** - Don't use long-term credentials
2. **Enable session logging** - Configure CloudWatch Logs in SSM
3. **Use MFA** - Require MFA for production access
4. **Limit permissions** - Use least privilege principle
5. **Rotate credentials** - Regularly rotate access keys
6. **Use tags** - Tag instances for better access control

## Getting Help

```bash
# General help
aws-ssm --help

# Command-specific help
aws-ssm session --help
aws-ssm list --help
aws-ssm port-forward --help

# Version information
aws-ssm version  # (if implemented)
```

