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
)

func getAsyncProcessedDSID(logger *logrus.Logger) string {
	cmd := exec.Command("gbr", "pool", "ls", "-d")
	cmdOut, err := cmd.Output()
	if err != nil {
		logger.Fatal(gbrAsyncProcessedDSErrLog)
	}
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
