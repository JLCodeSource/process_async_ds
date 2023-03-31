package main

func (ap *asyncProcessor) processFiles() {
	e = ap.env

	for i := range ap.files {
		ap.files[i].hasher()
		ap.files[i].oldHash = ap.files[i].hash
		// log
		ap.files[i].oldStagingPath = ap.files[i].stagingPath
		// log
		ap.files[i].move()
		// log (in Move)
		ap.files[i].hasher()
		if ap.files[i].compareHashes() {
			ap.files[i].success = true
			// log
		} // Add fail & log
		// log result
	}
}

func (f *file) compareHashes() bool {
	if f.oldHash != f.hash {
		return false
	}

	return true
}
