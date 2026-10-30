//go:build darwin

package fileops

import (
	"errors"
	"os"
	"syscall"

	"DuplicateFile/internal/macos"

	"golang.org/x/sys/unix"
)

func moveToRecycleBin(path string) error {
	err := macos.Trash(path)
	if errors.Is(err, macos.ErrTrashUnavailable) {
		return ErrNoRecycleBin
	}
	return err
}

func isCrossDevice(err error) bool { return errors.Is(err, syscall.EXDEV) }

// renameNoReplace uses RENAME_EXCL, so a file created at dest after the
// existence check is never replaced.
func renameNoReplace(src, dest string) error {
	err := unix.RenamexNp(src, dest, unix.RENAME_EXCL)
	if errors.Is(err, unix.EEXIST) {
		return ErrExists
	}
	if err != nil {
		return &os.LinkError{Op: "rename", Old: src, New: dest, Err: err}
	}
	return nil
}
