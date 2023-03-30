package main

import (
	"crypto/sha256"
	"io/fs"
	"net"
	"time"
)

// mockAsyncProcessor

type mockAsyncProcessor struct {
	Env   *env
	Files *[]File
}

func (m mockAsyncProcessor) getFiles() *[]File {
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

// getters

func (mf *mockFile) getID() string {
	return ""
}

func (mf *mockFile) getSmbName() string {
	return ""
}

func (mf *mockFile) getCreateTime() time.Time {
	return time.Now()
}

func (mf *mockFile) getSize() int64 {
	return 0
}

func (mf *mockFile) getDatasetID() string {
	return ""
}

func (mf *mockFile) getFanIP() net.IP {
	return nil
}

func (mf *mockFile) getStagingPath() string {
	return ""
}

func (mf *mockFile) getOldStagingPath() string {
	return ""
}

func (mf *mockFile) getHash() [32]byte {
	b := sha256.Sum256([]byte(""))
	return b
}

func (mf *mockFile) getOldHash() [32]byte {
	b := sha256.Sum256([]byte(""))
	return b
}

func (mf *mockFile) getFileInfo() fs.FileInfo {
	return nil
}

func (mf *mockFile) getSuccess() bool {
	return true
}

// setters

func (mf *mockFile) setOldHash(hash [32]byte) {
	mf.oldHash = hash
}

func (mf *mockFile) setOldStagingPath(stagingPath string) {
	mf.oldStagingPath = stagingPath
}

func (mf *mockFile) setSuccess(success bool) {
	mf.success = success
}

// methods

func (mf *mockFile) compareHashes() bool {
	return true
}

func (mf *mockFile) getByIDErrLog(err error) {}

func (mf *mockFile) hasher() {}

func (mf *mockFile) move() {}

func (mf *mockFile) parseMBFileNameByFileID(_ string) (_ string) {
	return ""
}

func (mf *mockFile) setMBDatasetByFileID(_ string) {}

func (mf *mockFile) verify() bool {
	return true
}

func (mf *mockFile) verifyCreateTime(_ time.Time) bool {
	return true
}

func (mf *mockFile) verifyEnvMatch() bool {
	return true
}

func (mf *mockFile) verifyFileIDName(_ string) bool {
	return true
}

func (mf *mockFile) verifyFileSize(_ int64) bool {
	return true
}

func (mf *mockFile) verifyGBMetadata() bool {
	return true
}

func (mf *mockFile) verifyIP() bool {
	return true
}

func (mf *mockFile) verifyInDataset(_ string) bool {
	return true
}

func (mf *mockFile) verifyMBDatasetByFileID() bool {
	return true
}

func (mf *mockFile) verifyMBFileNameByFileID() bool {
	return true
}

func (mf *mockFile) verifyStat() bool {
	return true
}

func (mf *mockFile) verifyTimeLimit() bool {
	return true
}
