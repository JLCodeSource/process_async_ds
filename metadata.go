package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	gbrCmd                  = "/usr/bin/gbr"
	gbrArgsPool             = "pool"
	gbrArgsLs               = "ls"
	gbrArgsDetails          = "--details"
	gbrAsyncProcessedDSErrLog = "gbr could not verify AsyncProcessed dataset"
	gbrAsyncProcessedDSLog    = "gbr verified asyncProcessedDataset as %v"
	gbrFileNameByIDLog      = "gbr verified file.id:%v as having filename:%v"
	gbrDatasetByIDLog       = "gbr verified file.id:%v as having dataset:%v"
)

func getAsyncProcessedDSID(logger *logrus.Logger) (string, error) {
	cmd := exec.Command(gbrCmd, gbrArgsPool, gbrArgsLs, gbrArgsDetails)
	cmdOut, err := cmd.Output()
	if err != nil {
		logger.Fatal(fmt.Sprintf(gbrAsyncProcessedDSErrLog))
		return "", errors.New(gbrAsyncProcessedDSErrLog)
	}
	return string(cmdOut), nil
}

func parseAsyncProcessedDSID(cmdOut string, logger *logrus.Logger) string {
	lines := strings.Split(string(cmdOut), "\n")
	for _, line := range lines {
		if strings.Contains(line, "ID") {
			asyncDelDS := line[len(line)-32:]
			logger.Info(fmt.Sprintf(gbrAsyncProcessedDSLog, asyncDelDS))
			return asyncDelDS
		}
	}
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
