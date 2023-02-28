package main

import (
	"net"
	"os"
	"testing"
	"testing/fstest"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
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
			smbName: "file.txt",
			fanIP:   ips[0],
		}
		testLogger, hook = setupLogs(t)
		assert.True(t, file.verifyIP(ips[0], testLogger))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "file.txt ip:" + file.fanIP.String() + " matches comparison ip:" + ips[0].String()

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
	t.Run("returns false if ip is not the same as the current machine", func(t *testing.T) {
		file = File{
			smbName: "file.txt",
			fanIP:   ips[0],
		}
		testLogger, hook = setupLogs(t)
		assert.False(t, file.verifyIP(ip, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := ("file.txt ip:" + file.fanIP.String() + " does not match comparison ip:" +
			ip.String() + "; skipping file")

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
			smbName:    "file.txt",
			createTime: now,
		}
		testLogger, hook = setupLogs(t)
		limit = now.Add(-((hours) * time.Hour))
		assert.True(t, file.verifyTimeLimit(limit, testLogger))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "file.txt createTime:" + file.createTime.String() + " is after timelimit:" + limit.String()

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
	t.Run("returns false if file.createTime is before time limit", func(t *testing.T) {
		file = File{
			smbName:    "file.txt",
			createTime: now,
		}
		limit = now.Add(24 * time.Hour)
		testLogger, hook := setupLogs(t)
		assert.False(t, file.verifyTimeLimit(limit, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := ("file.txt createTime:" + file.createTime.String() + " is before timelimit:" +
			limit.String() + "; skipping file")

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
			smbName:   "file.txt",
			id:        testFileID,
			datasetID: testDatasetID,
		}
		testLogger, hook = setupLogs(t)
		assert.True(t, file.verifyInProcessedDataset(testDatasetID, testLogger))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "file.txt datasetID:" + file.datasetID + " matches asyncProcessedDatasetID:" + testDatasetID

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
	t.Run("returns false if file.datasetID does not match asyncProcessedDatasetID", func(t *testing.T) {
		file = File{
			smbName:   "file.txt",
			id:        testFileID,
			datasetID: testDatasetID,
		}
		wrongDataset := "396862B0791111ECA62400155D014E11"

		testLogger, hook = setupLogs(t)
		assert.False(t, file.verifyInProcessedDataset(wrongDataset, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := ("file.txt datasetID:" + file.datasetID + " does not match asyncProcessedDatasetID:" +
			wrongDataset + "; skipping file")

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
		}
		fsys = fstest.MapFS{
			testPath: {Data: []byte("test")},
		}
		testLogger, hook = setupLogs(t)
		assert.True(t, file.verifyExists(fsys, testLogger))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := testName + " exists at " + file.stagingPath

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
	t.Run("returns false if file does not exist", func(t *testing.T) {
		file = File{
			smbName:     testName,
			stagingPath: "/data1/staging/not_the_real_path/test.txt",
		}

		fsys = fstest.MapFS{
			testPath: {Data: []byte("test")},
		}

		testLogger, hook = setupLogs(t)
		assert.False(t, file.verifyExists(fsys, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := (testName + " does not exist at " + file.stagingPath + "; skipping file")

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
