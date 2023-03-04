package main

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	fIPMatchTrueLog                 = "%v (file.id:%v) file.ip:%v matches comparison ip:%v"
	fIPMatchFalseLog                = "%v (file.id:%v) file.ip:%v does not match comparison ip:%v; skipping file"
	fCreateTimeAfterTimeLimitLog    = "%v (file.id:%v) file.createTime:%v is after timelimit:%v"
	fCreateTimeBeforeTimeLimitLog   = "%v (file.id:%v) file.createTime:%v is before timelimit:%v; skipping file"
	fDatasetMatchTrueLog            = "%v (file.id:%v) file.datasetID:%v matches asyncProcessedDataset:%v"
	fDatasetMatchFalseLog           = "%v (file.id:%v) file.datasetID:%v does not match asyncProcessedDataset:%v; skipping file"
	fExistsTrueLog                  = "%v (file.id:%v) exists at file.stagingPath:%v"
	fExistsFalseLog                 = "%v (file.id:%v) does not exist at file.stagingPath:%v; skipping file"
	fSizeMatchTrueLog               = "%v (file.id:%v) file.size:%v matches size in file.stagingPath size:%v"
	fSizeMatchFalseLog              = "%v (file.id:%v) file.size:%v does not match size in file.stagingPath size:%v; skipping file"
	fCreateTimeMatchTrueLog         = "%v (file.id: %v) file.createTime:%v matches comparator fileinfo.modTime:%v"
	fCreateTimeMatchFalseLog        = "%v (file.id: %v) file.createTime:%v does not match comparator fileinfo.modTime:%v; skipping file"
	fSmbNameMatchFileIDNameTrueLog  = "%v (file.id:%v) file.smbName:%v matches file.id name:%v"
	fSmbNameMatchFileIDNameFalseLog = "%v (file.id:%v) file.smbName:%v does not match file.id name:%v; skipping file"
	fStatMatchLog                   = "%v (file.id:%v) file.verifyStat passes all metadata checks for file.stagingPath:%v"
	fEnvMatchLog                    = "%v (file.id:%v) file.verfiyEnv passes all settings checks for file.stagingPath:%v"
	fGbrFileNameByFileIDLog         = "%v (file.id:%v) gbr verified file.id:%v as matching filename:%v"
	fGbrNoFileNameByFileIDLog       = "%v (file.id:%v) gbr could not find file.id:%v"
	fGbrDatasetByFileIDLog          = "%v (file.id:%v) gbr verified file.id:%v as matching dataset:%v"
)

// verify config metadata

func (f *File) verifyEnv(env Env, logger *logrus.Logger) bool {
	if !f.verifyIP(env.sysIP, logger) {
		return false
	}
	if !f.verifyTimeLimit(env.limit, logger) {
		return false
	}
	logger.Info(fmt.Sprintf(fEnvMatchLog, f.smbName, f.id, f.stagingPath))
	return true
}

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

// Verify GB internal metadata
/*
func (f *File) verifyGBMetadata(logger *logrus.Logger) bool {
	if !f.verifyInProcessedDataset(xxx, logger) {
		return true
	}
	return true
} */

func (f *File) getMBFileNameByFileID(logger *logrus.Logger) (string, bool) {
	id := f.id
	cmd := exec.Command("/usr/bin/gbr", "file", "ls", "-i", id)
	cmdOut, err := cmd.CombinedOutput()
	if err != nil {
		f.getByIDErrLog(err, logger)
		return "", false
	}
	out := string(cmdOut)
	out = cleanGbrOut(out)
	if out == "" {
		logger.Warn(fmt.Sprintf(fGbrNoFileNameByFileIDLog, f.smbName, id, id))
		return "", false
	}
	filename := f.parseMBFileNameByFileID(out, logger)
	return filename, true
}

func (f *File) getMBDatasetByFileID(logger *logrus.Logger) (string, bool) {
	id := f.id
	cmd := exec.Command("/usr/bin/gbr", "file", "ls", "-i", id, "-d")
	cmdOut, err := cmd.CombinedOutput()
	if err != nil {
		f.getByIDErrLog(err, logger)
	}
	out := string(cmdOut)
	out = cleanGbrOut(out)
	if out == "" {
		logger.Warn(fmt.Sprintf(fGbrNoFileNameByFileIDLog, f.smbName, id, id))
		return "", false
	}
	datasetID := f.parseMBDatasetByFileID(out, f.id, logger)
	return datasetID, true
}

