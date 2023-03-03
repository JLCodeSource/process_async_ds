package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	gbrAsyncProcessedDSErrLog = "gbr could not verify AsyncProcessed dataset"
	gbrAsyncProcessedDSLog    = "gbr verified asyncProcessedDataset as %v"
	gbrFileNameByIDLog      = "gbr verified file.id:%v as having filename:%v"
	gbrDatasetByIDLog       = "gbr verified file.id:%v as having dataset:%v"
)

func getAsyncProcessedDSID(logger *logrus.Logger) string {
	cmd := exec.Command("gbr", "pool", "ls", "-d")
	cmdOut, _ := cmd.Output()
	//if err != nil {
	//	logger.Fatal(gbrAsyncProcessedDSErrLog)
	//}
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
	cmd := exec.Command("gbr", "file", "ls", "-i", id)
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
	cmd := exec.Command("gbr", "file", "ls", "-i", id, "-d")
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
