package main

import (
	"bufio"
	"io/fs"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
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
		logger.Info("Processing: " + scanner.Text())
	}

	return lines

}

func parseLine(line string, logger *logrus.Logger) File {
	fileMetadata := strings.SplitAfter(line, "|")
	for i := 0; i < len(fileMetadata); i++ {
		// The first columns have a | after the content, the last one doesn't
		if i < (len(fileMetadata) - 1) {
			fileMetadata[i] = fileMetadata[i][0 : len(fileMetadata[i])-1]
		}
	}

	path := fileMetadata[0]
	logger.Info("path: " + path)
	datestring := fileMetadata[1]
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		logger.Fatal(err)
	}
	datetime, _ := time.ParseInLocation(time.UnixDate, datestring, loc)
	logger.Info("createTime: " + strconv.FormatInt(datetime.Unix(), 10))
	sizeStr := fileMetadata[2]
	logger.Info("size: " + sizeStr)
	id := fileMetadata[3]
	logger.Info("id: " + id)

	size, _ := strconv.ParseInt(sizeStr, 10, 64)

	file := File{
		path:       path,
		createTime: datetime,
		size:       size,
		id:         id}
	return file
}
