# Terminal User Interface (TUI)

The `aws-ssm tui` command provides an interactive, visually appealing terminal interface for managing AWS resources, inspired by k9s.

## Features

- **Unified Dashboard**: Access all aws-ssm functionality from a single interface
- **Vim-style Navigation**: Use `j/k` for up/down, `h/l` for left/right, or arrow keys
- **Real-time Updates**: View live status of EC2 instances, EKS clusters, and ASGs
- **Beautiful UI**: Colorful, bordered panels with k9s-inspired styling
- **Keyboard-driven**: Fast navigation without touching the mouse
- **Help System**: Press `?` anytime to see available keybindings

## Quick Start

```bash
# Launch TUI with default profile and region
aws-ssm tui

# Launch with specific region
aws-ssm tui --region us-west-2

# Launch with specific profile
aws-ssm tui --profile production

# Launch without colors (for terminals without color support)
aws-ssm tui --no-color
```

## Keybindings

### Global Keys
- `?` - Show help panel
- `q` or `Ctrl+C` - Quit
- `ESC` - Go back to previous view
- `↑/k` - Move cursor up
- `↓/j` - Move cursor down
- `Enter` - Select item

### Resource List Keys
- `g` - Jump to top of list
- `G` - Jump to bottom of list
- `r` - Refresh current view

## Navigation Flow

1. **Dashboard** - Main menu showing all available operations
   - EC2 Instances 🖥️
   - EKS Clusters ☸️
   - Auto Scaling Groups 📊
   - Help ❓

2. **Resource Views** - Select a resource type to view details
   - Browse instances, clusters, or ASGs
   - View real-time status and metadata
   - Navigate with vim keys or arrows

3. **Actions** - Select a resource to perform actions (coming soon)
   - Start SSM session
   - Scale resources
   - View detailed information

## Examples

### Browse EC2 Instances
```bash
aws-ssm tui
# Press Enter on "EC2 Instances"
# Use j/k to navigate the list
# Press ESC to return to dashboard
```

### View EKS Clusters
```bash
aws-ssm tui --region us-east-1
# Press Enter on "EKS Clusters"
# Browse cluster information
# Press r to refresh
```

### Multi-region Workflow
```bash
# Start in us-west-2
aws-ssm tui --region us-west-2
# View resources, then quit (q)

# Switch to us-east-1
aws-ssm tui --region us-east-1
```

## Aliases

The TUI command has convenient aliases:
- `aws-ssm tui`
- `aws-ssm ui`
- `aws-ssm interactive`

All three commands are equivalent.

## Status Bar

The bottom status bar shows:
- Current AWS region
- Active AWS profile
- Current view name

## Color Scheme

The TUI uses a k9s-inspired color palette:
- **Cyan** (#00D9FF) - Primary highlights and titles
- **Pink** (#FF79C6) - Secondary accents
- **Green** (#50FA7B) - Running/healthy states
- **Orange** (#FFB86C) - Pending/warning states
- **Red** (#FF5555) - Stopped/error states
- **Gray** (#6272A4) - Muted/terminated states

## Performance

The TUI is optimized for:
- Fast loading with async data fetching
- Efficient rendering with virtual scrolling
- Minimal AWS API calls with smart caching
- Responsive UI even with large resource lists

## Troubleshooting

### TUI doesn't display colors
Use the `--no-color` flag:
```bash
aws-ssm tui --no-color
```

### Screen rendering issues
Try resizing your terminal or restarting the TUI.

### AWS credentials not found
Ensure your AWS credentials are configured:
```bash
aws configure
# or
export AWS_PROFILE=your-profile
export AWS_REGION=us-west-2
```

## Future Enhancements

Planned features:
- Direct SSM session launch from TUI
- Resource filtering and search
- Multi-select operations
- Custom keybinding configuration
- Saved views and favorites
- Real-time metrics and monitoring

