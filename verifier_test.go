package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

const (
	testGbrFileIDDetailOut = ("1 - 05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56 (file id: D5B58980A3E311EBBA0AB026285E5610)\n" +
		"    version:            0\n" +
		"    type:               file\n" +
		"    parent folder id:   3E4FF671B44E11ED86FF00155D014E0D\n" +
		"    parent folder name: 6132\n" +
		"    parent id:          41545AB0788A11ECBD0700155D014E0D\n" +
		"    original file name: null\n" +
		"    file URI:           null\n" +
		"    fan URI:            ftp://user@192.168.101.210:2121/download/2023_02/1eeb7769-fdb1-4313-8d76-ec719ad7a44c/" +
		"mnt/nas01/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56\n" +
		"    pool id:            41545AB0788A11ECBD0700155D014E0D\n" +
		"    legal hold:         Enabled=false OwnerID=null MatterID=null Start=Sat Jan 01 05:00:00 EST 1 Release=Sat Jan 01 05:00:00 EST 1\n" +
		"    policies:\n" +
		"      RetentionDisposition(null, enabled='true')(start='Mon Feb 27 19:08:25 EST 2023', end='Sat Jan 27 19:08:25 EST 2029', neverDispose='true')\n" +
		"    file hash:\n")
	testGbrFileIDDetailOutLog = ("1 - 05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56 (file id: D5B58980A3E311EBBA0AB026285E5610);" +
		" version: 0; type: file; parent folder id: 3E4FF671B44E11ED86FF00155D014E0D; parent folder name: 6132;" +
		" parent id: 41545AB0788A11ECBD0700155D014E0D; original file name: null; file URI: null;" +
		" fan URI: ftp://user@192.168.101.210:2121/download/2023_02/1eeb7769-fdb1-4313-8d76-ec719ad7a44c/" +
		"mnt/nas01/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56; pool id: 41545AB0788A11ECBD0700155D014E0D;" +
		" legal hold: Enabled=false OwnerID=null MatterID=null Start=Sat Jan 01 05:00:00 EST 1 Release=Sat Jan 01 05:00:00 EST 1;" +
		" policies:; RetentionDisposition(null, enabled='true')(start='Mon Feb 27 19:08:25 EST 2023', end='Sat Jan 27 19:08:25 EST 2029'," +
		" neverDispose='true'); file hash:;")
	testGbrFileIDErrOut = ("java.lang.NumberFormatException: For input string: \"no\"\n" +
		"	at java.base/java.lang.NumberFormatException.forInputString(NumberFormatException.java:65)\n" +
		"	at java.base/java.lang.Integer.parseInt(Integer.java:652)\n" +
		"	at ttl.nds.java.common.StringUtils.fromHexString(StringUtils.java:139)\n" +
		"	at TTL.Nds.Mb.Objects.Common.Nuid.Nuid.<init>(Nuid.java:35)\n" +
		"	at com.trm.gb.restapi.client.commands.file.FileListCommand.call(FileListCommand.java:84)\n" +
		"	at com.trm.gb.restapi.client.commands.file.FileListCommand.call(FileListCommand.java:36)\n" +
		"	at picocli.CommandLine.executeUserObject(CommandLine.java:1933)\n" +
		"	at picocli.CommandLine.access$1200(CommandLine.java:145)\n" +
		"	at picocli.CommandLine$RunLast.executeUserObjectOfLastSubcommandWithSameParent(CommandLine.java:2332)\n" +
		"	at picocli.CommandLine$RunLast.handle(CommandLine.java:2326)\n" +
		"	at picocli.CommandLine$RunLast.handle(CommandLine.java:2291)\n" +
		"	at picocli.CommandLine$AbstractParseResultHandler.execute(CommandLine.java:2159)\n" +
		"	at picocli.CommandLine.execute(CommandLine.java:2058)\n" +
		"	at com.trm.gb.restapi.client.commands.GbrcCommand.main(GbrcCommand.java:71))\n")
	testGbrFileIDErrOutLog = ("java.lang.NumberFormatException: For input string: \"no\";" +
		" at java.base/java.lang.NumberFormatException.forInputString(NumberFormatException.java:65);" +
		" at java.base/java.lang.Integer.parseInt(Integer.java:652); at ttl.nds.java.common.StringUtils.fromHexString(StringUtils.java:139);" +
		" at TTL.Nds.Mb.Objects.Common.Nuid.Nuid.<init>(Nuid.java:35);" +
		" at com.trm.gb.restapi.client.commands.file.FileListCommand.call(FileListCommand.java:84);" +
		" at com.trm.gb.restapi.client.commands.file.FileListCommand.call(FileListCommand.java:36);" +
		" at picocli.CommandLine.executeUserObject(CommandLine.java:1933); at picocli.CommandLine.access$1200(CommandLine.java:145);" +
		" at picocli.CommandLine$RunLast.executeUserObjectOfLastSubcommandWithSameParent(CommandLine.java:2332);" +
		" at picocli.CommandLine$RunLast.handle(CommandLine.java:2326); at picocli.CommandLine$RunLast.handle(CommandLine.java:2291);" +
		" at picocli.CommandLine$AbstractParseResultHandler.execute(CommandLine.java:2159); at picocli.CommandLine.execute(CommandLine.java:2058);" +
		" at com.trm.gb.restapi.client.commands.GbrcCommand.main(GbrcCommand.java:71));")
	//testNotAFileID = "not_a_file_id"

	testContent              = "test"
	testLongerContent        = "longer than the test"
	testWrongDataset         = "396862B0791111ECA62400155D014E11"
	testFileIDInWrongDataset = "3E4FF671B44E11ED86FF00155D015E0D"
	testShortPath            = "staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56"

	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	guidBytes   = "0123456789abcdef"
	fileIDBytes = "0123456789ABCDEF"

	gbrList = "%v/gbr.list"
)

