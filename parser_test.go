package main

import (
	"os"
	"strconv"
	"testing"
	"testing/fstest"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

const (
	oneline   = "/data2/staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56{gbtmp-FD40CB70A63D11EBAB7FB02628E0E270}|Sun Apr 25 23:17:53 EDT 2021|0|95BA50C0A64211EB8B73B026285E5DA0\n"
	multiline = "/data2/staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56{gbtmp-FD40CB70A63D11EBAB7FB02628E0E270}|Sun Apr 25 23:17:53 EDT 2021|0|95BA50C0A64211EB8B73B026285E5DA0\n" +
		"/data2/staging/03bdd706-00000006-f8836565-60836565-2e095000-ab66ac56{gbtmp-E9DE7470A49311EBAB7FB02628E0E270}|Fri Apr 23 20:29:14 EDT 2021|0|24A80BC0A49411EB9275B026285E5440\n" +
		"/data1/staging/ffbb5588-00000006-a08893b2-608893b2-32645000-ee50a856{gbtmp-113E8140A7AA11EB94CCB02628E0E270}|Tue Apr 27 18:44:04 EDT 2021|0|55670DD0A7AD11EB985CB026285E5410\n"
)

var (
	multilineOut = []string{
		"/data2/staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56{gbtmp-FD40CB70A63D11EBAB7FB02628E0E270}|Sun Apr 25 23:17:53 EDT 2021|0|95BA50C0A64211EB8B73B026285E5DA0",
		"/data2/staging/03bdd706-00000006-f8836565-60836565-2e095000-ab66ac56{gbtmp-E9DE7470A49311EBAB7FB02628E0E270}|Fri Apr 23 20:29:14 EDT 2021|0|24A80BC0A49411EB9275B026285E5440",
		"/data1/staging/ffbb5588-00000006-a08893b2-608893b2-32645000-ee50a856{gbtmp-113E8140A7AA11EB94CCB02628E0E270}|Tue Apr 27 18:44:04 EDT 2021|0|55670DD0A7AD11EB985CB026285E5410"}
)

func TestParseFile(t *testing.T) {
	t.Run("parse file", func(t *testing.T) {
		testLogger, hook := setupLogs(t)
		fs := fstest.MapFS{
			"path/processed_files.out": {
				Data: []byte(oneline)},
		}
		onelineParsed := oneline[0 : len(oneline)-1]
		out := parseFile(fs, "path/processed_files.out", testLogger)

		got := out[0]
		want := onelineParsed
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "Processing: " + out[0]
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("parse multiline-file", func(t *testing.T) {
		testLogger, hook := setupLogs(t)
		fs := fstest.MapFS{
			"path/processed_files.out": {
				Data: []byte(multiline)},
		}
		got := parseFile(fs, "path/processed_files.out", testLogger)
		logs := hook.AllEntries()

		want := multilineOut

		for i := 0; i < len(want); i++ {
			assertCorrectString(t, got[i], want[i])

			gotLogMsg := logs[i].Message
			wantLogMsg := "Processing: " + got[i]
			assertCorrectString(t, gotLogMsg, wantLogMsg)

		}
	})

	t.Run("check it logs fsys error", func(t *testing.T) {
		fakeExit := func(int) {
			panic("os.Exit called")
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		testLogger, hook := setupLogs(t)
		fsys := fstest.MapFS{
			"path/processed_files.out": {
				Data: []byte(multiline)},
		}

		panic := func() {
			parseFile(fsys, "does_not_exist.file", testLogger)

		}

		assert.PanicsWithValue(t, "os.Exit called", panic, "os.Exit was not called")
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "open does_not_exist.file: file does not exist"

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestParseLine(t *testing.T) {
	t.Run("parse line", func(t *testing.T) {
		testLogger, _ := setupLogs(t)
		onelineParsed := oneline[0 : len(oneline)-1]
		workingFile := parseLine(onelineParsed, testLogger)

		got := workingFile.path
		want := "/data2/staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56{gbtmp-FD40CB70A63D11EBAB7FB02628E0E270}"
		assertCorrectString(t, got, want)

		got = strconv.FormatInt(workingFile.createTime.Unix(), 10)
		want = "1619407073"
		assertCorrectString(t, got, want)

		got = strconv.FormatInt(workingFile.size, 10)
		want = strconv.FormatInt(0, 10)
		assertCorrectString(t, got, want)

		got = workingFile.id
		want = "95BA50C0A64211EB8B73B026285E5DA0"
		assertCorrectString(t, got, want)

		//gotLogMsg := hook.LastEntry().Message
		//wantLogMsg := "Processing: " + out[0]
		//assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

}
