package main

import (
	"os"
	"path"
	"strings"
	"testing"
	"testing/fstest"
)

func TestMover(t *testing.T) {
	var files []File
	fsys = fstest.MapFS{}
	fsys, files = createFSTest(10)
	t.Run("should change path to xxx.processed", func(t *testing.T) {
		for _, f := range files {
			oldDir, fn := path.Split(f.stagingPath)
			parts := strings.Split(oldDir, string(os.PathSeparator))
			lastParts := parts[2:]
			firstParts := parts[:2]
			f.UpdatePath()
			fp := strings.Join(firstParts, string(os.PathSeparator))
			lp := strings.Join(lastParts, string(os.PathSeparator))
			want := fp + ".processed" + string(os.PathSeparator) + lp + fn
			assertCorrectString(t, f.stagingPath, want)
		}
	})
}