var (
	workDirs = []string{"/workspaces/process_async_ds/", "/usr/src/app/"}
)

func TestRootFSMap(t *testing.T) {
	ex, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	exPath := filepath.Dir(ex)
	parts := strings.Split(exPath, "/")
	dots := ""

	for i := 0; i < (len(parts) - 1); i++ {
		dots = dots + "../"
	}
	//fmt.Println(dots)
	err = os.Chdir(dots)
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Println(os.Executable())
	//fmt.Println(os.Getwd())
	fsys := os.DirFS("/")
	_, err = fs.ReadDir(fsys, ".")

	if err != nil {
		t.Fatal(err)
	}
}

// TestVerify encompasses all verification
func TestVerify(t *testing.T) {
	// setup server ip
	hostname, _ := os.Hostname()
	ips, _ := net.LookupIP(hostname)

	// Setup env

	now = time.Now()
	afterNow := now.Add(-10000 * time.Hour)

	fsys = fstest.MapFS{
		testShortPath: {Data: []byte(testContent),
			ModTime: now},
	}

	var files []file

	fsys, files = createFSTest(t, 10)

	e = new(env)
	e = &env{
		fsys:  fsys,
		limit: afterNow,
		sysIP: ips[0],
		//pwd:       testEnv.pwd,
		datasetID: testDatasetID,
	}

	e.logger, hook = setupLogs()
	ap = NewAsyncProcessor(e, files)

	t.Run("Gen verify", func(t *testing.T) {
		for _, f := range files {
			ok := f.verify()
			assert.True(t, ok)

			gotLogMsg := hook.LastEntry().Message
			wantLogMsg := fmt.Sprintf(fVerifiedLog, f.smbName, f.id)
			assertCorrectString(t, gotLogMsg, wantLogMsg)
		}
	})
}

