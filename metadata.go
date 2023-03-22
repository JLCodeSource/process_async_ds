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
	gbrAsyncProcessedDSErrLog   = "gbr could not verify AsyncProcessedDataset"
	gbrGetAsyncProcessedDSLog   = "gbr pool got output:%v"
	gbrParseAsyncProcessedDSLog = "gbr verified asyncProcessedDataset as %v"
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
	out = parseAsyncProcessedDSID(out, logger)

	return out
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

// Errors

func asyncProcessedDSIDErrLog(err error, logger *logrus.Logger) {
	// We want it to crash out on not finding the asyncProcessedDS
	err = errors.New(cleanGbrOut(err.Error()))
	logger.Error(err)
	logger.Fatal(gbrAsyncProcessedDSErrLog)
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
