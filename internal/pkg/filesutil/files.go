package filesutil

import (
	"os"
	"path/filepath"
)

func Exist(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

/*
Delete function deletes the file, specified in the argument and returns
the boolean value of operation.

If the file is deleted, true is returned.
If the file is not exist, true will be returned as well.

Otherwise, the function returns false
*/
func Delete(filePath string) bool {
	if err := os.Remove(filePath); err != nil {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return true
		}
		return false
	}
	return true
}

// ExtractFileName selects the file name from it's path
func ExtractFileName(filePath string) string {
	return filepath.Base(filePath)
}
