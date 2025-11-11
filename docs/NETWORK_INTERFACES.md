# Network Interfaces Documentation

This document explains how to list and inspect network interfaces on EC2 instances using the `aws-ssm interfaces` command.

## Overview

The `interfaces` command displays all network interfaces attached to EC2 instances in a formatted table. This is especially useful for:

- **Kubernetes/EKS nodes** with multiple network interfaces (Multus CNI)
- **Multi-homed instances** with interfaces in different subnets
- **Network troubleshooting** and verification
- **Security group auditing** across interfaces
- **Subnet mapping** and CIDR block inspection

## Basic Usage

### Syntax

```bash
aws-ssm interfaces [instance-identifier] [flags]
```

### Simple Examples

```bash
# Interactive fuzzy finder (no argument - recommended)
aws-ssm interfaces

# List interfaces for specific instance by ID
aws-ssm interfaces i-1234567890abcdef0

# List interfaces by instance name
aws-ssm interfaces web-server

# List interfaces by Kubernetes node name
aws-ssm interfaces ip-100-64-149-165.ec2.internal
```

### Interactive Fuzzy Finder

When you run `aws-ssm interfaces` without any arguments, an interactive fuzzy finder will open:

```bash
aws-ssm interfaces
```

This will:
1. Fetch all running EC2 instances in your current region
2. Display an interactive fuzzy finder interface (fzf-like)
3. Allow you to search/filter instances by typing (searches name, instance ID, IP, state)
4. Navigate with arrow keys (↑/↓)
5. Select an instance with Enter
6. Display network interfaces for the selected instance

The fuzzy finder provides the same experience as the `session` command, making it easy to discover and inspect instances without remembering IDs or names.

## Instance Identifiers

You can use any of the supported instance identifier types:

```bash
# By instance ID
aws-ssm interfaces i-1234567890abcdef0

# By instance name (Name tag)
aws-ssm interfaces web-server

# By tag (Key:Value format)
aws-ssm interfaces Environment:production

# By IP address
aws-ssm interfaces 10.0.1.100

# By DNS name (Kubernetes node name)
aws-ssm interfaces ip-100-64-149-165.ec2.internal
aws-ssm interfaces ip-100-64-149-165.us-west-2.compute.internal
```

## Output Format

The command displays a formatted table with the following columns:

- **Interface**: Interface name (ens5, ens6, etc.)
- **Subnet ID**: AWS subnet ID
- **CIDR**: Subnet CIDR block
- **SG ID**: Security group ID (first security group if multiple)

### Example Output

```
Instance: i-07792557b9c1167a4 | DNS Name: ip-100-64-149-165.ec2.internal | Instance Name: nk-rdc-upf-d-mg-worker-node
Interface | Subnet ID                 | CIDR               | SG ID
--------------------------------------------------------------------------------------
ens5      | subnet-06d8b73f0e116b342  | 100.64.128.0/19    | sg-00f82f14e7abe5298
ens6      | subnet-04f50f436a32a474f  | 10.2.9.0/26        | sg-0fd2dbe3c8853e657
ens7      | subnet-033edf3510a4e3f50  | 10.2.11.0/25       | sg-0fd2dbe3c8853e657
ens8      | subnet-03b757845ae511e01  | 10.2.11.128/25     | sg-0fd2dbe3c8853e657
ens9      | subnet-0b26e8854dfe9337e  | 10.2.12.0/25       | sg-0fd2dbe3c8853e657
ens10     | subnet-0b76e994204bdf6ad  | 10.2.12.128/25     | sg-0fd2dbe3c8853e657
ens11     | subnet-017813dcb57dea9da  | 10.2.13.0/25       | sg-0fd2dbe3c8853e657
ens12     | subnet-0de5b91765a4757f7  | 10.2.13.128/25     | sg-0fd2dbe3c8853e657

Total instances displayed: 1
```

## Interface Naming

