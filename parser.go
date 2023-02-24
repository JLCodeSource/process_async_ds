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
		if i < (len(fileMetadata) - 1) {
			fileMetadata[i] = fileMetadata[i][0 : len(fileMetadata[i])-1]
		}
	}

	path := fileMetadata[0]

	loc, _ := time.LoadLocation("America/New_York")
	datestring := fileMetadata[1]
	datetime, _ := time.ParseInLocation(time.UnixDate, datestring, loc)

	size, _ := strconv.ParseInt(fileMetadata[2], 10, 64)

	id := fileMetadata[3]

	file := File{
		path:       path,
		createTime: datetime,
		size:       size,
		id:         id}
	return file
}
