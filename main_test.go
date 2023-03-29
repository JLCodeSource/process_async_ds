package main

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"testing"
	"testing/fstest"
	"time"

	"bou.ke/monkey"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	log "github.com/JLCodeSource/process_async_ds/logger"
)

const (
	testDatasetID    = "41545AB0788A11ECBD0700155D014E0D"
	testFileID       = "D5B58980A3E311EBBA0AB026285E5610"
	testBadFileID    = "3D3D0900791F11ECB6BD00155D014E0D"
	testName         = "test.txt"
	testPath         = "data1/staging/test.txt"
	testBadPath      = "/not/a/path"
	testMismatchPath = "data1/staging/testMismatch.txt"
	testNotADataset  = "123"
	testSourceFile   = "%vtest.file"

	testArgsSourceFile = "-sourcefile=%vtest.file"
	testArgsDataset    = "-datasetid=%v"
	testArgsDays       = "-days=123"
	testArgsHelp       = "-help"

	testPostArgsSourceFile = "%vtest.file"
	testPostArgsDays       = int64(123)

	osPanicTrue  = "os.Exit called"
	osPanicFalse = "os.Exit was not called"

	testEmptyRootErr        = "stat %v: os: DirFS with empty root"
	testOpenDoesNotExistErr = "open %v: file does not exist"
	testRegexMatchErr       = "Regex match errored"
	testHostnameErr         = "os.Hostname err occurred"
	testChdirErr            = "os.Chdir err occurred"
	testGetwdErr            = "os.Getwd err occurred"
	testOsExecutableErr     = "os.Executable err occurred"
	testLookupIPErr         = "net.LookupIP err occurred"
	//testFileInfoErr         = "fs.FileInfo err occurred"

	testKarachiTime       = "Asia/Karachi"
	testKarachiDate       = "Mon Jan 30 17:55:14 PKT 2023"
	testKarachiDateParsed = "2023-01-30 17:55:14 +0500 PKT"
	testKarachiDateUTC    = "2023-01-30 12:55:14 +0000 UTC"
)

var (
	// setup logger
	testLogger *logrus.Logger
	hook       *test.Hook

	// setup env
	//testEnv env
	limit time.Time
	ip    net.IP

	// setup file
	file File
	now  time.Time

	// setup fsys
	fsys fstest.MapFS
)

type mockAsyncProcessor struct {
	Env   *env
	Files *[]File
}

func (m mockAsyncProcessor) getFiles() *[]File {
	return m.Files
}

func (m mockAsyncProcessor) getEnv() *env {
	return m.Env
}

func TestMainFunc(t *testing.T) {
	//files := &[]File{}
	//e = new(env)
	//ap = NewAsyncProcessor(e, files)
	t.Run("verify main args work", func(t *testing.T) {
		afs, files := createAferoTest(t, 5, true)
		hostname, _ := os.Hostname()
		ips, _ := net.LookupIP(hostname)
		workdir := getWorkDir()

		os.Args = append(os.Args, fmt.Sprintf(testArgsSourceFile, workdir))
		os.Args = append(os.Args, fmt.Sprintf(testArgsDataset, testDatasetID))
		os.Args = append(os.Args, testArgsDays)

		now = time.Now()
		limit = now.Add(-24 * time.Duration(testPostArgsDays) * time.Hour)

		dryrun = true
		testrun = false

		testLogger, hook = setupLogs()

		e.logger = testLogger
		e.afs = afs
		//e.sourceFile = sourceFile
		/* e = &env{
			logger: testLogger,
			//exePath:    exePath,
			//fsys:       fsys,
			afs: afs,
			//sysIP:      sysIP,
			sourceFile: sourceFile,
			//datasetID:  datasetID,
			//limit:      limit,
			//dryrun:     true,
			//testrun:    false,
		} */

		ap = mockAsyncProcessor{
			Env:   e,
			Files: &files,
		}

		main()

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(eMatchAsyncProcessedDSTrueLog, e.datasetID, testDatasetID)

		assertCorrectString(t, gotLogMsg, wantLogMsg)

		f, err := e.afs.Open(e.sourceFile)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		assert.Equal(t, e.sysIP, ips[0])

		_, err = f.Stat()
		assert.NoError(t, err)

		assertCorrectString(t, e.sourceFile, fmt.Sprintf(testPostArgsSourceFile, workdir))

		assertCorrectString(t, e.datasetID, testDatasetID)

		assertCorrectString(t, e.limit.Format(time.UnixDate), limit.Format(time.UnixDate))

		assert.True(t, e.dryrun)
		assert.False(t, e.testrun)
	})

	t.Run("verify main help out", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		os.Args = append(os.Args, testArgsHelp)

		panicFunc := func() { main() }
		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)
	})

	t.Run("verify hostname failure", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		fakeHostname := func() (string, error) {
			err := errors.New(testHostnameErr)
			return "", err
		}
		patch2 := monkey.Patch(os.Hostname, fakeHostname)
		defer patch2.Unpatch()

		os.Args = append(os.Args, testArgsSourceFile)
		os.Args = append(os.Args, fmt.Sprintf(testArgsDataset, testDatasetID))
		os.Args = append(os.Args, testArgsDays)

		panicFunc := func() { main() }
		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)
	})
}

