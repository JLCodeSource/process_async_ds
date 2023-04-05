package main

func statFiles(files []file) map[string]file {
	e = ap.getEnv()
	afs = e.afs

	m := make(map[string]file)

	for _, f := range files {
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
