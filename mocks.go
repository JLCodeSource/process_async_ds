package main

import (
	"io/fs"
	"net"
	"time"
)

// mockAsyncProcessor

type mockAsyncProcessor struct {
	Env   *env
	Files *[]file
}

func (m mockAsyncProcessor) getFiles() *[]file {
	return m.Files
}

func (m mockAsyncProcessor) getEnv() *env {
	return m.Env
}

func (m mockAsyncProcessor) setEnv(_ *env) {
	//m.Env = env
}

func (m mockAsyncProcessor) setFiles() {
}

func (m mockAsyncProcessor) processFiles() {
}

func (m mockAsyncProcessor) parseSourceFile() []string {
	return []string{}
}

func (m mockAsyncProcessor) parseLine(_ string) file {
	var f file
	return f
}

// mockFile

type mockFile struct {
	id             string
	smbName        string
	createTime     time.Time
	size           int64
	datasetID      string
	fanIP          net.IP
	stagingPath    string
	oldStagingPath string
	hash           [32]byte
	oldHash        [32]byte
	fileInfo       fs.FileInfo
	success        bool
}

func (mf *mockFile) compareHashes() bool

func (mf *mockFile) getByIDErrLog(err error)

func (mf *mockFile) hasher()

func (mf *mockFile) move()

func (mf *mockFile) parseMBFileNameByFileID(_ string) (_ string) {
	return ""
}

func (mf *mockFile) setMBDatasetByFileID(_ string)

func (mf *mockFile) verify() bool
func (mf *mockFile) verifyCreateTime(_ time.Time) bool

func (mf *mockFile) verifyEnvMatch() bool

func (mf *mockFile) verifyFileIDName(_ string) bool

func (mf *mockFile) verifyFileSize(_ int64) bool

func (mf *mockFile) verifyGBMetadata() bool

func (mf *mockFile) verifyIP() bool

func (mf *mockFile) verifyInDataset(_ string) bool

func (mf *mockFile) verifyMBDatasetByFileID() bool

func (mf *mockFile) verifyMBFileNameByFileID() bool

func (mf *mockFile) verifyStat() bool

func (mf *mockFile) verifyTimeLimit() bool
