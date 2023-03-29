package main

func (ap *asyncProcessor) processFiles() {
	e = ap.getEnv()
	for i := range *ap.Files {
		(*ap.Files)[i].Hasher()
		(*ap.Files)[i].Move(e.afs, e.logger)
		(*ap.Files)[i].Hasher()
	}
}