func TestNewAsyncProcessor(t *testing.T) {
	t.Run("should return the ap", func(t *testing.T) {
		testLogger, _ = setupLogs()
		e = new(env)
		e.logger = testLogger
		files = &[]File{}
		ap = NewAsyncProcessor(e, files)
		ap = mockAsyncProcessor{
			Env:   e,
			Files: files,
		}
		getEnv := ap.getEnv()
		getFiles := ap.getFiles()
		assert.Equal(t, testLogger, getEnv.logger)
		assert.Equal(t, e, getEnv)
		assert.Equal(t, files, getFiles)
	})
}

func TestOsWrapper(t *testing.T) {
	t.Run("wrapOsExecutable should return & log the path", func(t *testing.T) {
		testLogger, hook = setupLogs()

		pwd := wrapOs(testLogger, "os.Executable", os.Executable)
		ex, _ := os.Executable()

		assertCorrectString(t, pwd, ex)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(wrapOsLog, osExecutableLog, ex)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("wrapOs.Executable should panic & log an error on err", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		fakeOsExecutable := func() (string, error) {
			err := errors.New(testOsExecutableErr)
			return "", err
		}
		patch2 := monkey.Patch(os.Executable, fakeOsExecutable)
		defer patch2.Unpatch()

		testLogger, hook = setupLogs()
		panicFunc := func() { wrapOs(testLogger, osExecutableLog, os.Executable) }
		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := testOsExecutableErr
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("wrapOs.Hostname should return & log the hostname", func(t *testing.T) {
		testLogger, hook = setupLogs()

		out := wrapOs(testLogger, osHostnameLog, os.Hostname)
		hostname, _ := os.Hostname()

		assertCorrectString(t, out, hostname)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(wrapOsLog, osHostnameLog, hostname)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
	t.Run("wrapOs.Hostname should panic & log an error on err", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		fakeHostname := func() (string, error) {
			err := errors.New(testHostnameErr)
			return "", err
		}
		patch2 := monkey.Patch(os.Hostname, fakeHostname)
		defer patch2.Unpatch()

		testLogger, hook = setupLogs()
		panicFunc := func() { wrapOs(testLogger, osHostnameLog, os.Hostname) }
		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := testHostnameErr
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestWrapLookupIP(t *testing.T) {
	t.Run("wrapLookupIP should return & log the IP", func(t *testing.T) {
		testLogger, hook = setupLogs()

		hostname, _ := os.Hostname()
		ips, _ := net.LookupIP(hostname)

		ip := wrapLookupIP(testLogger, hostname, net.LookupIP)

		assertCorrectString(t, ip.String(), ips[0].String())

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(wrapLookupIPLog, hostname, ip)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("wrapLookupIP should panic & log the err on err", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		fakeLookupIP := func(string) ([]net.IP, error) {
			err := errors.New(testLookupIPErr)
			return nil, err
		}
		patch2 := monkey.Patch(net.LookupIP, fakeLookupIP)
		defer patch2.Unpatch()

		hostname, _ := os.Hostname()

		testLogger, hook = setupLogs()
		panicFunc := func() { wrapLookupIP(testLogger, hostname, net.LookupIP) }
		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := testLookupIPErr
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("wrapLookupIP should panic & log an err if there are more than one IP", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		fakeLookupIP := func(string) ([]net.IP, error) {
			var ips []net.IP
			ip1 := net.ParseIP("192.168.101.1")
			ip2 := net.ParseIP("192.168.101.2")
			ips = append(ips, ip1)
			ips = append(ips, ip2)
			return ips, nil
		}
		patch2 := monkey.Patch(net.LookupIP, fakeLookupIP)
		defer patch2.Unpatch()

		hostname, _ := os.Hostname()

		testLogger, hook = setupLogs()
		panicFunc := func() { wrapLookupIP(testLogger, hostname, net.LookupIP) }
		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := complexIPLog
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestSetSourceFile(t *testing.T) {
	e = new(env)
	t.Run("check for source file", func(t *testing.T) {
		e.logger, hook = setupLogs()
		e.fsys = fstest.MapFS{
			testPath: {Data: []byte(testContent)},
		}
		files := &[]File{}
		NewAsyncProcessor(e, files)
		e.setSourceFile("", testPath)

		got := e.sourceFile
		want := testPath
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(sourceLog, testPath)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("should handle full path", func(t *testing.T) {
		e.logger, hook = setupLogs()
		e.fsys = fstest.MapFS{
			testPath: {Data: []byte(testContent)},
		}
		fullpath := string(os.PathSeparator) + testPath
		files := &[]File{}
		NewAsyncProcessor(e, files)
		e.setSourceFile("", string(os.PathSeparator)+testPath)

		got := e.sourceFile
		want := string(os.PathSeparator) + testPath
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(sourceLog, fullpath)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("should warn to use full path on local path", func(t *testing.T) {
		e.logger, hook = setupLogs()
		ex, _ := os.Executable()
		dir, _ := path.Split(ex)
		path := dir + testName
		path = path[1:]
		e.fsys = fstest.MapFS{
			path: {Data: []byte(testContent)},
		}
		files := &[]File{}
		NewAsyncProcessor(e, files)
		e.setSourceFile(ex, testName)

		got := e.sourceFile
		want := testName
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(sourceLog, testName)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
	t.Run("check for empty root", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		e.logger, hook = setupLogs()
		e.fsys = os.DirFS("")
		files := &[]File{}
		NewAsyncProcessor(e, files)

		panicFunc := func() {
			e.setSourceFile("/", testDoesNotExistFile)
		}

		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(testEmptyRootErr, testDoesNotExistFile)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
	t.Run("error if file does not exist", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		e.logger, hook = setupLogs()
		e.fsys = fstest.MapFS{
			testMismatchPath: {Data: []byte(testContent)},
		}
		files := &[]File{}
		NewAsyncProcessor(e, files)

		panicFunc := func() {
			e.setSourceFile("/", testDoesNotExistFile)
		}

		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(testOpenDoesNotExistErr, testDoesNotExistFile)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestGetFileList(t *testing.T) {
	e = new(env)
	t.Run("getFileList should return a list of files", func(t *testing.T) {
		e.logger, hook = setupLogs()

		afs, want := createAferoTest(t, 10, true)
		files := &[]File{}

		e.afs = afs

		ap := NewAsyncProcessor(e, files)
		dir := getWorkDir()

		e.sourceFile = fmt.Sprintf(testSourceFile, dir)
		got := ap.getFiles()

		for i := range *got {
			assert.Equal(t, want[i].smbName, (*got)[i].smbName)
			assert.Equal(t, want[i].stagingPath, (*got)[i].stagingPath)
			assert.Equal(t, want[i].createTime.Unix(), (*got)[i].createTime.Unix())
			assert.Equal(t, want[i].size, (*got)[i].size)
			assert.Equal(t, want[i].id, (*got)[i].id)
			assert.Equal(t, want[i].fanIP, (*got)[i].fanIP)
			assert.Equal(t, want[i].fileInfo, (*got)[i].fileInfo)
		}
	})
	t.Run("getFileList should log properly", func(t *testing.T) {
		e.logger, hook = setupLogs()

		afs, want := createAferoTest(t, 1, true)
		files := &[]File{}

		e.afs = afs
		ap := NewAsyncProcessor(e, files)

		dir := getWorkDir()

		e.sourceFile = fmt.Sprintf(testSourceFile, dir)
		ap.getFiles()

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fAddedToListLog,
			want[0].smbName,
			want[0].id,
			want[0].stagingPath,
			want[0].createTime.Unix(),
			want[0].size,
			want[0].fanIP,
			want[0].fileInfo.Name())

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
	t.Run("getFileList should fatal if sourcefile does not exist", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		e.logger, hook = setupLogs()
		afs, _ := createAferoTest(t, 1, true)
		files := &[]File{}
		ap := NewAsyncProcessor(e, files)

		e.afs = afs

		e.sourceFile = testDoesNotExistFile
		panicFunc := func() { ap.getFiles() }

		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(testOpenDoesNotExistErr, testDoesNotExistFile)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestSetDatasetID(t *testing.T) {
	files := &[]File{}
	e = new(env)
	e.afs = afs
	ap := NewAsyncProcessor(e, files)

	t.Run("verify it returns the right dataset id", func(t *testing.T) {
		e.logger, _ = setupLogs()
		e.setDatasetID(testDatasetID)
		got := ap.getEnv().datasetID
		want := testDatasetID

		assertCorrectString(t, got, want)
	})

	t.Run("verify it logs the right dataset id", func(t *testing.T) {
		e.logger, hook = setupLogs()

		e.setDatasetID(testDatasetID)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(datasetLog, testDatasetID)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("verify that it exits if the datasetid is not the right format", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		e.logger, hook = setupLogs()

		panicFunc := func() { e.setDatasetID(testNotADataset) }

		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)
		gotLogMsg := hook.LastEntry().Message
		err := fmt.Sprintf(datasetRegexLog, testNotADataset, regexDatasetMatch)
		wantLogMsg := err

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("verify that it exits if the regex fails", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()
		fakeRegexMatch := func(string, string) (bool, error) {
			err := errors.New(testRegexMatchErr)
			return false, err
		}
		patch2 := monkey.Patch(regexp.MatchString, fakeRegexMatch)
		defer patch2.Unpatch()

		e.logger, hook = setupLogs()
		panicFunc := func() { e.setDatasetID(testNotADataset) }

		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)
		gotLogMsg := hook.LastEntry().Message
		err := testRegexMatchErr
		wantLogMsg := err

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
	t.Run("verify that it exits if the dataset doesn't match asyncprocessed", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		e.logger, hook = setupLogs()
		panicFunc := func() { e.setDatasetID(testID) }

		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(compareDatasetIDNotMatchLog, testID, testDatasetID)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestCompareDatasetId(t *testing.T) {
	files := &[]File{}
	e = new(env)
	NewAsyncProcessor(e, files)
	t.Run("Should return true if datasetid & asyncdelds check match & log it", func(t *testing.T) {
		e.logger, hook = setupLogs()
		e.compareDatasetID(testDatasetID)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(compareDatasetIDMatchLog, testDatasetID, testDatasetID)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
	t.Run("Should return false if datasetid & asyncdel metadata check do not match & log it", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		e.logger, hook = setupLogs()
		panicFunc := func() { e.compareDatasetID(testID) }
		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(compareDatasetIDNotMatchLog, testID, testDatasetID)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestSetTimeLimit(t *testing.T) {
	files := &[]File{}
	e = new(env)
	NewAsyncProcessor(e, files)
	t.Run("zero days", func(t *testing.T) {
		e.logger, hook = setupLogs()

		var days = int64(0)
		e.setTimeLimit(days)

		gotDays := e.limit
		wantDays := time.Time{}

		assertCorrectString(t, gotDays.String(), wantDays.String())

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := timelimitNoDaysLog
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("Multiple days", func(t *testing.T) {
		e.logger, hook = setupLogs()

		now = time.Now()
		days := int64(15)
		daysInTime := time.Duration(-15 * 24 * time.Hour)
		limit = now.Add(daysInTime)

		e.setTimeLimit(days)
		gotDays := e.limit
		wantDays := limit

		assertCorrectString(t, gotDays.Round(time.Millisecond).String(), wantDays.Round(time.Millisecond).String())

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(timelimitDaysLog, days, gotDays)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestSetDryRun(t *testing.T) {
	files := &[]File{}
	e = new(env)
	NewAsyncProcessor(e, files)
	t.Run("default dry run", func(t *testing.T) {
		e.logger, hook = setupLogs()

		e.setDryRun(true)
		got := e.dryrun
		assert.True(t, got)

		typ := reflect.TypeOf(e.afs)
		rofs := new(afero.ReadOnlyFs)
		assert.Equal(t, typ, reflect.TypeOf(rofs))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := dryRunTrueLog

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("non-dry run execute move", func(t *testing.T) {
		e.logger, hook = setupLogs()

		e.setDryRun(false)
		got := e.dryrun

		assert.False(t, got)

		typ := reflect.TypeOf(e.afs)
		osfs := new(afero.OsFs)
		assert.Equal(t, typ, reflect.TypeOf(osfs))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := dryRunFalseLog

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestSetTestRun(t *testing.T) {
	files := &[]File{}
	e = new(env)
	NewAsyncProcessor(e, files)
	t.Run("test run", func(t *testing.T) {
		e.logger, hook = setupLogs()

		e.setTestRun(true)
		got := e.testrun
		assert.True(t, got)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := testRunTrueLog

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("nontest run", func(t *testing.T) {
		e.logger, hook = setupLogs()

		e.setTestRun(false)
		got := e.testrun

		assert.False(t, got)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := testRunFalseLog

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
}

func TestSetPWD(t *testing.T) {
	files := &[]File{}
	e = new(env)
	NewAsyncProcessor(e, files)
	t.Run("setPWD should shift execution to root from current path", func(t *testing.T) {
		e.logger, _ = setupLogs()
		ex, _ := os.Executable()

		got := e.setPWD(ex)
		want := "/"
		assertCorrectString(t, got, want)
	})

	t.Run("setPWD should shift execution to root from any path", func(t *testing.T) {
		e.logger, _ = setupLogs()
		ex := "/workflows/process_async_processed/logger/logger.go"

		got := e.setPWD(ex)
		want := "/"
		assertCorrectString(t, got, want)
	})

	t.Run("setPWD should log an error & panic if it can't chdir", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()
		fakeChdir := func(string) error {
			err := errors.New(testChdirErr)
			return err
		}

		patch2 := monkey.Patch(os.Chdir, fakeChdir)
		defer patch2.Unpatch()
		e.logger, hook = setupLogs()

		panicFunc := func() { e.setPWD(testBadPath) }

		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := testChdirErr

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("setPWD should log an error & panic if it can't Getwd", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		fakeGetwd := func() (string, error) {
			err := errors.New(testGetwdErr)
			return "", err
		}

		patch2 := monkey.Patch(os.Getwd, fakeGetwd)
		defer patch2.Unpatch()
		e.logger, hook = setupLogs()

		panicFunc := func() { e.setPWD(testBadPath) }

		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := testGetwdErr

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

// TestVerifyEnvDataset

func TestVerifyDataset(t *testing.T) {
	t.Run("it should return true if env.datasetID matches asyncProcessed & log it", func(t *testing.T) {
		testLogger, hook = setupLogs()
		e = new(env)
		e.datasetID = testDatasetID
		assert.True(t, e.verifyDataset(testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(eMatchAsyncProcessedDSTrueLog, e.datasetID, testDatasetID)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
	t.Run("returns false if env.DsID does not match asyncProcessed & log it", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		testLogger, hook = setupLogs()
		e = new(env)
		e.datasetID = testWrongDataset

		panicFunc := func() { e.verifyDataset(testLogger) }
		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(eMatchAsyncProcessedDSFalseLog, e.datasetID, testDatasetID)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestFileMetadata(t *testing.T) {
	t.Run("Initial struct test", func(t *testing.T) {
		loc, err := time.LoadLocation(easternTime)
		datestring := testOldDate
		if err != nil {
			t.Fatalf(err.Error())
		}
		datetime, _ := time.ParseInLocation(time.UnixDate, datestring, loc)
		fanIP := net.ParseIP(testIP)
		fsys = fstest.MapFS{
			testPath: {Data: []byte(testContent)},
		}
		fileInfo, _ := fs.Stat(fsys, testPath)
		file = File{
			smbName:     testName,
			stagingPath: testPath,
			createTime:  datetime,
			size:        int64(1024),
			id:          testFileID,
			fanIP:       fanIP,
			datasetID:   testDatasetID,
			fileInfo:    fileInfo,
		}

		gotSmbName := file.smbName
		wantSmbName := testName
		assertCorrectString(t, gotSmbName, wantSmbName)

		gotPath := file.stagingPath
		wantPath := testPath
		assertCorrectString(t, gotPath, wantPath)

		// N.B. Need to sort out time zones
		gotCreateTime := file.createTime.String()
		wantCreateTime := testOldDateParsed
		assertCorrectString(t, gotCreateTime, wantCreateTime)

		gotCreateTimeUnix := strconv.FormatInt(file.createTime.Unix(), 10)
		wantCreateTimeUnix := strconv.FormatInt(1619407073, 10)
		assertCorrectString(t, gotCreateTimeUnix, wantCreateTimeUnix)

		gotCreateTimeUTC := file.createTime.UTC()
		wantCreateTimeUTC := testOldDateParsedUTC
		assertCorrectString(t, gotCreateTimeUTC.String(), wantCreateTimeUTC)

		gotSize := strconv.FormatInt(file.size, 10)
		wantSize := strconv.FormatInt(1024, 10)
		assertCorrectString(t, gotSize, wantSize)

		gotID := file.id
		wantID := testFileID
		assertCorrectString(t, gotID, wantID)

		gotFanIP := file.fanIP.String()
		wantFanIP := net.ParseIP(testIP).String()
		assertCorrectString(t, gotFanIP, wantFanIP)

		gotDatasetID := file.datasetID
		wantDatasetID := testDatasetID
		assertCorrectString(t, gotDatasetID, wantDatasetID)

		gotFileInfo := file.fileInfo
		wantFileInfo := fileInfo
		assertCorrectString(t, gotFileInfo.Name(), wantFileInfo.Name())
	})

	t.Run("PKT struct test", func(t *testing.T) {
		loc, err := time.LoadLocation(testKarachiTime)
		datestring := testKarachiDate
		if err != nil {
			t.Fatalf(err.Error())
		}
		datetime, _ := time.ParseInLocation(time.UnixDate, datestring, loc)
		file = File{
			stagingPath: testPath,
			createTime:  datetime,
			size:        int64(85512264),
			id:          testFileID}

		gotPath := file.stagingPath
		wantPath := testPath
		assertCorrectString(t, gotPath, wantPath)

		// N.B. Need to sort out time zones
		gotCreateTime := file.createTime.String()
		wantCreateTime := testKarachiDateParsed
		assertCorrectString(t, gotCreateTime, wantCreateTime)

		gotCreateTimeUnix := strconv.FormatInt(file.createTime.Unix(), 10)
		wantCreateTimeUnix := strconv.FormatInt(1675083314, 10)
		assertCorrectString(t, gotCreateTimeUnix, wantCreateTimeUnix)

		gotCreateTimeUTC := file.createTime.UTC()
		wantCreateTimeUTC := testKarachiDateUTC
		assertCorrectString(t, gotCreateTimeUTC.String(), wantCreateTimeUTC)

		gotSize := strconv.FormatInt(file.size, 10)
		wantSize := strconv.FormatInt(85512264, 10)
		assertCorrectString(t, gotSize, wantSize)

		gotID := file.id
		wantID := testFileID
		assertCorrectString(t, gotID, wantID)
	})
}

func assertCorrectString(t testing.TB, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("got '%s' want '%s'", got, want)
	}
}

func setupLogs() (testLogger *logrus.Logger, hook *test.Hook) {
	testLogger, hook = test.NewNullLogger()
	log.SetLogger(testLogger)

	return
}
