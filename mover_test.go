package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

const (
	tempdir1 = "/tmp/test1/"
	tempdir2 = "test2/"

	testAppFsMkdirAllErr = "operation not permitted"
)

func TestNewPath(t *testing.T) {
	fsys, files = createFSTest(10)
	t.Run("should return path of xxx.processed", func(t *testing.T) {
		for _, f := range files {
			oldDir, fn := path.Split(f.stagingPath)
			parts := strings.Split(oldDir, string(os.PathSeparator))
			lastParts := parts[2:]
			firstParts := parts[:2]

			got := newPath(&f)
			fp := strings.Join(firstParts, string(os.PathSeparator))
			lp := strings.Join(lastParts, string(os.PathSeparator))
			want := fp + ".processed" + string(os.PathSeparator) + lp + fn
			assertCorrectString(t, got, want)

		}
	})
}

func TestMoveFile(t *testing.T) {
	appFs, files := createAferoTest(t, 10)
	t.Run("should move file to new path & log it", func(t *testing.T) {
		for _, f := range files {
			testLogger, hook = setupLogs()
			oldPath := f.stagingPath
			newPath := newPath(&f)
			f.Move(appFs, testLogger)
			assert.NotEqual(t, oldPath, newPath)
			_, err := appFs.Stat(newPath)
			if err != nil {
				t.Fatal(err)
			}

			assertCorrectString(t, f.stagingPath, newPath)

			gotLogMsg := hook.LastEntry().Message
			wantLogMsg := fmt.Sprintf(fMoveFileLog,
				f.smbName,
				f.id,
				oldPath,
				newPath)

			assertCorrectString(t, gotLogMsg, wantLogMsg)

		}
	})

	t.Run("should report error if folder doesn't exist, log it & create it", func(t *testing.T) {
		for _, f := range files {
			fs := afero.NewMemMapFs()
			afs := &afero.Afero{Fs: fs}
			err := afero.WriteFile(afs, f.stagingPath, []byte{}, 0755)
			if err != nil {
				t.Fatal(err)
			}
			testLogger, hook = setupLogs()

			newPath := newPath(&f)
			dir, _ := path.Split(newPath)

			f.Move(fs, testLogger)

			gotLogMsg := hook.Entries[0].Message
			wantLogMsg := fmt.Sprintf(testFsysDoesNotExistErr, dir[:len(dir)-1])

			assertCorrectString(t, gotLogMsg, wantLogMsg)

		}
	})

}

func TestWrapAferoMkdirAll(t *testing.T) {

	t.Run("wrapAferoMkdirAll should return & log the path", func(t *testing.T) {
		var appFs = afero.NewMemMapFs()
		path := tempdir1 + tempdir2
		testLogger, hook = setupLogs()
		ok := wrapAferoMkdirAll(appFs, path, testLogger)
		defer appFs.RemoveAll(tempdir1)

		assert.True(t, ok)

	})

	t.Run("wrapAferoMkdirAll should panic & log the error on err", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		path := tempdir1 + tempdir2
		// Make the fs readonly to force error
		var appFs = afero.NewReadOnlyFs(afero.NewMemMapFs())
		testLogger, hook = setupLogs()

		panic := func() { wrapAferoMkdirAll(appFs, path, testLogger) }
		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := testAppFsMkdirAllErr

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

}

// Needs huge refactor; but necessary
func createAferoTest(t *testing.T, numFiles int) (afero.Fs, []File) {
	// handle gbr input

	// Create gbrList file
	var list string

	for _, d := range workDirs {
		list = fmt.Sprintf(gbrList, d)
		if err := os.Truncate(list, 0); err != nil {
			log.Printf("Failed to truncate: %v", err)
		} else {
			break
		}

	}

	out, err := os.OpenFile(list, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer out.Close()

	// Create AferoFs
	fs := afero.NewMemMapFs()
	afs := &afero.Afero{Fs: fs}
	var files []File
	var dirs = []string{}

	dirs = append(dirs, "mb/FAN/download/")
	for i := 1; i < 4; i++ {
		dirs = append(dirs, "datav"+strconv.Itoa(i)+"/staging/download/")
	}

	for _, d := range dirs {
		err = fs.MkdirAll(d, 0755)
		assert.Nil(t, err)
	}

	// Create Files
	for i := 0; i < numFiles; i++ {
		f := File{}
		// set name
		guid := genGUID()
		f.smbName = guid
		// set staging path
		dir := dirs[rand.Intn(len(dirs))] //#nosec - random testing code can be insecure
		gbtmp := "{gbtmp-" + string(genRandom(32, fileIDBytes)) + "}"
		f.stagingPath = dir + guid + gbtmp
		// set createTime
		now := time.Now()
		duration := time.Hour * time.Duration(rand.Intn(14)) * time.Duration(24) //#nosec - random testing code can be insecure
		beforeNow := now.Add(-duration)
		f.createTime = beforeNow
		// set size
		f.size = int64(rand.Intn(100000)) //#nosec - random testing code can be insecure
		// set content
		data := genRandom(f.size, letterBytes)
		// set id
		f.id = string(genRandom(32, fileIDBytes))
		// set fanIP
		hostname, _ := os.Hostname()
		ips, _ := net.LookupIP(hostname)
		f.fanIP = ips[0]
		// set datasetID
		f.datasetID = testDatasetID

		err := afero.WriteFile(afs, f.stagingPath, data, 0755)
		if err != nil {
			t.Fatal(err)
		}

		fs.Chtimes(f.stagingPath, beforeNow, beforeNow)

		fi, err := fs.Stat(f.stagingPath)
		f.fileInfo = fi
		assert.Nil(t, err)
		files = append(files, f)

		_, err = out.WriteString(fmt.Sprintf("%v,%v\n", f.id, f.smbName))
		if err != nil {
			log.Fatal(err)
		}

		assert.Nil(t, err)

	}

	return fs, files
}
