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
	"context"
	"fmt"
	"os"

	"github.com/oligoden/meta/entity"
	"github.com/oligoden/meta/project"
	"github.com/oligoden/meta/refmap"

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
		metaFileName, err := cmd.Flags().GetString("metafile")
		if err != nil {
			fmt.Println("error getting meta filename", err)
			return
		}

		f, err := os.Open(metaFileName)
		if err != nil {
			fmt.Println("error opening meta file", metaFileName, err)
			return
		}

		p, err := project.Load(f)
		if err != nil {
			fmt.Println("error loading file", metaFileName, err)
			return
		}
		f.Close()

		verboseValue, _ := cmd.Flags().GetInt("verbose")
		if verboseValue >= 1 {
			fmt.Println("Processing")
		}

		metaFolderName, err := cmd.Flags().GetString("metafolder")
		if err != nil {
			fmt.Println("error getting meta folder name", err)
			return
		}

		rm := refmap.Start()
		err = p.Process(project.BuildBranch, rm)
		if err != nil {
			fmt.Println("error processing project", err)
			return
		}
		rm.Evaluate()

		if verboseValue >= 1 {
			fmt.Println("Building")
		}

		destinationLocation, err := cmd.Flags().GetString("destination")
		if err != nil {
			fmt.Println("error getting destination location", err)
			return
		}
		forceFlag, _ := cmd.Flags().GetBool("force")
		ctx := context.WithValue(context.Background(), entity.ContextKey("source"), metaFolderName)
		ctx = context.WithValue(ctx, entity.ContextKey("destination"), destinationLocation)
		ctx = context.WithValue(ctx, entity.ContextKey("force"), forceFlag)
		ctx = context.WithValue(ctx, entity.ContextKey("watching"), false)
		ctx = context.WithValue(ctx, entity.ContextKey("verbose"), verboseValue)

		for _, ref := range rm.ChangedFiles() {
			err = ref.Perform(ctx)
			if err != nil {
				fmt.Println("error performing file actions,", err)
				return
			}
		}

		for _, ref := range rm.ChangedExecs() {
			err = ref.Perform(ctx)
			if err != nil {
				fmt.Println("error performing exec actions,", err)
				fmt.Println(ref.Identifier())
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().String("metafile", "meta.json", "The meta file")
	buildCmd.Flags().String("destination", "", "The destination folder")
	buildCmd.Flags().String("metafolder", "work", "The meta folder")
	buildCmd.Flags().BoolP("force", "f", false, "Force rebuilding of existing files")
	buildCmd.Flags().IntP("verbose", "v", 0, "Set verbosity to 1, 2 or 3")
}
