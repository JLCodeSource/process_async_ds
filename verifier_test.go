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
			smbName: "file.txt",
			fanIP:   ips[0],
		}
		testLogger, hook := setupLogs(t)
		assert.True(t, file.verifyIP(ips[0], testLogger))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "file.txt ip:" + file.fanIP.String() + " matches comparison ip:" + ips[0].String()

		assertCorrectString(t, gotLogMsg, wantLogMsg)

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
			smbName: "file.txt",
			fanIP:   ips[0],
		}
		testLogger, hook := setupLogs(t)
		ip := net.ParseIP("192.168.101.1")
		assert.False(t, file.verifyIP(ip, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := ("file.txt ip:" + file.fanIP.String() + " does not match comparison ip:" +
			ip.String() + "; skipping file")

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

}
