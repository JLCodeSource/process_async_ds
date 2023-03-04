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
)

// Getters

func TestGetAsyncProcessedDSID(t *testing.T) {
	t.Run("should return asyncprocessed dataset", func(t *testing.T) {
		testLogger, hook = setupLogs()

		got := getAsyncProcessedDSID(testLogger)
		want := testDatasetID
		assertCorrectString(t, got, want)

		gotLogMsg := hook.Entries[0].Message
		wantLogMsg := fmt.Sprintf(gbrGetAsyncProcessedDSLog, testGbrPoolOutLog)
		assertCorrectString(t, gotLogMsg, wantLogMsg)

		gotLogMsg = hook.Entries[1].Message
		wantLogMsg = fmt.Sprintf(gbrParseAsyncProcessedDSLog, testDatasetID)
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

			testLogger, hook = setupLogs()
			panic := func() { getAsyncProcessedDSID(testLogger) }

			assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)
			gotLogMsg := hook.LastEntry().Message
			wantLogMsg := gbrAsyncProcessedDSErrLog

			assertCorrectString(t, gotLogMsg, wantLogMsg)
		})*/
}

// Parsers

func TestParseAsyncProcessedDSID(t *testing.T) {
	t.Run("should parse output and return AsyncProcessedDSID", func(t *testing.T) {
		testLogger, hook = setupLogs()
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

		testLogger, hook = setupLogs()
		panic := func() { parseAsyncProcessedDSID("", testLogger) }

		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := gbrAsyncProcessedDSErrLog
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

		testLogger, hook = setupLogs()
		panic := func() { asyncProcessedDSIDErrLog(errors.New(osPanicTrue), testLogger) }

		assert.PanicsWithValue(t, osPanicTrue, panic, osPanicFalse)
		gotLogMsgs := hook.Entries
		wantLogMsg := osPanicTrue
		assertCorrectString(t, gotLogMsgs[0].Message, wantLogMsg)

		wantLogMsg = gbrAsyncProcessedDSErrLog
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
