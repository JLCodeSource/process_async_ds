package main

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"regexp"
	"strconv"
	"testing"
	"testing/fstest"
	"time"

	"bou.ke/monkey"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"

	log "github.com/JLCodeSource/process_async_ds/logger"
	"github.com/JLCodeSource/process_async_ds/mockfs"
)

const (
	testDatasetID    = "41545AB0788A11ECBD0700155D014E0D"
	testFileID       = "D5B58980A3E311EBBA0AB026285E5610"
	testName         = "test.txt"
	testPath         = "data1/staging/test.txt"
	testMismatchPath = "data1/staging/testMismatch.txt"
	testNotADataset  = "123"

	testArgsFile    = "-sourcefile=./README.md"
	testArgsDataset = "-datasetid=%v"
	testArgsDays    = "-days=123"
	testArgsHelp    = "-help"

	osPanicTrue  = "os.Exit called"
	osPanicFalse = "os.Exit was not called"

	testEmptyRootErr        = "stat %v: os: DirFS with empty root"
	testOpenDoesNotExistErr = "open %v: file does not exist"
	testRegexMatchErr       = "Regex match errored"

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
	testEnv Env
	limit   time.Time
	ip      net.IP

	// setup file
	file File
	now  time.Time

	// setup fsys
	fsys fstest.MapFS
	mfs  mockfs.MockFS
)

func TestMainFunc(t *testing.T) {

	t.Run("verify main args work", func(t *testing.T) {
		_, hook := setupLogs(t)

		os.Args = append(os.Args, testArgsFile)
		os.Args = append(os.Args, fmt.Sprintf(testArgsDataset, testDatasetID))
		os.Args = append(os.Args, testArgsDays)

		main()

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := dryRunTrueLog

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("verify main help out", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		os.Args = append(os.Args, testArgsHelp)

		panic := func() { main() }
		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)
	})

}

func TestGetSourceFile(t *testing.T) {
	t.Run("check for source file", func(t *testing.T) {
		testLogger, hook := setupLogs(t)
		fs := fstest.MapFS{
			testPath: {Data: []byte(testContent)},
		}
		file := getSourceFile(fs, testPath, testLogger)

		got := file.Name()
		want := testName
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(sourceLog, testPath)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
	t.Run("check for empty root", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		testLogger, hook := setupLogs(t)
		fsys := os.DirFS("")

		panic := func() {
			file := getSourceFile(fsys, testDoesNotExistFile, testLogger)
			println(file)
		}

		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)
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

		testLogger, hook := setupLogs(t)

		fs := fstest.MapFS{
			testMismatchPath: {Data: []byte(testContent)},
		}

		panic := func() {
			file := getSourceFile(fs, testDoesNotExistFile, testLogger)
			println(file)
		}

		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(testOpenDoesNotExistErr, testDoesNotExistFile)

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
		wantLogMsg := fmt.Sprintf(datasetLog, testDatasetID)

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})

	t.Run("verify that it exits if the datasetid is not the right format", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		testLogger, hook := setupLogs(t)
		panic := func() { getAsyncProcessedFolderID(testNotADataset, testLogger) }

		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)
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

		testLogger, hook := setupLogs(t)
		panic := func() { getAsyncProcessedFolderID(testNotADataset, testLogger) }

		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)
		gotLogMsg := hook.LastEntry().Message
		err := testRegexMatchErr
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
		wantLogMsg := timelimitNoDaysLog
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
		wantLogMsg := fmt.Sprintf(timelimitDaysLog, days, gotDays)

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
		wantLogMsg := dryRunTrueLog

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("non-dry run execute move", func(t *testing.T) {
		testLogger, hook := setupLogs(t)

		got := strconv.FormatBool(getNonDryRun(true, testLogger))
		want := strconv.FormatBool(true)

		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := dryRunFalseLog

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
		fsys := fstest.MapFS{
			testPath: {Data: []byte(testContent)},
		}
		fileInfo, _ := fs.Stat(fsys, testPath)
		file := File{
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
		file := File{
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

func setupLogs(t testing.TB) (testLogger *logrus.Logger, hook *test.Hook) {
	testLogger, hook = test.NewNullLogger()
	log.SetLogger(testLogger)
	return
}
