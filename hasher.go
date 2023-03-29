package main

import (
	"crypto/sha256"
	"fmt"

	"github.com/spf13/afero"
)

const (
	fHashLog = "%v (file.id:%v) file.hash: %x"
)

func (f *File) Hasher() {
	e = ap.getEnv()
	afs = e.afs
	logger := e.logger
	// fs.ReadFile handles close?
	content, err := afero.ReadFile(afs, f.stagingPath)
	if err != nil {
		// NB No need for fatal as if hash does not match, it will fail later
		logger.Error(err)
	}

	sha := sha256.Sum256(content)

	f.hash = sha

	logger.Info(fmt.Sprintf(fHashLog, f.smbName, f.id, f.hash))
}
