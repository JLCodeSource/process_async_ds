package main

func statFiles(tmpFiles []file) map[string]file {
	e = ap.getEnv()
	afs = e.afs

	m := make(map[string]file)

	for _, f := range tmpFiles {
		fi, err := afs.Stat(f.stagingPath)

		if err != nil {
			e.logger.Warn(err)
		}

		m[f.id] = file{
			id:          f.id,
			smbName:     fi.Name(),
			createTime:  fi.ModTime(),
			size:        fi.Size(),
			stagingPath: f.stagingPath,
			fileInfo:    fi,
		}

	}

	return m
}

func getCheckFilesMapMBMetadata(tmpFiles []file) map[string]file {
	e = ap.getEnv()
	afs = e.afs

	m := make(map[string]file)

	for _, f := range tmpFiles {
		out := f.getGBMetadata()
		f.setMBDatasetByFileID(out)
		f.smbName = f.parseMBFileNameByFileID(out)

		m[f.id] = file{
			id:        f.id,
			datasetID: f.datasetID,
			smbName:   f.smbName,
		}
	}

	return m
}
