package main

import (
	"bufio"
	"io/fs"
	"strings"

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
	vals := strings.SplitAfter(line, "|")
	for i := 0; i < len(vals); i++ {
		vals[i] = vals[i][0 : len(vals[i])-1]
	}
	file := File{path: vals[0]}
	return file
}
