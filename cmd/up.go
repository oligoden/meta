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
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/oligoden/meta/entity"
	"github.com/oligoden/meta/refmap"
	"github.com/spf13/cobra"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Bring the services up",
	Long:  `Using up will read and watch all files.`,
	Run: func(cmd *cobra.Command, args []string) {
		metaFileName, err := cmd.Flags().GetString("metafile")
		if err != nil {
			log.Fatalln("error getting meta filename", err)
			return
		}

		verboseValue, _ := cmd.Flags().GetInt("verbose")

		fmt.Println("loading metafile")
		f, err := os.Open(metaFileName)
		if err != nil {
			log.Fatalln(err)
			return
		}

		e := entity.NewProject()
		err = e.Load(f)
		if err != nil {
			log.Fatalln("error loading config", metaFileName, "->", err)
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
			err = e.Load(f)
			if err != nil {
				log.Fatalln("error loading file", metaFileName, err)
				return
			}
			f.Close()
		}

		workLocation, err := cmd.Flags().GetString("work")
		if err != nil {
			fmt.Println("error getting meta folder name", err)
			return
		}

		if workLocation == "" {
			workLocation = e.WorkLocation
		}

		destinationLocation, err := cmd.Flags().GetString("dest")
		if err != nil {
			fmt.Println("error getting destination location", err)
			return
		}

		if destinationLocation == "" {
			destinationLocation = e.DestLocation
		}

		ctx := context.WithValue(context.Background(), refmap.ContextKey("source"), workLocation)
		ctx = context.WithValue(ctx, refmap.ContextKey("destination"), destinationLocation)
		ctx = context.WithValue(ctx, refmap.ContextKey("verbose"), verboseValue)

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

		ctx = context.WithValue(ctx, refmap.ContextKey("watcher"), metafileWatcher)

		rm := refmap.Start()
		err = e.Process(&entity.ProjectBranch{}, rm, ctx)
		if err != nil {
			fmt.Println("error processing project", err)
			return
		}
		err = rm.Evaluate()
		if err != nil {
			fmt.Println("error evaluating graph", err)
			return
		}
		rm.Output()

		fmt.Println("building project...")
		for _, ref := range rm.ChangedRefs() {
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

		stopSignal := make(chan os.Signal, 1)
		signal.Notify(stopSignal, os.Interrupt, os.Kill)
		ctx = context.WithValue(ctx, refmap.ContextKey("force"), true)
		done := make(chan bool)

		// any changes to files are watched
		fmt.Println("watching for changes")

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
						relPath, err := filepath.Rel(workLocation, event.Name)
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
							fmt.Println("error opening meta file", err)
							run = false
							break
						}

						err = e.Load(f)
						if err != nil {
							fmt.Println("error loading file", metaFileName, err)
							run = false
							break
						}
						f.Close()

						f, err = os.Open(metaOverrideFileName)
						if err != nil {
							if !strings.Contains(err.Error(), "no such file or directory") {
								fmt.Println("error opening meta override file", err)
								return
							}
						} else {
							err = e.Load(f)
							if err != nil {
								fmt.Println("error loading meta override file", metaOverrideFileName, err)
								return
							}
							f.Close()
						}

						err = e.Process(&entity.ProjectBranch{}, rm, ctx)
						if err != nil {
							fmt.Println("error processing project", err)
							run = false
							break
						}
						rm.Evaluate()
						rm.Output()
					}
					rm.Propagate()

					fmt.Println("rebuilding")
					for _, ref := range rm.ChangedRefs() {
						err = ref.Perform(rm, ctx)
						if err != nil {
							fmt.Println("error performing actions on", ref.Identifier(), err)
							fmt.Println(ref.Output())
							rm.Finish()
							metafileChange = false
							fileChange = false
							break
						}

						if ref.Output() != "" {
							fmt.Println(ref.Output())
						}
					}

					if metafileChange || fileChange {
						rm.Finish()
						metafileChange = false
						fileChange = false
					}
				}
			}
			done <- true
		}()
		<-done
	},
}

func init() {
	rootCmd.AddCommand(upCmd)

	upCmd.Flags().String("metafile", "meta.json", "The meta file")
	upCmd.Flags().String("dest", "", "The destination directory")
	upCmd.Flags().String("work", "", "The meta work directory")
	upCmd.Flags().BoolP("force", "f", false, "Force rebuilding of existing files")
	upCmd.Flags().IntP("verbose", "v", 0, "Set verbosity to 1, 2 or 3")
}
