package main

func statFiles(files []file) map[string]file {
	e = ap.getEnv()
	afs = e.afs

	m := make(map[string]file)

	fi, err := afs.Stat(files[0].stagingPath)

	if err != nil {
		e.logger.Warn(err)
	}

	m[files[0].id] = file{
		id:          files[0].id,
		smbName:     fi.Name(),
		createTime:  fi.ModTime(),
		size:        fi.Size(),
		stagingPath: files[0].stagingPath,
		fileInfo:    fi,
	}

	return m
}
