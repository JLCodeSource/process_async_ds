package main

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

const (
	testGbrPoolOut = ("====== + Pools  in datalake 'nmr' ======" +
		"- pool01 ( disk pool, primary )\n" +
		"  == General ==\n" +
		"	   - description:\n" +
		"	   - creation date:             Tue Jan 18 23:12:53 EST 2022\n" +
		"	   - primary:                   true\n" +
		"	   - ID:                        41545AB0788A11ECBD0700155D014E0D\n" +
		"	   - parent datalake:           nmr (ID: 0E544860788911ECBD0700155D014E0D)")
)

var (
	execCommand = exec.Command
)

// need to add tests for failed lookups and errors...

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprintf(os.Stdout, testGbrPoolOut)
	os.Exit(0)
}

func TestGetAsyncProcessedDSID(t *testing.T) {
	t.Run("should return asyncprocessed dataset", func(t *testing.T) {
		testLogger, hook = setupLogs(t)
		//execCommand = fakeExecCommand
		got, err := getAsyncProcessedDSID(testLogger)
		if err != nil {
			t.Errorf("Expected nil error, got %#v", err)
		}
		want := testDatasetID
		assertCorrectString(t, got, want+"test")

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(gbrAsyncProcessedDSLog, testDatasetID)

		assertCorrectString(t, gotLogMsg, wantLogMsg+"test")

	})
	/*t.Run("should return asyncprocessed dataset with fakeExecCmd", func(t *testing.T) {
		testLogger, hook = setupLogs(t)
		execCommand = fakeExecCommand
		defer func() { execCommand = exec.Command }()
		got, err := getAsyncProcessedDSID(testLogger)
		if err != nil {
			t.Errorf("Expected nil error, got %#v", err)
		}
		want := testGbrPoolOut
		assertCorrectString(t, got, want)

		//gotLogMsg := hook.LastEntry().Message
		//wantLogMsg := fmt.Sprintf(gbrAsyncProcessedDSLog, testDatasetID)

		//assertCorrectString(t, gotLogMsg, wantLogMsg+"test")

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

func TestParseAsyncProcessedDSID(t *testing.T) {
	t.Run("should parse output and return AsyncProcessedDSID", func(t *testing.T) {
		testLogger, hook = setupLogs(t)
		got := parseAsyncProcessedDSID(testGbrPoolOut, testLogger)
		want := testDatasetID
		assertCorrectString(t, got, want+"test")
	})
}

func TestGetFilenameByID(t *testing.T) {
	t.Run("should return the filename by id", func(t *testing.T) {
		testLogger, hook = setupLogs(t)
		got := getFileNameByID(testFileID, testLogger)
		want := testSmbName
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(gbrFileNameByIDLog, testFileID, testSmbName)

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
