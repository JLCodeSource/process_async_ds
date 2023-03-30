package main

func (ap *asyncProcessor) processFiles() {
	e = ap.getEnv()
	for i := range *ap.Files {
		(*ap.Files)[i].Hasher()
		(*ap.Files)[i].oldHash = (*ap.Files)[i].hash
		// log
		(*ap.Files)[i].oldStagingPath = (*ap.Files)[i].stagingPath
		// log
		(*ap.Files)[i].Move()
		// log (in Move)
		(*ap.Files)[i].Hasher()
		if (*ap.Files)[i].compareHashes() {
			(*ap.Files)[i].success = true
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