The interface names follow the Amazon Linux 2023 naming convention:

- **ens5**: Device index 0 (primary interface)
- **ens6**: Device index 1
- **ens7**: Device index 2
- **ens8**: Device index 3
- And so on...

### Single Network Card Instances

For instances with a single network card, the formula is: `ens{device_index + 5}`

### Multiple Network Card Instances

For instances with multiple network cards (e.g., m6in.32xlarge, c6in.32xlarge), the ENS naming continues sequentially across network cards:

- **Network Card 0**: ens5-ens12 (device indices 0-7)
- **Network Card 1**: ens13-ens20 (device indices 0-7, continuing from card 0)
- **Network Card 2**: ens21-ens28 (device indices 0-7, continuing from card 1)

**Note**: Some network cards may not start at device index 0. For example, if network card 1 starts at device index 1, the interfaces will be ens14-ens20 (skipping ens13).

The tool automatically detects all network cards and calculates the correct ENS device names across all cards.

## Filtering Options

### By Node Name (Kubernetes)

List interfaces for specific Kubernetes nodes:

```bash
# Single node
aws-ssm interfaces -n ip-100-64-149-165.ec2.internal

# Multiple nodes
aws-ssm interfaces -n ip-100-64-149-165.ec2.internal -n ip-100-64-87-43.ec2.internal

# Also supports long DNS format
aws-ssm interfaces -n ip-100-64-149-165.us-west-2.compute.internal
```

### By Instance ID

List interfaces for specific instance IDs:

```bash
# Single instance
aws-ssm interfaces --instance-id i-1234567890abcdef0

# Multiple instances
aws-ssm interfaces --instance-id i-1234567890abcdef0 --instance-id i-0987654321fedcba0
```

### By Tag

Filter instances by tags:

```bash
# Single tag
aws-ssm interfaces --tag Environment:production

# Multiple tags
aws-ssm interfaces --tag Environment:production --tag Role:worker

# Tag with instance name
aws-ssm interfaces --tag Name:web-server
```

### Show All Instances

By default, only running instances are shown. To include stopped instances:

```bash
# Show all instances (running, stopped, stopping, pending)
aws-ssm interfaces --all
```

## Region and Profile

Use with specific AWS region and profile:

```bash
# Specific region
aws-ssm interfaces --region us-west-2

# Specific profile
aws-ssm interfaces --profile production

# Both region and profile
aws-ssm interfaces web-server --region eu-west-1 --profile staging
```

## Use Cases

### 1. Kubernetes/EKS Node Inspection

Verify network configuration for Kubernetes nodes:

```bash
# Check all worker nodes
aws-ssm interfaces --tag kubernetes.io/role:node

# Check specific node from kubectl
kubectl get nodes
# Copy node name (e.g., ip-100-64-149-165.ec2.internal)
aws-ssm interfaces ip-100-64-149-165.ec2.internal
```

### 2. Multus CNI Verification

Verify Multus network attachments:

```bash
# List all interfaces for a pod's node
aws-ssm interfaces ip-100-64-149-165.ec2.internal

# Verify subnet assignments match NetworkAttachmentDefinition
# Compare CIDR blocks with your Multus configuration
```

### 3. Network Troubleshooting

Debug network connectivity issues:

```bash
# Check which subnets an instance is connected to
aws-ssm interfaces web-server

# Verify security groups on each interface
aws-ssm interfaces web-server

# Compare network configuration across instances
aws-ssm interfaces --tag Environment:production
```

### 4. Security Audit

Audit security group assignments:

```bash
# List all interfaces for production instances
aws-ssm interfaces --tag Environment:production

# Check security groups across all worker nodes
aws-ssm interfaces --tag Role:worker

# Verify security group consistency
aws-ssm interfaces --all
```

### 5. Subnet Mapping

Map instances to subnets:

