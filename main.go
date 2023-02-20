package main

import (
	"regexp"

	log "github.com/JLCodeSource/process_async_ds/logger"

	"flag"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

const ()

func getAsyncProcessedFolderId(id string, logger *logrus.Logger) string {
	match, err := regexp.MatchString("^[A-F0-9]{32}$", id)
	if err != nil {
		logger.Errorf("DatasetId " + id + " & regex error ^[A-F0-9]{32}$")
		logger.Errorf(err.Error())
	}
	if match {
		logger.Info("DatasetId set to " + id)
		return id
	} else {
		logger.Fatal("DatasetId: 123 not of the form ^[A-F0-9]{32}$")
		return id
	}

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

func main() {

	log.Init()
	log.GetLogger()
	logger := log.GetLogger()

	datasetPtr := flag.String("dataset", "", "async processed dataset id (default '')")
	limitPtr := flag.Int64("days", 0, "number of days ago (default 0)")
	nonDryRunPtr := flag.Bool("non-dryrun", false, "execute non dry run (default false)")

	flag.Parse()

	getAsyncProcessedFolderId(*datasetPtr, logger)
	getTimeLimit(*limitPtr, logger)
	getNonDryRun(*nonDryRunPtr, logger)

}
