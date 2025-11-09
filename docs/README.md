# AWS SSM Manager CLI

A blazing-fast Golang CLI for managing AWS EC2 instances and EKS clusters with **zero dependencies**. Connect to instances and manage clusters using instance ID, DNS, IP, tags, or interactive selection - **no bastion host or AWS session-manager-plugin required!**

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## ✨ Key Features

- 🎯 **Pure Go Implementation** - Single binary, no external dependencies
- 🔍 **Enhanced Interactive Selection** - Fuzzy finder with multi-select and rich search
- 🚀 **Remote Command Execution** - Run commands without interactive sessions  
- 🌐 **Network Interface Inspection** - View all interfaces (Multus, EKS, etc.)
- ☸️ **EKS Cluster Management** - Interactive cluster selection and detailed info
- 🔌 **Port Forwarding** - Forward local ports to remote services
- 💾 **Smart Caching** - Performance-optimized with intelligent caching
- 🎨 **Rich Search Syntax** - `name:web state:running tag:Env=prod !tag:Env=dev`
- 📊 **Flexible Display** - Customizable columns and color themes
- 🔐 **Secure by Design** - Uses AWS SSM Session Manager

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

**From Source:**

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

**EKS Cluster Management:**

```bash
# Interactive cluster selection
aws-ssm eks

# Get specific cluster info
aws-ssm eks production-cluster
```

**List and inspect instances:**

```bash
aws-ssm list --tag Environment=production
aws-ssm interfaces web-server  # Network interfaces
```

## 📖 Core Commands

### `aws-ssm session` - EC2 Instance Sessions

**Interactive mode (fuzzy finder):**

```bash
aws-ssm session                    # Opens interactive selector
aws-ssm session --interactive      # Enhanced mode with multi-select
aws-ssm session --columns name,state,az  # Custom columns
aws-ssm session --favorites        # Show only bookmarked instances
```

**Rich search syntax:**

```bash
name:web state:running tag:Env=prod        # Filter by multiple criteria
!state:stopped                             # Negative filters
tag:Team=backend ip:10.0.1.*               # Complex queries
```

**Direct connection:**

```bash
aws-ssm session web-server                 # By name tag
aws-ssm session i-1234567890abcdef0        # By instance ID
aws-ssm session 10.0.1.100                 # By IP address
aws-ssm session ec2-54-123-45-67.compute   # By DNS
aws-ssm session Environment:production     # By tag
```

**Remote command execution:**

```bash
aws-ssm session web-server "docker ps"
aws-ssm session i-123 "systemctl status nginx"
aws-ssm session db "ps aux | grep postgres"
```

**Port forwarding:**

```bash
aws-ssm port-forward db-server --remote-port 3306 --local-port 3306
aws-ssm port-forward bastion --remote-port 5432 --local-port 5432
```

### `aws-ssm eks` - EKS Cluster Management

**Interactive cluster selection:**

```bash
aws-ssm eks                    # Opens interactive cluster selector
aws-ssm eks --region us-west-2 # Specific region
aws-ssm eks --profile prod     # Specific profile
```

**Direct cluster access:**

```bash
aws-ssm eks my-cluster         # By cluster name
aws-ssm eks production --region eu-west-1
```

Displays comprehensive cluster information including status, networking, node groups, and security configuration.

### `aws-ssm list` - Instance Listing

```bash
aws-ssm list                                    # All running instances
aws-ssm list --tag Environment=production       # Filter by tags
aws-ssm list --all                              # Include stopped instances
aws-ssm list --region eu-west-1 --profile prod  # Specific region/profile
```

### `aws-ssm interfaces` - Network Interface Inspection

```bash
aws-ssm interfaces                      # Interactive selection
aws-ssm interfaces web-server            # By instance name
aws-ssm interfaces i-1234567890abcdef0   # By instance ID
aws-ssm interfaces --tag Environment:prod # Filter by tags
```

Perfect for instances with multiple network interfaces (Multus, EKS, etc.).

## 🔧 Advanced Features

### Enhanced Search Syntax

The fuzzy finder supports powerful filtering:

- **Tag filters:** `tag:Environment=production`, `team:backend`
- **State filters:** `state:running`, `state:stopped`
- **Name/ID filters:** `name:web`, `id:i-123`
- **IP/DNS filters:** `ip:10.0.1.*`, `dns:*.compute.amazonaws.com`
- **Exclusion:** `!state:stopped`, `!tag:Env=dev`
- **Existence:** `has:Environment`, `missing:Team`
- **Fuzzy search:** `web prod backup` (space-separated terms)

### Multi-Select and Batch Operations

Interactive mode supports:

- **Space:** Toggle selection for multiple instances
- **Enter:** Connect to selected instances
- **c:** Run commands on selected instances
- **p:** Set up port forwarding

### Performance Optimizations

- **Intelligent Caching:** Reduces API calls with configurable TTL
- **Smart Filtering:** Client-side filtering for large instance sets
- **Rate Limiting:** Built-in AWS API rate limiting and retry logic
- **Streaming:** Efficient handling of large instance lists

## ⚡ Configuration

### Global Flags

All commands support:

- `--region, -r` - AWS region
- `--profile, -p` - AWS profile
- `--no-color` - Disable colored output
- `--width` - Set display width

### Configuration File

Create `~/.config/aws-ssm/config.yaml`:

```yaml
interactive:
  max_instances: 1000
  no_color: false
  width: 120
  
cache:
  enabled: true
  ttl_minutes: 30
  cache_dir: ~/.cache/aws-ssm
```

## 🔐 Requirements

### AWS Configuration

- AWS credentials configured (`~/.aws/credentials` or environment variables)
- IAM permissions for EC2, SSM, and EKS (see below)

### EC2 Instance Requirements

- SSM Agent 2.3.68.0+ installed and running
- IAM role with SSM permissions

### IAM Permissions

**User/Role Permissions:**

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeInstances",
        "ec2:DescribeNetworkInterfaces", 
        "ssm:StartSession",
        "ssm:TerminateSession",
        "ssm:SendCommand",
        "ssm:GetCommandInvocation",
        "eks:DescribeCluster",
        "eks:ListClusters"
      ],
      "Resource": "*"
    }
  ]
}
```

**EC2 Instance Role:**

```json
{
  "Version": "2012-10-17", 
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ssm:UpdateInstanceInformation",
        "ssmmessages:*"
      ],
      "Resource": "*"
    }
  ]
}
```

## 🔍 Troubleshooting

**No instances found:**

- Verify AWS credentials and region
- Check IAM permissions for EC2 DescribeInstances

**Connection fails:**

- Ensure SSM Agent is running on target instance
- Verify instance is in "running" state
- Check instance has SSM permissions

**Permission denied:**

- Review IAM permissions for your user/role
- Ensure EC2 instance has SSM agent permissions

**EKS cluster not found:**

- Verify `eks:DescribeCluster` and `eks:ListClusters` permissions
- Check cluster exists in specified region

## 🛠️ Development

```bash
# Build from source
git clone https://github.com/johnlam90/aws-ssm.git
cd aws-ssm
go build -o aws-ssm .

# Run tests
go test ./...

# Install development version
go install .
```

## 📚 Documentation

- [Installation Guide](docs/INSTALLATION.md) - Detailed setup for all platforms
- [Quick Reference](docs/QUICK_REFERENCE.md) - Command cheat sheet
- [Fuzzy Finder Guide](docs/FUZZY_FINDER.md) - Advanced search techniques
- [EKS Management](docs/EKS_MANAGEMENT.md) - Cluster management details
- [Architecture](docs/ARCHITECTURE.md) - Technical implementation details

## 🤝 Contributing

Contributions welcome! Please see [Contributing Guide](docs/CONTRIBUTING.md) for details.

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2) - AWS integration
- [Cobra](https://github.com/spf13/cobra) - CLI framework  
- [ssm-session-client](https://github.com/mmmorris1975/ssm-session-client) - SSM protocol
- [go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder) - Interactive selection
- [aws-gate](https://github.com/xen0l/aws-gate) - Original inspiration
