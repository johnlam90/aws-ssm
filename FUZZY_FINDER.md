# Interactive Fuzzy Finder

This document explains the interactive fuzzy finder feature for selecting EC2 instances.

## Overview

The fuzzy finder provides an **fzf-like terminal interface** for selecting EC2 instances interactively. It's the easiest and fastest way to connect to instances without remembering instance IDs or names.

## Features

- 🔍 **Real-time Filtering**: Type to filter instances as you type
- ⚡ **Fast Navigation**: Arrow keys for quick navigation
- 📊 **Rich Preview**: Detailed instance information in preview panel
- 🎯 **Smart Search**: Searches across name, instance ID, IP address, and state
- 🎨 **Clean Interface**: fzf-like terminal UI with syntax highlighting
- ⌨️ **Keyboard Shortcuts**: Familiar keyboard controls

## Usage

### Basic Usage

Simply run the session command without any arguments:

```bash
aws-ssm session
```

This will:
1. Fetch all running EC2 instances in your current region
2. Display an interactive fuzzy finder interface
3. Allow you to search and select an instance
4. Automatically connect to the selected instance

### With Region and Profile

```bash
# Specific region
aws-ssm session --region us-west-2

# Specific profile
aws-ssm session --profile production

# Both
aws-ssm session --region eu-west-1 --profile staging
```

## Interface

### Main Display

Each instance is displayed in a single line with the following format:

```
Name                           | Instance ID         | Private IP      | State
```

Example:
```
web-server-prod-1              | i-1234567890abcdef0 | 10.0.1.100      | running
database-primary               | i-0987654321fedcba0 | 10.0.2.50       | running
cache-redis-01                 | i-abcdef1234567890  | 10.0.3.25       | running
```

### Preview Panel

The preview panel (right side) shows detailed information about the currently selected instance:

```
Instance Details
================

Name:          web-server-prod-1
Instance ID:   i-1234567890abcdef0
State:         running
Instance Type: t3.medium
Private IP:    10.0.1.100
Public IP:     54.123.45.67
Private DNS:   ip-10-0-1-100.ec2.internal
Public DNS:    ec2-54-123-45-67.us-west-2.compute.amazonaws.com
AZ:            us-west-2a

Tags:
  Environment: production
  Role: web-server
  Team: backend
  ManagedBy: terraform
```

## Keyboard Controls

| Key | Action |
|-----|--------|
| **↑** / **↓** | Navigate up/down through instances |
| **Type** | Filter instances in real-time |
| **Enter** | Select instance and connect |
| **Esc** | Cancel and exit |
| **Ctrl+C** | Cancel and exit |
| **Ctrl+N** | Move down (alternative to ↓) |
| **Ctrl+P** | Move up (alternative to ↑) |

## Search/Filter Behavior

The fuzzy finder searches across multiple fields:

- **Instance Name** (from Name tag)
- **Instance ID**
- **Private IP Address**
- **State**

### Search Examples

**Search by name:**
```
Type: "web"
Matches: web-server-prod-1, web-server-staging-2, api-web-gateway
```

**Search by instance ID:**
```
Type: "i-1234"
Matches: i-1234567890abcdef0
```

**Search by IP:**
```
Type: "10.0.1"
Matches: All instances with IPs starting with 10.0.1
```

**Search by state:**
```
Type: "running"
Matches: All running instances
```

**Fuzzy matching:**
```
Type: "wbprd"
Matches: web-server-prod-1 (matches w-e-b-p-r-o-d)
```

## Filtering Behavior

### Only Running Instances

By default, the fuzzy finder only shows **running instances**. This is to prevent accidentally connecting to stopped or terminated instances.

If you have instances in other states (stopped, stopping, pending), you'll see a message like:

```
Error: no running instances found in region us-west-2 (found 5 instances in other states)
```

To connect to stopped instances, you need to:
1. Start the instance first
2. Or use the direct connection method with instance ID

### Empty Results

If no instances are found, you'll see:

```
Error: no instances found in region us-west-2
```

This could mean:
- No EC2 instances in the region
- No instances match your AWS credentials/permissions
- Wrong region selected

## Examples

