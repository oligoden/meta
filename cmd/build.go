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

	"github.com/fsnotify/fsnotify"
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
			log.Fatalln("error getting meta filename", err)
			return
		}

		verboseValue, _ := cmd.Flags().GetInt("verbose")
		if verboseValue > 0 {
			fmt.Println("verbosity level", verboseValue)
		}

		fmt.Println("loading metafile")
		f, err := os.Open(metaFileName)
		if err != nil {
			log.Fatalln(err)
			return
		}

		p, err := entity.Load(f)
		if err != nil {
			log.Fatalln("error loading file", metaFileName, err)
			return
		}
		f.Close()

		metaOverrideFileName := strings.TrimSuffix(metaFileName, filepath.Ext(metaFileName)) + ".override" + filepath.Ext(metaFileName)

		f, err = os.Open(metaOverrideFileName)
		if err != nil {
			if !strings.Contains(err.Error(), "no such file or directory") {
				log.Fatalln(err)
				return
			}
		} else {
			err = p.Load(f)
			if err != nil {
				log.Fatalln("error loading file", metaFileName, err)
				return
			}
			f.Close()
		}

		if verboseValue == 3 {
		}

		workLocation, err := cmd.Flags().GetString("work")
		if err != nil {
			fmt.Println("error getting meta folder name", err)
			return
		}

		if workLocation == "" {
			workLocation = p.WorkLocation
		}

		destinationLocation, err := cmd.Flags().GetString("dest")
		if err != nil {
			fmt.Println("error getting destination location", err)
			return
		}

		if destinationLocation == "" {
			destinationLocation = p.DestLocation
		}

		forceFlag, _ := cmd.Flags().GetBool("force")
		ctx := context.WithValue(context.Background(), entity.ContextKey("source"), workLocation)
		ctx = context.WithValue(ctx, entity.ContextKey("destination"), destinationLocation)
		ctx = context.WithValue(ctx, entity.ContextKey("force"), forceFlag)
		ctx = context.WithValue(ctx, entity.ContextKey("verbose"), verboseValue)

		// the configuration is processed and graph build
		fmt.Println("processing configuration")

		fileWatcher, err := fsnotify.NewWatcher()
		if err != nil {
			fmt.Println("error starting file watcher", err)
			return
		}
		defer fileWatcher.Close()

		metafileWatcher, err := fsnotify.NewWatcher()
		if err != nil {
			fmt.Println("error starting meta file watcher", err)
			return
		}
		defer metafileWatcher.Close()
		metafileWatcher.Add(metaFileName)

		ctx = context.WithValue(ctx, entity.ContextKey("watcher"), metafileWatcher)

		rm := refmap.Start()
		err = p.Process(&entity.ProjectBranch{}, rm, ctx)
		if err != nil {
			fmt.Println("error processing project", err)
			return
		}
		err = rm.Evaluate()
		if err != nil {
			fmt.Println("error evaluating graph", err)
			return
		}

		// the project is build
		fmt.Println("building project")

		for _, ref := range rm.ChangedRefs() {
			if verboseValue >= 1 {
				fmt.Println("performing", ref.Identifier())
			}

			if strings.HasPrefix(ref.Identifier(), "file:") {
				filename := filepath.Join(workLocation, ref.Identifier()[5:])
				fileWatcher.Add(filename)
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

		fmt.Println("READY")
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