// TestVerifyEnvSettings encompasses TestVerifyIP & TestVerifyTimeLimit
func TestVerifyEnvMatch(t *testing.T) {
	// setup server ip
	hostname, _ := os.Hostname()
	ips, _ := net.LookupIP(hostname)
	// set incorrect ip
	ip = net.ParseIP("192.168.101.1")

	now = time.Now()

	e = new(env)
	files = []file{}
	ap = NewAsyncProcessor(e, files)

	t.Run("returns true if config metadata matches", func(t *testing.T) {
		limit = now.Add(-24 * time.Hour)
		e = ap.getEnv()
		e = &env{
			sysIP: ips[0],
			limit: limit,
		}
		ap.setEnv(e)

		f = file{
			smbName:     testName,
			id:          testFileID,
			createTime:  now,
			stagingPath: testPath,
			fanIP:       ips[0],
		}
		e.logger, hook = setupLogs()
		assert.True(t, f.verifyEnvMatch())
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fEnvMatchLog, f.smbName, f.id, f.stagingPath)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
	t.Run("returns false if ip is not the same as the current machine", func(t *testing.T) {
		e = ap.getEnv()
		e = &env{
			sysIP: ip,
		}
		ap.setEnv(e)
		f = file{
			smbName: testName,
			id:      testFileID,
			fanIP:   ips[0],
		}
		e.logger, hook = setupLogs()
		assert.False(t, f.verifyEnvMatch())

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fIPMatchFalseLog, f.smbName, f.id, f.fanIP, ip)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("returns false if file.createTime is before time limit", func(t *testing.T) {
		f = file{
			smbName:    testName,
			id:         testFileID,
			createTime: now,
			fanIP:      ips[0],
		}
		limit = now.Add(24 * time.Hour)
		e = ap.getEnv()
		e = &env{
			limit: limit,
			sysIP: ips[0],
		}
		ap.setEnv(e)
		e.logger, hook = setupLogs()
		assert.False(t, f.verifyEnvMatch())

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fCreateTimeBeforeTimeLimitLog,
			f.smbName,
			f.id,
			f.createTime.Round(time.Millisecond),
			limit.Round(time.Millisecond))

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestVerifyIP(t *testing.T) {
	// setup server ip
	hostname, _ := os.Hostname()
	ips, _ := net.LookupIP(hostname)
	// set incorrect ip
	testIP := net.ParseIP("192.168.101.1")

	e = new(env)
	files = []file{}
	ap = NewAsyncProcessor(e, files)

	t.Run("returns true if ip is same as the current machine", func(t *testing.T) {
		f = file{
			smbName: testName,
			id:      testFileID,
			fanIP:   ips[0],
		}
		e.logger, hook = setupLogs()
		e.sysIP = ips[0]
		assert.True(t, f.verifyIP())
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fIPMatchTrueLog, f.smbName, f.id, f.fanIP, ips[0])
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
	t.Run("returns false if ip is not the same as the current machine", func(t *testing.T) {
		f = file{
			smbName: testName,
			id:      testFileID,
			fanIP:   ips[0],
		}
		e.logger, hook = setupLogs()
		e.sysIP = testIP
		assert.False(t, f.verifyIP())

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fIPMatchFalseLog, f.smbName, f.id, f.fanIP, testIP)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestVerifyTimeLimit(t *testing.T) {
	days := time.Duration(15)
	hours := time.Duration(days * 24)
	now := time.Now()

	e = new(env)
	files = []file{}
	ap = NewAsyncProcessor(e, files)

	t.Run("returns true if file.createTime is after time limit", func(t *testing.T) {
		f = file{
			smbName:    testName,
			id:         testFileID,
			createTime: now,
		}
		e.logger, hook = setupLogs()
		e.limit = now.Add(-((hours) * time.Hour))
		assert.True(t, f.verifyTimeLimit())
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(
			fCreateTimeAfterTimeLimitLog,
			f.smbName,
			f.id,
			f.createTime.Round(time.Millisecond),
			e.limit.Round(time.Millisecond),
		)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
	t.Run("returns false if file.createTime is before time limit", func(t *testing.T) {
		f = file{
			smbName:    testName,
			id:         testFileID,
			createTime: now,
		}
		e.limit = now.Add(24 * time.Hour)
		e.logger, hook = setupLogs()
		assert.False(t, f.verifyTimeLimit())

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fCreateTimeBeforeTimeLimitLog,
			f.smbName,
			f.id,
			f.createTime.Round(time.Millisecond),
			e.limit.Round(time.Millisecond))

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

// TestVerifyGBMetadata encompasses verifyInDataset, getMBFileName/DSByFileID
func TestVerifyGBMetadata(t *testing.T) {
	ex, _ := os.Executable()
	dir, _ := path.Split(ex)
	list := fmt.Sprintf(gbrList, dir)

	if err := os.Truncate(list, 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	out, err := os.OpenFile(list, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	e = new(env)
	files = []file{}
	ap = NewAsyncProcessor(e, files)

	t.Run("returns true if file.smbName matches filename", func(t *testing.T) {
		_, files := createFSTest(t, 1)
		e = new(env)
		e.datasetID = testDatasetID
		ap.setEnv(e)

		e.logger, hook = setupLogs()
		assert.True(t, files[0].verifyGBMetadata())
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(
			fSmbNameMatchFileIDNameTrueLog,
			files[0].smbName,
			files[0].id,
			files[0].smbName,
			files[0].smbName)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
	t.Run("returns false if file.datasetID does not match DatasetID", func(t *testing.T) {
		f = file{
			smbName:   testName,
			id:        testFileID,
			datasetID: testWrongDataset,
		}

		e.logger, hook = setupLogs()
		assert.False(t, f.verifyGBMetadata())

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fDatasetMatchFalseLog, f.smbName, f.id, f.datasetID, testDatasetID)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("returns false if file.smbName does not match MB filename", func(t *testing.T) {
		f = file{
			smbName:   testName,
			id:        testFileID,
			datasetID: testDatasetID,
		}
		e.logger, hook = setupLogs()
		assert.False(t, f.verifyGBMetadata())

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(
			fSmbNameMatchFileIDNameFalseLog, testName, testFileID, testName, testSmbName)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("returns false if file.datasetID does not match MB dataset", func(t *testing.T) {
		f = file{
			smbName:   testSmbName,
			id:        testFileIDInWrongDataset,
			datasetID: testWrongDataset,
		}

		e.datasetID = testDatasetID
		e.logger, hook = setupLogs()
		assert.False(t, f.verifyGBMetadata())

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(
			fDatasetMatchFalseLog, testSmbName, testFileIDInWrongDataset, testWrongDataset, testDatasetID)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestGetMBFilenameByFileID(t *testing.T) {
	e = new(env)
	files = []file{}
	ap = NewAsyncProcessor(e, files)

	t.Run("should return true if it exists", func(t *testing.T) {
		f = file{
			smbName: testSmbName,
			id:      testFileID,
		}
		e.logger, hook = setupLogs()
		ok := f.verifyMBFileNameByFileID(f.getGBMetadata())
		assert.True(t, ok)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(
			fSmbNameMatchFileIDNameTrueLog, testSmbName, testFileID, testSmbName, testSmbName)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("should return false if MB file has different name", func(t *testing.T) {
		f = file{
			smbName: testName,
			id:      testFileID,
		}
		e.logger, hook = setupLogs()
		ok := f.verifyMBFileNameByFileID(f.getGBMetadata())
		assert.False(t, ok)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(
			fSmbNameMatchFileIDNameFalseLog, testName, testFileID, testName, testSmbName)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("should return false if no MB file exists", func(t *testing.T) {
		f = file{
			smbName: testSmbName,
			id:      testBadFileID,
		}
		e.logger, hook = setupLogs()
		ok := f.verifyMBFileNameByFileID(f.getGBMetadata())
		assert.False(t, ok)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(
			fGbrNoFileNameByFileIDLog, testSmbName, testBadFileID, testBadFileID)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestGetMBDatasetByFileID(t *testing.T) {
	e = new(env)
	files = []file{}
	ap = NewAsyncProcessor(e, files)

	t.Run("should return the dataset by id if it exists", func(t *testing.T) {
		f = file{
			smbName:   testSmbName,
			id:        testFileID,
			datasetID: testDatasetID,
		}

		e.datasetID = testDatasetID
		e.logger, hook = setupLogs()
		ok := f.verifyMBDatasetByFileID(f.getGBMetadata())
		assert.True(t, ok)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fDatasetMatchTrueLog, testSmbName, testFileID, testDatasetID, testDatasetID)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("should return empty if no file exists", func(t *testing.T) {
		f = file{
			smbName: testSmbName,
			id:      testBadFileID,
		}
		e.logger, hook = setupLogs()
		ok := f.verifyMBDatasetByFileID(f.getGBMetadata())
		assert.False(t, ok)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fGbrNoFileNameByFileIDLog, testSmbName, testBadFileID, testBadFileID)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestParseFileNameByID(t *testing.T) {
	e = new(env)
	files = []file{}
	ap = NewAsyncProcessor(e, files)

	t.Run("should parse output and return filename", func(t *testing.T) {
		f = file{
			smbName:   testSmbName,
			id:        testFileID,
			datasetID: testDatasetID,
		}
		e.logger, hook = setupLogs()
		got := f.parseMBFileNameByFileID(testGbrFileIDDetailOutLog)
		want := testSmbName
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(
			fGbrFileNameByFileIDLog, testSmbName, testFileID, testFileID, testSmbName)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestSetFileDatasetByID(t *testing.T) {
	e = new(env)
	files = []file{}
	ap = NewAsyncProcessor(e, files)

	t.Run("should set f.datasetID if it exists", func(t *testing.T) {
		f = file{
			smbName: testSmbName,
			id:      testFileID,
		}
		e.logger, hook = setupLogs()
		f.setMBDatasetByFileID(testGbrFileIDDetailOutLog)
		got := f.datasetID
		want := testDatasetID
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(
			fGbrDatasetByFileIDLog, testSmbName, testFileID, testFileID, testDatasetID)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("should return '' if the file does not exist", func(t *testing.T) {
		f = file{
			smbName: testSmbName,
			id:      testBadFileID,
		}
		e.logger, hook = setupLogs()
		f.setMBDatasetByFileID("")
		got := f.datasetID
		want := ""
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(
			fGbrNoFileNameByFileIDLog, testSmbName, testBadFileID, testBadFileID)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestGetByIDErrLog(t *testing.T) {
	e = new(env)
	files = []file{}
	ap = NewAsyncProcessor(e, files)

	t.Run("should log err and gbrNoFileNameByID on err", func(t *testing.T) {
		f = file{
			smbName: testSmbName,
			id:      testFileID,
		}
		fakeLoggerError := func(args ...interface{}) {
			panic(testGbrFileIDErrOut)
		}
		patch := monkey.Patch(e.logger.Error, fakeLoggerError)
		defer patch.Unpatch()

		e.logger, hook = setupLogs()
		f.getByIDErrLog(errors.New(testGbrFileIDErrOut))

		gotLogMsgs := hook.Entries
		wantLogMsg := testGbrFileIDErrOutLog
		assertCorrectString(t, gotLogMsgs[0].Message, wantLogMsg)

		wantLogMsg = fmt.Sprintf(
			fGbrNoFileNameByFileIDLog, testSmbName, testFileID, testFileID)
		assertCorrectString(t, gotLogMsgs[1].Message, wantLogMsg)
	})
}

func TestVerifyInProcessedDataset(t *testing.T) {
	e = new(env)
	files = []file{}
	ap = NewAsyncProcessor(e, files)

	t.Run("returns true if file.datasetID matches asyncProcessedDatasetID", func(t *testing.T) {
		f = file{
			smbName:   testName,
			id:        testFileID,
			datasetID: testDatasetID,
		}
		e.logger, hook = setupLogs()
		assert.True(t, f.verifyInDataset(testDatasetID))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fDatasetMatchTrueLog, f.smbName, f.id, f.datasetID, testDatasetID)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
	t.Run("returns false if file.datasetID does not match asyncProcessedDatasetID", func(t *testing.T) {
		f = file{
			smbName:   testName,
			id:        testFileID,
			datasetID: testDatasetID,
		}

		e.logger, hook = setupLogs()
		assert.False(t, f.verifyInDataset(testWrongDataset))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fDatasetMatchFalseLog, f.smbName, f.id, f.datasetID, testWrongDataset)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestVerifyStat(t *testing.T) {
	e = new(env)
	files = []file{}
	ap = NewAsyncProcessor(e, files)

	t.Run("returns true if file matches", func(t *testing.T) {
		fsys = fstest.MapFS{
			testPath: {Data: []byte(testContent)},
		}
		info, _ := fsys.Stat(testPath)
		size := int64(4)
		f = file{
			smbName:     testName,
			id:          testFileID,
			stagingPath: testPath,
			size:        size,
			fileInfo:    info,
		}
		e.logger, hook = setupLogs()
		e.fsys = fsys
		assert.True(t, f.verifyStat())
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fStatMatchLog, f.smbName, f.id, f.stagingPath)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("returns false if file does not exist", func(t *testing.T) {
		f = file{
			smbName:     testName,
			stagingPath: testMismatchPath,
			id:          testFileID,
		}

		fsys = fstest.MapFS{
			testPath: {Data: []byte(testContent)},
		}

		e.logger, hook = setupLogs()
		e.fsys = fsys
		assert.False(t, f.verifyStat())

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fExistsFalseLog, f.smbName, f.id, f.stagingPath)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("returns false if file.size does not match comparator", func(t *testing.T) {
		fsys = fstest.MapFS{
			testPath:         {Data: []byte(testContent)},
			testMismatchPath: {Data: []byte(testLongerContent)},
		}
		info, _ := fsys.Stat(testMismatchPath)
		f = file{
			smbName:     testName,
			id:          testFileID,
			stagingPath: testPath,
			size:        4,
			fileInfo:    info,
		}
		e.logger, hook = setupLogs()
		e.fsys = fsys
		assert.False(t, f.verifyStat())

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fSizeMatchFalseLog, f.smbName, f.id, f.size, f.fileInfo.Size())

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("returns false if file.CreateTime does not match comparator", func(t *testing.T) {
		fsys = fstest.MapFS{}
		now := time.Now()
		afterNow := now.Add(5 * time.Second)
		fsys[testName] = &fstest.MapFile{
			ModTime: afterNow,
		}

		fileInfo, err := fs.Stat(fsys, testName)
		if err != nil {
			fmt.Print(err.Error())
		}

		f = file{
			smbName:     testName,
			id:          testFileID,
			stagingPath: testName,
			createTime:  now,
			fileInfo:    fileInfo,
		}

		e.logger, hook = setupLogs()
		e.fsys = fsys
		assert.False(t, f.verifyStat())

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fCreateTimeMatchFalseLog,
			f.smbName,
			f.id,
			f.createTime.Round(time.Millisecond),
			f.fileInfo.ModTime().Round(time.Millisecond))

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestVerifyFileSize(t *testing.T) {
	e = new(env)
	files = []file{}
	ap = NewAsyncProcessor(e, files)

	t.Run("returns true if file.size matches comparator", func(t *testing.T) {
		fsys = fstest.MapFS{
			testPath: {Data: []byte(testContent)},
		}
		info, _ := fsys.Stat(testPath)
		size := int64(4)
		f = file{
			smbName:     testName,
			id:          testFileID,
			stagingPath: testPath,
			size:        size,
			fileInfo:    info,
		}
		e.logger, hook = setupLogs()
		e.fsys = fsys
		assert.True(t, f.verifyFileSize(size))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fSizeMatchTrueLog, f.smbName, f.id, f.size, f.fileInfo.Size())

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
	t.Run("returns false if file.size does not match comparator", func(t *testing.T) {
		fsys = fstest.MapFS{
			testPath:         {Data: []byte(testContent)},
			testMismatchPath: {Data: []byte(testLongerContent)},
		}
		fileInfo, _ := fsys.Stat(testMismatchPath)
		size := int64(4)
		f = file{
			smbName:     testName,
			id:          testFileID,
			stagingPath: testPath,
			size:        size,
			fileInfo:    fileInfo,
		}
		e.logger, hook = setupLogs()
		e.fsys = fsys
		assert.False(t, f.verifyStat())

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fSizeMatchFalseLog, f.smbName, f.id, f.size, f.fileInfo.Size())

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestVerifyFileCreateTime(t *testing.T) {
	e = new(env)
	files = []file{}
	ap = NewAsyncProcessor(e, files)

	t.Run("returns true if file.createTime matches comparator", func(t *testing.T) {
		fsys = fstest.MapFS{}
		now := time.Now()
		fsys[testName] = &fstest.MapFile{
			ModTime: now,
		}

		fileInfo, err := fs.Stat(fsys, testName)
		if err != nil {
			fmt.Print(err.Error())
		}

		f = file{
			smbName:     testName,
			id:          testFileID,
			stagingPath: testName,
			createTime:  now,
			fileInfo:    fileInfo,
		}
		e.logger, hook = setupLogs()
		e.fsys = fsys
		assert.True(t, f.verifyCreateTime(now))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fCreateTimeMatchTrueLog,
			f.smbName,
			f.id,
			f.createTime.Round(time.Millisecond),
			f.fileInfo.ModTime().Round(time.Millisecond))

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
	t.Run("returns false if file.CreateTime does not match comparator", func(t *testing.T) {
		fsys = fstest.MapFS{}
		now := time.Now()
		afterNow := now.Add(5 * time.Second)
		fsys[testName] = &fstest.MapFile{
			ModTime: afterNow,
		}

		fileInfo, err := fs.Stat(fsys, testName)
		if err != nil {
			fmt.Print(err.Error())
		}

		f = file{
			smbName:     testName,
			id:          testFileID,
			stagingPath: testName,
			createTime:  now,
			fileInfo:    fileInfo,
		}
		e.logger, hook = setupLogs()
		e.fsys = fsys
		assert.False(t, f.verifyCreateTime(afterNow))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fCreateTimeMatchFalseLog,
			f.smbName,
			f.id,
			f.createTime.Round(time.Millisecond),
			f.fileInfo.ModTime().Round(time.Millisecond))

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestVerifyFileIDName(t *testing.T) {
	e = new(env)
	files = []file{}
	ap = NewAsyncProcessor(e, files)

	t.Run("returns true if file.smbname matches file.id filename", func(t *testing.T) {
		f = file{
			smbName: testName,
			id:      testFileID,
		}
		e.logger, hook = setupLogs()
		e.fsys = fsys
		assert.True(t, f.verifyFileIDName(testName))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fSmbNameMatchFileIDNameTrueLog, f.smbName, f.id, f.smbName, testName)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
	t.Run("returns false if file.smbname matches file.id filename", func(t *testing.T) {
		f = file{
			smbName: testSmbName,
			id:      testFileID,
		}

		e.logger, hook = setupLogs()
		e.fsys = fsys
		assert.False(t, f.verifyFileIDName(testName))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(fSmbNameMatchFileIDNameFalseLog, f.smbName, f.id, f.smbName, testName)
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func createFSTest(t *testing.T, numFiles int) (fstest.MapFS, []file) {
	// Create gbrList file
	var list string

	// Create err
	var err error

	dir := getWorkDir()

	list = fmt.Sprintf(gbrList, dir)
	if err := os.Truncate(list, 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	out, err := os.OpenFile(list, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer out.Close()

	fsys = fstest.MapFS{}

	var files []file

	var dirs = []string{}

	dirs = append(dirs, "mb/FAN/")
	dirs = append(dirs, "mb/FAN/download/")

	for i := 1; i < 4; i++ {
		dirs = append(dirs, "datav"+strconv.Itoa(i)+"/staging/")
		dirs = append(dirs, "datav"+strconv.Itoa(i)+"/staging/download/")
	}

	for i := 0; i < numFiles; i++ {
		f := file{}
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
		// set hash
		f.hash = sha256.Sum256(data)

		// set id
		f.id = string(genRandom(32, fileIDBytes))
		// set fanIP
		hostname, _ := os.Hostname()
		ips, _ := net.LookupIP(hostname)
		f.fanIP = ips[0]
		// set datasetID
		f.datasetID = testDatasetID

		fsys[f.stagingPath] = &fstest.MapFile{
			Data:    []byte(data),
			ModTime: f.createTime,
		}
		fi, err := fs.Stat(fsys, f.stagingPath)
		f.fileInfo = fi

		if err != nil {
			fmt.Print(err.Error())
		}

		files = append(files, f)

		_, err = out.WriteString(fmt.Sprintf("%v,%v\n", f.id, f.smbName))
		if err != nil {
			t.Fatal(err)
		}
	}

	return fsys, files
}

func genRandom(i int64, s string) (random []byte) {
	random = make([]byte, i)
	for j := range random {
		random[j] = s[rand.Intn(len(s))] //#nosec - random testing code can be insecure
	}

	return
}

func genGUID() (guid string) {
	for i := 0; i < 6; i++ {
		if i == 1 {
			guid = guid + "00000006-"
		} else if i == 5 {
			guid = guid + string(genRandom(6, guidBytes))
		} else {
			guid = guid + string(genRandom(6, guidBytes)) + "-"
		}
	}

	return
}
