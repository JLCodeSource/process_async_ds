package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

const (
	testProcessedFilesOut   = "path/processed_files.out"
	testDoesNotExistFile    = "does_not_exist.file"
	testFsysDoesNotExistErr = "open %v: file does not exist"
	testDateNotIntErr       = "strconv.ParseInt: parsing \"%v\": invalid syntax"
	testSmbName             = "05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56"
	testStagingPath         = "/data2/staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56{gbtmp-FD40CB70A63D11EBAB7FB02628E0E270}"
	testSize                = "0"
	testID                  = "95BA50C0A64211EB8B73B026285E5DA0"
	testIP                  = "192.168.101.210"
	testOldDate             = "Sun Apr 25 23:17:53 EDT 2021"
	testOldDateParsed       = "2021-04-25 23:17:53 -0400 EDT"
	testOldDateParsedUTC    = "2021-04-26 03:17:53 +0000 UTC"
	testTimeParseInLocErr   = "time.ParseInLocation errored"
	testTimeLoadLocError    = "time.Loadloc errored"

	oneline        = "05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56|/data2/staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56{gbtmp-FD40CB70A63D11EBAB7FB02628E0E270}|1619407073|0|95BA50C0A64211EB8B73B026285E5DA0|192.168.101.210|\n"
	onelineOldDate = "05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56|/data2/staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56{gbtmp-FD40CB70A63D11EBAB7FB02628E0E270}|Sun Apr 25 23:17:53 EDT 2021|0|95BA50C0A64211EB8B73B026285E5DA0|192.168.101.210|\n"
	multiline      = "/data2/staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56{gbtmp-FD40CB70A63D11EBAB7FB02628E0E270}|Sun Apr 25 23:17:53 EDT 2021|0|95BA50C0A64211EB8B73B026285E5DA0|\n" +
		"/data2/staging/03bdd706-00000006-f8836565-60836565-2e095000-ab66ac56{gbtmp-E9DE7470A49311EBAB7FB02628E0E270}|Fri Apr 23 20:29:14 EDT 2021|0|24A80BC0A49411EB9275B026285E5440|\n" +
		"/data1/staging/ffbb5588-00000006-a08893b2-608893b2-32645000-ee50a856{gbtmp-113E8140A7AA11EB94CCB02628E0E270}|Tue Apr 27 18:44:04 EDT 2021|0|55670DD0A7AD11EB985CB026285E5410|\n"
)

var (
	testCreateTimeUnix = time.Unix(1619407073, 0)

	onelineParsed = []string{"05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56|/data2/staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56{gbtmp-FD40CB70A63D11EBAB7FB02628E0E270}|1619407073|0|95BA50C0A64211EB8B73B026285E5DA0|192.168.101.210|"}

	multilineParsed = []string{
		"/data2/staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56{gbtmp-FD40CB70A63D11EBAB7FB02628E0E270}|Sun Apr 25 23:17:53 EDT 2021|0|95BA50C0A64211EB8B73B026285E5DA0|",
		"/data2/staging/03bdd706-00000006-f8836565-60836565-2e095000-ab66ac56{gbtmp-E9DE7470A49311EBAB7FB02628E0E270}|Fri Apr 23 20:29:14 EDT 2021|0|24A80BC0A49411EB9275B026285E5440|",
		"/data1/staging/ffbb5588-00000006-a08893b2-608893b2-32645000-ee50a856{gbtmp-113E8140A7AA11EB94CCB02628E0E270}|Tue Apr 27 18:44:04 EDT 2021|0|55670DD0A7AD11EB985CB026285E5410|"}
)

