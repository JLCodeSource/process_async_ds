package main

import (
	"crypto/sha256"
	"fmt"

	"github.com/spf13/afero"
)

const (
	fHashLog = "%v (file.id:%v) %v-move file.hash: %x"
)

func (f *file) hasher() {
	var prePost string

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

	if f.oldStagingPath == "" {
		prePost = "pre"
	} else {
		prePost = "post"
	}

	logger.Info(fmt.Sprintf(fHashLog, f.smbName, f.id, prePost, f.hash))
}
