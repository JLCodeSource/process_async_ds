package main

import (
	"bufio"
	"io/fs"

	"github.com/sirupsen/logrus"
)

func parseFile(fsys fs.FS, f string, logger *logrus.Logger) []string {

	file, err := fsys.Open(f)

	if err != nil {
		logger.Fatal(err)
	}

	out := []string{}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		out = append(out, scanner.Text())
		logger.Info("Processing: " + scanner.Text())
	}

	return out

}
