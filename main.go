package main

import (
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"regexp"

	log "github.com/JLCodeSource/process_async_ds/logger"

	"flag"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	sourceLog          = "sourceFile: %v"
	datasetLog         = "datasetID: %v"
	datasetRegexLog    = "datasetID: %v not of the form %v"
	timelimitNoDaysLog = "timelimit: No days set; processing all processed files"
	timelimitDaysLog   = "timelimit: Days time limit set to %v days ago which is %v"
	dryRunTrueLog      = "dryrun: true; skipping exeecute move"
	dryRunFalseLog     = "dryrun: false; executing move"

	regexDatasetMatch = "^[A-F0-9]{32}$"

	sourceFileArgTxt  = "sourcefile"
	sourceFileArgHelp = "source path/file (default '')"
	datasetIDArgTxt   = "datasetid"
	datasetIDArgHelp  = "async processed dataset id (default '')"
	timelimitArgTxt   = "days"
	timelimitArgHelp  = "number of days ago (default 0)"
	nondryrunArgTxt   = "non-dryrun"
	nondryrunArgHelp  = "execute non dry run (default false)"
)

var (
	sourceFile string
	datasetID  string
	days       int64
	nondryrun  bool
)

// File type is a struct which holds its relevant metadata
type File struct {
	smbName     string
	stagingPath string
	createTime  time.Time
	size        int64
	id          string
	fanIP       net.IP
	datasetID   string
	fileInfo    fs.FileInfo
}

func getSourceFile(filesystem fs.FS, f string, logger *logrus.Logger) fs.FileInfo {
	file, err := fs.Stat(filesystem, f)
	if err != nil {
		logger.Fatal(err.Error())
	}
	logger.Info(fmt.Sprintf(sourceLog, f))
	return file
}

func getAsyncProcessedFolderID(id string, logger *logrus.Logger) string {
	match, err := regexp.MatchString(regexDatasetMatch, id)
	if err != nil {
		logger.Fatal(err.Error())
	}
	if !match {
		logger.Fatal(fmt.Sprintf(datasetRegexLog, id, regexDatasetMatch))
		return ""
	}

	logger.Info(fmt.Sprintf(datasetLog, id))
	return id
}

func getTimeLimit(days int64, logger *logrus.Logger) (limit int64) {

	limit = 0

	if days == 0 {
		logger.Warn(timelimitNoDaysLog)
		return
	}
	now := time.Now().Unix()
	limit = now - days*86400
	logger.Info(fmt.Sprintf(timelimitDaysLog, days, limit))
	return

}

func getNonDryRun(nondryrun bool, logger *logrus.Logger) bool {
	if nondryrun {
		logger.Warn(dryRunFalseLog)
	} else {
		logger.Info(dryRunTrueLog)
	}

	return nondryrun
}

func init() {

	log.Init()
	log.GetLogger()

	flag.StringVar(&sourceFile, sourceFileArgTxt, "", sourceFileArgHelp)
	flag.StringVar(&datasetID, datasetIDArgTxt, "", datasetIDArgHelp)
	flag.Int64Var(&days, timelimitArgTxt, 0, timelimitArgHelp)
	flag.BoolVar(&nondryrun, nondryrunArgTxt, false, nondryrunArgHelp)

}

func main() {

	logger := log.GetLogger()

	flag.Parse()

	dir, f := filepath.Split((sourceFile))
	fsys := os.DirFS(dir)

	getSourceFile(fsys, f, logger)
	getAsyncProcessedFolderID(datasetID, logger)
	getTimeLimit(days, logger)
	getNonDryRun(nondryrun, logger)

}
