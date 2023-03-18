package main

import (
	"crypto/sha256"
	"fmt"
	"io/fs"

	"github.com/sirupsen/logrus"
)

const (
	fHashLog = "%v (file.id:%v) file.hash: %x"
)

func (f *File) hasher(fsys fs.FS, logger *logrus.Logger) {
	// fs.ReadFile handles close?
	content, err := fs.ReadFile(fsys, f.stagingPath)
	if err != nil {
		// NB No need for fatal as if hash does not match, it will fail later
		logger.Error(err)
	}
	sha := sha256.Sum256(content)
	logger.Info(fmt.Sprintf(fHashLog, f.smbName, f.id, f.hash))
	f.hash = sha
}
