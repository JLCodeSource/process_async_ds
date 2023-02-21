package main

import (
	"os"

	log "github.com/JLCodeSource/process_async_ds/logger"

	"strconv"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

const (
	testDatasetId = "41545AB0788A11ECBD0700155D014E0D"
	testFileId    = "D5B58980A3E311EBBA0AB026285E5610"
)

func TestMainFunc(t *testing.T) {

	t.Run("verify main args work", func(t *testing.T) {
		_, hook := setupLogs(t)

		os.Args = append(os.Args, "-dataset=41545AB0788A11ECBD0700155D014E0D")
		os.Args = append(os.Args, "-days=123")

		main()

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "Setting dryrun to true; skipping exeecute move"

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

}

func TestGetAsyncProcessedFolderId(t *testing.T) {

	t.Run("verify it returns the right dataset id", func(t *testing.T) {
		testLogger, _ := setupLogs(t)
		got := getAsyncProcessedFolderId(testDatasetId, testLogger)
		want := testDatasetId

		assertCorrectString(t, got, want)
	})

	t.Run("verify it logs the right dataset id", func(t *testing.T) {
		testLogger, hook := setupLogs(t)
		_ = getAsyncProcessedFolderId(testDatasetId, testLogger)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "DatasetId set to " + testDatasetId

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})

	/* 	t.Run("verify that the dataset id is of the right format", func(t *testing.T) {
		testLogger, hook := setupLogs(t)
		datasetId := getAsyncProcessedFolderId("123", testLogger)

		got := datasetId
		want := strconv.FormatInt(123, 10)

		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		err := "DatasetId: 123 not of the form ^[A-F0-9]{32}$"
		wantLogMsg := err

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	}) */

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
		file := File{
			path:       "/path/file",
			createTime: datetime,
			size:       int64(1024),
			id:         testFileId}

		gotPath := file.path
		wantPath := "/path/file"
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

		gotId := file.id
		wantId := testFileId
		assertCorrectString(t, gotId, wantId)

	})

	t.Run("PKT struct test", func(t *testing.T) {
		loc, err := time.LoadLocation("Asia/Karachi")
		datestring := "Mon Jan 30 17:55:14 PKT 2023"
		if err != nil {
			t.Fatalf(err.Error())
		}
		datetime, _ := time.ParseInLocation(time.UnixDate, datestring, loc)
		file := File{
			path:       "/path/file",
			createTime: datetime,
			size:       int64(85512264),
			id:         testFileId}

		gotPath := file.path
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

		gotId := file.id
		wantId := testFileId
		assertCorrectString(t, gotId, wantId)

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
