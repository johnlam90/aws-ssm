# AWS SSM Manager CLI

Fast, dependency-free CLI for managing AWS EC2 instances, EKS clusters, and Auto Scaling Groups via AWS Systems Manager. Single static binary—no session-manager-plugin, no SSH bastions, no external runtime.

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue.svg)](https://golang.org/dl/)
[![CI Pipeline](https://github.com/johnlam90/aws-ssm/actions/workflows/ci.yml/badge.svg)](https://github.com/johnlam90/aws-ssm/actions/workflows/ci.yml)
[![coverage](https://img.shields.io/codecov/c/github/johnlam90/aws-ssm?label=coverage&logo=codecov)](https://codecov.io/gh/johnlam90/aws-ssm)
[![go report](https://goreportcard.com/badge/github.com/johnlam90/aws-ssm)](https://goreportcard.com/report/github.com/johnlam90/aws-ssm)
[![release](https://img.shields.io/github/v/release/johnlam90/aws-ssm?label=release)](https://github.com/johnlam90/aws-ssm/releases)
[![downloads](https://img.shields.io/github/downloads/johnlam90/aws-ssm/total?label=downloads)](https://github.com/johnlam90/aws-ssm/releases)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

> Not affiliated with Amazon Web Services. Uses official AWS SDK v2.

## ✨ Key Features

- 🎯 **Pure Go** - Single binary, zero external dependencies
- 🔍 **Interactive Fuzzy Finder** - Multi-select with rich search syntax
- 🚀 **Remote Command Execution** - Run commands without interactive sessions
- ☸️ **EKS Management** - Scale nodegroups, update launch templates, view cluster details
- 📈 **ASG Scaling** - Interactive Auto Scaling Group management
- 🌐 **Network Inspection** - View all interfaces (Multus, EKS, etc.)
- 🔌 **Port Forwarding** - Forward local ports to remote services
- 💾 **Smart Caching** - Configurable TTL & region-scoped caching
- 🔐 **Secure** - SSM tunnels, no inbound ports or bastions required

## 🚀 Quick Start

### Installation

**macOS (Homebrew):**

```bash
brew tap johnlam90/tap
brew install aws-ssm
```

**Linux/macOS (Manual):**

```bash
curl -L https://github.com/johnlam90/aws-ssm/releases/latest/download/aws-ssm-linux-amd64.tar.gz | tar xz
sudo mv aws-ssm-linux-amd64 /usr/local/bin/aws-ssm
```

**Windows:**
Download from [GitHub Releases](https://github.com/johnlam90/aws-ssm/releases/latest)

**From Source (Go 1.23+):**

```bash
go install github.com/johnlam90/aws-ssm@latest
```

### Basic Usage

**Connect to EC2 instances:**

```bash
# Interactive selection (recommended)
aws-ssm session

# Direct connection
aws-ssm session web-server
aws-ssm session i-1234567890abcdef0
aws-ssm session 10.0.1.100

# Execute commands
aws-ssm session web-server "uptime"
```

**EKS Management:**

```bash
# Interactive cluster selection
aws-ssm eks

# Scale nodegroups
aws-ssm eks nodegroup scale

# Update launch template version
aws-ssm eks nodegroup update-lt

# Get specific cluster info
aws-ssm eks production-cluster
```

**Auto Scaling Groups:**

```bash
# Interactive ASG scaling
aws-ssm asg scale
```

**List and inspect:**

```bash
aws-ssm list --tag Environment=production
aws-ssm interfaces web-server  # Network interfaces
```

## 📖 Core Commands

### EC2 Sessions

```bash
# Interactive selection
aws-ssm session

# Direct connection
aws-ssm session web-server
aws-ssm session i-1234567890abcdef0

# Execute commands
aws-ssm session web-server "docker ps"

# Port forwarding
aws-ssm port-forward db-server --remote-port 3306 --local-port 3306
```

**Search syntax:** `name:web state:running tag:Env=prod !state:stopped`

### EKS Management

```bash
# View clusters
aws-ssm eks                          # Interactive selection
aws-ssm eks production-cluster       # Specific cluster

# Nodegroup operations
aws-ssm eks nodegroup scale          # Interactive scaling with retry navigation
aws-ssm eks nodegroup update-lt      # Update launch template version
aws-ssm eks nodegroup scale my-cluster my-nodegroup --desired 5
```

**New in v0.8.0:** Improved navigation flow—press ESC or type "back" to return to selection without restarting the command.

### Auto Scaling Groups

```bash
# Interactive scaling
aws-ssm asg scale                    # Select ASG and scale interactively

# Direct scaling
aws-ssm asg scale my-asg --desired 10 --min 5 --max 20
```

### Instance Management

```bash
# List instances
aws-ssm list --tag Environment=production

# Network interfaces
aws-ssm interfaces web-server
```

## 🔧 Advanced Features

### Search Syntax

- **Tag filters:** `tag:Environment=production`
- **State filters:** `state:running`, `!state:stopped`
- **Name/ID filters:** `name:web`, `id:i-123`
- **IP filters:** `ip:10.0.1.*`
- **Fuzzy search:** `web prod backup`

### Interactive Controls

- **Space:** Toggle multi-select
- **Enter:** Confirm selection
- **ESC/back:** Return to previous selection (EKS/ASG commands)
- **c:** Run commands on selected instances
- **p:** Port forwarding

## ⚡ Configuration

### Global Flags

- `--region, -r` - AWS region
- `--profile, -p` - AWS profile
- `--no-color` - Disable colored output

### Config File

Create `~/.aws-ssm/config.yaml`:

```yaml
cache:
  enabled: true
  ttl_minutes: 30
```

Precedence: CLI flags > Environment variables > Config file > Defaults

## 🔐 Requirements

### Prerequisites

- AWS credentials configured (`~/.aws/credentials` or environment)
- SSM Agent 2.3.68.0+ on target EC2 instances
- IAM permissions for EC2, SSM, EKS, and Auto Scaling

### IAM Permissions

**User/Role (minimum required):**

```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "ec2:DescribeInstances",
      "ec2:DescribeNetworkInterfaces",
      "ec2:DescribeLaunchTemplateVersions",
      "ssm:StartSession",
      "ssm:TerminateSession",
      "ssm:SendCommand",
      "eks:DescribeCluster",
      "eks:ListClusters",
      "eks:DescribeNodegroup",
      "eks:ListNodegroups",
      "eks:UpdateNodegroupVersion",
      "autoscaling:DescribeAutoScalingGroups",
      "autoscaling:UpdateAutoScalingGroup"
    ],
    "Resource": "*"
  }]
}
```

**EC2 Instance Role:**

```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "ssm:UpdateInstanceInformation",
      "ssmmessages:*"
    ],
    "Resource": "*"
  }]
}
```

## 🔍 Troubleshooting

| Issue | Solution |
|-------|----------|
| No instances found | Verify AWS credentials, region, and IAM permissions |
| Connection fails | Ensure SSM Agent is running and instance is in "running" state |
| Permission denied | Review IAM permissions for user/role and instance |
| EKS cluster not found | Check `eks:DescribeCluster` permissions and region |

## 🛠️ Development

```bash
git clone https://github.com/johnlam90/aws-ssm.git
cd aws-ssm
go build -o aws-ssm .
go test ./...
```

## 📄 License

MIT License - see [LICENSE](LICENSE)

Not affiliated with AWS. AWS trademarks belong to their owners.

## 🙏 Acknowledgments

Built with [AWS SDK Go v2](https://github.com/aws/aws-sdk-go-v2), [Cobra](https://github.com/spf13/cobra), [ssm-session-client](https://github.com/mmmorris1975/ssm-session-client), and [go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder).
