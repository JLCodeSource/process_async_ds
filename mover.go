package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const (
	fMoveFileLog = "%v: (file.id:%v) oldPath:%v, newPath:%v"
)

func (f *File) Move(afsys afero.Fs, logger *logrus.Logger) {
	oldLocation := f.stagingPath
	newLocation := newPath(f)
	dir, _ := path.Split(newLocation)
	_, err := afsys.Stat(dir)
	if err != nil {
		logger.Warn(err)
		wrapAferoMkdirAll(afsys, dir, logger)
	}
	err = afsys.Rename(oldLocation, newLocation)
	if err != nil {
		logger.Fatal(err)
	}
	f.stagingPath = newLocation
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

func wrapAferoMkdirAll(afsys afero.Fs, path string, logger *logrus.Logger) bool {
	err := afsys.MkdirAll(path, 0755)
	if err != nil && !os.IsExist(err) {
		logger.Fatal(err)
	}
	return true
}
