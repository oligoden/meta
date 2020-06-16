// +build linux darwin

package example_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/oligoden/meta/entity"
	"github.com/oligoden/meta/project"
	"github.com/oligoden/meta/refmap"
)

func ExampleBuild() {
	f, err := os.Open("meta.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	p, err := project.Load(f)
	if err != nil {
		fmt.Println("error loading file meta.json", err)
		return
	}

	ctx := context.WithValue(context.Background(), entity.ContextKey("source"), "work")
	ctx = context.WithValue(ctx, entity.ContextKey("destination"), ".")
	ctx = context.WithValue(ctx, entity.ContextKey("force"), true)
	ctx = context.WithValue(ctx, entity.ContextKey("watch"), false)
	ctx = context.WithValue(ctx, entity.ContextKey("verbose"), 0)

	rm := refmap.Start()
	err = p.Process(project.BuildBranch, rm, ctx)
	if err != nil {
		fmt.Println("error processing project,", err)
		return
	}
	rm.Evaluate()

	for _, ref := range rm.ChangedFiles() {
		err = ref.Perform(rm, ctx)
		if err != nil {
			fmt.Println("error performing file actions,", err)
			return
		}
	}

	for _, ref := range rm.ChangedExecs() {
		err = ref.Perform(rm, ctx)
		if err != nil {
			fmt.Println("error performing exec actions,", err)
			return
		}
	}

	content, err := ioutil.ReadFile("a/aa.ext")
	if err != nil {
		fmt.Println("error reading file,", err)
		return
	}
	fmt.Println(string(content))

	os.RemoveAll("a/")

	//Output:
	//test
}
