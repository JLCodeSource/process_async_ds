package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	parseFileLog   = "Processing: %v"
	smbNameLog     = "%v; file.smbName: %v"
	stagingPathLog = "%v; file.stagingPath: %v"
	createTimeLog  = "%v; file.createTime: %s#v"
	sizeLog        = "%v; file.size: %v"
	idLog          = "%v; file.id: %v"
	fanIPLog       = "%v; file.fanIP: %v"
)

func parseFile(fsys fs.FS, f string, logger *logrus.Logger) []string {

	file, err := fsys.Open(f)

	if err != nil {
		logger.Fatal(err)
	}

	lines := []string{}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		logger.Info(fmt.Sprintf(parseFileLog, scanner.Text()))
	}

	return lines

}

func parseLine(line string, logger *logrus.Logger) File {
	var dateTime time.Time

	fileMetadata := strings.SplitAfter(line, "|")
	for i := 0; i < len(fileMetadata); i++ {
		// The first columns have a | after the content, the last one doesn't
		if i < (len(fileMetadata) - 1) {
			fileMetadata[i] = fileMetadata[i][0 : len(fileMetadata[i])-1]
		}
	}

	id := fileMetadata[4]
	processing := fmt.Sprintf(parseFileLog, id)

	smbName := fileMetadata[0]
	logger.Info(fmt.Sprintf(smbNameLog, processing, smbName))
	stagingPath := fileMetadata[1]
	logger.Info(fmt.Sprintf(stagingPathLog, processing, stagingPath))
	dateTimeString := fileMetadata[2]
	dateTimeInt, err := strconv.ParseInt(dateTimeString, 10, 64)
	if err != nil {
		logger.Warn(err)
		loc, err := time.LoadLocation("America/New_York")
		if err != nil {
			logger.Fatal(err)
		}
		dateTime, err = time.ParseInLocation(time.UnixDate, dateTimeString, loc)
		if err != nil {
			logger.Fatal(err)
		}

	} else {
		dateTime = time.Unix(dateTimeInt, 0)
	}
	logger.Info(fmt.Sprintf(createTimeLog, processing, dateTime.UTC()))

	sizeStr := fileMetadata[3]
	logger.Info(fmt.Sprintf(sizeLog, processing, sizeStr))
	// Set above
	logger.Info(fmt.Sprintf(idLog, processing, id))
	fanIP := net.ParseIP(fileMetadata[5])
	logger.Info(fmt.Sprintf(fanIPLog, processing, fanIP))

	size, _ := strconv.ParseInt(sizeStr, 10, 64)

	file := File{
		smbName:     smbName,
		stagingPath: stagingPath,
		createTime:  dateTime,
		size:        size,
		id:          id,
		fanIP:       fanIP}
	return file
}
