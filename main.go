package main

import (
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
	complexIPLog       = "net.LookupIP: unexpected; more ips than expected"
	wrapOsLog          = "%v: %v"
	osHostnameLog      = "os.Hostname"
	osExecutableLog    = "os.Executable"
	wrapLookupIPLog    = "net.LookupIP: %v=%v"

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
	numDays    int64
	nondryrun  bool
	env        *Env
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

// Env type holds config and environment settings
type Env struct {
	fsys       fs.FS
	sourceFile string
	datasetID  string
	//days       int64
	limit     time.Time
	nondryrun bool
	sysIP     net.IP
	pwd       string
}

func getSourceFile(filesystem fs.FS, f string, logger *logrus.Logger) fs.FileInfo {
	var path string
	if strings.HasPrefix(f, string(os.PathSeparator)) {
		path = f[1:]
	} else {
		path = f
	}
	file, err := fs.Stat(filesystem, path)
	if err != nil {
		logger.Fatal(err.Error())
	}
	logger.Info(fmt.Sprintf(sourceLog, f))
	return file
}

func getDatasetID(id string, logger *logrus.Logger) string {
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

func getTimeLimit(days int64, logger *logrus.Logger) (limit time.Time) {

	limit = time.Time{}

	if days == 0 {
		logger.Warn(timelimitNoDaysLog)
		return
	}
	now := time.Now()
	limit = now.Add(-24 * time.Duration(days) * time.Hour)
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

func setPWD(ex string, logger *logrus.Logger) string {
	// job needs to run in root dir

	exPath := filepath.Dir(ex)

	parts := strings.Split(exPath, "/")
	dots := ""
	for i := 0; i < (len(parts) - 1); i++ {
		dots = dots + "../"
	}
	err := os.Chdir(dots)
	if err != nil {
		logger.Fatal(err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		logger.Fatal(err)
	}

	return pwd
}

//func getEnv() *Env {
//	return env
//}

func init() {

	log.Init()
	log.GetLogger()

	flag.StringVar(&sourceFile, sourceFileArgTxt, "", sourceFileArgHelp)
	flag.StringVar(&datasetID, datasetIDArgTxt, "", datasetIDArgHelp)
	flag.Int64Var(&numDays, timelimitArgTxt, 0, timelimitArgHelp)
	flag.BoolVar(&nondryrun, nondryrunArgTxt, false, nondryrunArgHelp)
}

func main() {

	logger := log.GetLogger()

	flag.Parse()

	ex := wrapOs(logger, osExecutableLog, os.Executable)

	root := setPWD(ex, logger)

	fsys := os.DirFS(root)

	getSourceFile(fsys, sourceFile, logger)
	ds := getDatasetID(datasetID, logger)
	l := getTimeLimit(numDays, logger)
	ndr := getNonDryRun(nondryrun, logger)

	hostname := wrapOs(logger, osHostnameLog, os.Hostname)

	ip := wrapLookupIP(logger, hostname, net.LookupIP)
	/* ips, err := net.LookupIP(hostname)
	if err != nil {
		logger.Fatal(err)
	}
	if len(ips) > 1 {
		logger.Fatal(complexIPLog)
	} */

	env = new(Env)
	env = &Env{
		fsys:       fsys,
		sourceFile: sourceFile,
		datasetID:  ds,
		limit:      l,
		nondryrun:  ndr,
		sysIP:      ip,
	}

}

func wrapOs(logger *logrus.Logger, wrapped string, f func() (string, error)) string {
	out, err := f()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info(fmt.Sprintf(wrapOsLog, wrapped, out))
	return out
}

func wrapLookupIP(logger *logrus.Logger, hostname string, f func(string) ([]net.IP, error)) net.IP {
	ips, err := f(hostname)
	if err != nil {
		logger.Fatal(err)
	} else if len(ips) > 1 {
		logger.Fatal(complexIPLog)
	}

	ip := ips[0]
	logger.Info(fmt.Sprintf(wrapLookupIPLog, hostname, ip.String()))
	return ip
}

/*
//go:generate mockery --name osWrapper
type Wrapper interface {
	wrapOs(logger *logrus.Logger, f func() (string, error)) string
}
*/
