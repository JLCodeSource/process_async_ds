package main

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
)

/* func TestWorker(t *testing.T) {
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
			oldPaths = append(oldPaths, files[i].getStagingPath())
			newPaths = append(newPaths, newPath(files[i]))
			content, err := afero.ReadFile(afs, files[i].getStagingPath())
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
			assert.Equal(t, newPaths[i], files[i].getStagingPath())
			assert.Equal(t, oldPaths[i], files[i].getOldStagingPath())
			hash := files[i].getHash()
			assert.Equal(t, oldHashes[i], string(hash[:]))
			assert.True(t, files[i].getSuccess())
		}

	})
} */

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
