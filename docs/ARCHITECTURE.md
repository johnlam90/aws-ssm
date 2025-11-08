# Architecture Overview

This document provides an overview of the AWS SSM Manager CLI architecture and design decisions.

## Project Structure

```
aws-ssm/
├── main.go                          # Application entry point
├── cmd/                             # CLI commands (Cobra)
│   ├── root.go                     # Root command with global flags
│   ├── list.go                     # List instances command
│   └── session.go                  # Start SSM session command
├── pkg/
│   └── aws/                        # AWS SDK wrappers
│       ├── client.go               # AWS client initialization
│       ├── instance.go             # EC2 instance discovery
│       └── session.go              # SSM session management
├── examples/                        # Example IAM policies
│   ├── iam-policy.json             # User/role permissions
│   └── instance-role-policy.json   # EC2 instance role permissions
├── go.mod                          # Go module dependencies
├── Makefile                        # Build automation
├── README.md                       # Main documentation
├── QUICKSTART.md                   # Quick start guide
└── ARCHITECTURE.md                 # This file

```

## Design Principles

### 1. Native Go Implementation
- No external dependencies except AWS SDK and Cobra
- Single binary distribution
- Fast startup and execution
- Cross-platform compatibility

### 2. AWS SDK v2
- Uses the latest AWS SDK for Go (v2)
- Better performance and smaller binary size
- Modern context-based API
- Improved error handling

### 3. Flexible Instance Discovery
- Multiple identifier types supported
- Automatic identifier type detection
- Smart fallback mechanisms
- User-friendly error messages

### 4. Clean Architecture
- Separation of concerns (CLI, AWS logic, business logic)
- Testable components
- Minimal coupling between packages
- Clear interfaces

## Component Details

### Main Entry Point (`main.go`)

Simple entry point that delegates to the command package:
- Initializes the CLI
- Handles top-level errors
- Sets exit codes

### Command Package (`cmd/`)

Implements CLI commands using Cobra framework:

**root.go**
- Defines global flags (region, profile)
- Sets up command hierarchy
- Provides help text

**list.go**
- Lists EC2 instances
- Supports tag filtering
- Displays results in table format
- Handles multiple states (running, stopped, etc.)

**session.go**
- Starts SSM sessions
- Validates instance state
- Handles multiple identifier types
- Provides user feedback

### AWS Package (`pkg/aws/`)

Encapsulates AWS SDK interactions:

**client.go**
- Creates AWS clients (EC2, SSM)
- Handles configuration (region, profile)
- Manages credentials
- Supports environment variables

**instance.go**
- Queries EC2 instances
- Implements identifier detection logic
- Supports multiple query types:
  - Instance ID
  - DNS names (public/private)
  - IP addresses (public/private)
  - Tags (Key:Value format)
  - Name tag (shorthand)
- Returns structured instance data

**session.go**
- Starts SSM sessions
- Validates session-manager-plugin
- Handles session lifecycle
- Manages signal handling (Ctrl+C)
- Terminates sessions on exit

## Instance Identifier Detection

The CLI automatically detects the type of identifier:

```
Identifier Pattern          → Detection Logic
─────────────────────────────────────────────────
i-xxxxxxxxxxxxxxxxx         → Instance ID (starts with "i-")
Key:Value                   → Tag filter (contains ":")
10.0.1.100                  → IP address (valid IP format)
ec2-*.compute.amazonaws.com → DNS name (contains domain)
ip-*.compute.internal       → Private DNS (contains domain)
web-server                  → Name tag (fallback)
```

## AWS API Interactions

### EC2 DescribeInstances

Used for instance discovery with filters:
- `instance-id`: Direct instance lookup
- `tag:Key`: Tag-based filtering
- `private-ip-address`: Private IP lookup
- `ip-address`: Public IP lookup
- `private-dns-name`: Private DNS lookup
- `dns-name`: Public DNS lookup
- `instance-state-name`: State filtering

### SSM StartSession

Initiates SSM session:
1. Calls `StartSession` API
2. Receives session credentials
3. Launches `session-manager-plugin`
4. Passes session data to plugin
5. Handles session termination

## Configuration

### AWS Credentials

Supports multiple credential sources (in order):
1. Command-line flags (`--profile`, `--region`)
2. Environment variables (`AWS_PROFILE`, `AWS_REGION`)
3. AWS credentials file (`~/.aws/credentials`)
4. AWS config file (`~/.aws/config`)
5. IAM role (for EC2 instances)

### Region Selection

Region is determined by:
1. `--region` flag
2. `AWS_REGION` environment variable
3. Profile's default region
4. AWS SDK default region

## Error Handling

### User-Friendly Errors

All errors are wrapped with context:
```go
return fmt.Errorf("failed to find instance: %w", err)
```

### Validation

- Instance state validation (must be running)
- Plugin availability check
- Multiple instance detection
- Permission errors with helpful messages

## Security Considerations

### IAM Permissions

Minimal required permissions:
- `ec2:DescribeInstances` - Instance discovery
- `ssm:StartSession` - Session initiation
- `ssm:TerminateSession` - Session cleanup

### Session Security

- Uses AWS SSM Session Manager (encrypted)
- No SSH keys required
- All sessions logged in CloudTrail
- No open ports required

### Credential Handling

- Never stores credentials
- Uses AWS SDK credential chain
- Supports temporary credentials (STS)
- Compatible with IAM roles

## Performance

### Optimizations

- Parallel API calls where possible
- Efficient filtering at API level
- Minimal memory footprint
- Fast binary startup

### Scalability

- Handles large instance fleets
- Pagination support (via AWS SDK)
- Efficient tag filtering
- No local caching (always fresh data)

## Future Enhancements

Potential improvements:

1. **Port Forwarding**: Support for local port forwarding
2. **SSH Tunneling**: SSH over SSM support
3. **Config File**: Store connection aliases
4. **Batch Operations**: Connect to multiple instances
5. **Session History**: Track recent connections
6. **Auto-completion**: Shell completion scripts
7. **Interactive Mode**: TUI for instance selection
8. **Logging**: Structured logging support
9. **Metrics**: Connection statistics
10. **Testing**: Comprehensive test suite

## Dependencies

### Direct Dependencies

- `github.com/aws/aws-sdk-go-v2` - AWS SDK
- `github.com/aws/aws-sdk-go-v2/config` - AWS configuration
- `github.com/aws/aws-sdk-go-v2/service/ec2` - EC2 service
- `github.com/aws/aws-sdk-go-v2/service/ssm` - SSM service
- `github.com/spf13/cobra` - CLI framework

### External Tools

- `session-manager-plugin` - AWS Session Manager plugin (required at runtime)

## Building and Distribution

### Build Process

```bash
go build -o aws-ssm .
```

### Cross-Compilation

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o aws-ssm-linux .

# macOS
GOOS=darwin GOARCH=amd64 go build -o aws-ssm-darwin .

# Windows
GOOS=windows GOARCH=amd64 go build -o aws-ssm.exe .
```

### Binary Size

Typical binary size: ~20-30 MB (includes AWS SDK)

## Testing Strategy

### Unit Tests

Test individual components:
- Instance identifier detection
- Tag parsing
- Error handling

### Integration Tests

Test AWS interactions:
- EC2 API calls
- SSM session creation
- Credential handling

### Manual Testing

Test real-world scenarios:
- Various instance identifiers
- Different AWS regions
- Multiple AWS profiles
- Error conditions

## Contributing

When contributing, maintain:
- Clean architecture
- Comprehensive error handling
- User-friendly messages
- Documentation updates
- Backward compatibility

