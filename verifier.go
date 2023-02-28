package main

import (
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