```bash
# See which instances are in which subnets
aws-ssm interfaces

# Check subnet distribution for a specific tag
aws-ssm interfaces --tag Application:database
```

## Integration with kubectl

When working with Kubernetes, you can combine `kubectl` with `aws-ssm interfaces`:

```bash
# Get node name from kubectl
NODE_NAME=$(kubectl get nodes -o jsonpath='{.items[0].metadata.name}')

# List interfaces for that node
aws-ssm interfaces $NODE_NAME

# Or in one line
aws-ssm interfaces $(kubectl get nodes -o jsonpath='{.items[0].metadata.name}')
```

## Scripting Examples

### Check All Nodes in a Cluster

```bash
#!/bin/bash

# Get all node names
NODES=$(kubectl get nodes -o jsonpath='{.items[*].metadata.name}')

# List interfaces for each node
for node in $NODES; do
    echo "=== Node: $node ==="
    aws-ssm interfaces $node
    echo ""
done
```

### Export to CSV

```bash
# List interfaces and save to file
aws-ssm interfaces > network-interfaces.txt

# Or for specific instances
aws-ssm interfaces --tag Environment:production > prod-interfaces.txt
```

### Verify Subnet Consistency

```bash
#!/bin/bash

# Check if all worker nodes have the same number of interfaces
aws-ssm interfaces --tag Role:worker | grep "^ens" | wc -l
```

## Comparison with Python Script

This Go implementation provides the same functionality as the original Python script (`list-interfaces.py`) with these advantages:

| Feature | Python Script | Go CLI |
|---------|--------------|--------|
| **Dependencies** | Requires Python + boto3 | Single binary, no dependencies |
| **Speed** | Slower (Python runtime) | Faster (compiled Go) |
| **Installation** | pip install boto3 | Just download binary |
| **Integration** | Standalone script | Part of unified CLI |
| **Identifier Support** | Limited | Full (ID, name, DNS, IP, tags) |
| **Output Format** | Same | Same |

### Migration from Python Script

If you were using the Python script:

```bash
# Old Python script
./scripts/list-multus-interfaces.py -n ip-100-64-149-165.ec2.internal

# New Go CLI (same output)
aws-ssm interfaces -n ip-100-64-149-165.ec2.internal
```

All the same flags are supported:

- `-n, --node-name`: Kubernetes node DNS name
- `--instance-id`: Specific instance IDs
- `--region`: AWS region
- `--tag`: Tag filters

## Troubleshooting

### No Instances Found

**Problem**: `No instances found matching the criteria.`

**Solutions**:
1. Verify the instance identifier is correct
2. Check you're using the correct region: `--region us-west-2`
3. Check you're using the correct profile: `--profile production`
4. Ensure the instance is running (or use `--all` flag)

### Permission Denied

**Problem**: Permission errors when listing instances

**Solutions**:
1. Ensure your IAM user/role has `ec2:DescribeInstances` permission
2. Ensure you have `ec2:DescribeSubnets` permission for CIDR lookup
3. Check your AWS credentials are configured correctly

### Incorrect Interface Names

**Problem**: Interface names don't match what you see on the instance

**Solutions**:
1. The naming convention (ens5, ens6, etc.) is for Amazon Linux 2023
2. For other OS distributions, interface names may differ (eth0, eth1, etc.)
3. The device index is always correct, just the naming convention varies

### Missing CIDR or Subnet Information

**Problem**: CIDR shows as "N/A"

**Solutions**:
1. Ensure you have `ec2:DescribeSubnets` permission
2. Check if the subnet still exists
3. Verify network connectivity to AWS API

## See Also

- [README.md](README.md) - Main documentation
- [QUICK_REFERENCE.md](QUICK_REFERENCE.md) - Command reference
- [FUZZY_FINDER.md](FUZZY_FINDER.md) - Interactive instance selection
- [COMMAND_EXECUTION.md](COMMAND_EXECUTION.md) - Remote command execution

