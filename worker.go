package main

func (ap *asyncProcessor) processFiles() {
	e = ap.getEnv()
	for i := range *ap.Files {
		(*ap.Files)[i].Hasher()
		(*ap.Files)[i].oldHash = (*ap.Files)[i].hash
		(*ap.Files)[i].oldStagingPath = (*ap.Files)[i].stagingPath
		(*ap.Files)[i].Move()
		(*ap.Files)[i].Hasher()
		if (*ap.Files)[i].compareHashes() {
			(*ap.Files)[i].success = true
		}
	}
}

func (f *File) compareHashes() bool {
	if f.oldHash != f.hash {
		return false
	}

	return true
}
