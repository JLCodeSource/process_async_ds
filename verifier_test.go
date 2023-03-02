package main

import (
	"fmt"
	"net"
	"os"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/JLCodeSource/process_async_ds/mockfs"
)

const (
	testContent       = "test"
	testLongerContent = "longer than the test"
	testWrongDataset  = "396862B0791111ECA62400155D014E11"
	testShortPath     = "staging"
)

func TestVerifyEnvSettings(t *testing.T) {
	// setup server ip
	hostname, _ := os.Hostname()
	ips, _ := net.LookupIP(hostname)
	// set incorrect ip
	ip = net.ParseIP("192.168.101.1")

	now = time.Now()

	t.Run("returns true if config metadata matches", func(t *testing.T) {
		limit = now.Add(-24 * time.Hour)
		env = Env{
			sysIP: ips[0],
			limit: limit,
		}
		file = File{
			smbName:     testName,
			id:          testFileID,
			createTime:  now,
			stagingPath: testPath,
			fanIP:       ips[0],
		}
		testLogger, hook = setupLogs(t)
		assert.True(t, file.verifyEnv(env, testLogger))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fEnvMatchLog, file.smbName, file.id, file.stagingPath)
		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
	t.Run("returns false if ip is not the same as the current machine", func(t *testing.T) {
		env = Env{
			sysIP: ip,
		}
		file = File{
			smbName: testName,
			id:      testFileID,
			fanIP:   ips[0],
		}
		testLogger, hook = setupLogs(t)
		assert.False(t, file.verifyEnv(env, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fIPMatchFalseLog, file.smbName, file.id, file.fanIP, ip)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("returns false if file.createTime is before time limit", func(t *testing.T) {
		file = File{
			smbName:    testName,
			id:         testFileID,
			createTime: now,
			fanIP:      ips[0],
		}
		limit = now.Add(24 * time.Hour)
		env = Env{
			limit: limit,
			sysIP: ips[0],
		}
		testLogger, hook := setupLogs(t)
		assert.False(t, file.verifyEnv(env, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fCreateTimeBeforeTimeLimitLog,
			file.smbName,
			file.id,
			file.createTime.Round(time.Millisecond),
			limit.Round(time.Millisecond))

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

}

func TestVerifyIP(t *testing.T) {
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

func TestVerifyStat(t *testing.T) {

	t.Run("returns true if file matches", func(t *testing.T) {
		fsys = fstest.MapFS{
			testPath: {Data: []byte(testContent)},
		}
		info, _ := fsys.Stat(testPath)
		size := int64(4)
		file = File{
			smbName:     testName,
			id:          testFileID,
			stagingPath: testPath,
			size:        size,
			fileInfo:    info,
		}
		testLogger, hook = setupLogs(t)
		assert.True(t, file.verifyStat(fsys, testLogger))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fStatMatchLog, file.smbName, file.id, file.stagingPath)

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
		assert.False(t, file.verifyStat(fsys, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fExistsFalseLog, file.smbName, file.id, file.stagingPath)

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
		assert.False(t, file.verifyStat(fsys, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fSizeMatchFalseLog, file.smbName, file.id, file.size, file.fileInfo.Size())

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("returns false if file.CreateTime does not match comparator", func(t *testing.T) {
		mfs = mockfs.MockFS{}
		now := time.Now()
		afterNow := now.Add(5 * time.Second)
		mf := mockfs.MockFile{
			FS:        mfs,
			MFModTime: afterNow,
			MFName:    testName,
		}
		mfs = mockfs.MockFS{
			mockfs.NewFile(mf),
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
		assert.False(t, file.verifyStat(mfs, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fCreateTimeMatchFalseLog,
			file.smbName,
			file.id,
			file.createTime.Round(time.Millisecond),
			file.fileInfo.ModTime().Round(time.Millisecond))

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

}

func TestVerifyFileSize(t *testing.T) {
	t.Run("returns true if file.size matches comparator", func(t *testing.T) {
		fsys = fstest.MapFS{
			testPath: {Data: []byte(testContent)},
		}
		info, _ := fsys.Stat(testPath)
		size := int64(4)
		file = File{
			smbName:     testName,
			id:          testFileID,
			stagingPath: testPath,
			size:        size,
			fileInfo:    info,
		}
		testLogger, hook = setupLogs(t)
		assert.True(t, file.verifyFileSize(size, testLogger))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fSizeMatchTrueLog, file.smbName, file.id, file.size, file.fileInfo.Size())

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})

	t.Run("returns false if file.size does not match comparator", func(t *testing.T) {
		fsys = fstest.MapFS{
			testPath:         {Data: []byte(testContent)},
			testMismatchPath: {Data: []byte(testLongerContent)},
		}
		fileInfo, _ := fsys.Stat(testMismatchPath)
		size := int64(4)
		file = File{
			smbName:     testName,
			id:          testFileID,
			stagingPath: testPath,
			size:        size,
			fileInfo:    fileInfo,
		}

		testLogger, hook = setupLogs(t)
		assert.False(t, file.verifyStat(fsys, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fSizeMatchFalseLog, file.smbName, file.id, file.size, file.fileInfo.Size())

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestVerifyFileCreateTime(t *testing.T) {
	t.Run("returns true if file.createTime matches comparator", func(t *testing.T) {
		mfs = mockfs.MockFS{}
		now := time.Now()
		mf := mockfs.MockFile{
			FS:        mfs,
			MFModTime: now,
			MFName:    testName,
		}
		mfs = mockfs.MockFS{
			mockfs.NewFile(mf),
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
		assert.True(t, file.verifyCreateTime(now, testLogger))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fCreateTimeMatchTrueLog,
			file.smbName,
			file.id,
			file.createTime.Round(time.Millisecond),
			file.fileInfo.ModTime().Round(time.Millisecond))

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})

	t.Run("returns false if file.CreateTime does not match comparator", func(t *testing.T) {
		mfs = mockfs.MockFS{}
		now := time.Now()
		afterNow := now.Add(5 * time.Second)
		mf := mockfs.MockFile{
			FS:        mfs,
			MFModTime: afterNow,
			MFName:    testName,
		}
		mfs = mockfs.MockFS{
			mockfs.NewFile(mf),
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
		assert.False(t, file.verifyCreateTime(afterNow, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fCreateTimeMatchFalseLog,
			file.smbName,
			file.id,
			file.createTime.Round(time.Millisecond),
			file.fileInfo.ModTime().Round(time.Millisecond))

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestVerifyFileIDName(t *testing.T) {
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
