package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

const (
	tempdir1 = "/tmp/test1/"
	tempdir2 = "test2/"

	testOsMkdirAllErr = "os.MkdirAll err occurred"
)

func TestNewPath(t *testing.T) {
	fsys, files = createFSTest(10)
	t.Run("should return path of xxx.processed", func(t *testing.T) {
		for _, f := range files {
			oldDir, fn := path.Split(f.stagingPath)
			parts := strings.Split(oldDir, string(os.PathSeparator))
			lastParts := parts[2:]
			firstParts := parts[:2]

			got := newPath(&f)
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
			newPath := newPath(&f)
			f.Move(fsys, testLogger)
			assert.NotEqual(t, oldPath, newPath)
			_, err := fsys.Stat(newPath)
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

	t.Run("should report error if folder doesn't exist, log it & create it", func(t *testing.T) {
		for _, f := range files {
			testLogger, hook = setupLogs()
			newPath := newPath(&f)

			f.Move(fsys, testLogger)

			_, err := fsys.Stat(newPath)
			assert.Error(t, err)

			gotLogMsg := hook.Entries[0].Message
			wantLogMsg := fmt.Sprintf(testFsysDoesNotExistErr, newPath)

			assertCorrectString(t, gotLogMsg, wantLogMsg)

		}
	})

}

func TestWrapOsMkdirAll(t *testing.T) {

	t.Run("wrapOSMkdirAll should return & log the path", func(t *testing.T) {
		path := tempdir1 + tempdir2
		testLogger, hook = setupLogs()
		ok := wrapOsMkdirAll(path, testLogger)
		defer os.RemoveAll(tempdir1)

		assert.True(t, ok)

	})

	t.Run("wrapOSMkdirAll should panic & log the error on err", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		fakeOsMkdirAll := func(string, fs.FileMode) error {
			err := errors.New(testOsMkdirAllErr)
			return err
		}
		patch2 := monkey.Patch(os.MkdirAll, fakeOsMkdirAll)
		defer patch2.Unpatch()

		path := tempdir1 + tempdir2
		testLogger, hook = setupLogs()

		panic := func() { wrapOsMkdirAll(path, testLogger) }
		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := testOsMkdirAllErr

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

}
