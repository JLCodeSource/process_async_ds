package main

import (
	"errors"
	"net"
	"os"
	"regexp"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	log "github.com/JLCodeSource/process_async_ds/logger"

	"strconv"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

const (
	testDatasetID = "41545AB0788A11ECBD0700155D014E0D"
	testFileID    = "D5B58980A3E311EBBA0AB026285E5610"
)

func TestMainFunc(t *testing.T) {

	t.Run("verify main args work", func(t *testing.T) {
		_, hook := setupLogs(t)

		os.Args = append(os.Args, "-file=./README.md")
		os.Args = append(os.Args, "-datasetid="+testDatasetID)
		os.Args = append(os.Args, "-days=123")

		main()

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "Setting dryrun to true; skipping exeecute move"

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("verify main help out", func(t *testing.T) {
		fakeExit := func(int) {
			panic("help output")
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		os.Args = append(os.Args, "-help")

		panic := func() { main() }
		assert.PanicsWithValue(t, "help output", panic, "help output not called")
	})

}

func TestGetSourceFile(t *testing.T) {
	t.Run("check for source file", func(t *testing.T) {
		testLogger, hook := setupLogs(t)
		fs := fstest.MapFS{
			"path/file.txt": {Data: []byte("test")},
		}
		file := getSourceFile(fs, "path/file.txt", testLogger)

		got := file.Name()
		want := "file.txt"
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "SourceFile: path/file.txt"
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
	t.Run("check for empty root", func(t *testing.T) {
		fakeExit := func(int) {
			panic("os.Exit called")
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		testLogger, hook := setupLogs(t)
		fsys := os.DirFS("")

		panic := func() {
			file := getSourceFile(fsys, "does_not_exist.file", testLogger)
			println(file)
		}

		assert.PanicsWithValue(t, "os.Exit called", panic, "os.Exit was not called")
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "stat does_not_exist.file: os: DirFS with empty root"

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
	t.Run("error if file does not exist", func(t *testing.T) {
		fakeExit := func(int) {
			panic("os.Exit called")
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		testLogger, hook := setupLogs(t)

		fs := fstest.MapFS{
			"notapath/file.txt": {Data: []byte("test")},
		}

		panic := func() {
			file := getSourceFile(fs, "doesnotexist.txt", testLogger)
			println(file)
		}

		assert.PanicsWithValue(t, "os.Exit called", panic, "os.Exit was not called")
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "open doesnotexist.txt: file does not exist"

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestGetAsyncProcessedFolderId(t *testing.T) {

	t.Run("verify it returns the right dataset id", func(t *testing.T) {
		testLogger, _ := setupLogs(t)
		got := getAsyncProcessedFolderID(testDatasetID, testLogger)
		want := testDatasetID

		assertCorrectString(t, got, want)
	})

	t.Run("verify it logs the right dataset id", func(t *testing.T) {
		testLogger, hook := setupLogs(t)
		_ = getAsyncProcessedFolderID(testDatasetID, testLogger)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "DatasetId set to " + testDatasetID

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})

	t.Run("verify that it exits if the datasetid is not the right format", func(t *testing.T) {
		fakeExit := func(int) {
			panic("os.Exit called")
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		testLogger, hook := setupLogs(t)
		panic := func() { getAsyncProcessedFolderID("123", testLogger) }

		assert.PanicsWithValue(t, "os.Exit called", panic, "os.Exit was not called")
		gotLogMsg := hook.LastEntry().Message
		err := "DatasetId: 123 not of the form ^[A-F0-9]{32}$"
		wantLogMsg := err

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("verify that it exits if the regex fails", func(t *testing.T) {
		fakeExit := func(int) {
			panic("os.Exit called")
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()
		fakeRegexMatch := func(string, string) (bool, error) {
			err := errors.New("Regex match errored")
			return false, err
		}
		patch2 := monkey.Patch(regexp.MatchString, fakeRegexMatch)
		defer patch2.Unpatch()

		testLogger, hook := setupLogs(t)
		panic := func() { getAsyncProcessedFolderID("not_a_dataset", testLogger) }

		assert.PanicsWithValue(t, "os.Exit called", panic, "os.Exit was not called")
		gotLogMsg := hook.LastEntry().Message
		err := "Regex match errored"
		wantLogMsg := err

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestGetTimeLimit(t *testing.T) {

	t.Run("zero days", func(t *testing.T) {
		testLogger, hook := setupLogs(t)

		var days = int64(0)
		gotDays := strconv.FormatInt(getTimeLimit(days, testLogger), 10)
		wantDays := strconv.FormatInt(int64(0), 10)

		assertCorrectString(t, gotDays, wantDays)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "No days time limit set; processing all processed files"
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("Multiple days", func(t *testing.T) {
		testLogger, hook := setupLogs(t)

		var now = time.Now().Unix()
		days := int64(15)
		limit := now - days*86400
		gotDays := strconv.FormatInt(getTimeLimit(days, testLogger), 10)
		wantDays := strconv.FormatInt(limit, 10)

		assertCorrectString(t, gotDays, wantDays)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "Days time limit set to " +
			strconv.FormatInt(days, 10) +
			" days ago which is " +
			strconv.FormatInt(limit, 10) +
			" in epoch time"

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestGetNonDryRun(t *testing.T) {

	t.Run("default dry run", func(t *testing.T) {
		testLogger, hook := setupLogs(t)

		got := strconv.FormatBool(getNonDryRun(false, testLogger))
		want := strconv.FormatBool(false)

		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "Setting dryrun to true; skipping exeecute move"

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("non-dry run execute move", func(t *testing.T) {
		testLogger, hook := setupLogs(t)

		got := strconv.FormatBool(getNonDryRun(true, testLogger))
		want := strconv.FormatBool(true)

		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "Setting dryrun to false; executing move"

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
}

func TestFileMetadata(t *testing.T) {
	t.Run("Initial struct test", func(t *testing.T) {
		loc, err := time.LoadLocation("America/New_York")
		datestring := "Wed Apr 28 20:41:34 EDT 2021"
		if err != nil {
			t.Fatalf(err.Error())
		}
		datetime, _ := time.ParseInLocation(time.UnixDate, datestring, loc)
		fanIP := net.ParseIP("192.168.101.210")
		file := File{
			smbName:     "file.txt",
			stagingPath: "/path/file.txt",
			createTime:  datetime,
			size:        int64(1024),
			id:          testFileID,
			fanIP:       fanIP,
		}

		gotSmbName := file.smbName
		wantSmbName := "file.txt"
		assertCorrectString(t, gotSmbName, wantSmbName)

		gotPath := file.stagingPath
		wantPath := "/path/file.txt"
		assertCorrectString(t, gotPath, wantPath)

		// N.B. Need to sort out time zones
		gotCreateTime := file.createTime.String()
		wantCreateTime := "2021-04-28 20:41:34 -0400 EDT"
		assertCorrectString(t, gotCreateTime, wantCreateTime)

		gotCreateTimeUnix := strconv.FormatInt(file.createTime.Unix(), 10)
		wantCreateTimeUnix := strconv.FormatInt(1619656894, 10)
		assertCorrectString(t, gotCreateTimeUnix, wantCreateTimeUnix)

		gotCreateTimeUTC := file.createTime.UTC()
		wantCreateTimeUTC := "2021-04-29 00:41:34 +0000 UTC"
		assertCorrectString(t, gotCreateTimeUTC.String(), wantCreateTimeUTC)

		gotSize := strconv.FormatInt(file.size, 10)
		wantSize := strconv.FormatInt(1024, 10)
		assertCorrectString(t, gotSize, wantSize)

		gotID := file.id
		wantID := testFileID
		assertCorrectString(t, gotID, wantID)

		gotFanIP := file.fanIP.String()
		wantFanIP := net.ParseIP("192.168.101.210").String()
		assertCorrectString(t, gotFanIP, wantFanIP)

	})

	t.Run("PKT struct test", func(t *testing.T) {
		loc, err := time.LoadLocation("Asia/Karachi")
		datestring := "Mon Jan 30 17:55:14 PKT 2023"
		if err != nil {
			t.Fatalf(err.Error())
		}
		datetime, _ := time.ParseInLocation(time.UnixDate, datestring, loc)
		file := File{
			stagingPath: "/path/file",
			createTime:  datetime,
			size:        int64(85512264),
			id:          testFileID}

		gotPath := file.stagingPath
		wantPath := "/path/file"
		assertCorrectString(t, gotPath, wantPath)

		// N.B. Need to sort out time zones
		gotCreateTime := file.createTime.String()
		wantCreateTime := "2023-01-30 17:55:14 +0500 PKT"
		assertCorrectString(t, gotCreateTime, wantCreateTime)

		gotCreateTimeUnix := strconv.FormatInt(file.createTime.Unix(), 10)
		wantCreateTimeUnix := strconv.FormatInt(1675083314, 10)
		assertCorrectString(t, gotCreateTimeUnix, wantCreateTimeUnix)

		gotCreateTimeUTC := file.createTime.UTC()
		wantCreateTimeUTC := "2023-01-30 12:55:14 +0000 UTC"
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

func setupLogs(t testing.TB) (testLogger *logrus.Logger, hook *test.Hook) {
	testLogger, hook = test.NewNullLogger()
	log.SetLogger(testLogger)
	return
}
