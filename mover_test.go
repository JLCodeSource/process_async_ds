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
	fsys, files = createFSTest(t, 10)

	t.Run("should return path of xxx.processed", func(t *testing.T) {
		for _, f := range *files {
			oldDir, fn := path.Split(f.stagingPath)
			parts := strings.Split(oldDir, string(os.PathSeparator))
			lastParts := parts[2:]
			firstParts := parts[:2]

			got := newPath(&f) //#nosec - testing code can be insecure
			fp := strings.Join(firstParts, string(os.PathSeparator))
			lp := strings.Join(lastParts, string(os.PathSeparator))
			want := fp + ".processed" + string(os.PathSeparator) + lp + fn
			assertCorrectString(t, got, want)
		}
	})
}

func TestMoveFile(t *testing.T) {
	appFs, files := createAferoTest(t, 10, false)
	t.Run("should move file to new path & log it", func(t *testing.T) {
		for _, f := range files {
			testLogger, hook = setupLogs()
			oldPath := f.stagingPath
			newPath := newPath(&f) //#nosec - testing code can be insecure
			e = &env{
				dryrun: false,
			}
			f.Move(appFs, testLogger)
			assert.NotEqual(t, oldPath, newPath)
			_, err := appFs.Stat(newPath)
			if err != nil {
				t.Fatal(err)
			}

			assertCorrectString(t, f.stagingPath, newPath)

			gotLogMsg := hook.Entries[0].Message
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

			newPath := newPath(&f) //#nosec - testing code can be insecure
			dir, _ := path.Split(newPath)

			f.Move(fs, testLogger)

			gotLogMsg := hook.LastEntry().Message
			wantLogMsg := fmt.Sprintf(testFsysDoesNotExistErr, dir[:len(dir)-1])

			assertCorrectString(t, gotLogMsg, wantLogMsg)
		}
	})

	t.Run("should check for dryrun & log not executing move", func(t *testing.T) {
		for _, f := range files {
			fs := afero.NewMemMapFs()
			afs := &afero.Afero{Fs: fs}
			err := afero.WriteFile(afs, f.stagingPath, []byte{}, 0755)
			if err != nil {
				t.Fatal(err)
			}
			testLogger, hook = setupLogs()

			e = &env{
				dryrun: true,
			}

			f.Move(fs, testLogger)

			gotLogMsg := hook.LastEntry().Message
			wantLogMsg := fmt.Sprintf(fMoveDryRunTrueLog, f.smbName, f.id)

			assertCorrectString(t, gotLogMsg, wantLogMsg)
		}
	})

	t.Run("should check for nondryrun & log executing move", func(t *testing.T) {
		for _, f := range files {
			fs := afero.NewMemMapFs()
			afs := &afero.Afero{Fs: fs}
			err := afero.WriteFile(afs, f.stagingPath, []byte{}, 0755)
			if err != nil {
				t.Fatal(err)
			}
			testLogger, hook = setupLogs()

			e = &env{
				dryrun: false,
			}

			f.Move(fs, testLogger)

			gotLogMsg := hook.Entries[1].Message
			wantLogMsg := fmt.Sprintf(fMoveDryRunFalseLog, f.smbName, f.id)

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

		err := appFs.RemoveAll(tempdir1)
		if err != nil {
			t.Fatal(err)
		}
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

		panicFunc := func() { wrapAferoMkdirAll(appFs, path, testLogger) }
		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := testAppFsMkdirAllErr

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func createAferoTest(t *testing.T, numFiles int, createTestFile bool) (afero.Fs, []File) {
	// createTestFile
	var outSourceFile afero.File

	// Create gbrList file
	var list string

	var err error

	dir := getWorkDir()

	list = fmt.Sprintf(gbrList, dir)

	if err := os.Truncate(list, 0); err != nil {
		t.Errorf("Failed to truncate: %v", err)
	}

	out, err := os.OpenFile(list, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}

	defer out.Close()

	// Create AferoFs
	fs := afero.NewMemMapFs()
	afs := &afero.Afero{Fs: fs}

	if createTestFile {
		testSF := fmt.Sprintf(testSourceFile, dir)

		outSourceFile, err = fs.OpenFile(testSF, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

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

		err = fs.Chtimes(f.stagingPath, beforeNow, beforeNow)
		if err != nil {
			t.Fatal(err)
		}

		fi, err := fs.Stat(f.stagingPath)
		f.fileInfo = fi

		assert.Nil(t, err)

		files = append(files, f)

		_, err = out.WriteString(fmt.Sprintf("%v,%v\n", f.id, f.smbName))
		if err != nil {
			log.Fatal(err)
		}

		assert.Nil(t, err)

		if createTestFile {
			_, err = outSourceFile.WriteString(
				fmt.Sprintf("%v|%v|%v|%v|%v|%v|\n",
					f.smbName,
					f.stagingPath,
					f.createTime.Unix(),
					f.size,
					f.id,
					f.fanIP))
		}

		if err != nil {
			t.Fatal(err)
		}
	}

	return fs, files
}

func getWorkDir() (dir string) {
	for _, dir = range workDirs {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			break
		}
	}

	return
}
