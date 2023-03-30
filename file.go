package main

import (
	"io/fs"
	"net"
	"time"
)

// File interface is the interface for File
type File interface {
	compareHashes() bool

	getID() string
	getSmbName() string
	getCreateTime() time.Time
	getSize() int64
	getDatasetID() string
	getFanIP() net.IP
	getStagingPath() string
	getOldStagingPath() string
	getHash() [32]byte
	getOldHash() [32]byte
	getFileInfo() fs.FileInfo
	getSuccess() bool

	setOldHash([32]byte)
	setOldStagingPath(string)
	setSuccess(bool)

	getByIDErrLog(err error)
	hasher()
	move()
	parseMBFileNameByFileID(cmdOut string) (filename string)
	setMBDatasetByFileID(cmdOut string)
	verify() bool
	verifyCreateTime(t time.Time) bool
	verifyEnvMatch() bool
	verifyFileIDName(fileName string) bool
	verifyFileSize(size int64) bool
	verifyGBMetadata() bool
	verifyIP() bool
	verifyInDataset(datasetID string) bool
	verifyMBDatasetByFileID() bool
	verifyMBFileNameByFileID() bool
	verifyStat() bool
	verifyTimeLimit() bool
}

// file type is a struct holding file metadata
type file struct {
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

func (f *file) getID() string {
	return f.id
}

func (f *file) getSmbName() string {
	return f.smbName
}

func (f *file) getCreateTime() time.Time {
	return f.createTime
}

func (f *file) getSize() int64 {
	return f.size
}

func (f *file) getDatasetID() string {
	return f.datasetID
}

func (f *file) getFanIP() net.IP {
	return f.fanIP
}

func (f *file) getStagingPath() string {
	return f.stagingPath
}

func (f *file) getOldStagingPath() string {
	return f.oldStagingPath
}

func (f *file) getHash() [32]byte {
	return f.hash
}

func (f *file) getOldHash() [32]byte {
	return f.oldHash
}

func (f *file) getFileInfo() fs.FileInfo {
	return f.fileInfo
}

func (f *file) getSuccess() bool {
	return f.success
}

func (f *file) setOldHash(hash [32]byte) {
	f.oldHash = hash
}

func (f *file) setOldStagingPath(stagingPath string) {
	f.oldStagingPath = stagingPath
}

func (f *file) setSuccess(success bool) {
	f.success = success
}
