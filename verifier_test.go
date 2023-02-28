package main

import (
	"net"
	"os"
	"testing"
	"time"

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

func TestVerifyTimeLimit(t *testing.T) {

	t.Run("returns true if file.date is after time limit", func(t *testing.T) {
		file := File{
			smbName:    "file.txt",
			createTime: time.Now(),
		}
		days := time.Duration(15)
		hours := time.Duration(days * 24)
		now := time.Now()
		limit := now.Add(-((hours) * time.Hour))
		testLogger, hook := setupLogs(t)
		assert.True(t, file.verifyTimeLimit(limit, testLogger))
		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := "file.txt createTime:" + file.createTime.String() + " is after timelimit:" + limit.String()

		assertCorrectString(t, gotLogMsg, wantLogMsg)

	})
	t.Run("returns false if file.date is before time limit", func(t *testing.T) {
		file := File{
			smbName:    "file.txt",
			createTime: time.Now(),
		}
		limit := file.createTime.Add(24 * time.Hour)
		testLogger, hook := setupLogs(t)
		assert.False(t, file.verifyTimeLimit(limit, testLogger))

		gotLogMsg := hook.LastEntry().Message
		wantLogMsg := ("file.txt createTime:" + file.createTime.String() + " is before timelimit:" +
			limit.String() +
			"; skipping file")

		assertCorrectString(t, gotLogMsg, wantLogMsg)
	})

}
