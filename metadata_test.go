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
	testGbrPoolOutLog = "====== + Pools in datalake 'nmr' ======;;- pool01 ( disk pool, primary ); == General ==; - description:; - creation date: Tue Jan 18 23:12:53 EST 2022; - primary: true; - ID: 41545AB0788A11ECBD0700155D014E0D; - parent datalake: nmr (ID: 0E544860788911ECBD0700155D014E0D);"
	testNotAFileID    = "not_a_file_id"
)

// need to add tests for failed lookups and errors...

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
		wantLogMsg := fmt.Sprintf(osPanicTrue)
		assertCorrectString(t, gotLogMsgs[0].Message, wantLogMsg)

		wantLogMsg = fmt.Sprintf(gbrAsyncProcessedDSErrLog)
		assertCorrectString(t, gotLogMsgs[1].Message, wantLogMsg)

	})
}

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
		wantLogMsg := fmt.Sprintf(gbrAsyncProcessedDSErrLog)
		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
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

func TestGetDatasetByID(t *testing.T) {
	t.Run("should return the dataset by id", func(t *testing.T) {
		testLogger, hook = setupLogs(t)
		got := getDatasetByID(testFileID, testLogger)
		want := testDatasetID
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(gbrDatasetByIDLog, testFileID, testDatasetID)

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
}
