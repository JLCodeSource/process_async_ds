package main

import (
	"fmt"
	"io/fs"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/afero"
	log "github.com/JLCodeSource/process_async_ds/logger"

	"flag"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	sourceLog                   = "sourceFile: %v"
	datasetLog                  = "datasetID: %v"
	datasetRegexLog             = "datasetID: %v not of the form %v"
	compareDatasetIDMatchLog    = "datasetID: %v matches asyncProcessedDataset: %v"
	compareDatasetIDNotMatchLog = "datasetID: %v does not match asyncProcessedDataset: %v"
	timelimitNoDaysLog          = "timelimit: No days set; processing all processed files"
	timelimitDaysLog            = "timelimit: Days time limit set to %v days ago which is %v"
	dryRunTrueLog               = "dryrun: true; skipping exeecute move"
	dryRunFalseLog              = "dryrun: false; executing move"
	complexIPLog                = "net.LookupIP: unexpected; more ips than expected"
	wrapOsLog                   = "%v: %v"
	osHostnameLog               = "os.Hostname"
	osExecutableLog             = "os.Executable"
	wrapLookupIPLog             = "net.LookupIP: %v=%v"

	eMatchAsyncProcessedDSTrueLog  = "env.datasetID:%v matches asyncProcessedDataset: %v"
	eMatchAsyncProcessedDSFalseLog = "env.datasetID:%v does not match asyncProcessedDataset: %v"

	fAddedToListLog = "%v (file.id:%v) added to list with file.stagingPath:%v, file.createTime:%v, file.size:%v, file.fanIP:%v, file.fileInfo:%v"

	regexDatasetMatch = "^[A-F0-9]{32}$"

	sourceFileArgTxt  = "sourcefile"
	sourceFileArgHelp = "source path/file (default '')"
	datasetIDArgTxt   = "datasetid"
	datasetIDArgHelp  = "async processed dataset id (default '')"
	timelimitArgTxt   = "days"
	timelimitArgHelp  = "number of days ago (default 0)"
	dryrunArgTxt      = "dryrun"
	dryrunArgHelp     = "execute as dry run (default true)"
	testrunArgTxt     = "test"
	testrunArgHelp    = "execute with test fs (default false)"
)

var (
	sourceFile string
	datasetID  string
	numDays    int64
	dryrun     bool
	testrun    bool
	afs        afero.Fs
	e          *env
	fileList   *[]File
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
	hash        [32]byte
}

/*
type E interface {
	GetEnv() *Env
}
*/

/*
// files type is a slice of Files
type files struct {
	files []File
}*/

// env type holds config and environment settings
type env struct {
	fsys       fs.FS
	afs        afero.Fs
	sourceFile string
	datasetID  string
	limit      time.Time
	dryrun     bool
	sysIP      net.IP
	//pwd        string
	//days       int64

}

// AsyncProcessor is the async processing instance
type AsyncProcessor struct {
	Logger *logrus.Logger
	Env    *env
	Files  *[]File
}

// NewAsyncProcessor returns a pointer to an AsyncProcessor
func NewAsyncProcessor(Logger *logrus.Logger, Env *env, Files *[]File) *AsyncProcessor {
	return &AsyncProcessor{
		Logger: Logger,
		Env:    Env,
		Files:  Files,
	}
}

// verify env
func (e *env) verifyDataset(logger *logrus.Logger) bool {
	ds := getAsyncProcessedDSID(logger)
	if e.datasetID != ds {
		logger.Fatal(fmt.Sprintf(eMatchAsyncProcessedDSFalseLog, e.datasetID, ds))
		return false
	}

	logger.Info(fmt.Sprintf(eMatchAsyncProcessedDSTrueLog, e.datasetID, ds))

	return true
}

func (ap *AsyncProcessor) getSourceFile(ex string, f string) fs.FileInfo {
	var pth string
	filesystem := ap.Env.fsys

	dir, fn := path.Split(f)

	if strings.HasPrefix(f, string(os.PathSeparator)) {
		pth = f[1:]
	} else if dir == "./" || dir == "" {
		dir, _ = path.Split(ex)
		pth = dir + fn
		pth = pth[1:]
	} else {
		pth = f
	}

	file, err := fs.Stat(filesystem, pth)

	if err != nil {
		ap.Logger.Fatal(err.Error())
	}

	ap.Logger.Info(fmt.Sprintf(sourceLog, f))

	return file
}

