package main

import (
	"testing"

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

		for i := range files {
			oldPaths = append(oldPaths, files[i].stagingPath)
			newPaths = append(newPaths, newPath(&files[i]))
		}

		ap.processFiles()

		for i := range oldPaths {
			assert.NotEqual(t, oldPaths[i], newPaths[i])
			assert.Equal(t, newPaths[i], files[i].stagingPath)
		}

	})
}