### Example 1: Quick Connection

```bash
$ aws-ssm session
Opening interactive instance selector...
(Use arrow keys to navigate, type to filter, Enter to select, Esc to cancel)

# Fuzzy finder appears
# Type "web" to filter
# Press Enter on desired instance
# Automatically connects
```

### Example 2: Multi-Region Search

```bash
# Search in us-east-1
$ aws-ssm session --region us-east-1

# Search in eu-west-1
$ aws-ssm session --region eu-west-1

# Search in ap-southeast-1
$ aws-ssm session --region ap-southeast-1
```

### Example 3: Multi-Account Search

```bash
# Development account
$ aws-ssm session --profile dev

# Staging account
$ aws-ssm session --profile staging

# Production account
$ aws-ssm session --profile prod
```

## Comparison with Direct Connection

| Feature | Fuzzy Finder | Direct Connection |
|---------|--------------|-------------------|
| **Ease of Use** | ⭐⭐⭐⭐⭐ Very easy | ⭐⭐⭐ Moderate |
| **Speed** | ⭐⭐⭐⭐ Fast | ⭐⭐⭐⭐⭐ Very fast |
| **Discovery** | ⭐⭐⭐⭐⭐ Excellent | ⭐⭐ Limited |
| **Preview** | ⭐⭐⭐⭐⭐ Full details | ⭐ None |
| **Filtering** | ⭐⭐⭐⭐⭐ Real-time | ⭐⭐⭐ Manual |
| **Best For** | Exploring, browsing | Known instance ID/name |

## Tips and Tricks

### 1. Quick Filter Patterns

Use short, unique patterns to quickly find instances:

```bash
# Instead of typing full name "web-server-production-1"
# Just type: "wbprd1"
```

### 2. Use Tags Effectively

Tag your instances with meaningful names for easier discovery:

```bash
# Good tags
Name: web-server-prod-1
Environment: production
Role: web-server

# Bad tags
Name: i-1234567890abcdef0
Environment: env1
Role: server
```

### 3. Combine with Shell Aliases

Create shell aliases for common workflows:

```bash
# In ~/.bashrc or ~/.zshrc
alias ssm-prod='aws-ssm session --profile production'
alias ssm-dev='aws-ssm session --profile development'
alias ssm-us='aws-ssm session --region us-east-1'
alias ssm-eu='aws-ssm session --region eu-west-1'
```

### 4. Use Region-Specific Searches

If you have instances in multiple regions, use region flags:

```bash
# Quick search in specific region
aws-ssm session -r us-west-2
```

## Troubleshooting

### Fuzzy Finder Doesn't Appear

**Problem**: Command exits immediately without showing fuzzy finder

**Solutions**:
1. Make sure you're not providing an instance identifier argument
2. Check that you have running instances in the region
3. Verify your AWS credentials are configured

### No Instances Shown

**Problem**: Fuzzy finder shows "no instances found"

**Solutions**:
1. Check you're using the correct region: `--region <region>`
2. Verify your AWS profile: `--profile <profile>`
3. Ensure instances are in "running" state
4. Check IAM permissions for EC2 describe actions

### Terminal Display Issues

**Problem**: Fuzzy finder display is garbled or not rendering correctly

**Solutions**:
1. Ensure your terminal supports ANSI colors
2. Try resizing your terminal window
3. Update your terminal emulator
4. Check `TERM` environment variable is set correctly

### Slow Performance

**Problem**: Fuzzy finder is slow to load or filter

**Solutions**:
1. You may have many instances - this is normal
2. Use more specific region to reduce instance count
3. Consider using direct connection for known instances

## Implementation Details

The fuzzy finder is implemented using:

- **Library**: [ktr0731/go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder)
- **Terminal UI**: tcell-based terminal interface
- **Search Algorithm**: Fuzzy matching algorithm similar to fzf
- **Preview**: Real-time preview panel with instance details

## See Also

- [README.md](README.md) - Main documentation
- [QUICK_REFERENCE.md](QUICK_REFERENCE.md) - Command reference
- [NATIVE_IMPLEMENTATION.md](NATIVE_IMPLEMENTATION.md) - Pure Go implementation details

