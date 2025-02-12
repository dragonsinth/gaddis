package examples

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const slash = string(filepath.Separator)

func TestExamples(t *testing.T) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	_, file, _, _ := runtime.Caller(0)
	root := filepath.Dir(file)
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
		testname := strings.TrimPrefix(path, root)
		testname = strings.ReplaceAll(testname, slash, "_")
		t.Run(testname, func(t *testing.T) {
			if err := RunTest(ctx, t, path); errors.Is(err, context.Canceled) {
				t.Skip(err)
			} else if err != nil {
				t.Error(err)
			}
		})
		return err
	})
	if err != nil {
		t.Error(err)
	}
}
