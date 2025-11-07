# Pure Go Native Implementation

This document explains the pure Go implementation of AWS SSM Session Manager protocol, which eliminates the need for the external `session-manager-plugin` binary.

## Overview

By default, this CLI uses a **pure Go implementation** of the SSM protocol, powered by the [mmmorris1975/ssm-session-client](https://github.com/mmmorris1975/ssm-session-client) library. This means:

- ✅ **No external dependencies** - no need to install session-manager-plugin
- ✅ **Single binary** - just download and run
- ✅ **Cross-platform** - works on Linux, macOS, and Windows
- ✅ **Faster startup** - no subprocess overhead
- ✅ **Better error handling** - native Go error messages

## How It Works

### Traditional Approach (Plugin-Based)

The traditional AWS CLI approach requires two components:

1. **AWS CLI/SDK** - Calls `StartSession` API to get session credentials
2. **session-manager-plugin** - External binary that:
   - Establishes WebSocket connection to AWS SSM
   - Handles encryption/decryption of session data
   - Manages terminal I/O
   - Implements the SSM protocol

```
┌─────────────┐      StartSession API      ┌─────────────┐
│   AWS CLI   │ ──────────────────────────> │  AWS SSM    │
└─────────────┘                             └─────────────┘
       │                                            │
       │ spawn subprocess                           │
       ▼                                            │
┌─────────────────────┐    WebSocket + SSM         │
│ session-manager-    │ <──────────────────────────┘
│ plugin (external)   │
└─────────────────────┘
```

### Pure Go Approach (Native)

Our implementation combines both components into a single Go binary:

```
┌─────────────────────────────────────────┐
│         aws-ssm CLI (Go)                │
│  ┌────────────────────────────────────┐ │
│  │  AWS SDK v2                        │ │
│  │  - StartSession API                │ │
│  └────────────────────────────────────┘ │
│  ┌────────────────────────────────────┐ │
│  │  ssm-session-client library        │ │
│  │  - WebSocket connection            │ │
│  │  - SSM protocol implementation     │ │
│  │  - Terminal I/O handling           │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
              │
              │ WebSocket + SSM protocol
              ▼
        ┌─────────────┐
        │  AWS SSM    │
        └─────────────┘
```

## Implementation Details

### Key Components

1. **pkg/aws/session_native.go** - Native session implementation
   - `StartNativeSession()` - Starts interactive shell session
   - `StartPortForwardingSession()` - Starts port forwarding session

2. **mmmorris1975/ssm-session-client library** - Core SSM protocol implementation
   - WebSocket connection management
   - SSM protocol encoding/decoding
   - Terminal PTY handling
   - Signal handling (Ctrl+C, window resize, etc.)

### Code Example

```go
// Start a native SSM session (no plugin required)
func (c *Client) StartNativeSession(ctx context.Context, instanceID string) error {
    // The ssm-session-client library accepts AWS SDK v2 config directly
    if err := ssmclient.ShellSession(c.Config, instanceID); err != nil {
        return fmt.Errorf("failed to start native session: %w", err)
    }
    return nil
}
```

### Port Forwarding

The native implementation also supports port forwarding:

```go
func (c *Client) StartPortForwardingSession(ctx context.Context, instanceID string, remotePort, localPort int) error {
    input := &ssmclient.PortForwardingInput{
        Target:     instanceID,
        RemotePort: remotePort,
        LocalPort:  localPort,
    }
    
    if err := ssmclient.PortForwardingSession(c.Config, input); err != nil {
        return fmt.Errorf("failed to start port forwarding: %w", err)
    }
    return nil
}
```

## Features Supported

The native implementation supports all major SSM features:

- ✅ **Interactive Shell Sessions** - Full terminal support with PTY
- ✅ **Port Forwarding** - Forward local ports to remote services
- ✅ **Signal Handling** - Ctrl+C, window resize, etc.
- ✅ **Session Logging** - CloudWatch Logs integration (if configured in AWS)
- ✅ **Session Encryption** - TLS + SSM protocol encryption
- ✅ **Multi-region Support** - Works across all AWS regions

## Switching Between Native and Plugin Mode

You can switch between native and plugin-based modes using the `--native` flag:

```bash
# Use native Go implementation (default)
aws-ssm session web-server

# Explicitly enable native mode
aws-ssm session web-server --native

# Use plugin-based mode (requires session-manager-plugin)
aws-ssm session web-server --native=false
```

## Why Use Plugin Mode?

You might want to use plugin mode (`--native=false`) if:

- You already have session-manager-plugin installed and configured
- You need features specific to the official plugin
- You're troubleshooting connection issues
- You want to compare behavior between implementations

## Performance Comparison

| Metric | Native Go | Plugin-Based |
|--------|-----------|--------------|
| Startup Time | ~100ms | ~500ms |
| Binary Size | ~15MB | ~5MB CLI + ~10MB plugin |
| Memory Usage | ~20MB | ~30MB (both processes) |
| Dependencies | None | session-manager-plugin |
| Installation | Single binary | Two components |

## Security Considerations

Both implementations are equally secure:

- ✅ Both use TLS for transport encryption
- ✅ Both use AWS IAM for authentication
- ✅ Both support session logging to CloudWatch
- ✅ Both implement the same SSM protocol

The native implementation uses the same AWS SDK and follows the same security best practices.

## Troubleshooting

### Native Mode Issues

If you encounter issues with native mode:

1. **Check SSM Agent version** on the instance (must be 2.3.68.0+)
   ```bash
   # On the instance
   sudo systemctl status amazon-ssm-agent
   ```

2. **Verify IAM permissions** - same permissions required as plugin mode

3. **Try plugin mode** to isolate the issue:
   ```bash
   aws-ssm session <instance> --native=false
   ```

4. **Check network connectivity** - ensure WebSocket connections are allowed

### Common Errors

**"failed to start native session: websocket: bad handshake"**
- Usually indicates network/firewall issues
- Check that outbound HTTPS (443) is allowed
- Verify SSM endpoints are accessible

**"failed to start native session: TargetNotConnected"**
- SSM Agent is not running on the instance
- Instance doesn't have proper IAM role
- Instance is not registered with SSM

## Credits

This implementation is powered by:

- [mmmorris1975/ssm-session-client](https://github.com/mmmorris1975/ssm-session-client) - Pure Go SSM protocol implementation
- [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2) - AWS API client
- [gorilla/websocket](https://github.com/gorilla/websocket) - WebSocket implementation

## References

- [AWS Systems Manager Session Manager](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager.html)
- [SSM Session Manager Plugin Source](https://github.com/aws/session-manager-plugin)
- [SSM Protocol Documentation](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-sessions-start.html)

