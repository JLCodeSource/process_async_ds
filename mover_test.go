package main

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdatePath(t *testing.T) {
	fsys, files = createFSTest(10)
	t.Run("should change path to xxx.processed", func(t *testing.T) {
		for _, f := range files {
			oldDir, fn := path.Split(f.stagingPath)
			parts := strings.Split(oldDir, string(os.PathSeparator))
			lastParts := parts[2:]
			firstParts := parts[:2]
			got := f.UpdatePath()
			fp := strings.Join(firstParts, string(os.PathSeparator))
			lp := strings.Join(lastParts, string(os.PathSeparator))
			want := fp + ".processed" + string(os.PathSeparator) + lp + fn
			assertCorrectString(t, got, want)
		}
	})
}

func TestMoveFile(t *testing.T) {
	fsys, files = createFSTest(10)
	t.Run("should move file to new path & log it", func(t *testing.T) {
		for _, f := range files {
			testLogger, hook = setupLogs()
			oldPath := f.stagingPath
			newPath := f.UpdatePath()
			f.Move(newPath, testLogger)
			assert.NotEqual(t, oldPath, newPath)
			_, err := fsys.Stat(f.stagingPath)
			assert.Nil(t, err)

			gotLogMsg := hook.LastEntry().Message
			wantLogMsg := fmt.Sprintf(fMoveFileLog,
				f.smbName,
				f.id,
				oldPath,
				newPath)

			assertCorrectString(t, gotLogMsg, wantLogMsg)

		}
	})

}
