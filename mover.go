package main

import (
	"os"
	"path"
	"strings"
)

func (f *File) UpdatePath() {
	oldDir, fn := path.Split(f.stagingPath)
	parts := strings.Split(oldDir, string(os.PathSeparator))
	lastParts := parts[2:]
	firstParts := parts[:2]
	fp := strings.Join(firstParts, string(os.PathSeparator))
	lp := strings.Join(lastParts, string(os.PathSeparator))
	f.stagingPath = fp + ".processed" + string(os.PathSeparator) + lp + fn
	return
}
