package main

import (
	"crypto/sha256"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestWorker(t *testing.T) {
	t.Run("given a list of multiple files, it processes them", func(t *testing.T) {
		afs, files := createAferoTest(t, 10, false)
		e = new(env)

		e.logger, hook = setupLogs()
		e.afs = afs
		ap = NewAsyncProcessor(e, &files)

		var oldPaths []string
		var newPaths []string
		var oldHashes []string

		for i := range files {
			oldPaths = append(oldPaths, files[i].stagingPath)
			newPaths = append(newPaths, newPath(&files[i]))
			content, err := afero.ReadFile(afs, files[i].stagingPath)
			if err != nil {
				t.Fatal(err)
			}
			b := sha256.Sum256(content)
			s := string(b[:])
			oldHashes = append(oldHashes, s)
		}

		ap.processFiles()

		for i := range oldPaths {
			assert.NotEqual(t, oldPaths[i], newPaths[i])
			assert.Equal(t, newPaths[i], files[i].stagingPath)
			assert.Equal(t, oldPaths[i], files[i].oldStagingPath)
			assert.Equal(t, oldHashes[i], string(files[i].hash[:]))
		}

	})
}
