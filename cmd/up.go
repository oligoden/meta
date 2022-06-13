package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/oligoden/meta/entity"
	"github.com/oligoden/meta/refmap"
	"github.com/spf13/cobra"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Continuously run Meta and watch for changes",
	Long: `meta up will stay running until Ctrl-C is pressed.
It will watch for changes to files or the config and rebuild
dependent nodes if an update is detected.
	
See https://oligoden.com/meta for more information.`,

	Run: func(cmd *cobra.Command, args []string) {
		metaFileName, err := cmd.Flags().GetString("metafile")
		if err != nil {
			fmt.Println("error getting config filename flag,", err)
			os.Exit(1)
		}

		verboseValue, _ := cmd.Flags().GetInt("verbose")
		if verboseValue > 0 {
			fmt.Println("verbosity level", verboseValue)
		}

		_, err = os.Stat(metaFileName)
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("project config \"%s\" not found\n", metaFileName)
			os.Exit(0)
		}

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

		origLocation, err := cmd.Flags().GetString("orig")
		if err != nil {
			fmt.Println("error getting origin flag", err)
			return
		}

		if origLocation == "" {
			origLocation = e.OrigLocation
		}

		destLocation, err := cmd.Flags().GetString("dest")
		if err != nil {
			fmt.Println("error getting destination flag", err)
			return
		}

		if destLocation == "" {
			destLocation = e.DestLocation
		}

		ctx := context.WithValue(context.Background(), refmap.ContextKey("orig"), origLocation)
		ctx = context.WithValue(ctx, refmap.ContextKey("dest"), destLocation)
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
		rm.Assess()
		rm.Output()

		fmt.Println("building project...")
		for _, ref := range rm.ChangedRefs() {
			if strings.HasPrefix(ref.Identifier(), "file:") {
				filename := filepath.Join(origLocation, ref.Identifier()[5:])
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
		signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)
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
						relPath, err := filepath.Rel(origLocation, event.Name)
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

					}

					err = e.Process(&entity.ProjectBranch{}, rm, ctx)
					if err != nil {
						fmt.Println("error processing project", err)
						run = false
						break
					}

					if metafileChange {
						rm.Evaluate()
					}

					rm.Propagate()
					rm.Assess()
					rm.Output()

					fmt.Println("rebuilding")
					for _, ref := range rm.ChangedRefs() {
						err = ref.Perform(rm, ctx)
						if err != nil {
							fmt.Println("error performing actions on", ref.Identifier(), err)
							fmt.Println(ref.Output())
							metafileChange = false
							fileChange = false
							break
						}

						if ref.Output() != "" {
							fmt.Println(ref.Output())
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
	rootCmd.AddCommand(upCmd)

	upCmd.Flags().String("metafile", "meta.json", "The meta file")
	upCmd.Flags().String("dest", "", "The base destination directory")
	upCmd.Flags().String("orig", "", "The base origin directory")
	upCmd.Flags().BoolP("force", "f", false, "Force rebuilding of existing files")
	upCmd.Flags().IntP("verbose", "v", 0, "Set verbosity to 1, 2 or 3")
}
