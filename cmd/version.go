package cmd

import (
	"fmt"

	"github.com/johnlam90/aws-ssm/pkg/version"
	"github.com/spf13/cobra"
)

var (
	versionShort bool
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Display version information including build details.`,
	Run:   runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&versionShort, "short", "s", false, "Print only the version number")
}

func runVersion(_ *cobra.Command, _ []string) {
	if versionShort {
		fmt.Println(version.GetVersion())
		return
	}

	fmt.Printf("aws-ssm version %s\n", version.GetFullVersion())
	fmt.Println()
	fmt.Println("Build Information:")
	buildInfo := version.GetBuildInfo()
	fmt.Printf("  Version:    %s\n", buildInfo["version"])
	fmt.Printf("  Commit:     %s\n", buildInfo["commit"])
	fmt.Printf("  Build Date: %s\n", buildInfo["build_date"])
	fmt.Printf("  Go Version: %s\n", buildInfo["go_version"])
	fmt.Printf("  OS/Arch:    %s/%s\n", buildInfo["os"], buildInfo["arch"])
}
