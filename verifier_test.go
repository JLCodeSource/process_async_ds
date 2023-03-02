package main

import (
	"fmt"
	"net"
	"os"
	"testing"
	"testing/fstest"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

const (
	testContent       = "test"
	testLongerContent = "longer than the test"
	testWrongDataset  = "396862B0791111ECA62400155D014E11"
	testShortPath     = "staging"
)

func TestVerifyFileIP(t *testing.T) {
	// setup logger
	var testLogger *logrus.Logger
	var hook *test.Hook

	// setup file
	var file File

	// setup server ip
	hostname, _ := os.Hostname()
	ips, _ := net.LookupIP(hostname)
	// set incorrect ip
	ip := net.ParseIP("192.168.101.1")

	t.Run("returns true if ip is same as the current machine", func(t *testing.T) {
		file = File{
			smbName: testName,
			id:      testFileID,
			fanIP:   ips[0],
		}
		testLogger, hook = setupLogs(t)
		assert.True(t, file.verifyIP(ips[0], testLogger))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fIPMatchTrueLog, file.smbName, file.id, file.fanIP, ips[0])
		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
	t.Run("returns false if ip is not the same as the current machine", func(t *testing.T) {
		file = File{
			smbName: testName,
			id:      testFileID,
			fanIP:   ips[0],
		}
		testLogger, hook = setupLogs(t)
		assert.False(t, file.verifyIP(ip, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fIPMatchFalseLog, file.smbName, file.id, file.fanIP, ip)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

}

func TestVerifyTimeLimit(t *testing.T) {
	// setup logger
	var testLogger *logrus.Logger
	var hook *test.Hook

	// setup file
	var file File

	// setup limit
	var limit time.Time

	days := time.Duration(15)
	hours := time.Duration(days * 24)
	now := time.Now()

	t.Run("returns true if file.createTime is after time limit", func(t *testing.T) {
		file = File{
			smbName:    testName,
			id:         testFileID,
			createTime: now,
		}
		testLogger, hook = setupLogs(t)
		limit = now.Add(-((hours) * time.Hour))
		assert.True(t, file.verifyTimeLimit(limit, testLogger))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(
			fCreateTimeAfterTimeLimitLog,
			file.smbName,
			file.id,
			file.createTime.Round(time.Millisecond),
			limit.Round(time.Millisecond),
		)

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
	t.Run("returns false if file.createTime is before time limit", func(t *testing.T) {
		file = File{
			smbName:    testName,
			id:         testFileID,
			createTime: now,
		}
		limit = now.Add(24 * time.Hour)
		testLogger, hook := setupLogs(t)
		assert.False(t, file.verifyTimeLimit(limit, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fCreateTimeBeforeTimeLimitLog,
			file.smbName,
			file.id,
			file.createTime.Round(time.Millisecond),
			limit.Round(time.Millisecond))

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

}

func TestVerifyInProcessedDataset(t *testing.T) {
	// setup logger
	var testLogger *logrus.Logger
	var hook *test.Hook

	// setup file
	var file File

	t.Run("returns true if file.datasetID matches asyncProcessedDatasetID", func(t *testing.T) {
		file = File{
			smbName:   testName,
			id:        testFileID,
			datasetID: testDatasetID,
		}
		testLogger, hook = setupLogs(t)
		assert.True(t, file.verifyInProcessedDataset(testDatasetID, testLogger))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fDatasetMatchTrueLog, file.smbName, file.id, file.datasetID, testDatasetID)
		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
	t.Run("returns false if file.datasetID does not match asyncProcessedDatasetID", func(t *testing.T) {
		file = File{
			smbName:   testName,
			id:        testFileID,
			datasetID: testDatasetID,
		}

		testLogger, hook = setupLogs(t)
		assert.False(t, file.verifyInProcessedDataset(testWrongDataset, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fDatasetMatchFalseLog, file.smbName, file.id, file.datasetID, testWrongDataset)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

}

func TestVerifyFileExists(t *testing.T) {
	// setup logger
	var testLogger *logrus.Logger
	var hook *test.Hook

	// setup file
	var file File

	// setup fs
	var fsys fstest.MapFS

	t.Run("returns true if file exists", func(t *testing.T) {
		file = File{
			smbName:     testName,
			stagingPath: testPath,
			id:          testFileID,
		}
		fsys = fstest.MapFS{
			testPath: {Data: []byte(testContent)},
		}
		testLogger, hook = setupLogs(t)
		assert.True(t, file.verifyExists(fsys, testLogger))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fExistsTrueLog, file.smbName, file.id, file.stagingPath)

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
	t.Run("returns false if file does not exist", func(t *testing.T) {
		file = File{
			smbName:     testName,
			stagingPath: testMismatchPath,
			id:          testFileID,
		}

		fsys = fstest.MapFS{
			testPath: {Data: []byte(testContent)},
		}

		testLogger, hook = setupLogs(t)
		assert.False(t, file.verifyExists(fsys, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fExistsFalseLog, file.smbName, file.id, file.stagingPath)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

}

func TestVerifyFileSize(t *testing.T) {
	// setup logger
	var testLogger *logrus.Logger
	var hook *test.Hook

	// setup file
	var file File

	// setup fs
	var fsys fstest.MapFS

	t.Run("returns true if file.size matches comparator", func(t *testing.T) {
		fsys = fstest.MapFS{
			testPath: {Data: []byte(testContent)},
		}
		info, _ := fsys.Stat(testPath)
		file = File{
			smbName:     testName,
			id:          testFileID,
			stagingPath: testPath,
			size:        4,
			fileInfo:    info,
		}
		testLogger, hook = setupLogs(t)
		assert.True(t, file.verifyFileSize(fsys, testLogger))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fSizeMatchTrueLog, file.smbName, file.id, file.size, file.fileInfo.Size())

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})

	t.Run("returns false if file.size does not match comparator", func(t *testing.T) {
		fsys = fstest.MapFS{
			testPath:         {Data: []byte(testContent)},
			testMismatchPath: {Data: []byte(testLongerContent)},
		}
		info, _ := fsys.Stat(testMismatchPath)
		file = File{
			smbName:     testName,
			id:          testFileID,
			stagingPath: testPath,
			size:        4,
			fileInfo:    info,
		}

		testLogger, hook = setupLogs(t)
		assert.False(t, file.verifyFileSize(fsys, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fSizeMatchFalseLog, file.smbName, file.id, file.size, file.fileInfo.Size())

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("it should error on fs.Stat issue", func(t *testing.T) {
		fsys = fstest.MapFS{
			testPath: {Data: []byte(testContent)},
		}
		file = File{
			smbName:     testName,
			id:          testFileID,
			stagingPath: testMismatchPath,
		}

		testLogger, hook = setupLogs(t)
		assert.False(t, file.verifyFileSize(fsys, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(testFsysDoesNotExistErr, file.stagingPath)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

}

func TestVerifyFileCreateTime(t *testing.T) {
	// setup logger
	var testLogger *logrus.Logger
	var hook *test.Hook

	// setup file
	var file File

	// setup fs
	var fsys fstest.MapFS
	var mfs MockFS

	t.Run("returns true if file.createTime matches comparator", func(t *testing.T) {
		mfs = MockFS{}
		now := time.Now()
		mf := MockFile{
			FS:        mfs,
			MFModTime: now,
			name:      testName,
		}
		mfs = MockFS{
			NewFile(mf),
		}

		fileInfo, _ := mf.Stat()
		file = File{
			smbName:     testName,
			id:          testFileID,
			stagingPath: testName,
			createTime:  now,
			fileInfo:    fileInfo,
		}
		testLogger, hook = setupLogs(t)
		assert.True(t, file.verifyCreateTime(mfs, testLogger))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fCreateTimeMatchTrueLog,
			file.smbName,
			file.id,
			file.createTime.Round(time.Millisecond),
			file.fileInfo.ModTime().Round(time.Millisecond))

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})

	t.Run("returns false if file.CreateTime does not match comparator", func(t *testing.T) {
		mfs = MockFS{}
		now := time.Now()
		mf := MockFile{
			FS:        mfs,
			MFModTime: now.Add(5 * time.Second),
			name:      testName,
		}
		mfs = MockFS{
			NewFile(mf),
		}

		fileInfo, _ := mf.Stat()
		file = File{
			smbName:     testName,
			id:          testFileID,
			stagingPath: testName,
			createTime:  now,
			fileInfo:    fileInfo,
		}

		testLogger, hook = setupLogs(t)
		assert.False(t, file.verifyCreateTime(mfs, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fCreateTimeMatchFalseLog,
			file.smbName,
			file.id,
			file.createTime.Round(time.Millisecond),
			file.fileInfo.ModTime().Round(time.Millisecond))

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("it should error on fs.Stat issue", func(t *testing.T) {
		fsys = fstest.MapFS{
			testPath: {Data: []byte(testContent)},
		}
		file = File{
			smbName:     testName,
			id:          testFileID,
			stagingPath: testMismatchPath,
		}

		testLogger, hook = setupLogs(t)
		assert.False(t, file.verifyCreateTime(fsys, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(testFsysDoesNotExistErr, file.stagingPath)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

}

func TestVerifyFileIDName(t *testing.T) {
	// setup logger
	var testLogger *logrus.Logger
	var hook *test.Hook

	// setup file
	var file File

	t.Run("returns true if file.smbname matches file.id filename", func(t *testing.T) {
		file = File{
			smbName: testName,
			id:      testFileID,
		}
		testLogger, hook = setupLogs(t)
		assert.True(t, file.verifyFileIDName(testName, testLogger))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fSmbNameMatchFileIDNameTrueLog, file.smbName, file.id, file.smbName, testName)
		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
	t.Run("returns false if file.smbname matches file.id filename", func(t *testing.T) {
		file = File{
			smbName: testSmbName,
			id:      testFileID,
		}

		testLogger, hook = setupLogs(t)
		assert.False(t, file.verifyFileIDName(testName, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fSmbNameMatchFileIDNameFalseLog, file.smbName, file.id, file.smbName, testName)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

/*
Need to work this out for the future
type function func(File, interface{}, *logrus.Logger) bool

func TestVerify(t *testing.T) {
	// setup logger
	var testLogger *logrus.Logger
	var hook *test.Hook

	// setup file
	var file File

	// setup server ip
	hostname, _ := os.Hostname()
	ips, _ := net.LookupIP(hostname)
	// set incorrect ip
	ip := net.ParseIP("192.168.101.1")

	verifyTests := []struct {
		name     string
		file     File
		verify   bool
		function function
		log      string
	}{
		{
			name: "returns true if ip is same as the current machine",
			file: File{
				smbName: "file.txt",
				fanIP:   ips[0],
			},
			verify:   true,
			function: File.verifyIP(file, ips[0], testLogger),
			log:      "file.txt ip:" + file.fanIP.String() + " matches comparison ip:" + ips[0].String(),
		},
	}

}*/
