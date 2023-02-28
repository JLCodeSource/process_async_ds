package main

import (
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyFileIP(t *testing.T) {

	t.Run("returns true if ip is same as the current machine", func(t *testing.T) {
		hostname, err := os.Hostname()
		if err != nil {
			panic(err)
		}
		ips, err := net.LookupIP(hostname)
		if err != nil {
			panic(err)
		}
		file := File{
			fanIP: ips[0],
		}
		assert.True(t, file.verifyIP(ips[0]))

	})
	t.Run("returns false if ip is not the same as the current machine", func(t *testing.T) {
		hostname, err := os.Hostname()
		if err != nil {
			panic(err)
		}
		ips, err := net.LookupIP(hostname)
		if err != nil {
			panic(err)
		}
		file := File{
			fanIP: ips[0],
		}
		ip := net.IP("192.168.101.1")
		assert.False(t, file.verifyIP(ip))
	})

}
