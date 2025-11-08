# Remote Command Execution

This document explains how to execute commands on EC2 instances via AWS SSM without starting an interactive shell session.

## Overview

The command execution feature allows you to run commands on EC2 instances and receive the output directly in your terminal, without the need for an interactive SSH or SSM session. This is perfect for:

- Quick health checks
- Automated scripts
- CI/CD pipelines
- Monitoring and diagnostics
- One-off administrative tasks

## Basic Usage

### Syntax

```bash
aws-ssm session <instance-identifier> "<command>"
```

### Simple Examples

```bash
# Check uptime
aws-ssm session web-server "uptime"

# Check disk usage
aws-ssm session i-1234567890abcdef0 "df -h"

# View memory usage
aws-ssm session web-server "free -h"

# Check running processes
aws-ssm session web-server "ps aux"
```

## Instance Identifiers

You can use any of the supported instance identifier types:

```bash
# By instance ID
aws-ssm session i-1234567890abcdef0 "hostname"

# By instance name (Name tag)
aws-ssm session web-server "hostname"

# By tag (Key:Value format)
aws-ssm session Environment:production "hostname"

# By IP address
aws-ssm session 10.0.1.100 "hostname"

# By DNS name
aws-ssm session ec2-54-123-45-67.us-west-2.compute.amazonaws.com "hostname"
```

## Complex Commands

### Commands with Pipes

```bash
# Filter processes
aws-ssm session web-server "ps aux | grep nginx"

# Count log entries
aws-ssm session web-server "cat /var/log/app.log | wc -l"

# Find large files
aws-ssm session web-server "du -sh /var/* | sort -rh | head -10"
```

### Commands with Multiple Arguments

```bash
# Check service status
aws-ssm session web-server "systemctl status nginx"

# Search in files
aws-ssm session web-server "grep -r 'ERROR' /var/log/app/"

# Docker commands
aws-ssm session web-server "docker ps -a"
```

### Commands with Quotes

When your command contains quotes, use proper escaping:

```bash
# Find files modified in last 24 hours
aws-ssm session web-server "find /var/log -name '*.log' -mtime -1"

# Search with grep
aws-ssm session web-server "grep -E 'error|warning' /var/log/syslog"

# Complex awk command
aws-ssm session web-server "ps aux | awk '{print \$1, \$11}'"
```

## Region and Profile

Execute commands with specific AWS region and profile:

```bash
# Specific region
aws-ssm session web-server "uptime" --region us-west-2

# Specific profile
aws-ssm session web-server "uptime" --profile production

# Both region and profile
aws-ssm session web-server "docker ps" --region eu-west-1 --profile staging
```

## How It Works

### Execution Flow

1. **Command Submission**: The CLI sends your command to AWS SSM using the `SendCommand` API
2. **Command Execution**: AWS SSM executes the command on the target instance
3. **Status Polling**: The CLI polls for command completion (every 2 seconds)
4. **Output Retrieval**: Once complete, stdout and stderr are retrieved
5. **Display Results**: Output is displayed in your terminal

### Timeout

- **Default timeout**: 2 minutes
- If the command doesn't complete within 2 minutes, the CLI will exit with a timeout error
- The command may still be running on the instance after timeout

### Command Document

The CLI uses the `AWS-RunShellScript` SSM document, which:
- Executes commands in a shell environment (`/bin/bash` on Linux)
- Captures stdout and stderr
- Returns exit codes

## Output Handling

### Standard Output

```bash
$ aws-ssm session web-server "echo 'Hello World'"
Searching for instance: web-server
Executing command on instance:
  ID:          i-1234567890abcdef0
  Name:        web-server
  Command:     echo 'Hello World'

Command ID: abc123-def456-ghi789
Waiting for command to complete...

Hello World
```

### Standard Error

If the command produces stderr output, it will be displayed after stdout:

```bash
$ aws-ssm session web-server "ls /nonexistent"
Searching for instance: web-server
Executing command on instance:
  ID:          i-1234567890abcdef0
  Name:        web-server
  Command:     ls /nonexistent

Command ID: abc123-def456-ghi789
Waiting for command to complete...

Stderr:
ls: cannot access '/nonexistent': No such file or directory
```

### Failed Commands

If a command fails, an error is returned:

```bash
$ aws-ssm session web-server "exit 1"
Error: failed to execute command: command failed: <error details>
```

## Use Cases

### 1. Health Checks

```bash
# Check if web server is responding
aws-ssm session web-server "curl -s localhost:8080/health"

# Check database connectivity
aws-ssm session db-server "pg_isready -h localhost"

# Check service status
aws-ssm session app-server "systemctl is-active myapp"
```

### 2. Log Inspection

