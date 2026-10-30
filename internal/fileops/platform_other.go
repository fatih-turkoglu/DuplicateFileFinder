//go:build !windows && !darwin

package fileops

import (
	"errors"
	"os"
	"syscall"
)

func moveToRecycleBin(path string) error {
	return ErrNoRecycleBin
}

func isCrossDevice(err error) bool {
	return errors.Is(err, syscall.EXDEV)
}

// renameNoReplace: unsupported platforms only re-check before renaming.
func renameNoReplace(src, dest string) error {
	if exists(dest) {
		return ErrExists
	}
	return os.Rename(src, dest)
}
