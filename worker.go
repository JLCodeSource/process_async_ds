package main

func (ap *asyncProcessor) processFiles() {
	e = ap.getEnv()
	for i := range *ap.Files {
		(*ap.Files)[i].hasher()
		(*ap.Files)[i].setOldHash((*ap.Files)[i].getHash())
		// log
		(*ap.Files)[i].setOldStagingPath((*ap.Files)[i].getStagingPath())
		// log
		(*ap.Files)[i].move()
		// log (in Move)
		(*ap.Files)[i].hasher()
		if (*ap.Files)[i].compareHashes() {
			(*ap.Files)[i].setSuccess(true)
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
