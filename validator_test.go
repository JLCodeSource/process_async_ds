package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatFiles(t *testing.T) {
	t.Run("statFiles should return the Stat for a file", func(t *testing.T) {
		e.logger, hook = setupLogs()

		afs, files := createAferoTest(t, 10, true)

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

func TestGetCheckFileMapMBMetadata(t *testing.T) {
	t.Run("getMBFileMetadata should return the MB file metadata for a file", func(t *testing.T) {
		e.logger, hook = setupLogs()

		afs, files := createAferoTest(t, 1, true)

		e.afs = afs
		ap = NewAsyncProcessor(e, files)

		testMap := make(map[string]file)

		for _, f := range files {
			testMap[f.id] = file{
				id:        f.id,
				datasetID: testDatasetID,
				smbName:   f.smbName,
			}
		}

		filesMap := getCheckFileMapMBMetadata(files)

		assert.Equal(t, testMap, filesMap)
	})
}
