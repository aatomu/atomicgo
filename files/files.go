package files

import (
	"os"
	"path/filepath"
	"runtime"
)

// Get Current Dir?
func CurrentDir() string {
	_, file, _, _ := runtime.Caller(1)
	return filepath.Dir(file)
}

// Change Workdir
func SetWorkDir(dirPath string) {
	os.Chdir(dirPath)
}

// Access file/dir
func IsAccess(filePath string) (ok bool) {
	_, err := os.Stat(filePath)
	return err == nil
}

// Create File/Directory
func Create(path string, isDir bool) error {
	if isDir {
		return os.Mkdir(path, 0666)
	}
	_, err := os.Create(path)
	return err
}

// WriteFile flash 0x666
func WriteFileFlash(path string, data []byte) (err error) {
	return os.WriteFile(path, data, 0666)
}
