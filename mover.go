package main

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	fMoveFileLog = "%v: (file.id:%v) oldPath:%v, newPath:%v"
)

func (f *File) UpdatePath() string {
	oldDir, fn := path.Split(f.stagingPath)
	parts := strings.Split(oldDir, string(os.PathSeparator))
	lastParts := parts[2:]
	firstParts := parts[:2]
	fp := strings.Join(firstParts, string(os.PathSeparator))
	lp := strings.Join(lastParts, string(os.PathSeparator))
	return fp + ".processed" + string(os.PathSeparator) + lp + fn
}

func (f *File) Move(newLocation string, logger *logrus.Logger) {
	oldLocation := f.stagingPath
	_ = os.Rename(oldLocation, newLocation)
	//if err != nil {
	//	logger.Fatal(err)
	//}
	logger.Info(fmt.Sprintf(fMoveFileLog, f.smbName, f.id, oldLocation, newLocation))
}
