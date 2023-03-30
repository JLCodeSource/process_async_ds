package main

import (
	"io/fs"
	"net"
	"time"
)

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
