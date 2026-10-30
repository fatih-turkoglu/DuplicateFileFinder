//go:build !darwin && !windows

package pathutil

func IsProtected(path string) bool { return false }
