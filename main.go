package main

import (
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"regexp"

	log "github.com/JLCodeSource/process_async_ds/logger"

	"flag"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	file      string
	datasetid string
	days      int64
	nondryrun bool
)

// File type is a struct which holds its relevant metadata
type File struct {
	smbName     string
	stagingPath string
	createTime  time.Time
	size        int64
	id          string
	fanIP       net.IP
}

func getSourceFile(filesystem fs.FS, f string, logger *logrus.Logger) fs.FileInfo {
	file, err := fs.Stat(filesystem, f)
	if err != nil {
		logger.Fatal(err.Error())
	}
	logger.Info("SourceFile: " + f)
	return file
}

func getAsyncProcessedFolderID(id string, logger *logrus.Logger) string {
	match, err := regexp.MatchString("^[A-F0-9]{32}$", id)
	if err != nil {
		logger.Fatal(err.Error())
	}
	if !match {
		logger.Fatal("DatasetId: " + id + " not of the form ^[A-F0-9]{32}$")
		return ""
	}

	logger.Info("DatasetId set to " + id)
	return id
}

func getTimeLimit(days int64, logger *logrus.Logger) (limit int64) {

	limit = 0

	if days == 0 {
		logger.Warn("No days time limit set; processing all processed files")
		return
	}
	now := time.Now().Unix()
	limit = now - days*86400
	logger.Info("Days time limit set to " +
		strconv.FormatInt(days, 10) +
		" days ago which is " +
		strconv.FormatInt(limit, 10) +
		" in epoch time")
	return

}

func getNonDryRun(nondryrun bool, logger *logrus.Logger) bool {
	if nondryrun {
		logger.Warn("Setting dryrun to false; executing move")
	} else {
		logger.Info("Setting dryrun to true; skipping exeecute move")
	}

	return nondryrun
}

func init() {

	log.Init()
	log.GetLogger()

	flag.StringVar(&file, "file", "", "source path/file (default '')")
	flag.StringVar(&datasetid, "datasetid", "", "async processed dataset id (default '')")
	flag.Int64Var(&days, "days", 0, "number of days ago (default 0)")
	flag.BoolVar(&nondryrun, "non-dryrun", false, "execute non dry run (default false)")

}

func main() {

	logger := log.GetLogger()

	flag.Parse()

	dir, f := filepath.Split((file))
	fsys := os.DirFS(dir)

	getSourceFile(fsys, f, logger)
	getAsyncProcessedFolderId(datasetid, logger)
	getTimeLimit(days, logger)
	getNonDryRun(nondryrun, logger)

}
