package utils

import (
	"os"
)

func IsDir(name string) bool {
	if f, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return f.IsDir()
	}
	return true
}

func IsNotExist(src string) bool {
	_, err := os.Stat(src)

	return os.IsNotExist(err)
}
