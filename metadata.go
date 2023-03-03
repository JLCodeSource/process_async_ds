package main

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	gbrAsyncProcessedDSErrLog   = "gbr could not verify AsyncProcessed dataset"
	gbrGetAsyncProcessedDSLog   = "gbr pool got output:%v"
	gbrParseAsyncProcessedDSLog = "gbr verified asyncProcessedDataset as %v"
	gbrFileNameByIDLog        = "gbr verified file.id:%v as having filename:%v"
	gbrNoFileNameByIDLog      = "gbr could not find file.id:%v"
	gbrFileDatasetByIDLog     = "gbr verified file.id:%v as having dataset:%v"
)

// Getters

func getAsyncProcessedDSID(logger *logrus.Logger) string {
	cmd := exec.Command("/usr/bin/gbr", "pool", "ls", "-d")
	cmdOut, err := cmd.CombinedOutput()
	if err != nil {
		asyncProcessedDSIDErrLog(err, logger)
	}
	out := string(cmdOut)
	out = cleanGbrOut(out)
	logger.Info(fmt.Sprintf(gbrGetAsyncProcessedDSLog, out))

	return out
}

func getFileNameByID(id string, logger *logrus.Logger) (string, bool) {
	cmd := exec.Command("/usr/bin/gbr", "file", "ls", "-i", id)
	cmdOut, err := cmd.CombinedOutput()
	if err != nil {
		getByIDErrLog(id, err, logger)
		return "", false
	}
	out := string(cmdOut)
	out = cleanGbrOut(out)
	if out == "" {
		logger.Warn(fmt.Sprintf(gbrNoFileNameByIDLog, id))
		return "", false
	}
	filename := parseFileNameByID(out, id, logger)
	logger.Info(fmt.Sprintf(gbrFileNameByIDLog, id, filename))
	return filename, true
}

func getFileDatasetByID(id string, logger *logrus.Logger) (string, bool) {
	cmd := exec.Command("/usr/bin/gbr", "file", "ls", "-i", id, "-d")
	cmdOut, err := cmd.Output()
	if err != nil {
		getByIDErrLog(id, err, logger)
	}
	out := string(cmdOut)
	out = cleanGbrOut(out)
	if out == "" {
		logger.Warn(fmt.Sprintf(gbrNoFileNameByIDLog, id))
		return "", false
	}
	datasetID := parseFileDatasetByID(out, id, logger)
	logger.Info(fmt.Sprintf(gbrFileDatasetByIDLog, id, datasetID))
	return datasetID, true
}

// Parsers

func parseAsyncProcessedDSID(cmdOut string, logger *logrus.Logger) (asyncDelDS string) {
	lines := strings.Split(string(cmdOut), ";")
	for _, line := range lines {
		if strings.Contains(line, "ID") {
			asyncDelDS = line[len(line)-32:]
			logger.Info(fmt.Sprintf(gbrParseAsyncProcessedDSLog, asyncDelDS))
			return
		}
	}
	// We want it to crash out on not finding the asyncProcessedDS
	logger.Fatal(gbrAsyncProcessedDSErrLog)
	return
}

func parseFileNameByID(cmdOut, id string, logger *logrus.Logger) (filename string) {
	line := strings.Split(cmdOut, " ")
	filename = line[2]
	logger.Info(fmt.Sprintf(gbrFileNameByIDLog, id, filename))
	return
}

func parseFileDatasetByID(cmdOut, id string, logger *logrus.Logger) (parentDS string) {
	lines := strings.Split(string(cmdOut), ";")
	for _, line := range lines {
		if strings.Contains(line, "parent id") {
			parentDS = line[len(line)-32:]
			logger.Info(fmt.Sprintf(gbrFileDatasetByIDLog, id, parentDS))
			return
		}
	}
	logger.Warn(fmt.Sprintf(gbrNoFileNameByIDLog, id))
	return
}

// Errors

func asyncProcessedDSIDErrLog(err error, logger *logrus.Logger) {
	// We want it to crash out on not finding the asyncProcessedDS
	err = errors.New(cleanGbrOut(err.Error()))
	logger.Error(err)
	logger.Fatal(gbrAsyncProcessedDSErrLog)
}

func getByIDErrLog(id string, err error, logger *logrus.Logger) {
	err = errors.New(cleanGbrOut(err.Error()))
	logger.Warn(err)
	logger.Warn(fmt.Sprintf(gbrNoFileNameByIDLog, id))
}

// Cleaners

func cleanGbrOut(out string) string {
	// Remove new lines from out
	out = strings.Replace(out, "\n", ";", -1)
	// Remove duplicate spaces from out
	space := regexp.MustCompile(`\s+`)
	out = space.ReplaceAllString(out, " ")
	return out
}
