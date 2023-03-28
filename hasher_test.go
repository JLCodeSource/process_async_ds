package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"testing"
	"testing/fstest"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

const (
	testFsReadFileErr = "fs.ReadFile err occurred"
)

func TestHasher(t *testing.T) {
	var files *[]File

	fsys = fstest.MapFS{}

	t.Run("should return the hash of the file & log it", func(t *testing.T) {
		fsys, files = createFSTest(t, 10)
		for _, f := range *files {
			testLogger, hook = setupLogs()

			content, _ := fs.ReadFile(fsys, f.stagingPath)
			sha := sha256.Sum256(content)
			f.hasher(fsys, testLogger)
			assert.Equal(t, sha, f.hash)

			gotLogMsg := hook.LastEntry().Message
			wantLogMsg := fmt.Sprintf(fHashLog, f.smbName, f.id, f.hash)
			assertCorrectString(t, gotLogMsg, wantLogMsg)
		}
	})
	t.Run("should log an error on failure to hash", func(t *testing.T) {
		fakeFsReadFile := func(fsys fs.FS, name string) ([]byte, error) {
			err := errors.New(testFsReadFileErr)
			return nil, err
		}
		patch := monkey.Patch(fs.ReadFile, fakeFsReadFile)
		defer patch.Unpatch()

		fsys, files = createFSTest(t, 10)
		for _, f := range *files {
			testLogger, hook = setupLogs()

			f.hasher(fsys, testLogger)

			gotLogMsg := hook.Entries[0].Message
			wantLogMsg := testFsReadFileErr
			assertCorrectString(t, gotLogMsg, wantLogMsg)
		}
	})
}
