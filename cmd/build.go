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
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/oligoden/meta/entity"
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
			log.Fatalln("error getting config filename flag ->", err)
		}

		verboseValue, _ := cmd.Flags().GetInt("verbose")
		if verboseValue > 0 {
			fmt.Println("verbosity level", verboseValue)
		}

		e := entity.NewProject()
		err = e.LoadFile(metaFileName)
		if err != nil {
			log.Fatalln("error loading project config ->", err)
		}

		metaOverrideFileName := strings.TrimSuffix(metaFileName, filepath.Ext(metaFileName)) + ".override" + filepath.Ext(metaFileName)
		if _, err := os.Stat(metaOverrideFileName); err == nil {
			err = e.LoadFile(metaOverrideFileName)
			if err != nil {
				log.Fatalln("error loading project config ->", err)
			}
		}

		if verboseValue >= 1 {
			if e.Environment != "" {
				fmt.Println("environment:", e.Environment)
			} else {
				fmt.Println("no environment set")
			}
		}

		workLocation, err := cmd.Flags().GetString("work")
		if err != nil {
			log.Fatalln("error getting work location flag ->", err)
		}

		if workLocation == "" {
			workLocation = e.WorkLocation
		}

		destLocation, err := cmd.Flags().GetString("dest")
		if err != nil {
			log.Fatalln("error getting destination location flag ->", err)
		}

		if destLocation == "" {
			destLocation = e.DestLocation
		}

		ctx := context.WithValue(context.Background(), refmap.ContextKey("source"), workLocation)
		ctx = context.WithValue(ctx, refmap.ContextKey("destination"), destLocation)
		ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), verboseValue)

		// the configuration is processed and graph build
		fmt.Println("processing...")
		rm := refmap.Start()

		err = e.Process(&entity.ProjectBranch{}, rm, ctx)
		if err != nil {
			log.Fatalln("error processing project ->", err)
		}

		err = rm.Evaluate()
		if err != nil {
			log.Fatalln("error evaluating graph ->", err)
		}
		rm.Output()

		fmt.Println("building...")
		for _, ref := range rm.ChangedRefs() {
			if verboseValue >= 2 {
				fmt.Println("performing", ref.Identifier())
			}

			err = ref.Perform(rm, ctx)
			if err != nil {
				fmt.Println("error performing actions on", ref.Identifier(), err)
				fmt.Println(ref.Output())
			}

			if ref.Output() != "" {
				fmt.Println(ref.Output())
			}
		}
		rm.Finish()

		fmt.Println("done")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().String("metafile", "meta.json", "The meta file")
	buildCmd.Flags().String("dest", "", "The destination directory")
	buildCmd.Flags().String("work", "", "The meta work directory")
	buildCmd.Flags().BoolP("force", "f", false, "Force rebuilding of existing files")
	buildCmd.Flags().IntP("verbose", "v", 0, "Set verbosity to 1, 2 or 3")
}