func getFileList(fsys afero.Fs, sourcefile string, logger *logrus.Logger) []File {
	var files = []File{}

	_, err := fsys.Stat(sourcefile)
	if err != nil {
		logger.Fatal(err)
	}

	lines := parseFile(fsys, sourcefile, logger)

	for _, line := range lines {
		newFile := parseLine(line, logger)
		newFile.fileInfo, err = fsys.Stat(newFile.stagingPath)

		if err != nil {
			// Need to add testing
			logger.Error(err)
			continue
		}

		files = append(files, newFile)
		logger.Info(fmt.Sprintf(fAddedToListLog,
			newFile.smbName,
			newFile.id,
			newFile.stagingPath,
			newFile.createTime.Unix(),
			newFile.size,
			newFile.fanIP,
			newFile.fileInfo.Name()))
	}

	return files
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

	ok, _ := compareDatasetID(id, logger)
	if !ok {
		return ""
	}

	logger.Info(fmt.Sprintf(datasetLog, id))

	return id
}

func compareDatasetID(datasetID string, logger *logrus.Logger) (bool, string) {
	asyncProcessedDS := getAsyncProcessedDSID(logger)
	if asyncProcessedDS != datasetID {
		logger.Fatal(fmt.Sprintf(compareDatasetIDNotMatchLog, datasetID, asyncProcessedDS))
		return false, ""
	}

	logger.Info(fmt.Sprintf(compareDatasetIDMatchLog, datasetID, asyncProcessedDS))

	return true, asyncProcessedDS
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

func getDryRun(dryrun bool, logger *logrus.Logger) bool {
	if dryrun {
		e.afs = afero.NewReadOnlyFs(afero.NewOsFs())

		logger.Info(dryRunTrueLog)
	} else {
		e.afs = afero.NewOsFs()
		logger.Warn(dryRunFalseLog)
	}

	return dryrun
}

func setPWD(ex string, logger *logrus.Logger) string {
	// job needs to run in root dir
	exPath := filepath.Dir(ex)

	parts := strings.Split(exPath, string(os.PathSeparator))
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

func getEnv() *env {
	return e
}

/*
type GetAfs interface {
	getAfs()
}
*/

func getAfs(afs afero.Fs) afero.Fs {
	return afs
}

func init() {
	log.Init()
	log.GetLogger()

	flag.StringVar(&sourceFile, sourceFileArgTxt, "", sourceFileArgHelp)
	flag.StringVar(&datasetID, datasetIDArgTxt, "", datasetIDArgHelp)
	flag.Int64Var(&numDays, timelimitArgTxt, 0, timelimitArgHelp)
	flag.BoolVar(&dryrun, dryrunArgTxt, true, dryrunArgHelp)
	flag.BoolVar(&testrun, testrunArgTxt, false, testrunArgHelp)
}

func main() {
	// Set logger
	logger := log.GetLogger()

	// Parse flags
	flag.Parse()

	// Get pointer to new Env
	e = new(env)

	// Get executable path
	ex := wrapOs(logger, osExecutableLog, os.Executable)

	// Set PWD to root
	root := setPWD(ex, logger)

	e.fsys = os.DirFS(root)
	e.afs = afero.NewOsFs()

	files := []File{}

	ap := NewAsyncProcessor(logger, e, &files)

	ap.getSourceFile(ex, sourceFile)
	ds := getDatasetID(datasetID, logger)
	l := getTimeLimit(numDays, logger)
	ndr := getDryRun(dryrun, logger)

	if testrun {
		e.afs = getAfs(nil)
	}

	hostname := wrapOs(logger, osHostnameLog, os.Hostname)

	ip := wrapLookupIP(logger, hostname, net.LookupIP)

	e = &env{
		fsys:       e.fsys,
		afs:        afs,
		sourceFile: sourceFile,
		datasetID:  ds,
		limit:      l,
		dryrun:     ndr,
		sysIP:      ip,
	}

	e.verifyDataset(logger)
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
