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
	testRunTrueLog              = "testrun: setting to true"
	testRunFalseLog             = "testrun: setting to false"
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
	ap         AsyncProcessor
	e          *env
	afs        afero.Fs
	files      *[]File

	// testIntegrationTestSetup
	testIntegrationTestSetup AsyncProcessor
)

// env type holds config and environment settings
type env struct {
	logger  *logrus.Logger
	exePath string
	fsys    fs.FS
	afs     afero.Fs
	sysIP   net.IP

	sourceFile string
	datasetID  string
	limit      time.Time
	dryrun     bool
	testrun    bool
}

// AsyncProcessor interface is the interface for AD
type AsyncProcessor interface {
	getEnv() *env
	getFiles() *[]File
	setEnv(*env)
	setFiles()
	processFiles()
	parseSourceFile() []string
	//parseLine(string) file
}

// asyncProcessor is the async processing instance
type asyncProcessor struct {
	Env   *env
	Files *[]File
}

// NewAsyncProcessor returns a pointer to an AsyncProcessor
func NewAsyncProcessor(Env *env, Files *[]File) AsyncProcessor {
	return &asyncProcessor{
		Env:   Env,
		Files: Files,
	}
}

// verify env
func (e *env) verifyDataset() bool {
	ds := getAsyncProcessedDSID(e.logger)
	if e.datasetID != ds {
		e.logger.Fatal(fmt.Sprintf(eMatchAsyncProcessedDSFalseLog, e.datasetID, ds))
		return false
	}

	e.logger.Info(fmt.Sprintf(eMatchAsyncProcessedDSTrueLog, e.datasetID, ds))

	return true
}

func (e *env) setSourceFile(ex string, f string) {
	var pth string

	filesystem := e.fsys

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

	_, err := fs.Stat(filesystem, pth)
	if err != nil {
		e.logger.Fatal(err.Error())
	}

	e.sourceFile = f

	e.logger.Info(fmt.Sprintf(sourceLog, f))
}

func (e *env) setDatasetID(id string) {
	logger := e.logger
	match, err := regexp.MatchString(regexDatasetMatch, id)

	if err != nil {
		logger.Fatal(err.Error())
	}

	if !match {
		logger.Fatal(fmt.Sprintf(datasetRegexLog, id, regexDatasetMatch))
	}

	// if no fatal datsets match
	ok := e.compareDatasetID(id)
	if !ok {
		return
	}

	e.datasetID = id

	logger.Info(fmt.Sprintf(datasetLog, id))
}

func (e *env) compareDatasetID(datasetID string) bool {
	logger := e.logger
	asyncProcessedDS := getAsyncProcessedDSID(logger)

	if asyncProcessedDS != datasetID {
		logger.Fatal(fmt.Sprintf(compareDatasetIDNotMatchLog, datasetID, asyncProcessedDS))
		return false
	}

	logger.Info(fmt.Sprintf(compareDatasetIDMatchLog, datasetID, asyncProcessedDS))

	return true
}

func (e *env) setTimeLimit(days int64) {
	logger := e.logger

	if days == 0 {
		logger.Warn(timelimitNoDaysLog)
		return
	}

	now := time.Now()
	limit := now.Add(-24 * time.Duration(days) * time.Hour)

	e.limit = limit

	logger.Info(fmt.Sprintf(timelimitDaysLog, days, limit))
}

func (e *env) setDryRun(dryrun bool) {
	logger := e.logger

	if dryrun {
		e.afs = afero.NewReadOnlyFs(afero.NewOsFs())
		e.dryrun = true

		logger.Info(dryRunTrueLog)
	} else {
		e.afs = afero.NewOsFs()
		e.dryrun = false

		logger.Warn(dryRunFalseLog)
	}
}

func (e *env) setTestRun(testrun bool) bool {
	logger := e.logger

	e.testrun = testrun

	if testrun {
		logger.Info(testRunTrueLog)
	} else {
		logger.Warn(testRunFalseLog)
	}

	return testrun
}

func (e *env) setSysIP() {
	hostname := wrapOs(e.logger, osHostnameLog, os.Hostname)

	ip := wrapLookupIP(e.logger, hostname, net.LookupIP)

	e.sysIP = ip
}

func (e *env) setPWD(ex string) string {
	// job needs to run in root dir
	exPath := filepath.Dir(ex)

	parts := strings.Split(exPath, string(os.PathSeparator))
	dots := ""

	for i := 0; i < (len(parts) - 1); i++ {
		dots = dots + "../"
	}

	err := os.Chdir(dots)
	if err != nil {
		e.logger.Fatal(err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		e.logger.Fatal(err)
	}

	return pwd
}

func (ap *asyncProcessor) getEnv() *env {
	return ap.Env
}

func (ap *asyncProcessor) getFiles() *[]File {
	return ap.Files
}

func (ap *asyncProcessor) setEnv(env *env) {
	ap.Env = env
}

func (ap *asyncProcessor) setFiles() {
	e = ap.Env
	afs := e.afs
	logger := e.logger

	_, err := afs.Stat(e.sourceFile)
	if err != nil {
		logger.Fatal(err)
	}

	lines := ap.parseSourceFile()

	for _, line := range lines {
		newFile := parseLine(line, e)
		newFile.fileInfo, err = afs.Stat(newFile.stagingPath)

		if err != nil {
			// Need to add testing
			logger.Error(err)
			continue
		}

		*ap.Files = append(*ap.Files, &newFile)
		logger.Info(fmt.Sprintf(fAddedToListLog,
			newFile.smbName,
			newFile.id,
			newFile.stagingPath,
			newFile.createTime.Unix(),
			newFile.size,
			newFile.fanIP,
			newFile.fileInfo.Name()))
	}
}

func init() {
	// Get pointer to new Env
	e = new(env)

	log.Init()

	// Set logger
	e.logger = log.GetLogger()

	flag.StringVar(&sourceFile, sourceFileArgTxt, "", sourceFileArgHelp)
	flag.StringVar(&datasetID, datasetIDArgTxt, "", datasetIDArgHelp)
	flag.Int64Var(&numDays, timelimitArgTxt, 0, timelimitArgHelp)
	flag.BoolVar(&dryrun, dryrunArgTxt, true, dryrunArgHelp)
	flag.BoolVar(&testrun, testrunArgTxt, false, testrunArgHelp)

	// Get executable path
	e.exePath = wrapOs(e.logger, osExecutableLog, os.Executable)

	// Set PWD to root
	root := e.setPWD(e.exePath)

	e.fsys = os.DirFS(root)
	e.afs = afero.NewOsFs()

	ap = NewAsyncProcessor(e, files)
}

func main() {
	// Parse flags
	flag.Parse()

	if e.setTestRun(testrun) {
		ap = testIntegrationTestSetup
	}

	e.setSourceFile(e.exePath, sourceFile)
	e.setDatasetID(datasetID)
	e.setTimeLimit(numDays)
	e.setDryRun(dryrun)

	e.setSysIP()

	e.verifyDataset()

	ap.setFiles()
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
