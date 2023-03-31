package main

import "fmt"

var (
	adSetOldHashLog        = "%v (file.id:%v) setting f.oldHash:%v"
	adSetOldStagingPathLog = "%v (file.id:%v) setting f.oldStagingPath:%v"
	adSetSuccessLog        = "%v (file.id:%v) setting f.success:%v"
	adReadyForProcessingLog  = "%v (file.id:%v) f.stagingPath:%v is ready for processing"
)

func (ap *asyncProcessor) processFiles() {
	e = ap.env

	for i := range ap.files {
		ap.files[i].hasher()
		ap.files[i].oldHash = ap.files[i].hash
		e.logger.Info(fmt.Sprintf(adSetOldHashLog, ap.files[i].smbName, ap.files[i].id, ap.files[i].hash))
		ap.files[i].oldStagingPath = ap.files[i].stagingPath
		e.logger.Info(fmt.Sprintf(adSetOldStagingPathLog, ap.files[i].smbName, ap.files[i].id, ap.files[i].stagingPath))
		ap.files[i].move()
		// log (in Move)
		ap.files[i].hasher()
		if ap.files[i].compareHashes() {
			ap.files[i].success = true
		} else {
			ap.files[i].success = false
		}
		e.logger.Info(fmt.Sprintf(adSetSuccessLog, ap.files[i].smbName, ap.files[i].id, ap.files[i].success))
		e.logger.Info(fmt.Sprintf(adReadyForProcessingLog, ap.files[i].smbName, ap.files[i].id, ap.files[i].stagingPath))
	}
}

func (f *file) compareHashes() bool {
	if f.oldHash != f.hash {
		return false
	}

	return true
}
