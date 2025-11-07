package cmd

import (
	"github.com/spf13/cobra"
)

var (
	region  string
	profile string
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
}
