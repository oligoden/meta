package cmd

import (
	"context"
	"errors"
	"fmt"
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
	Short: "Build the source code and exit",
	Long: `Use build to do a once-off build and exit.
Refer to 'meta up' to keep running and watch for changes.

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
			fmt.Printf(`project config "%s" not found\n`, metaFileName)
			os.Exit(0)
		}

		e := entity.NewProject()
		err = e.LoadFile(metaFileName)
		if err != nil {
			fmt.Println("error loading project config,", err)
			os.Exit(1)
		}

		metaOverrideFileName := strings.TrimSuffix(metaFileName, filepath.Ext(metaFileName)) + ".override" + filepath.Ext(metaFileName)
		if _, err := os.Stat(metaOverrideFileName); err == nil {
			err = e.LoadFile(metaOverrideFileName)
			if err != nil {
				fmt.Println("error loading project config,", err)
				os.Exit(1)
			}
		} else if errors.Is(err, os.ErrNotExist) {
			if verboseValue >= 1 {
				fmt.Println("no config override file used")
			}
		} else {
			fmt.Println("error loading project config override,", err)
			fmt.Println("continuing with normal config")
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
			fmt.Println("error getting work location flag,", err)
			os.Exit(1)
		}

		if workLocation == "" {
			workLocation = e.WorkLocation
		}

		destLocation, err := cmd.Flags().GetString("dest")
		if err != nil {
			fmt.Println("error getting destination location flag,", err)
			os.Exit(1)
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

		pb := &entity.ProjectBranch{}
		err = e.Process(pb, rm, ctx)
		if err != nil {
			fmt.Println("error processing project,", err)
			os.Exit(1)
		}

		err = rm.Evaluate()
		if err != nil {
			fmt.Println("error evaluating graph,", err)
			os.Exit(1)
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
		rm.Assess()
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