func (f *File) parseMBFileNameByFileID(cmdOut string, logger *logrus.Logger) (filename string) {
	line := strings.Split(cmdOut, " ")
	filename = line[2]
	logger.Info(fmt.Sprintf(fGbrFileNameByFileIDLog, f.smbName, f.id, f.id, filename))
	return
}

func (f *File) parseMBDatasetByFileID(cmdOut, id string, logger *logrus.Logger) (parentDS string) {
	lines := strings.Split(string(cmdOut), ";")
	for _, line := range lines {
		if strings.Contains(line, "parent id") {
			parentDS = line[len(line)-32:]
			logger.Info(fmt.Sprintf(fGbrDatasetByFileIDLog, f.smbName, f.id, f.id, parentDS))
			return
		}
	}
	// Should never happen as caught with previous checks
	logger.Warn(fmt.Sprintf(fGbrNoFileNameByFileIDLog, f.smbName, f.id, f.id))
	return
}

func (f *File) getByIDErrLog(err error, logger *logrus.Logger) {
	err = errors.New(cleanGbrOut(err.Error()))
	logger.Warn(err)
	logger.Warn(fmt.Sprintf(fGbrNoFileNameByFileIDLog, f.smbName, f.id, f.id))
}

func (f *File) verifyInProcessedDataset(datasetID string, logger *logrus.Logger) bool {
	if f.datasetID == datasetID {
		logger.Info(fmt.Sprintf(fDatasetMatchTrueLog, f.smbName, f.id, f.datasetID, datasetID))
	} else {
		logger.Warn(fmt.Sprintf(fDatasetMatchFalseLog, f.smbName, f.id, f.datasetID, datasetID))
	}
	return f.datasetID == datasetID
}

func (f *File) verifyFileIDName(fileName string, logger *logrus.Logger) bool {
	if f.smbName == fileName {
		logger.Info(fmt.Sprintf(
			fSmbNameMatchFileIDNameTrueLog, f.smbName, f.id, f.smbName, fileName))
	} else {
		logger.Warn(fmt.Sprintf(
			fSmbNameMatchFileIDNameFalseLog, f.smbName, f.id, f.smbName, fileName))
	}
	return f.smbName == fileName
}

// Verify local FS metadata

func (f *File) verifyStat(fsys fs.FS, logger *logrus.Logger) bool {
	fileInfo, err := fs.Stat(fsys, f.stagingPath)
	if err != nil {
		logger.Warn(fmt.Sprintf(fExistsFalseLog, f.smbName, f.id, f.stagingPath))
		return false
	}
	logger.Info(fmt.Sprintf(fExistsTrueLog, f.smbName, f.id, f.stagingPath))
	if !f.verifyFileSize(fileInfo.Size(), logger) {
		return false
	}
	if !f.verifyCreateTime(fileInfo.ModTime(), logger) {
		return false
	}
	logger.Info(fmt.Sprintf(fStatMatchLog, f.smbName, f.id, f.stagingPath))
	return true
}

/* func (f *File) verifyExists(fsys fs.FS, logger *logrus.Logger) bool {
	_, err := fs.Stat(fsys, f.stagingPath)
	if err != nil {
		logger.Warn(fmt.Sprintf(fExistsFalseLog, f.smbName, f.id, f.stagingPath))
		return false
	}
	logger.Info(fmt.Sprintf(fExistsTrueLog, f.smbName, f.id, f.stagingPath))
	return true
} */

func (f *File) verifyFileSize(size int64, logger *logrus.Logger) bool {
	if size != f.fileInfo.Size() {
		logger.Warn(fmt.Sprintf(fSizeMatchFalseLog, f.smbName, f.id, f.size, f.fileInfo.Size()))
		return false
	}
	logger.Info(fmt.Sprintf(fSizeMatchTrueLog, f.smbName, f.id, f.size, f.fileInfo.Size()))
	return true
}

func (f *File) verifyCreateTime(t time.Time, logger *logrus.Logger) bool {
	if !t.Equal(f.createTime) {
		logger.Warn(fmt.Sprintf(fCreateTimeMatchFalseLog,
			f.smbName,
			f.id,
			f.createTime.Round(time.Millisecond),
			t.Round(time.Millisecond)))
		return false
	}
	logger.Info(fmt.Sprintf(
		fCreateTimeMatchTrueLog,
		f.smbName,
		f.id,
		f.createTime.Round(time.Millisecond),
		t.Round(time.Millisecond)))
	return true
}
