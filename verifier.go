package main

import (
	"io/fs"
	"net"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
)

func (f File) verifyIP(ip net.IP, logger *logrus.Logger) bool {
	if reflect.DeepEqual(f.fanIP, ip) {
		logger.Info(f.smbName + " ip:" + f.fanIP.String() + " matches comparison ip:" + ip.String())
	} else {
		logger.Warn(f.smbName + " ip:" + f.fanIP.String() + " does not match comparison ip:" + ip.String() + "; skipping file")

	}
	return reflect.DeepEqual(f.fanIP, ip)
}

func (f File) verifyTimeLimit(limit time.Time, logger *logrus.Logger) bool {
	if f.createTime.After(limit) {
		logger.Info(f.smbName + " createTime:" + f.createTime.String() + " is after timelimit:" + limit.String())
	} else {
		logger.Warn(f.smbName + " createTime:" + f.createTime.String() + " is before timelimit:" + limit.String() + "; skipping file")
	}
	return f.createTime.After(limit)
}

func (f File) verifyInProcessedDataset(datasetID string, logger *logrus.Logger) bool {
	if f.datasetID == datasetID {
		logger.Info(f.smbName + " datasetID:" + f.datasetID + " matches asyncProcessedDatasetID:" + datasetID)
	} else {
		logger.Warn(f.smbName + " datasetID:" + f.datasetID + " does not match asyncProcessedDatasetID:" + datasetID +
			"; skipping file")
	}
	return f.datasetID == datasetID
}

func (f File) verifyExists(fsys fs.FS, logger *logrus.Logger) bool {
	_, err := fs.Stat(fsys, f.stagingPath)
	if err != nil {
		logger.Warn(f.smbName + " does not exist at " + f.stagingPath + "; skipping file")
		return false
	}
	logger.Info(f.smbName + " exists at " + f.stagingPath)
	return true
}
