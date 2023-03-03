package main

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

const (
	testGbrPoolOut = ("====== + Pools  in datalake 'nmr' ======\n\n" +
		"- pool01 ( disk pool, primary )\n" +
		" == General ==\n" +
		"   - description:\n" +
		"   - creation date:             Tue Jan 18 23:12:53 EST 2022\n" +
		"   - primary:                   true\n" +
		"   - ID:                        41545AB0788A11ECBD0700155D014E0D\n" +
		"   - parent datalake:           nmr (ID: 0E544860788911ECBD0700155D014E0D)\n")
	testGbrPoolOutLog = ("====== + Pools in datalake 'nmr' ======;;- pool01 ( disk pool, primary ); == General ==; - description:;" +
		" - creation date: Tue Jan 18 23:12:53 EST 2022; - primary: true; - ID: 41545AB0788A11ECBD0700155D014E0D;" +
		" - parent datalake: nmr (ID: 0E544860788911ECBD0700155D014E0D);")
	testGbrFileIDOut       = "1 - 05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56 (file id: D5B58980A3E311EBBA0AB026285E5610)"
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
	testNotAFileID = "not_a_file_id"
)

// Getters

func TestGetAsyncProcessedDSID(t *testing.T) {
	t.Run("should return asyncprocessed dataset", func(t *testing.T) {
		testLogger, hook = setupLogs(t)

		got := getAsyncProcessedDSID(testLogger)
		want := testGbrPoolOutLog
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(gbrGetAsyncProcessedDSLog, testGbrPoolOutLog)
		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})

	/*
		Monkey patchinig exec.Command probably doesn't work
		as it's outside of golang control.
		Need to mock exec.Commamd
		t.Run("should panic on error", func(t *testing.T) {
			fakeExit := func(string, ...string) *exec.Cmd {
				panic(osPanicTrue)
			}
			patch := monkey.Patch(exec.Command, fakeExit)
			defer patch.Unpatch()

			testLogger, hook = setupLogs(t)
			panic := func() { getAsyncProcessedDSID(testLogger) }

			assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)
			gotLogMsg := hook.LastEntry().Message
			wantLogMsg := gbrAsyncProcessedDSErrLog

			assertCorrectString(t, gotLogMsg, wantLogMsg)
		})*/
}

func TestGetFilenameByID(t *testing.T) {
	t.Run("should return the filename by id if it exists", func(t *testing.T) {
		testLogger, hook = setupLogs(t)
		got, ok := getFileNameByID(testFileID, testLogger)
		want := testSmbName
		assert.True(t, ok)
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(gbrFileNameByIDLog, testFileID, testSmbName)
		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})

	t.Run("should return empty if no file exists", func(t *testing.T) {
		testLogger, hook = setupLogs(t)
		got, ok := getFileNameByID(testBadFileID, testLogger)
		want := ""
		assert.False(t, ok)
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(gbrNoFileNameByIDLog, testBadFileID)
		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})

}

