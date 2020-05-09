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
	"os/signal"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/oligoden/meta/entity"
	"github.com/oligoden/meta/project"
	"github.com/oligoden/meta/refmap"
	"github.com/spf13/cobra"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch meta files for any changes and update source code",
	Long: `Use watch to monitor your meta files for any changes and update source code.
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
		ctx = context.WithValue(ctx, entity.ContextKey("watching"), true)
		ctx = context.WithValue(ctx, entity.ContextKey("verbose"), verboseValue)

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

		for _, ref := range rm.ChangedFiles() {
			filename := filepath.Join(metaFolderName, ref.Identifier())
			fmt.Println("monitoring", filename)
			fileWatcher.Add(filename)
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
				return
			}
		}

		rm.Finish()

		stopSignal := make(chan os.Signal, 1)
		signal.Notify(stopSignal, os.Interrupt, os.Kill)

		ctx = context.WithValue(ctx, entity.ContextKey("force"), true)
		done := make(chan bool)
		go func() {
			run := true
			metafileChange := false
			fileChange := false

			for run {
				select {
				case <-stopSignal:
					fmt.Println()
					fmt.Println("stopping")
					run = false
				case event := <-metafileWatcher.Events:
					fmt.Println("fs event", event.Op, event.Name)
					if event.Op&fsnotify.Write == fsnotify.Write ||
						event.Op&fsnotify.Chmod == fsnotify.Chmod {
						metafileChange = true
					}
				case err := <-metafileWatcher.Errors:
					fmt.Println("error on meta file watcher", err)
					run = false
				case event := <-fileWatcher.Events:
					fmt.Println("fs event", event.Op, event.Name)
					if event.Op&fsnotify.Write == fsnotify.Write ||
						event.Op&fsnotify.Chmod == fsnotify.Chmod {
						relPath, err := filepath.Rel(metaFolderName, event.Name)
						if err != nil {
							fmt.Println("error finding relative path", err)
							continue
						}
						rm.SetUpdate("file:" + relPath)
						fileChange = true
					}
				case err := <-fileWatcher.Errors:
					fmt.Println("error on meta file watcher", err)
					run = false
				case <-time.After(400 * time.Millisecond):
					if !(metafileChange || fileChange) {
						continue
					}
					if metafileChange {
						f, err := os.Open(metaFileName)
						if err != nil {
							fmt.Println("error opening meta file", metaFileName, err)
							run = false
							break
						}

						p, err = p.Load(f)
						f.Close()
						if err != nil {
							fmt.Println("error loading file", metaFileName, err)
							run = false
							break
						}

						err = p.Process(project.BuildBranch, rm)
						if err != nil {
							fmt.Println("error processing project", err)
							run = false
							break
						}
						rm.Evaluate()
					}
					rm.Propagate()

					for _, ref := range rm.ChangedFiles() {
						fmt.Println("rebuilding", ref.Identifier())
						err = ref.Perform(ctx)
						if err != nil {
							fmt.Println("error performing file actions,", err)
							run = false
							break
						}
					}

					if !run {
						break
					}

					for _, ref := range rm.ChangedExecs() {
						err = ref.Perform(ctx)
						if err != nil {
							fmt.Println("error performing exec actions,", err)
							run = false
							break
						}
					}

					rm.Finish()

					metafileChange = false
					fileChange = false
				}
			}
			done <- true
		}()

		<-done
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().String("metafile", "meta.json", "The meta file")
	watchCmd.Flags().String("destination", "", "The destination folder")
	watchCmd.Flags().String("metafolder", "work", "The meta folder")
	watchCmd.Flags().BoolP("force", "f", false, "Force rebuilding of existing files")
	watchCmd.Flags().IntP("verbose", "v", 0, "Set verbosity to 1, 2 or 3")
}
