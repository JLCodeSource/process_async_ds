package main

import (
	"fmt"
	"io/fs"
	"net"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	fIPMatchTrueLog               = "%v (file.id:%v) file.ip:%v matches comparison ip:%v"
	fIPMatchFalseLog              = "%v (file.id:%v) file.ip:%v does not match comparison ip:%v; skipping file"
	fCreateTimeAfterTimeLimitLog  = "%v (file.id:%v) file.createTime:%v is after timelimit:%v"
	fCreateTimeBeforeTimeLimitLog = "%v (file.id:%v) file.createTime:%v is before timelimit:%v; skipping file"
	fDatasetMatchTrueLog          = "%v (file.id:%v) file.datasetID:%v matches asyncProcessedDataset:%v"
	fDatasetMatchFalseLog         = "%v (file.id:%v) file.datasetID:%v does not match asyncProcessedDataset:%v; skipping file"
	fExistsTrueLog                = "%v (file.id:%v) exists at file.stagingPath:%v"
	fExistsFalseLog               = "%v (file.id:%v) does not exist at file.stagingPath:%v; skipping file"
	fSizeMatchTrueLog             = "%v (file.id:%v) file.size:%v matches size in file.stagingPath size:%v"
	fSizeMatchFalseLog            = "%v (file.id:%v) file.size:%v does not match size in file.stagingPath size:%v; skipping file"
	fCreateTimeMatchTrueLog       = "%v (file.id: %v) file.createTime:%v matches comparator fileinfo.modTime:%v"
	fCreateTimeMatchFalseLog      = "%v (file.id: %v) file.createTime:%v does not match comparator fileinfo.modTime:%v; skipping file"
)

func (f *File) verifyIP(ip net.IP, logger *logrus.Logger) bool {
	if reflect.DeepEqual(f.fanIP, ip) {
		logger.Info(fmt.Sprintf(fIPMatchTrueLog, f.smbName, f.id, f.fanIP, ip))
	} else {
		logger.Warn(fmt.Sprintf(fIPMatchFalseLog, f.smbName, f.id, f.fanIP, ip))

	}
	return reflect.DeepEqual(f.fanIP, ip)
}

func (f *File) verifyTimeLimit(limit time.Time, logger *logrus.Logger) bool {
	if f.createTime.After(limit) {
		logger.Info(fmt.Sprintf(
			fCreateTimeAfterTimeLimitLog,
			f.smbName,
			f.id,
			f.createTime.Round(time.Millisecond),
			limit.Round(time.Millisecond)))
	} else {
		logger.Warn(fmt.Sprintf(
			fCreateTimeBeforeTimeLimitLog,
			f.smbName,
			f.id,
			f.createTime.Round(time.Millisecond),
			limit.Round(time.Millisecond)))
	}
	return f.createTime.After(limit)
}

func (f *File) verifyInProcessedDataset(datasetID string, logger *logrus.Logger) bool {
	if f.datasetID == datasetID {
		logger.Info(fmt.Sprintf(fDatasetMatchTrueLog, f.smbName, f.id, f.datasetID, datasetID))
	} else {
		logger.Warn(fmt.Sprintf(fDatasetMatchFalseLog, f.smbName, f.id, f.datasetID, datasetID))
	}
	return f.datasetID == datasetID
}

func (f *File) verifyExists(fsys fs.FS, logger *logrus.Logger) bool {
	_, err := fs.Stat(fsys, f.stagingPath)
	if err != nil {
		logger.Warn(fmt.Sprintf(fExistsFalseLog, f.smbName, f.id, f.stagingPath))
		return false
	}
	logger.Info(fmt.Sprintf(fExistsTrueLog, f.smbName, f.id, f.stagingPath))
	return true
}

func (f *File) verifyFileSize(fsys fs.FS, logger *logrus.Logger) bool {
	info, err := fs.Stat(fsys, f.stagingPath)
	if err != nil {
		logger.Warn(err)
		return false
	}
	if info.Size() != f.fileInfo.Size() {
		logger.Warn(fmt.Sprintf(fSizeMatchFalseLog, f.smbName, f.id, f.size, f.fileInfo.Size()))
		return false
	}
	logger.Info(fmt.Sprintf(fSizeMatchTrueLog, f.smbName, f.id, f.size, f.fileInfo.Size()))
	return true
}

func (f *File) verifyCreateTime(fsys fs.FS, logger *logrus.Logger) bool {
	fileInfo, err := fs.Stat(fsys, f.stagingPath)
	if err != nil {
		logger.Warn(err)
		return false
	}
	if !fileInfo.ModTime().Equal(f.createTime) {
		logger.Warn(fmt.Sprintf(fCreateTimeMatchFalseLog,
			f.smbName,
			f.id,
			f.createTime.Round(time.Millisecond),
			f.fileInfo.ModTime().Round(time.Millisecond)))
		return false
	}
	logger.Info(fmt.Sprintf(
		fCreateTimeMatchTrueLog,
		f.smbName,
		f.id,
		f.createTime.Round(time.Millisecond),
		fileInfo.ModTime().Round(time.Millisecond)))
	return true
}
