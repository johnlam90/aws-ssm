package cmd

import (
	"github.com/spf13/cobra"
)

var (
	region          string
	profile         string
	interactive     bool
	interactiveCols []string
	noColor         bool
	width           int
	favorites       bool
	outputFormat    string
	configPath      string
)

var rootCmd = &cobra.Command{
	Use:   "aws-ssm",
	Short: "AWS SSM Session Manager CLI",
	Long: `A native Golang CLI tool for managing AWS SSM sessions.
Connect to EC2 instances using instance ID, DNS name, IP address, or tags.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&region, "region", "r", "", "AWS region (defaults to AWS_REGION env var or default profile region)")
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "AWS profile to use (defaults to AWS_PROFILE env var or default profile)")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to config file (defaults to ~/.aws-ssm/config.yaml; XDG config directory supported if provided)")

	// Enhanced interactive flags
	rootCmd.PersistentFlags().BoolVarP(&interactive, "interactive", "i", false, "Enable enhanced interactive mode with multi-select support")
	rootCmd.PersistentFlags().StringSliceVar(&interactiveCols, "columns", []string{"name", "instance-id", "private-ip", "state"}, "Columns to display in interactive mode (name, instance-id, private-ip, state, type, az)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colors in output")
	rootCmd.PersistentFlags().IntVar(&width, "width", 0, "Terminal width override (0 = auto-detect)")
	rootCmd.PersistentFlags().BoolVar(&favorites, "favorites", false, "Show only bookmarked instances (applies to interactive mode)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format (json) for non-interactive use")
}