func TestParseFile(t *testing.T) {
	e = new(env)
	files = &[]File{}
	ap = NewAsyncProcessor(e, files)
	t.Run("test parseFile", func(t *testing.T) {
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
				log:     parseFileLog,
			},
			{
				name:    "parse multiline file",
				content: multiline,
				want:    multilineParsed,
				log:     parseFileLog,
			},
		}

		for _, tt := range parsingTests {
			t.Run(tt.name, func(t *testing.T) {
				fs := afero.NewMemMapFs()
				e.afs = &afero.Afero{Fs: fs}
				dir, _ := path.Split(testProcessedFilesOut)
				err := fs.MkdirAll(dir, 0755)
				if err != nil {
					t.Fatal(err)
				}
				err = afero.WriteFile(e.afs, testProcessedFilesOut, []byte(tt.content), 0755)
				if err != nil {
					t.Fatal(err)
				}

				e.sourceFile = testProcessedFilesOut
				e.logger, hook = setupLogs()

				got := ap.parseSourceFile()

				logs := hook.AllEntries()

				for i := 0; i < len(tt.want); i++ {
					assertCorrectString(t, got[i], tt.want[i])

					gotLogMsg := logs[i].Message
					wantLogMsg := fmt.Sprintf(parseFileLog, got[i])
					assertCorrectString(t, gotLogMsg, wantLogMsg)
				}
			})
		}
	})

	t.Run("check it logs fsys error", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		fs := afero.NewMemMapFs()
		afs := &afero.Afero{Fs: fs}

		e.afs = afs
		e.sourceFile = testDoesNotExistFile
		e.logger, hook = setupLogs()

		panicFunc := func() {
			ap.parseSourceFile()
		}

		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := fmt.Sprintf(testFsysDoesNotExistErr, testDoesNotExistFile)

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}

func TestParseLine(t *testing.T) {
	e = new(env)
	files = &[]File{}
	ap = NewAsyncProcessor(e, files)
	t.Run("verify ParseLine", func(t *testing.T) {
		e.logger, hook = setupLogs()
		onelineParsed := oneline
		workingFile := parseLine(onelineParsed, e)

		parsingTests := []struct {
			name string
			got  string
			want string
			log  string
		}{
			{
				name: "verify smbName",
				got:  workingFile.smbName,
				want: testSmbName,
				log:  smbNameLog,
			},
			{
				name: "verify path",
				got:  workingFile.stagingPath,
				want: testStagingPath,
				log:  stagingPathLog,
			},
			{
				name: "verify createTime",
				got:  workingFile.createTime.UTC().String(),
				want: testCreateTimeUnix.UTC().String(),
				log:  createTimeLog,
			},
			{
				name: "verify size",
				got:  strconv.FormatInt(workingFile.size, 10),
				want: testSize,
				log:  sizeLog,
			},
			{
				name: "verify id",
				got:  workingFile.id,
				want: testID,
				log:  idLog,
			},
			{
				name: "verify fanIp",
				got:  workingFile.fanIP.String(),
				want: testIP,
				log:  fanIPLog,
			},
		}
		for i, tt := range parsingTests {
			t.Run(tt.name, func(t *testing.T) {
				testParseFileLog := fmt.Sprintf(parseFileLog, testID)
				gotLogMsg := hook.Entries[i].Message
				wantLogMsg := fmt.Sprintf(tt.log, testParseFileLog, tt.got)
				assert.Equal(t, tt.got, tt.want)
				assertCorrectString(t, gotLogMsg, wantLogMsg)
			})
		}
	})

	t.Run("it should warn if strconv.ParseInt on dateTime fails", func(t *testing.T) {
		strconvParseIntErr := fmt.Sprintf(testDateNotIntErr, testOldDate)

		e.logger, hook = setupLogs()
		parseLine(onelineOldDate, e)

		gotLogMsg := hook.Entries[2].Message
		err := strconvParseIntErr
		wantLogMsg := err

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("it should panic if time.LoadLocation fails", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		fakeLoadLoc := func(string) (*time.Location, error) {
			err := errors.New(testTimeLoadLocError)
			return nil, err
		}
		patch2 := monkey.Patch(time.LoadLocation, fakeLoadLoc)
		defer patch2.Unpatch()

		e.logger, hook = setupLogs()
		panicFunc := func() { parseLine(onelineOldDate, e) }

		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)

		gotLogMsg := hook.LastEntry().Message
		err := testTimeLoadLocError
		wantLogMsg := err

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

	t.Run("it should panic if time.ParseInLocation fails", func(t *testing.T) {
		fakeExit := func(int) {
			panic(osPanicTrue)
		}
		patch := monkey.Patch(os.Exit, fakeExit)
		defer patch.Unpatch()

		fakeParseInLoc := func(string, string, *time.Location) (time.Time, error) {
			err := errors.New(testTimeParseInLocErr)
			return time.Time{}, err
		}
		patch2 := monkey.Patch(time.ParseInLocation, fakeParseInLoc)
		defer patch2.Unpatch()

		e.logger, hook = setupLogs()
		panicFunc := func() { parseLine(onelineOldDate, e) }

		assert.PanicsWithValue(t, osPanicTrue, panicFunc, osPanicFalse)

		gotLogMsg := hook.LastEntry().Message
		err := testTimeParseInLocErr
		wantLogMsg := err

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}
