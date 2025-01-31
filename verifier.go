package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os/exec"
	"reflect"
	"strings"
	"time"
)

const (
	fIPMatchTrueLog                 = "%v (file.id:%v) file.ip:%v matches comparison ip:%v"
	fIPMatchFalseLog                = "%v (file.id:%v) file.ip:%v does not match comparison ip:%v; skipping file"
	fCreateTimeAfterTimeLimitLog    = "%v (file.id:%v) file.createTime:%v is after timelimit:%v"
	fCreateTimeBeforeTimeLimitLog   = "%v (file.id:%v) file.createTime:%v is before timelimit:%v; skipping file"
	fDatasetMatchTrueLog            = "%v (file.id:%v) file.datasetID:%v matches Dataset:%v"
	fDatasetMatchFalseLog           = "%v (file.id:%v) file.datasetID:%v does not match Dataset:%v; skipping file"
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
	fGbrFileNameByFileIDLog         = "%v (file.id:%v) gbr verified file.id:%v as matching MB filename:%v"
	fGbrNoFileNameByFileIDLog       = "%v (file.id:%v) gbr could not find MB file.id:%v"
	fGbrDatasetByFileIDLog          = "%v (file.id:%v) gbr verified & set file.id:%v to dataset:%v"
	fVerifiedLog                    = "%v (file.id:%v) verified as ready to be migrated in preparation for removal!"
)

// verify all

func (f *file) verify() bool {
	e = ap.getEnv()

	if !f.verifyEnvMatch() {
		return false
	}

	if !f.verifyGBMetadata() {
		return false
	}

	if !f.verifyStat() {
		return false
	}

	e.logger.Info(fmt.Sprintf(fVerifiedLog, f.smbName, f.id))

	return true
}

// verify config metadata
func (f *file) verifyEnvMatch() bool {
	e = ap.getEnv()

	if !f.verifyIP() {
		return false
	}

	if !f.verifyTimeLimit() {
		return false
	}

	e.logger.Info(fmt.Sprintf(fEnvMatchLog, f.smbName, f.id, f.stagingPath))

	return true
}

func (f *file) verifyIP() bool {
	e = ap.getEnv()

	if reflect.DeepEqual(f.fanIP, e.sysIP) {
		e.logger.Info(fmt.Sprintf(fIPMatchTrueLog, f.smbName, f.id, f.fanIP, e.sysIP))
	} else {
		e.logger.Warn(fmt.Sprintf(fIPMatchFalseLog, f.smbName, f.id, f.fanIP, e.sysIP))
	}

	return reflect.DeepEqual(f.fanIP, e.sysIP)
}

func (f *file) verifyTimeLimit() bool {
	e = ap.getEnv()
	if f.createTime.After(e.limit) {
		e.logger.Info(fmt.Sprintf(
			fCreateTimeAfterTimeLimitLog,
			f.smbName,
			f.id,
			f.createTime.Round(time.Millisecond),
			e.limit.Round(time.Millisecond)))
	} else {
		e.logger.Warn(fmt.Sprintf(
			fCreateTimeBeforeTimeLimitLog,
			f.smbName,
			f.id,
			f.createTime.Round(time.Millisecond),
			e.limit.Round(time.Millisecond)))
	}

	return f.createTime.After(e.limit)
}

func (f *file) getGBMetadata() string {
	e = ap.getEnv()
	id := f.id
	cmd := exec.Command("/usr/bin/gbr", "file", "ls", "-i", id, "-d")
	cmdOut, err := cmd.CombinedOutput()

	if err != nil {
		f.getByIDErrLog(err)
	}

	out := string(cmdOut)

	return cleanGbrOut(out)
}

// Verify GB internal metadata
func (f *file) verifyGBMetadata() bool {
	out := f.getGBMetadata()
	// Gets file MBDS & compares with e.DS
	if !f.verifyMBDatasetByFileID(out) {
		return false
	}

	return f.verifyMBFileNameByFileID(out)
}

func (f *file) verifyMBFileNameByFileID(out string) bool {
	id := f.id
	if out == "" {
		e.logger.Warn(fmt.Sprintf(fGbrNoFileNameByFileIDLog, f.smbName, id, id))
		return false
	}

	filename := f.parseMBFileNameByFileID(out)

	return f.verifyFileIDName(filename)
}