```bash
# View recent logs
aws-ssm session web-server "tail -n 100 /var/log/nginx/access.log"

# Search for errors
aws-ssm session app-server "grep ERROR /var/log/app.log | tail -20"

# Count error occurrences
aws-ssm session web-server "grep -c 'ERROR' /var/log/app.log"
```

### 3. System Diagnostics

```bash
# Check disk space
aws-ssm session web-server "df -h"

# Check memory usage
aws-ssm session web-server "free -h"

# Check CPU load
aws-ssm session web-server "uptime"

# Check network connections
aws-ssm session web-server "netstat -tuln"
```

### 4. Application Management

```bash
# Check Docker containers
aws-ssm session web-server "docker ps"

# View application version
aws-ssm session app-server "cat /app/VERSION"

# Check running Java processes
aws-ssm session app-server "ps aux | grep java"
```

### 5. File Operations

```bash
# Check if file exists
aws-ssm session web-server "test -f /etc/nginx/nginx.conf && echo 'exists' || echo 'not found'"

# View file contents
aws-ssm session web-server "cat /etc/hosts"

# Count files in directory
aws-ssm session web-server "ls -1 /var/log | wc -l"
```

## Scripting and Automation

### In Shell Scripts

```bash
#!/bin/bash

# Check multiple servers
for server in web-1 web-2 web-3; do
    echo "Checking $server..."
    aws-ssm session $server "uptime"
done
```

### In CI/CD Pipelines

```bash
# GitHub Actions example
- name: Check deployment
  run: |
    aws-ssm session production-web "curl -f localhost:8080/health" \
      --region us-west-2 \
      --profile production
```

### Capture Output

```bash
# Capture output to variable
DISK_USAGE=$(aws-ssm session web-server "df -h / | tail -1 | awk '{print \$5}'")
echo "Disk usage: $DISK_USAGE"

# Save output to file
aws-ssm session web-server "cat /var/log/app.log" > local-app.log
```

## Comparison with Interactive Sessions

| Feature | Command Execution | Interactive Session |
|---------|------------------|---------------------|
| **Use Case** | One-off commands | Interactive work |
| **Speed** | Fast (no session setup) | Slower (session setup) |
| **Output** | Captured and returned | Real-time display |
| **Timeout** | 2 minutes max | No timeout |
| **Scripting** | Easy to automate | Harder to automate |
| **Interactivity** | None | Full shell access |

## Limitations

1. **Timeout**: Commands must complete within 2 minutes
2. **No Interactivity**: Cannot handle commands requiring user input
3. **No TTY**: Commands requiring a TTY (like `top`, `vim`) won't work properly
4. **Output Size**: Very large outputs may be truncated
5. **Environment**: Limited environment variables compared to interactive sessions

## Best Practices

### 1. Use Quotes Properly

```bash
# Good - command is properly quoted
aws-ssm session web-server "ps aux | grep nginx"

# Bad - shell will interpret pipe locally
aws-ssm session web-server ps aux | grep nginx
```

### 2. Handle Errors

```bash
# Check exit code
if aws-ssm session web-server "systemctl is-active nginx"; then
    echo "Nginx is running"
else
    echo "Nginx is not running"
fi
```

### 3. Keep Commands Simple

```bash
# Good - simple, focused command
aws-ssm session web-server "df -h /"

# Better for complex tasks - use interactive session
aws-ssm session web-server
```

### 4. Use Timeouts Wisely

For long-running commands, consider using an interactive session instead:

```bash
# This might timeout
aws-ssm session web-server "find / -name '*.log'"

# Better approach - use interactive session
aws-ssm session web-server
# Then run: find / -name '*.log'
```

## Troubleshooting

### Command Times Out

**Problem**: Command doesn't complete within 2 minutes

**Solutions**:
1. Use an interactive session for long-running commands
2. Optimize the command to run faster
3. Break the command into smaller parts

### Command Not Found

**Problem**: `command not found` error

**Solutions**:
1. Ensure the command is installed on the instance
2. Use full path to the command: `/usr/bin/docker ps`
3. Check the instance's PATH environment variable

### Permission Denied

**Problem**: Permission errors when executing commands

**Solutions**:
1. Commands run as `ssm-user` by default
2. Use `sudo` if needed: `aws-ssm session web-server "sudo systemctl status nginx"`
3. Ensure SSM agent has necessary permissions

### No Output

**Problem**: Command executes but produces no output

**Solutions**:
1. Check if the command actually produces output
2. Verify the command syntax is correct
3. Check stderr for error messages

## See Also

- [README.md](README.md) - Main documentation
- [QUICK_REFERENCE.md](QUICK_REFERENCE.md) - Command reference
- [FUZZY_FINDER.md](FUZZY_FINDER.md) - Interactive instance selection

