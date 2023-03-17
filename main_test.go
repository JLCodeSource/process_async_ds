package main

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"reflect"
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

	testArgsFile    = "-sourcefile=%v/README.md"
	testArgsDataset = "-datasetid=%v"
	testArgsDays    = "-days=123"
	testArgsHelp    = "-help"

	testPostArgsFile = "%v/README.md"
	testPostArgsDays = int64(123)

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
)

func TestMainFunc(t *testing.T) {

	t.Run("verify main args work", func(t *testing.T) {
		_, hook = setupLogs()
		hostname, _ := os.Hostname()
		ips, _ := net.LookupIP(hostname)
		pwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		os.Args = append(os.Args, fmt.Sprintf(testArgsFile, pwd[1:]))
		os.Args = append(os.Args, fmt.Sprintf(testArgsDataset, testDatasetID))
		os.Args = append(os.Args, testArgsDays)

		now = time.Now()

		//_, file := filepath.Split((testPostArgsFile))
		fsys := os.DirFS("//")

		main()

		// Set for other tests
		testEnv.pwd = pwd
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(wrapLookupIPLog, hostname, ips[0].String())

		assertCorrectString(t, gotLogMsg, wantLogMsg)

		f, err := env.fsys.Open(env.sourceFile)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		got, err := f.Stat()
		if err != nil {
			t.Fatal(err)
		}

		f, err = fsys.Open(env.sourceFile)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		want, err := f.Stat()
		if err != nil {
			t.Fatal(err)
		}
		ok := reflect.DeepEqual(got, want)

		assert.True(t, ok)

		assertCorrectString(t, env.sourceFile, fmt.Sprintf(testPostArgsFile, pwd[1:]))

		assertCorrectString(t, env.datasetID, testDatasetID)

		limit := now.Add(-24 * time.Duration(testPostArgsDays) * time.Hour).Format(time.UnixDate)

		assertCorrectString(t, env.limit.Format(time.UnixDate), limit)

		assert.Equal(t, env.nondryrun, false)
		assert.Equal(t, env.sysIP, ips[0])
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

		os.Args = append(os.Args, testArgsFile)
		os.Args = append(os.Args, fmt.Sprintf(testArgsDataset, testDatasetID))
		os.Args = append(os.Args, testArgsDays)

		panic := func() { main() }
		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)
	})

	//	t.Run("verify lookup IP err", func(t *testing.T) {

	//	}

}

type SpyWrapper struct {
	Calls int
}

func (s *SpyWrapper) wrapOS(*logrus.Logger) {
	s.Calls++
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
		panic := func() { wrapOs(testLogger, osExecutableLog, os.Executable) }
		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := testOsExecutableErr
		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})

	t.Run("wrapOs.Hostname should return & log the path", func(t *testing.T) {
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
		panic := func() { wrapOs(testLogger, osHostnameLog, os.Hostname) }
		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)

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
		panic := func() { wrapLookupIP(testLogger, hostname, net.LookupIP) }
		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)

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
		panic := func() { wrapLookupIP(testLogger, hostname, net.LookupIP) }
		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := complexIPLog
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestGetSourceFile(t *testing.T) {
	t.Run("check for source file", func(t *testing.T) {
		testLogger, hook = setupLogs()
		fsys = fstest.MapFS{
			testPath: {Data: []byte(testContent)},
		}
		file := getSourceFile(fsys, testPath, testLogger)

		got := file.Name()
		want := testName
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(sourceLog, testPath)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("should handle full path", func(t *testing.T) {
		testLogger, hook = setupLogs()
		fsys = fstest.MapFS{
			testPath: {Data: []byte(testContent)},
		}
		fullpath := string(os.PathSeparator) + testPath
		file := getSourceFile(fsys, string(os.PathSeparator)+testPath, testLogger)

		got := file.Name()
		want := testName
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(sourceLog, fullpath)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
	t.Run("check for empty root", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		testLogger, hook = setupLogs()
		fs := os.DirFS("")

		panic := func() {
			getSourceFile(fs, testDoesNotExistFile, testLogger)
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

		testLogger, hook = setupLogs()

		fsys = fstest.MapFS{
			testMismatchPath: {Data: []byte(testContent)},
		}

		panic := func() {
			file := getSourceFile(fsys, testDoesNotExistFile, testLogger)
			println(file)
		}

		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(testOpenDoesNotExistErr, testDoesNotExistFile)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestGetDatasetID(t *testing.T) {

	t.Run("verify it returns the right dataset id", func(t *testing.T) {
		testLogger, _ = setupLogs()
		got := getDatasetID(testDatasetID, testLogger)
		want := testDatasetID

		assertCorrectString(t, got, want)
	})

	t.Run("verify it logs the right dataset id", func(t *testing.T) {
		testLogger, hook = setupLogs()
		_ = getDatasetID(testDatasetID, testLogger)

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

		testLogger, hook = setupLogs()
		panic := func() { getDatasetID(testNotADataset, testLogger) }

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

		testLogger, hook = setupLogs()
		panic := func() { getDatasetID(testNotADataset, testLogger) }

		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)
		gotLogMsg := hook.LastEntry().Message
		err := testRegexMatchErr
		wantLogMsg := err

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestGetTimeLimit(t *testing.T) {

	t.Run("zero days", func(t *testing.T) {
		testLogger, hook = setupLogs()

		var days = int64(0)
		gotDays := getTimeLimit(days, testLogger)
		wantDays := time.Time{}

		assertCorrectString(t, gotDays.String(), wantDays.String())

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := timelimitNoDaysLog
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("Multiple days", func(t *testing.T) {
		testLogger, hook = setupLogs()

		now = time.Now()
		days := int64(15)
		daysInTime := time.Duration(-15 * 24 * time.Hour)
		limit = now.Add(daysInTime)
		gotDays := getTimeLimit(days, testLogger)
		wantDays := limit

		assertCorrectString(t, gotDays.Round(time.Millisecond).String(), wantDays.Round(time.Millisecond).String())

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(timelimitDaysLog, days, gotDays)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestGetNonDryRun(t *testing.T) {

	t.Run("default dry run", func(t *testing.T) {
		testLogger, hook = setupLogs()

		got := strconv.FormatBool(getNonDryRun(false, testLogger))
		want := strconv.FormatBool(false)

		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := dryRunTrueLog

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("non-dry run execute move", func(t *testing.T) {
		testLogger, hook = setupLogs()

		got := strconv.FormatBool(getNonDryRun(true, testLogger))
		want := strconv.FormatBool(true)

		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := dryRunFalseLog

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
}

func TestSetPWD(t *testing.T) {
	t.Run("setPWD should shift execution to root from current path", func(t *testing.T) {
		testLogger, _ = setupLogs()
		ex, _ := os.Executable()

		got := setPWD(ex, testLogger)
		want := "/"
		assertCorrectString(t, got, want)

	})

	t.Run("setPWD should shift execution to root from any path", func(t *testing.T) {
		testLogger, _ = setupLogs()
		ex := "/workflows/process_async_processed/logger/logger.go"

		got := setPWD(ex, testLogger)
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
		testLogger, hook = setupLogs()

		panic := func() { setPWD(testBadPath, testLogger) }

		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)
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
		testLogger, hook = setupLogs()

		panic := func() { setPWD(testBadPath, testLogger) }

		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := testGetwdErr

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