func TestGetFileDatasetByID(t *testing.T) {
	t.Run("should return the dataset by id if it exists", func(t *testing.T) {
		testLogger, hook = setupLogs(t)
		got, ok := getFileDatasetByID(testFileID, testLogger)
		want := testDatasetID
		assert.True(t, ok)
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(gbrFileDatasetByIDLog, testFileID, testDatasetID)

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})

	t.Run("should return empty if no file exists", func(t *testing.T) {
		testLogger, hook = setupLogs(t)
		got, ok := getFileDatasetByID(testBadFileID, testLogger)
		want := ""
		assert.False(t, ok)
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(gbrNoFileNameByIDLog, testBadFileID)
		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
}

// Parsers

func TestParseAsyncProcessedDSID(t *testing.T) {
	t.Run("should parse output and return AsyncProcessedDSID", func(t *testing.T) {
		testLogger, hook = setupLogs(t)
		got := parseAsyncProcessedDSID(testGbrPoolOutLog, testLogger)
		want := testDatasetID
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(gbrParseAsyncProcessedDSLog, testDatasetID)
		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})

	t.Run("should parse output and fatal out if no asyncdelDS match", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		testLogger, hook = setupLogs(t)
		panic := func() { parseAsyncProcessedDSID("", testLogger) }

		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := gbrAsyncProcessedDSErrLog
		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
}

func TestParseFileNameByID(t *testing.T) {
	t.Run("should parse output and return filename", func(t *testing.T) {
		testLogger, hook = setupLogs(t)
		got := parseFileNameByID(testGbrFileIDDetailOutLog, testFileID, testLogger)
		want := testSmbName
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(gbrFileNameByIDLog, testFileID, testSmbName)
		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
}

func TestParseFileDatasetByID(t *testing.T) {
	t.Run("should return the dataset by id if it exists", func(t *testing.T) {
		testLogger, hook = setupLogs(t)
		got := parseFileDatasetByID(testGbrFileIDDetailOutLog, testFileID, testLogger)
		want := testDatasetID
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(gbrFileDatasetByIDLog, testFileID, testDatasetID)
		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})

	t.Run("should return '' if the file does not exist", func(t *testing.T) {
		testLogger, hook = setupLogs(t)
		got := parseFileDatasetByID("", testBadFileID, testLogger)
		want := ""
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(gbrNoFileNameByIDLog, testBadFileID)
		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
}

// Errors

func TestAsyncProcessedDSIDErrLog(t *testing.T) {
	t.Run("should log cmd error and fatal", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		testLogger, hook = setupLogs(t)
		panic := func() { asyncProcessedDSIDErrLog(errors.New(osPanicTrue), testLogger) }

		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)
		gotLogMsgs := hook.Entries
		wantLogMsg := osPanicTrue
		assertCorrectString(t, gotLogMsgs[0].Message, wantLogMsg)

		wantLogMsg = gbrAsyncProcessedDSErrLog
		assertCorrectString(t, gotLogMsgs[1].Message, wantLogMsg)

	})
}

func TestGetByIDErrLog(t *testing.T) {
	t.Run("should log err and gbrNoFileNameByID on err", func(t *testing.T) {
		testLogger, hook = setupLogs(t)
		fakeLoggerError := func(args ...interface{}) {
			panic(testGbrFileIDErrOut)
		}
		patch := monkey.Patch(testLogger.Error, fakeLoggerError)
		defer patch.Unpatch()

		getByIDErrLog(testFileID, errors.New(testGbrFileIDErrOut), testLogger)

		gotLogMsgs := hook.Entries
		wantLogMsg := fmt.Sprintf(testGbrFileIDErrOutLog)
		assertCorrectString(t, gotLogMsgs[0].Message, wantLogMsg)

		wantLogMsg = fmt.Sprintf(gbrNoFileNameByIDLog, testFileID)
		assertCorrectString(t, gotLogMsgs[1].Message, wantLogMsg)

	})
}

// Cleaners

func TestCleanGbrOut(t *testing.T) {
	t.Run("should strip \n and dupe white spaces from gbr out", func(t *testing.T) {

		cleaningTests := []struct {
			name string
			got  string
			want string
		}{
			{
				name: "clean gbr pool out",
				got:  testGbrPoolOut,
				want: testGbrPoolOutLog,
			},
			{
				name: "clean gbr file detail out",
				got:  testGbrFileIDDetailOut,
				want: testGbrFileIDDetailOutLog,
			},
			{
				name: "clean gbr file err out",
				got:  testGbrFileIDErrOut,
				want: testGbrFileIDErrOutLog,
			},
		}

		for _, tt := range cleaningTests {
			t.Run(tt.name, func(t *testing.T) {
				got := cleanGbrOut(tt.got)
				want := tt.want
				assertCorrectString(t, got, want)
			})
		}

	})
}
