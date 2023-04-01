package main

import "fmt"

var (
	adHasherErrLog            = "%v (file.id:%v) f.hasher error:%v; continuing"
	adSetOldHashLog           = "%v (file.id:%v) setting f.oldHash:%v"
	adSetOldStagingPathLog    = "%v (file.id:%v) setting f.oldStagingPath:%v"
	adSetSuccessLog           = "%v (file.id:%v) setting f.success:%v"
	adCompareHashesNoMatchLog = "%v (file.id:%v) f.oldHash:%v does not match f.hash:%v; fatal"
	adCompareHashesMatchLog   = "%v (file.id:%v) f.oldHash:%v matches f.hash:%v"

	adReadyForProcessingLog = "%v (file.id:%v) f.stagingPath:%v is ready for processing"
)

func (ap *asyncProcessor) processFiles() {
	e = ap.env

	for i := range ap.files {
		err := ap.files[i].hasher()
		if err != nil {
			e.logger.Warn(fmt.Sprintf(adHasherErrLog, ap.files[i].smbName, ap.files[i].id, err))
			continue
		}
		ap.files[i].oldHash = ap.files[i].hash
		e.logger.Info(fmt.Sprintf(adSetOldHashLog, ap.files[i].smbName, ap.files[i].id, ap.files[i].hash))
		ap.files[i].oldStagingPath = ap.files[i].stagingPath
		e.logger.Info(fmt.Sprintf(adSetOldStagingPathLog, ap.files[i].smbName, ap.files[i].id, ap.files[i].stagingPath))
		ap.files[i].move()
		// log (in Move)
		err = ap.files[i].hasher()
		if err != nil {
			e.logger.Warn(fmt.Sprintf(adHasherErrLog, ap.files[i].smbName, ap.files[i].id, err))
			continue
		}

		if ap.files[i].compareHashes() {
			ap.files[i].success = true
			e.logger.Info(fmt.Sprintf(
				adCompareHashesMatchLog, ap.files[i].smbName, ap.files[i].id, ap.files[i].oldHash, ap.files[i].hash))
		} else {
			ap.files[i].success = false
			// Should never happen (assert?)
			e.logger.Fatal(fmt.Sprintf(
				adCompareHashesNoMatchLog, ap.files[i].smbName, ap.files[i].id, ap.files[i].oldHash, ap.files[i].hash))
		}
		e.logger.Info(fmt.Sprintf(adSetSuccessLog, ap.files[i].smbName, ap.files[i].id, ap.files[i].success))
		e.logger.Info(fmt.Sprintf(adReadyForProcessingLog, ap.files[i].smbName, ap.files[i].id, ap.files[i].stagingPath))
	}
}

func (f *file) compareHashes() bool {
	return f.oldHash == f.hash
}
