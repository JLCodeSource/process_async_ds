package main

import (
	"net"
	"reflect"
)

func (f File) verifyIP(ip net.IP) bool {
	return reflect.DeepEqual(f.fanIP, ip)
}
