package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"

	"bou.ke/monkey"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

const (
	testFsReadFileErr = "fs.ReadFile err occurred"
)

func TestHasher(t *testing.T) {
	t.Run("should return the hash of 'pre'file & log it", func(t *testing.T) {
		e = new(env)
		afs, files := createAferoTest(t, 1, false)
		e.afs = afs
		ap = NewAsyncProcessor(e, &files)

		for _, f := range files {
			e.logger, hook = setupLogs()
			content, err := afero.ReadFile(afs, f.getStagingPath())
			if err != nil {
				t.Fatal(err)
			}
			prePost := "pre"
			sha := sha256.Sum256(content)
			f.hasher()
			assert.Equal(t, sha, f.getHash())

			gotLogMsg := hook.LastEntry().Message
			wantLogMsg := fmt.Sprintf(fHashLog, f.getSmbName(), f.getID(), prePost, f.getHash())
			assertCorrectString(t, gotLogMsg, wantLogMsg)
		}
	})
	t.Run("should log an error on failure to hash", func(t *testing.T) {
		fakeFsReadFile := func(afs afero.Fs, name string) ([]byte, error) {
			err := errors.New(testFsReadFileErr)
			return nil, err
		}
		patch := monkey.Patch(afero.ReadFile, fakeFsReadFile)
		defer patch.Unpatch()

		afs, files := createAferoTest(t, 10, false)
		e.afs = afs
		for _, f := range files {
			e.logger, hook = setupLogs()

			f.hasher()

			gotLogMsg := hook.Entries[0].Message
			wantLogMsg := testFsReadFileErr
			assertCorrectString(t, gotLogMsg, wantLogMsg)
		}
	})
}
