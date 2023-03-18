package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	fMoveFileLog = "%v: (file.id:%v) oldPath:%v, newPath:%v"
)

func (f *File) Move(fsys fs.FS, logger *logrus.Logger) {
	oldLocation := f.stagingPath
	newLocation := newPath(f)
	_, err := fsys.Open(newLocation)
	if err != nil {
		logger.Warn(err)
		wrapOsMkdirAll(newLocation, logger)
	}
	err = os.Rename(oldLocation, newLocation)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info(fmt.Sprintf(fMoveFileLog, f.smbName, f.id, oldLocation, newLocation))
}

func newPath(f *File) string {
	oldDir, fn := path.Split(f.stagingPath)
	parts := strings.Split(oldDir, string(os.PathSeparator))
	lastParts := parts[2:]
	firstParts := parts[:2]
	fp := strings.Join(firstParts, string(os.PathSeparator))
	lp := strings.Join(lastParts, string(os.PathSeparator))
	return fp + ".processed" + string(os.PathSeparator) + lp + fn
}

func wrapOsMkdirAll(path string, logger *logrus.Logger) bool {
	err := os.MkdirAll(path, 0755)
	if err != nil && !os.IsExist(err) {
		logger.Fatal(err)
	}
	return true
}
