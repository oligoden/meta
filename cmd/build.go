/*
Copyright © 2020 Anro le Roux <anro@oligoden.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the source code and return",
	Long: `Use build to do a once-off build of your source code.
By default, only files that do not exist will be build.
Use the force flag (-f) to force rebuilding of all files.`,
	Run: func(cmd *cobra.Command, args []string) {
		// p := project.Load(cmd.Flags().GetString("metafile"))
		// rm := refmap.Start(cmd.Flags().GetString("metafolder"))
		// p.Process(project.BranchBuilder, rm)
		// ctx := context.WithValue(context.Background(), "source", cmd.Flags().GetString("metafolder"))
		// ctx = context.WithValue(ctx, "destination", cmd.Flags().GetString("destination"))
		// ctx = context.WithValue(ctx, "force", cmd.Flags().GetBool("force"))
		// meta.Build(ctx, rm)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.PersistentFlags().String("metafile", "meta.yml", "The meta file")
	buildCmd.PersistentFlags().String("metafolder", "meta", "The meta folder")
	buildCmd.Flags().BoolP("force", "f", false, "Force rebuilding of existing files")
}
