//go:build windows

package pathutil

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// IsProtected identifies Windows system data, installed applications and
// per-user application data. Resolve junctions and symlinks too, so selecting
// a linked folder cannot bypass it.
func IsProtected(path string) bool {
	if protectedWinPath(path) {
		return true
	}
	resolved, err := finalPath(path)
	if err != nil {
		// A file that does not exist yet is judged by its folder.
		dir, err := finalPath(filepath.Dir(path))
		if err != nil {
			return false
		}
		resolved = filepath.Join(dir, filepath.Base(path))
	}
	return !strings.EqualFold(resolved, path) && protectedWinPath(resolved)
}

const volumeNameDOS = 0x0 // GetFinalPathNameByHandle: C:\... rather than a volume GUID

// finalPath resolves every symlink and junction in path. filepath.EvalSymlinks
// does not follow junctions (mount points) since Go 1.23, so the handle is asked.
func finalPath(path string) (string, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}
	h, err := windows.CreateFile(name, 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(h)
	buf := make([]uint16, 512)
	for {
		n, err := windows.GetFinalPathNameByHandle(h, &buf[0], uint32(len(buf)), volumeNameDOS)
		if err != nil {
			return "", err
		}
		if int(n) < len(buf) {
			s := windows.UTF16ToString(buf[:n])
			if rest, ok := strings.CutPrefix(s, `\\?\UNC\`); ok {
				return `\\` + rest, nil
			}
			return strings.TrimPrefix(s, `\\?\`), nil
		}
		buf = make([]uint16, n) // n is the required size including NUL
	}
}

// Top-level folders protected on every drive, so system folders of another
// Windows installation or a mounted disk are recognised as well.
var protectedWinRoots = map[string]bool{
	"windows": true, "windows.old": true, "program files": true, "program files (x86)": true,
	"programdata": true, "$recycle.bin": true, "system volume information": true, "recovery": true,
	"boot": true, "config.msi": true, "msocache": true, "$windows.~bt": true, "$windows.~ws": true,
	"$winreagent": true,
}

func protectedWinPath(path string) bool {
	path = strings.ToLower(filepath.Clean(path))
	parts := strings.Split(strings.Trim(path[len(filepath.VolumeName(path)):], `\`), `\`)
	if parts[0] == "" {
		return false // drive root: walk it, skipping the folders below
	}
	if protectedWinRoots[parts[0]] {
		return true
	}
	// Windows may live outside C:\Windows or C:\Program Files.
	for _, env := range []string{"SystemRoot", "ProgramFiles", "ProgramFiles(x86)", "ProgramW6432", "ProgramData"} {
		if root := os.Getenv(env); root != "" && Within(path, root) {
			return true
		}
	}
	// Users\<name>\AppData is the counterpart of ~/Library; its Temp folder
	// stays usable for explicitly chosen disposable files (including tests).
	if len(parts) >= 3 && parts[0] == "users" && parts[2] == "appdata" {
		return len(parts) < 5 || parts[3] != "local" || parts[4] != "temp"
	}
	return false
}
