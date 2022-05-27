package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the release version",
	Long: `Shows the release version.

See https://oligoden.com/meta for more information.`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("v0.0.17")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
