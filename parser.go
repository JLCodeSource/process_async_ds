package main

import (
	"bufio"
	"io/fs"

	"github.com/sirupsen/logrus"
)

func parseFile(fsys fs.FS, f string, logger *logrus.Logger) string {

	file, _ := fsys.Open(f)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	scanner.Scan()
	out := scanner.Text()

	logger.Info("Processing: " + out)
	return out

}
