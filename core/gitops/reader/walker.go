package reader

import (
	"io/fs"
)

const (
	jsonExt = ".json"
	yamlExt = ".yaml"
)

func WalkDir(entries *[]string) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !d.IsDir() {
			*entries = append(*entries, path)
		}

		return nil
	}
}
