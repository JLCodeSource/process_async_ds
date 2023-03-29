package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorker(t *testing.T) {
	t.Run("given a list of 1 file, it processes it", func(t *testing.T) {
		afs, files := createAferoTest(t, 1, false)
		e = new(env)

		e.logger, hook = setupLogs()
		e.afs = afs
		ap = NewAsyncProcessor(e, &files)

		oldPath := files[0].stagingPath
		newPath := newPath(&files[0])

		ap.processFiles()

		assert.NotEqual(t, oldPath, newPath)
		assert.Equal(t, newPath, files[0].stagingPath)
	})
}
