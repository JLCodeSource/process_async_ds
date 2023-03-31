package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	parseFileLog   = "Processing: %v"
	smbNameLog     = "%v; file.smbName: %v"
	stagingPathLog = "%v; file.stagingPath: %v"
	createTimeLog  = "%v; file.createTime: %v"
	sizeLog        = "%v; file.size: %v"
	idLog          = "%v; file.id: %v"
	fanIPLog       = "%v; file.fanIP: %v"

	easternTime = "America/New_York"
)

func parseSourceFile(e *env) []string {
	afs = e.afs
	sf := e.sourceFile
	logger := e.logger

	_, err := afs.Stat(sf)
	if err != nil {
		e.logger.Fatal(err)
	}

	file, err := afs.Open(sf)
	if err != nil {
		e.logger.Fatal(err)
	}

	//defer file.Close()

	lines := []string{}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		logger.Info(fmt.Sprintf(parseFileLog, scanner.Text()))
	}

	err = file.Close()
	if err != nil {
		logger.Fatal(err)
	}

	return lines
}

func parseLine(line string, e *env) file {
	var dateTime time.Time

	fileMetadata := strings.SplitAfter(line, "|")
	// len-1 because the last split is empty
	for i := 0; i < len(fileMetadata)-1; i++ {
		fileMetadata[i] = fileMetadata[i][0 : len(fileMetadata[i])-1]
	}

	id := fileMetadata[4]
	processing := fmt.Sprintf(parseFileLog, id)

	smbName := fileMetadata[0]

	e.logger.Info(fmt.Sprintf(smbNameLog, processing, smbName))

	stagingPath := fileMetadata[1]

	e.logger.Info(fmt.Sprintf(stagingPathLog, processing, stagingPath))

	dateTimeString := fileMetadata[2]
	dateTimeInt, err := strconv.ParseInt(dateTimeString, 10, 64)

	if err != nil {
		e.logger.Warn(err)
		loc, err := time.LoadLocation(easternTime)

		if err != nil {
			e.logger.Fatal(err)
		}

		dateTime, err = time.ParseInLocation(time.UnixDate, dateTimeString, loc)
		if err != nil {
			e.logger.Fatal(err)
		}
	} else {
		dateTime = time.Unix(dateTimeInt, 0)
	}

	e.logger.Info(fmt.Sprintf(createTimeLog, processing, dateTime.UTC()))

	sizeStr := fileMetadata[3]
	e.logger.Info(fmt.Sprintf(sizeLog, processing, sizeStr))
	// Set above
	e.logger.Info(fmt.Sprintf(idLog, processing, id))

	fanIP := net.ParseIP(fileMetadata[5])
	e.logger.Info(fmt.Sprintf(fanIPLog, processing, fanIP))

	size, _ := strconv.ParseInt(sizeStr, 10, 64)

	file := file{
		smbName:     smbName,
		stagingPath: stagingPath,
		createTime:  dateTime,
		size:        size,
		id:          id,
		fanIP:       fanIP}

	return file
}
