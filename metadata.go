package main

import (
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
	gbrDatasetByIDLog         = "gbr verified file.id:%v as having dataset:%v"
)

func getAsyncProcessedDSID(logger *logrus.Logger) string {
	cmd := exec.Command("/usr/bin/gbr", "pool", "ls", "-d")
	cmdOut, err := cmd.CombinedOutput()
	if err != nil {
		asyncProcessedDSIDErrLog(err, logger)
	}
	out := string(cmdOut)
	// Remove new lines from out
	out = strings.Replace(out, "\n", ";", -1)
	// Remove duplicate spaces from out
	space := regexp.MustCompile(`\s+`)
	out = space.ReplaceAllString(out, " ")
	logger.Info(fmt.Sprintf(gbrGetAsyncProcessedDSLog, out))

	return out
}

func asyncProcessedDSIDErrLog(err error, logger *logrus.Logger) {
	// We want it to crash out on not finding the asyncProcessedDS
	logger.Error(err)
	logger.Fatal(gbrAsyncProcessedDSErrLog)
}

func parseAsyncProcessedDSID(cmdOut string, logger *logrus.Logger) string {
	lines := strings.Split(string(cmdOut), ";")
	for _, line := range lines {
		if strings.Contains(line, "ID") {
			asyncDelDS := line[len(line)-32:]
			logger.Info(fmt.Sprintf(gbrParseAsyncProcessedDSLog, asyncDelDS))
			return asyncDelDS
		}
	}
	// We want it to crash out on not finding the asyncProcessedDS
	logger.Fatal(gbrAsyncProcessedDSErrLog)
	return ""
}

func getFileNameByID(id string, logger *logrus.Logger) string {
	cmd := exec.Command("/usr/bin/gbr", "file", "ls", "-i", id)
	cmdOut, _ := cmd.Output()
	//if err != nil {
	//		logger.Fatal(gbrAsyncProcessedDSErrLog)
	//	}
	line := strings.Split(string(cmdOut), " ")
	filename := line[2]
	logger.Info(fmt.Sprintf(gbrFileNameByIDLog, id, filename))
	return filename
}

func getDatasetByID(id string, logger *logrus.Logger) string {
	cmd := exec.Command("/usr/bin/gbr", "file", "ls", "-i", id, "-d")
	cmdOut, _ := cmd.Output()
	//if err != nil {
	//		logger.Fatal(gbrAsyncProcessedDSErrLog)
	//	}
	lines := strings.Split(string(cmdOut), "\n")
	for _, line := range lines {
		if strings.Contains(line, "parent id") {
			parentDS := line[len(line)-32:]
			logger.Info(fmt.Sprintf(gbrDatasetByIDLog, id, parentDS))
			return parentDS
		}
	}
	return ""
}
