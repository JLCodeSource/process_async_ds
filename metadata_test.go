package main

import (
	"fmt"
	"testing"
)

// need to add tests for failed lookups and errors...

func TestGetAsyncProcessedDSID(t *testing.T) {
	t.Run("should return asyncprocessed dataset", func(t *testing.T) {
		testLogger, hook = setupLogs(t)
		got := getAsyncProcessedDSID(testLogger)
		want := testDatasetID
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(gbrAsyncProcessedDSLog, testDatasetID)

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
