package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatFiles(t *testing.T) {
	t.Run("statFiles should return the Stat for a file", func(t *testing.T) {
		e.logger, hook = setupLogs()

		afs, files := createAferoTest(t, 1, true)

		e.afs = afs
		ap = NewAsyncProcessor(e, files)
		fileMap := statFiles(files)

		testMap := make(map[string]file)
		fi, err := afs.Stat(files[0].stagingPath)
		if err != nil {
			t.Fatal(err)
		}
		testMap[files[0].id] = file{
			id:          files[0].id,
			smbName:     fi.Name(),
			createTime:  fi.ModTime(),
			size:        fi.Size(),
			stagingPath: files[0].stagingPath,
			fileInfo:    fi,
		}

		assert.Equal(t, fileMap, testMap)

	})
}
