package main

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestWorker(t *testing.T) {
	t.Run("given a list of multiple files, it processes them", func(t *testing.T) {
		afs, files := createAferoTest(t, 1, false)
		e = new(env)

		e.logger, hook = setupLogs()
		e.afs = afs
		ap = NewAsyncProcessor(e, files)

		var oldPaths []string
		var newPaths []string
		var oldHashes []string

		for i := range files {
			oldPaths = append(oldPaths, files[i].stagingPath)
			newPaths = append(newPaths, newPath(files[i]))
			content, err := afero.ReadFile(afs, files[i].stagingPath)
			if err != nil {
				t.Fatal(err)
			}
			b := sha256.Sum256(content)
			s := string(b[:])
			oldHashes = append(oldHashes, s)
		}

		ap.processFiles()

		for i := 0; i < len(oldPaths); i++ {
			assert.NotEqual(t, oldPaths[i], newPaths[i])
			assert.Equal(t, newPaths[i], files[i].stagingPath)
			assert.Equal(t, oldPaths[i], files[i].oldStagingPath)
			assert.Equal(t, oldHashes[i], string(files[i].hash[:]))
			assert.True(t, files[i].success)

			logs := hook.Entries

			gotLogMsg := logs[1].Message
			wantLogMsg := fmt.Sprintf(adSetOldHashLog, files[i].smbName, files[i].id, files[i].hash)
			assertCorrectString(t, gotLogMsg, wantLogMsg)

			gotLogMsg = logs[2].Message
			wantLogMsg = fmt.Sprintf(adSetOldStagingPathLog, files[i].smbName, files[i].id, files[i].oldStagingPath)
			assertCorrectString(t, gotLogMsg, wantLogMsg)

			gotLogMsg = logs[7].Message
			wantLogMsg = fmt.Sprintf(adSetSuccessLog, files[i].smbName, files[i].id, true)
			assertCorrectString(t, gotLogMsg, wantLogMsg)

			gotLogMsg = logs[8].Message
			wantLogMsg = fmt.Sprintf(adReadyForProcessingLog, files[i].smbName, files[i].id, files[i].stagingPath)
			assertCorrectString(t, gotLogMsg, wantLogMsg)
		}
	})
}

func TestCompareHashes(t *testing.T) {
	t.Run("matching hashes should return true", func(t *testing.T) {
		var f file
		var hash [32]byte
		var b []byte

		b = []byte("test")
		hash = sha256.Sum256(b)
		f.hash = hash
		f.oldHash = hash
		assert.True(t, f.compareHashes())
	})
	t.Run("non-matching hashes should return false", func(t *testing.T) {
		var f file
		var hash [32]byte
		var b []byte

		b = []byte("test")
		hash = sha256.Sum256(b)
		f.hash = hash
		b = []byte("difftest")
		hash = sha256.Sum256(b)
		f.oldHash = hash
		assert.False(t, f.compareHashes())
	})
}
