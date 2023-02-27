package main

import (
	"errors"
	"os"
	"strconv"
	"testing"
	"testing/fstest"
	"time"

	"bou.ke/monkey"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

const (
	oneline        = "05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56|/data2/staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56{gbtmp-FD40CB70A63D11EBAB7FB02628E0E270}|1619407073|0|95BA50C0A64211EB8B73B026285E5DA0|192.168.101.210\n"
	onelineOldDate = "05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56|/data2/staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56{gbtmp-FD40CB70A63D11EBAB7FB02628E0E270}|Sun Apr 25 23:17:53 EDT 2021|0|95BA50C0A64211EB8B73B026285E5DA0|192.168.101.210\n"
	multiline      = "/data2/staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56{gbtmp-FD40CB70A63D11EBAB7FB02628E0E270}|Sun Apr 25 23:17:53 EDT 2021|0|95BA50C0A64211EB8B73B026285E5DA0\n" +
		"/data2/staging/03bdd706-00000006-f8836565-60836565-2e095000-ab66ac56{gbtmp-E9DE7470A49311EBAB7FB02628E0E270}|Fri Apr 23 20:29:14 EDT 2021|0|24A80BC0A49411EB9275B026285E5440\n" +
		"/data1/staging/ffbb5588-00000006-a08893b2-608893b2-32645000-ee50a856{gbtmp-113E8140A7AA11EB94CCB02628E0E270}|Tue Apr 27 18:44:04 EDT 2021|0|55670DD0A7AD11EB985CB026285E5410\n"
)

var (
	onelineParsed = []string{"05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56|/data2/staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56{gbtmp-FD40CB70A63D11EBAB7FB02628E0E270}|1619407073|0|95BA50C0A64211EB8B73B026285E5DA0|192.168.101.210"}

	multilineParsed = []string{
		"/data2/staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56{gbtmp-FD40CB70A63D11EBAB7FB02628E0E270}|Sun Apr 25 23:17:53 EDT 2021|0|95BA50C0A64211EB8B73B026285E5DA0",
		"/data2/staging/03bdd706-00000006-f8836565-60836565-2e095000-ab66ac56{gbtmp-E9DE7470A49311EBAB7FB02628E0E270}|Fri Apr 23 20:29:14 EDT 2021|0|24A80BC0A49411EB9275B026285E5440",
		"/data1/staging/ffbb5588-00000006-a08893b2-608893b2-32645000-ee50a856{gbtmp-113E8140A7AA11EB94CCB02628E0E270}|Tue Apr 27 18:44:04 EDT 2021|0|55670DD0A7AD11EB985CB026285E5410"}
)

func TestParseFile(t *testing.T) {

	t.Run("test parseFile", func(t *testing.T) {
		var testLogger *logrus.Logger
		var hook *test.Hook
		fs := fstest.MapFS{}

		parsingTests := []struct {
			name    string
			content string
			want    []string
			log     string
		}{
			{
				name:    "parsefile",
				content: oneline,
				want:    onelineParsed,
				log:     "Processing: ",
			},
			{
				name:    "parse multiline file",
				content: multiline,
				want:    multilineParsed,
				log:     "Processing: ",
			},
		}

		for _, tt := range parsingTests {
			t.Run(tt.name, func(t *testing.T) {
				testLogger, hook = setupLogs(t)
				fs = fstest.MapFS{
					"path/processed_files.out": {
						Data: []byte(tt.content)},
				}

				got := parseFile(fs, "path/processed_files.out", testLogger)

				logs := hook.AllEntries()

				for i := 0; i < len(tt.want); i++ {
					assertCorrectString(t, got[i], tt.want[i])

					gotLogMsg := logs[i].Message
					wantLogMsg := "Processing: " + got[i]
					assertCorrectString(t, gotLogMsg, wantLogMsg)

				}
			})
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

	t.Run("verify ParseLine", func(t *testing.T) {
		testLogger, hook := setupLogs(t)
		onelineParsed := oneline[0 : len(oneline)-1]
		workingFile := parseLine(onelineParsed, testLogger)

		parsingTests := []struct {
			name string
			got  string
			want string
			log  string
		}{
			{
				name: "verify smbName",
				got:  workingFile.smbName,
				want: "05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56",
				log:  "smbName: ",
			},
			{
				name: "verify path",
				got:  workingFile.stagingPath,
				want: "/data2/staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56{gbtmp-FD40CB70A63D11EBAB7FB02628E0E270}",
				log:  "stagingPath: ",
			},
			{
				"verify createTime",
				strconv.Itoa(int(workingFile.createTime.Unix())),
				"1619407073",
				"createTime: ",
			},
			{
				"verify size",
				strconv.FormatInt(workingFile.size, 10),
				"0",
				"size: ",
			},
			{
				"verify id",
				workingFile.id,
				"95BA50C0A64211EB8B73B026285E5DA0",
				"id: ",
			},
			{
				"verify fanIp",
				workingFile.fanIP.String(),
				"192.168.101.210",
				"fanIP: ",
			},
		}

		for i, tt := range parsingTests {

			t.Run(tt.name, func(t *testing.T) {
				gotLogMsg := hook.Entries[i].Message
				wantLogMsg := tt.log + tt.got
				assertCorrectString(t, tt.got, tt.want)
				assertCorrectString(t, gotLogMsg, wantLogMsg)
			})

		}
	})

	t.Run("it should warn if strconv.ParseInt on dateTime fails", func(t *testing.T) {
		strconvParseIntErr := "strconv.ParseInt: parsing \"Sun Apr 25 23:17:53 EDT 2021\": invalid syntax"

		testLogger, hook := setupLogs(t)
		parseLine(onelineOldDate, testLogger)

		gotLogMsg := hook.Entries[2].Message
		err := strconvParseIntErr
		wantLogMsg := err

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})

	t.Run("it should panic if time.LoadLocation fails", func(t *testing.T) {
		fakeExit := func(int) {
			panic("os.Exit called")
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()
		timeLoadLocError := "time.Loadloc errored"
		fakeLoadLoc := func(string) (*time.Location, error) {
			err := errors.New(timeLoadLocError)
			return nil, err
		}
		patch2 := monkey.Patch(time.LoadLocation, fakeLoadLoc)
		defer patch2.Unpatch()

		testLogger, hook := setupLogs(t)
		panic := func() { parseLine(onelineOldDate, testLogger) }

		assert.PanicsWithValue(t, "os.Exit called", panic, "os.Exit was not called")

		gotLogMsg := hook.LastEntry().Message
		err := timeLoadLocError
		wantLogMsg := err

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})

	t.Run("it should panic if time.ParseInLocation fails", func(t *testing.T) {
		fakeExit := func(int) {
			panic("os.Exit called")
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()
		timeParseInLocErr := "time.ParseInLocation errored"
		fakeParseInLoc := func(string, string, *time.Location) (time.Time, error) {
			err := errors.New(timeParseInLocErr)
			return time.Time{}, err
		}
		patch2 := monkey.Patch(time.ParseInLocation, fakeParseInLoc)
		defer patch2.Unpatch()

		testLogger, hook := setupLogs(t)
		panic := func() { parseLine(onelineOldDate, testLogger) }

		assert.PanicsWithValue(t, "os.Exit called", panic, "os.Exit was not called")

		gotLogMsg := hook.LastEntry().Message
		err := timeParseInLocErr
		wantLogMsg := err

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
}
