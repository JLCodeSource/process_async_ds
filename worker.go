package main

func (ap *asyncProcessor) processFiles() {
	e = ap.getEnv()
	for i := range *ap.Files {
		(*ap.Files)[i].Hasher()
		(*ap.Files)[i].Move()
		(*ap.Files)[i].Hasher()
	}
}
