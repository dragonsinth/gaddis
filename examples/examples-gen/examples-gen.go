package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dragonsinth/gaddis/examples"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
)

// TODO: instead of a separate command, this could just be options on the main gaddis command
// -r (recursive) -i (capture input)

const slash = string(filepath.Separator)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	_, file, _, _ := runtime.Caller(0)
	root := filepath.Dir(filepath.Dir(file))
	if !strings.HasSuffix(root, slash) {
		root += slash
	}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".gad") {
			return nil
		}
		if _, err := os.Stat(path + ".in"); err == nil {
			fmt.Println("skipping:", path)
			return nil
		}
		fmt.Println("running:", path)
		if err := examples.RunTest(ctx, path); errors.Is(err, context.Canceled) {
			os.Exit(1)
		} else if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
