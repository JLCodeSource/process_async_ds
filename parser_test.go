package main

import (
	"testing"
	"testing/fstest"
)

func TestParseFile(t *testing.T) {
	t.Run("parse file", func(t *testing.T) {
		testLogger, hook := setupLogs(t)
		fs := fstest.MapFS{
			"path/processed_files.out": {
				Data: []byte("/data2/staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56{gbtmp-FD40CB70A63D11EBAB7FB02628E0E270}|Sun Apr 25 23:17:53 EDT 2021|0|95BA50C0A64211EB8B73B026285E5DA0"),
				//	"/data2/staging/03bdd706-00000006-f8836565-60836565-2e095000-ab66ac56{gbtmp-E9DE7470A49311EBAB7FB02628E0E270}|Fri Apr 23 20:29:14 EDT 2021|0|24A80BC0A49411EB9275B026285E5440" +
				//	"/data1/staging/ffbb5588-00000006-a08893b2-608893b2-32645000-ee50a856{gbtmp-113E8140A7AA11EB94CCB02628E0E270}|Tue Apr 27 18:44:04 EDT 2021|0|55670DD0A7AD11EB985CB026285E5410"),
			},
		}
		out := parseFile(fs, "path/processed_files.out", testLogger)

		got := out
		want := "/data2/staging/05043fe1-00000006-2f8630d0-608630d0-67d25000-ab66ac56{gbtmp-FD40CB70A63D11EBAB7FB02628E0E270}|Sun Apr 25 23:17:53 EDT 2021|0|95BA50C0A64211EB8B73B026285E5DA0"
		assertCorrectString(t, got, want)

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "Processing: " + out
		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})
}
