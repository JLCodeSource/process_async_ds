package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatFiles(t *testing.T) {
	t.Run("statFiles should return the Stat for a file", func(t *testing.T) {
		e.logger, hook = setupLogs()

		afs, files := createAferoTest(t, 2, true)

		e.afs = afs
		ap = NewAsyncProcessor(e, files)
		fileMap := statFiles(files)

		testMap := make(map[string]file)

		for _, f := range files {

			fi, err := afs.Stat(f.stagingPath)
			if err != nil {
				t.Fatal(err)
			}
			testMap[f.id] = file{
				id:          f.id,
				smbName:     fi.Name(),
				createTime:  fi.ModTime(),
				size:        fi.Size(),
				stagingPath: f.stagingPath,
				fileInfo:    fi,
			}

		}
		assert.Equal(t, fileMap, testMap)

	})
}
