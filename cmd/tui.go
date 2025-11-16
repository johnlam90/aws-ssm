package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/johnlam90/aws-ssm/pkg/aws"
	"github.com/johnlam90/aws-ssm/pkg/ui/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:     "tui",
	Aliases: []string{"ui", "interactive"},
	Short:   "Launch interactive TUI for managing AWS resources",
	Long: `Launch an interactive Terminal User Interface (TUI) for managing AWS resources.

The TUI provides a unified, visually appealing interface for:
- Managing EC2 instances
- Viewing and managing EKS clusters
- Scaling Auto Scaling Groups
- And more...

Features:
- Vim-style keybindings (j/k for up/down, h/l for left/right)
- Real-time resource status updates
- Intuitive navigation with ESC to go back
- Beautiful, colorful interface inspired by k9s

Examples:
  # Launch TUI with default profile and region
  aws-ssm tui

  # Launch TUI with specific region
  aws-ssm tui --region us-west-2

  # Launch TUI with specific profile
  aws-ssm tui --profile production

  # Launch TUI without colors
  aws-ssm tui --no-color`,
	RunE: runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTUI(_ *cobra.Command, _ []string) error {
	// Create a context that can be cancelled with Ctrl+C
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Create AWS client
	client, err := aws.NewClient(ctx, region, profile, configPath)
	if err != nil {
		return fmt.Errorf("failed to create AWS client: %w", err)
	}

	// Get actual region from client
	actualRegion := client.GetRegion()

	// Use the profile from the flag/env since SDK doesn't expose it
	actualProfile := profile
	if actualProfile == "" {
		actualProfile = "default"
	}

	// Create TUI config
	config := tui.Config{
		Region:     actualRegion,
		Profile:    actualProfile,
		ConfigPath: configPath,
		NoColor:    noColor,
	}

	// Create TUI model
	model := tui.NewModel(ctx, client, config)

	// Create Bubble Tea program with alt screen
	// Note: Mouse support is disabled to allow native terminal text selection
	// Users can select text with mouse and copy with Cmd+C (macOS) or Ctrl+Shift+C (Linux)
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(), // Use alternate screen buffer
	)

	// Run the program
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Check if there was an error in the final model
	m, ok := finalModel.(tui.Model)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}

	if m.GetError() != nil {
		return fmt.Errorf("TUI error: %w", m.GetError())
	}

	// Handle post-exit actions
	// Check if we need to start an SSM session
	if instanceID := m.GetPendingSSMSession(); instanceID != nil {
		fmt.Printf("\nStarting session with instance %s...\n\n", *instanceID)
		if err := client.StartSession(ctx, *instanceID); err != nil {
			return fmt.Errorf("failed to start SSM session: %w", err)
		}
		return nil
	}

	// Check if we need to show EKS cluster details
	if clusterName := m.GetPendingEKSCluster(); clusterName != nil {
		fmt.Printf("\nFetching cluster details for %s...\n\n", *clusterName)
		cluster, err := client.DescribeCluster(ctx, *clusterName)
		if err != nil {
			return fmt.Errorf("failed to fetch cluster details: %w", err)
		}
		tui.DisplayClusterInfo(cluster)
		return nil
	}

	return nil
}
