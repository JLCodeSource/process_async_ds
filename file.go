package main

import (
	"io/fs"
	"net"
	"time"
)

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
