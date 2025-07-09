package main

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestWorker(t *testing.T) {
	// N.B. Need to add failure tests
	t.Run("given a file, it processes it", func(t *testing.T) {
		afs, files := createAferoTest(t, 10, false)
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
		}

		// Confirm first logs correct
		logs := hook.Entries

		// logs[0] checked in f.hasher

		gotLogMsg := logs[1].Message
		wantLogMsg := fmt.Sprintf(adSetOldHashLog, files[0].smbName, files[0].id, files[0].hash)
		assertCorrectString(t, gotLogMsg, wantLogMsg)

		gotLogMsg = logs[2].Message
		wantLogMsg = fmt.Sprintf(adSetOldStagingPathLog, files[0].smbName, files[0].id, files[0].oldStagingPath)
		assertCorrectString(t, gotLogMsg, wantLogMsg)

		// logs[3:5] checked in f.move
		// logs[6] checked in f.hasher

		gotLogMsg = logs[7].Message
		wantLogMsg = fmt.Sprintf(adCompareHashesMatchLog, files[0].smbName, files[0].id, files[0].oldHash, files[0].hash)
		assertCorrectString(t, gotLogMsg, wantLogMsg)

		gotLogMsg = logs[8].Message
		wantLogMsg = fmt.Sprintf(adSetSuccessLog, files[0].smbName, files[0].id, true)
		assertCorrectString(t, gotLogMsg, wantLogMsg)

		gotLogMsg = logs[9].Message
		wantLogMsg = fmt.Sprintf(adReadyForProcessingLog, files[0].smbName, files[0].id, files[0].stagingPath)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestCompareHashes(t *testing.T) {
	t.Run("matching hashes should return true", func(t *testing.T) {
		var f file

		var hash [32]byte

		b := []byte("test")
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
