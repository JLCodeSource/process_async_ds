package main

import (
	"net"
	"reflect"

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