func (f *file) verifyMBDatasetByFileID(out string) bool {
	id := f.id

	if out == "" {
		e.logger.Warn(fmt.Sprintf(fGbrNoFileNameByFileIDLog, f.smbName, id, id))
		return false
	}

	// set f.datasetID
	f.setMBDatasetByFileID(out)

	// get env datasetID
	datasetID := e.datasetID

	// Compare f.datasetID & env.datasetID
	return f.verifyInDataset(datasetID)
}

func (f *file) parseMBFileNameByFileID(cmdOut string) (filename string) {
	e = ap.getEnv()
	line := strings.Split(cmdOut, " ")
	filename = line[2]
	e.logger.Info(fmt.Sprintf(fGbrFileNameByFileIDLog, f.smbName, f.id, f.id, filename))

	return
}

func (f *file) setMBDatasetByFileID(cmdOut string) {
	e = ap.getEnv()
	lines := strings.Split(string(cmdOut), ";")

	for _, line := range lines {
		if strings.Contains(line, "parent id") {
			parentDS := line[len(line)-32:]
			f.datasetID = parentDS
			e.logger.Info(fmt.Sprintf(fGbrDatasetByFileIDLog, f.smbName, f.id, f.id, parentDS))

			return
		}
	}
	// Should never happen as caught with previous checks
	e.logger.Warn(fmt.Sprintf(fGbrNoFileNameByFileIDLog, f.smbName, f.id, f.id))
}

func (f *file) getByIDErrLog(err error) {
	e = ap.getEnv()
	err = errors.New(cleanGbrOut(err.Error()))
	e.logger.Warn(err)
	e.logger.Warn(fmt.Sprintf(fGbrNoFileNameByFileIDLog, f.smbName, f.id, f.id))
}

func (f *file) verifyInDataset(datasetID string) bool {
	e = ap.getEnv()
	if f.datasetID == datasetID {
		e.logger.Info(fmt.Sprintf(fDatasetMatchTrueLog, f.smbName, f.id, f.datasetID, datasetID))
	} else {
		e.logger.Warn(fmt.Sprintf(fDatasetMatchFalseLog, f.smbName, f.id, f.datasetID, datasetID))
	}

	return f.datasetID == datasetID
}

func (f *file) verifyFileIDName(fileName string) bool {
	e = ap.getEnv()
	if f.smbName == fileName {
		e.logger.Info(fmt.Sprintf(
			fSmbNameMatchFileIDNameTrueLog, f.smbName, f.id, f.smbName, fileName))
	} else {
		e.logger.Warn(fmt.Sprintf(
			fSmbNameMatchFileIDNameFalseLog, f.smbName, f.id, f.smbName, fileName))
	}

	return f.smbName == fileName
}

// Verify local FS metadata
func (f *file) verifyStat() bool {
	e = ap.getEnv()
	fileInfo, err := fs.Stat(e.fsys, f.stagingPath)

	if err != nil {
		e.logger.Warn(fmt.Sprintf(fExistsFalseLog, f.smbName, f.id, f.stagingPath))
		return false
	}

	e.logger.Info(fmt.Sprintf(fExistsTrueLog, f.smbName, f.id, f.stagingPath))

	if !f.verifyFileSize(fileInfo.Size()) {
		return false
	}

	if !f.verifyCreateTime(fileInfo.ModTime()) {
		return false
	}

	e.logger.Info(fmt.Sprintf(fStatMatchLog, f.smbName, f.id, f.stagingPath))

	return true
}

func (f *file) verifyFileSize(size int64) bool {
	e = ap.getEnv()
	if size != f.fileInfo.Size() {
		e.logger.Warn(fmt.Sprintf(fSizeMatchFalseLog, f.smbName, f.id, f.size, f.fileInfo.Size()))
		return false
	}

	e.logger.Info(fmt.Sprintf(fSizeMatchTrueLog, f.smbName, f.id, f.size, f.fileInfo.Size()))

	return true
}

func (f *file) verifyCreateTime(t time.Time) bool {
	e = ap.getEnv()
	if !t.Equal(f.createTime) {
		e.logger.Warn(fmt.Sprintf(fCreateTimeMatchFalseLog,
			f.smbName,
			f.id,
			f.createTime.Round(time.Millisecond),
			t.Round(time.Millisecond)))

		return false
	}

	e.logger.Info(fmt.Sprintf(
		fCreateTimeMatchTrueLog,
		f.smbName,
		f.id,
		f.createTime.Round(time.Millisecond),
		t.Round(time.Millisecond)))

	return true
}
