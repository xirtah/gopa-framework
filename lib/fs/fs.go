package fs

import (
	"os"
	"path/filepath"
)

func GetFullPath(fileName string) string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	return exPath + "/" + fileName
}
