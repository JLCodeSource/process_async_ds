package main

func (ap *asyncProcessor) processFiles() {
	e = ap.getEnv()
	for i := range *ap.Files {
		(*ap.Files)[i].Hasher()
		(*ap.Files)[i].oldStagingPath = (*ap.Files)[i].stagingPath
		(*ap.Files)[i].Move()
		(*ap.Files)[i].Hasher()
		// Compare hashes
		// Confirm success
		// Set success bool on File
	}
}
